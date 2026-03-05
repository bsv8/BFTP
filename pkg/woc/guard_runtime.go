package woc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GuardRuntimeOptions 描述 guard 的自动托管参数。
// 约束：guard 地址固定为本地默认地址，业务只需传 network。
type GuardRuntimeOptions struct {
	Network         string
	ProtectInterval time.Duration
	StartupTimeout  time.Duration
}

// EnsureGuardRunning 保证 guard 可用：
// - 优先复用已在运行的 guard；
// - 若本地 guard 不可达，则在当前主进程内自动拉起并返回 stop 函数。
func EnsureGuardRunning(ctx context.Context, opt GuardRuntimeOptions) (baseURL string, stop func(), err error) {
	baseURL = DefaultGuardBaseURL

	if err := guardHealthCheck(ctx, baseURL); err == nil {
		return baseURL, func() {}, nil
	}

	interval := opt.ProtectInterval
	if interval <= 0 {
		interval = 1 * time.Second
	}
	timeout := opt.StartupTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	stop, err = ensureInprocGuard(baseURL, NewGuardServer(opt.Network, "", interval))
	if err != nil {
		return "", nil, fmt.Errorf("guard unavailable and auto-start failed: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctxOrBackground(ctx), timeout)
	defer cancel()
	if err := waitGuardReady(waitCtx, baseURL); err != nil {
		if stop != nil {
			stop()
		}
		return "", nil, fmt.Errorf("guard auto-start failed: %w", err)
	}

	if stop == nil {
		stop = func() {}
	}
	return baseURL, stop, nil
}

func guardHealthCheck(ctx context.Context, baseURL string) error {
	if status, body, handled, err := tryInprocGuardRequest(ctx, baseURL, http.MethodGet, "/healthz", nil, ""); handled {
		if err != nil {
			return err
		}
		return validateGuardHealthResponse(status, body)
	}

	reqCtx, cancel := context.WithTimeout(ctxOrBackground(ctx), 1200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/healthz", nil)
	if err != nil {
		return err
	}
	resp, err := (&http.Client{Timeout: 1500 * time.Millisecond}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read healthz body: %w", err)
	}
	return validateGuardHealthResponse(resp.StatusCode, body)
}

func validateGuardHealthResponse(statusCode int, body []byte) error {
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("http %d", statusCode)
	}
	var out struct {
		OK      bool   `json:"ok"`
		Service string `json:"service"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("decode healthz response: %w", err)
	}
	if !out.OK {
		return fmt.Errorf("healthz not ok")
	}
	if strings.TrimSpace(out.Service) != GuardServiceName {
		return fmt.Errorf("unexpected healthz service: %s", strings.TrimSpace(out.Service))
	}
	if strings.TrimSpace(out.Version) == "" {
		return fmt.Errorf("unexpected healthz version: empty")
	}
	return nil
}

func waitGuardReady(ctx context.Context, baseURL string) error {
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := guardHealthCheck(ctx, baseURL); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func ctxOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
