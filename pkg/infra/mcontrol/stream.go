package mcontrol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// 兼容现有 BitFS managed 控制变量名，统一给 gateway/arbiter/domain 复用。
	ControlEndpointEnv = "BITFS_MANAGED_CONTROL_ENDPOINT"
	ControlTokenEnv    = "BITFS_MANAGED_CONTROL_TOKEN"
)

type RuntimeEvent struct {
	Seq            uint64         `json:"seq"`
	RuntimeEpoch   string         `json:"runtime_epoch"`
	Topic          string         `json:"topic"`
	Scope          string         `json:"scope"`
	OccurredAtUnix int64          `json:"occurred_at_unix"`
	Producer       string         `json:"producer"`
	TraceID        string         `json:"trace_id"`
	Payload        map[string]any `json:"payload,omitempty"`
}

type commandFrame struct {
	Type       string         `json:"type"`
	CommandID  string         `json:"command_id"`
	Action     string         `json:"action"`
	SentAtUnix int64          `json:"sent_at_unix,omitempty"`
	Payload    map[string]any `json:"payload,omitempty"`
}

type CommandFrame struct {
	CommandID string
	Action    string
	Payload   map[string]any
}

type commandHandler func(CommandFrame)

type Stream interface {
	Emit(topic, scope, producer, traceID string, payload map[string]any)
	StartCommandLoop(handler func(CommandFrame)) error
	Close() error
}

func NewFromEnv() (Stream, error) {
	endpoint := strings.TrimSpace(os.Getenv(ControlEndpointEnv))
	if endpoint == "" {
		return noopStream{}, nil
	}
	token := strings.TrimSpace(os.Getenv(ControlTokenEnv))
	if token == "" {
		return nil, fmt.Errorf("managed control token is required")
	}
	network, address, err := parseEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout(network, address, 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("dial managed control endpoint failed: %w", err)
	}
	s := &socketStream{
		conn:         conn,
		token:        token,
		runtimeEpoch: fmt.Sprintf("rt-%d", time.Now().UnixNano()),
	}
	if err := s.writeFrameLocked(map[string]any{
		"type":          "hello",
		"token":         token,
		"runtime_epoch": s.runtimeEpoch,
	}); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("write managed control hello failed: %w", err)
	}
	return s, nil
}

type noopStream struct{}

func (noopStream) Emit(string, string, string, string, map[string]any) {}

func (noopStream) StartCommandLoop(func(CommandFrame)) error { return nil }

func (noopStream) Close() error { return nil }

type socketStream struct {
	mu            sync.Mutex
	conn          net.Conn
	token         string
	seq           uint64
	runtimeEpoch  string
	commandLoopOn bool
	handler       commandHandler
}

func (s *socketStream) Emit(topic, scope, producer, traceID string, payload map[string]any) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		return
	}
	s.seq++
	frame := RuntimeEvent{
		Seq:            s.seq,
		RuntimeEpoch:   s.runtimeEpoch,
		Topic:          topic,
		Scope:          normalizeScope(scope),
		OccurredAtUnix: time.Now().Unix(),
		Producer:       strings.TrimSpace(producer),
		TraceID:        strings.TrimSpace(traceID),
		Payload:        clonePayload(payload),
	}
	_ = s.writeFrameLocked(frame)
}

func (s *socketStream) StartCommandLoop(handler func(CommandFrame)) error {
	if handler == nil {
		return fmt.Errorf("managed control command handler is required")
	}
	s.mu.Lock()
	if s.commandLoopOn {
		s.mu.Unlock()
		return nil
	}
	s.commandLoopOn = true
	s.handler = handler
	conn := s.conn
	s.mu.Unlock()
	if conn == nil {
		return fmt.Errorf("managed control connection is not ready")
	}
	go s.readLoop(conn)
	return nil
}

func (s *socketStream) Close() error {
	s.mu.Lock()
	conn := s.conn
	s.conn = nil
	s.mu.Unlock()
	if conn == nil {
		return nil
	}
	return conn.Close()
}

func (s *socketStream) readLoop(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var raw commandFrame
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return
		}
		if strings.TrimSpace(raw.Type) != "command" || strings.TrimSpace(raw.CommandID) == "" || strings.TrimSpace(raw.Action) == "" {
			return
		}
		s.mu.Lock()
		handler := s.handler
		s.mu.Unlock()
		if handler != nil {
			handler(CommandFrame{
				CommandID: strings.TrimSpace(raw.CommandID),
				Action:    strings.TrimSpace(raw.Action),
				Payload:   clonePayload(raw.Payload),
			})
		}
	}
}

func (s *socketStream) writeFrameLocked(frame any) error {
	if s.conn == nil {
		return fmt.Errorf("managed control connection is not ready")
	}
	raw, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	_, err = s.conn.Write(append(raw, '\n'))
	return err
}

func clonePayload(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func normalizeScope(scope string) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "public":
		return "public"
	default:
		return "private"
	}
}

func parseEndpoint(raw string) (network string, address string, err error) {
	endpoint := strings.TrimSpace(raw)
	if endpoint == "" {
		return "", "", fmt.Errorf("managed control endpoint is required")
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", fmt.Errorf("invalid managed control endpoint: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(u.Scheme)) {
	case "tcp":
		host := strings.TrimSpace(u.Host)
		if host == "" {
			return "", "", fmt.Errorf("managed control tcp host is required")
		}
		return "tcp", host, nil
	case "unix":
		path := strings.TrimSpace(u.Path)
		if path == "" {
			return "", "", fmt.Errorf("managed control unix path is required")
		}
		return "unix", path, nil
	default:
		return "", "", fmt.Errorf("unsupported managed control endpoint scheme: %s", u.Scheme)
	}
}

