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
	// DB 只保留给迁移中的旧代码使用。
	// 设计说明：
	// - 新代码必须优先走 Actor；
	// - 这轮改造还没把所有 sqlite 访问点一次性迁完，所以先把 owner actor 立起来，
	//   再逐步清掉旧的直连路径；
	// - 这里显式保留 DB 字段，是为了让迁移期间能编译通过，而不是鼓励继续扩散直连。
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
