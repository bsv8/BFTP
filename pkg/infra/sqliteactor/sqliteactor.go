package sqliteactor

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

type Opened struct {
	// DB 只保留给迁移中的旧访问点使用。
	// 设计说明：
	// - 新代码必须优先走 Actor；
	// - 这轮改造还没把所有 sqlite 访问点一次性迁完，所以先把 owner actor 立起来，
	//   再逐步清掉旧的裸 DB 调用；
	// - 这里显式保留 DB 字段，是为了让迁移期间能编译通过，而不是鼓励继续扩散裸 DB 写法。
	DB *sql.DB

	// Actor 是正式入口。
	// 规则：
	// - 所有新的运行时代码都应该通过 Actor 串行执行；
	// - 不允许把 Rows / Tx / Stmt 带出闭包，否则会重新引入生命周期失控问题。
	Actor *Actor
}

type Actor struct {
	db *sql.DB

	requests chan actorRequest
	closedCh chan struct{}

	closeOnce sync.Once
	wg        sync.WaitGroup
	mu        sync.RWMutex
	closed    bool
}

type actorRequest struct {
	ctx context.Context
	run func(*sql.DB) (any, error)
	out chan actorResponse
}

type actorResponse struct {
	value any
	err   error
}

// Open 打开 sqlite 的单 owner actor 入口。
// 设计规则：
// - 一个库文件 = 一个 owner = 一个连接 = 一个串行执行入口；
// - `*sql.DB` 虽然并发安全，但它默认是连接池，这对 sqlite 往往不是正确模型；
// - 这里强制 `SetMaxOpenConns(1)` / `SetMaxIdleConns(1)`，把物理连接压成 1 条；
// - 保留 WAL，但 WAL 只改善读写关系，不代表可以随便并发写；
// - `busy_timeout` / retry 不作为主方案，主方案是结构上避免竞争。
func Open(path string) (*Opened, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("sqlite path is empty")
	}
	dsn := appendPragma(path, "journal_mode(WAL)")
	dsn = appendQueryValue(dsn, "_txlock", "immediate")
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	actor, err := New(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Opened{
		DB:    db,
		Actor: actor,
	}, nil
}

func New(db *sql.DB) (*Actor, error) {
	if db == nil {
		return nil, fmt.Errorf("sqlite db is nil")
	}
	// 设计说明：
	// - actor 只允许一个 owner goroutine 真正触碰 sqlite；
	// - 所有访问都通过请求队列进来，避免多个 goroutine 直接并发打 sqlite；
	// - 这样可以把“偶发锁竞争”变成明确的串行时序。
	a := &Actor{
		db:       db,
		requests: make(chan actorRequest),
		closedCh: make(chan struct{}),
	}
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.loop()
	}()
	return a, nil
}

func (a *Actor) Do(ctx context.Context, fn func(*sql.DB) error) error {
	if fn == nil {
		return fmt.Errorf("sqlite actor do func is nil")
	}
	_, err := callValue(ctx, a, func(db *sql.DB) (struct{}, error) {
		return struct{}{}, fn(db)
	})
	return err
}

func (a *Actor) Tx(ctx context.Context, fn func(*sql.Tx) error) error {
	if fn == nil {
		return fmt.Errorf("sqlite actor tx func is nil")
	}
	_, err := TxValue(ctx, a, func(tx *sql.Tx) (struct{}, error) {
		return struct{}{}, fn(tx)
	})
	return err
}

func DoValue[T any](ctx context.Context, a *Actor, fn func(*sql.DB) (T, error)) (T, error) {
	return callValue(ctx, a, fn)
}

func TxValue[T any](ctx context.Context, a *Actor, fn func(*sql.Tx) (T, error)) (T, error) {
	var zero T
	if a == nil {
		return zero, fmt.Errorf("sqlite actor is nil")
	}
	if fn == nil {
		return zero, fmt.Errorf("sqlite actor tx func is nil")
	}
	result, err := a.call(ctx, func(db *sql.DB) (any, error) {
		// 设计规则：
		// - 事务闭包必须在 actor 内部完整执行完；
		// - 不允许在事务中回头再走别的库级 `db.*`，否则单连接模式下会自锁；
		// - 不允许把 `*sql.Tx`、`*sql.Rows`、`*sql.Stmt` 带出闭包。
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		value, runErr := fn(tx)
		if runErr != nil {
			_ = tx.Rollback()
			return nil, runErr
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return value, nil
	})
	if err != nil {
		return zero, err
	}
	value, ok := result.(T)
	if !ok {
		return zero, fmt.Errorf("sqlite actor tx result type mismatch")
	}
	return value, nil
}

func (a *Actor) Close() error {
	if a == nil {
		return nil
	}
	var err error
	a.closeOnce.Do(func() {
		a.mu.Lock()
		a.closed = true
		a.mu.Unlock()
		close(a.closedCh)
		a.wg.Wait()
		err = a.db.Close()
	})
	return err
}

func (a *Actor) loop() {
	for {
		select {
		case <-a.closedCh:
			return
		case req := <-a.requests:
			if req.out == nil {
				continue
			}
			if req.ctx != nil && req.ctx.Err() != nil {
				req.out <- actorResponse{err: req.ctx.Err()}
				continue
			}
			value, err := req.run(a.db)
			req.out <- actorResponse{value: value, err: err}
		}
	}
}

func (a *Actor) call(ctx context.Context, fn func(*sql.DB) (any, error)) (any, error) {
	if a == nil {
		return nil, fmt.Errorf("sqlite actor is nil")
	}
	if fn == nil {
		return nil, fmt.Errorf("sqlite actor call func is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	a.mu.RLock()
	closed := a.closed
	a.mu.RUnlock()
	if closed {
		return nil, fmt.Errorf("sqlite actor is closed")
	}
	req := actorRequest{
		ctx: ctx,
		run: fn,
		out: make(chan actorResponse, 1),
	}
	// 设计说明：
	// - actor 是正式串行入口，不允许业务代码绕过它并发触碰 sqlite；
	// - 真正的稳定性来自“没有竞争”，不是“竞争后多等一会儿再试”。
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-a.closedCh:
		return nil, fmt.Errorf("sqlite actor is closed")
	case a.requests <- req:
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-req.out:
		return resp.value, resp.err
	}
}

func callValue[T any](ctx context.Context, a *Actor, fn func(*sql.DB) (T, error)) (T, error) {
	var zero T
	if a == nil {
		return zero, fmt.Errorf("sqlite actor is nil")
	}
	if fn == nil {
		return zero, fmt.Errorf("sqlite actor value func is nil")
	}
	result, err := a.call(ctx, func(db *sql.DB) (any, error) {
		return fn(db)
	})
	if err != nil {
		return zero, err
	}
	value, ok := result.(T)
	if !ok {
		return zero, fmt.Errorf("sqlite actor result type mismatch")
	}
	return value, nil
}

func appendPragma(dsn string, pragma string) string {
	return appendQueryValue(dsn, "_pragma", pragma)
}

func appendQueryValue(dsn string, key string, value string) string {
	dsn = strings.TrimSpace(dsn)
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if dsn == "" || key == "" || value == "" {
		return dsn
	}
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + key + "=" + value
}
