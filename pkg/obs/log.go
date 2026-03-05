package obs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	LevelDebug     = "debug"
	LevelInfo      = "info"
	LevelBusiness  = "business"
	LevelImportant = "important"
	LevelError     = "error"

	CategorySystem   = "system"
	CategoryBusiness = "business"
)

type Event struct {
	TS       string         `json:"ts"`
	Level    string         `json:"level"`
	Category string         `json:"category"`
	Service  string         `json:"service"`
	Name     string         `json:"name"`
	Fields   map[string]any `json:"fields,omitempty"`
}

type loggerState struct {
	mu             sync.Mutex
	file           *os.File
	consoleMinRank int
	nextListenerID int
	listeners      map[int]func(Event)
}

var state = &loggerState{consoleMinRank: rank(LevelBusiness)}

func Init(filePath string, consoleMinLevel string) error {
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.file != nil {
		_ = state.file.Close()
		state.file = nil
	}
	if filePath != "" {
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return err
		}
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		state.file = f
	}
	state.consoleMinRank = rank(consoleMinLevel)
	return nil
}

func Close() error {
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.file == nil {
		return nil
	}
	err := state.file.Close()
	state.file = nil
	return err
}

// AddListener 注册一个事件监听器（进程内“消息系统”挂钩点）。
//
// 典型用途：E2E/in-proc 集成测试在不“手把手执行业务动作”的前提下，等待并验证业务流程是否真正发生。
//
// 注意：回调会在写日志之后被调用；回调不应阻塞太久（必要时自行起 goroutine）。
func AddListener(fn func(Event)) (remove func()) {
	if fn == nil {
		return func() {}
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.listeners == nil {
		state.listeners = map[int]func(Event){}
	}
	state.nextListenerID++
	id := state.nextListenerID
	state.listeners[id] = fn
	return func() {
		state.mu.Lock()
		defer state.mu.Unlock()
		if state.listeners == nil {
			return
		}
		delete(state.listeners, id)
	}
}

func Info(service, name string, fields map[string]any) {
	write(LevelInfo, CategorySystem, service, name, fields)
}

func Debug(service, name string, fields map[string]any) {
	write(LevelDebug, CategorySystem, service, name, fields)
}

func Business(service, name string, fields map[string]any) {
	write(LevelBusiness, CategoryBusiness, service, name, fields)
}

func Important(service, name string, fields map[string]any) {
	write(LevelImportant, CategoryBusiness, service, name, fields)
}

func Error(service, name string, fields map[string]any) {
	write(LevelError, CategorySystem, service, name, fields)
}

func write(level, category, service, name string, fields map[string]any) {
	var compacted map[string]any
	if fields != nil {
		if m, ok := CompactAny(fields).(map[string]any); ok {
			compacted = m
		} else {
			compacted = map[string]any{"_": CompactAny(fields)}
		}
	}
	ev := Event{
		TS:       time.Now().UTC().Format(time.RFC3339),
		Level:    level,
		Category: category,
		Service:  service,
		Name:     name,
		Fields:   compacted,
	}
	b, _ := json.Marshal(ev)
	line := string(b)

	var ls []func(Event)
	state.mu.Lock()

	if state.file != nil {
		_, _ = fmt.Fprintln(state.file, line)
	}
	if rank(level) >= state.consoleMinRank {
		fmt.Println(line)
	}
	if len(state.listeners) > 0 {
		ls = make([]func(Event), 0, len(state.listeners))
		for _, fn := range state.listeners {
			ls = append(ls, fn)
		}
	}
	state.mu.Unlock()

	for _, fn := range ls {
		func() {
			defer func() { _ = recover() }()
			fn(ev)
		}()
	}
}

func rank(level string) int {
	switch level {
	case LevelDebug:
		return 10
	case LevelInfo:
		return 20
	case LevelBusiness:
		return 30
	case LevelImportant:
		return 40
	case LevelError:
		return 50
	default:
		return 30
	}
}
