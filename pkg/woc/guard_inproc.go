package woc

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

type inprocGuardEntry struct {
	handler http.Handler
	refs    int
}

var inprocGuardRegistry struct {
	mu      sync.RWMutex
	entries map[string]*inprocGuardEntry
}

func ensureInprocGuard(baseURL string, srv *GuardServer) (func(), error) {
	if srv == nil {
		return nil, nil
	}
	key := normalizeGuardBaseURL(baseURL)
	handler := srv.Handler()

	inprocGuardRegistry.mu.Lock()
	if inprocGuardRegistry.entries == nil {
		inprocGuardRegistry.entries = map[string]*inprocGuardEntry{}
	}
	if existing, ok := inprocGuardRegistry.entries[key]; ok && existing != nil {
		existing.refs++
		inprocGuardRegistry.mu.Unlock()
		return makeInprocGuardStop(key), nil
	}
	inprocGuardRegistry.entries[key] = &inprocGuardEntry{
		handler: handler,
		refs:    1,
	}
	inprocGuardRegistry.mu.Unlock()
	return makeInprocGuardStop(key), nil
}

func makeInprocGuardStop(key string) func() {
	var once sync.Once
	return func() {
		once.Do(func() {
			inprocGuardRegistry.mu.Lock()
			defer inprocGuardRegistry.mu.Unlock()
			entry, ok := inprocGuardRegistry.entries[key]
			if !ok || entry == nil {
				return
			}
			entry.refs--
			if entry.refs <= 0 {
				delete(inprocGuardRegistry.entries, key)
			}
		})
	}
}

func getInprocGuardHandler(baseURL string) (http.Handler, bool) {
	key := normalizeGuardBaseURL(baseURL)
	inprocGuardRegistry.mu.RLock()
	entry, ok := inprocGuardRegistry.entries[key]
	inprocGuardRegistry.mu.RUnlock()
	if !ok || entry == nil || entry.handler == nil {
		return nil, false
	}
	return entry.handler, true
}

// tryInprocGuardRequest 如果 baseURL 对应进程内 guard，则直接在内存中执行请求。
func tryInprocGuardRequest(ctx context.Context, baseURL string, method string, path string, body []byte, contentType string) (statusCode int, respBody []byte, handled bool, err error) {
	handler, ok := getInprocGuardHandler(baseURL)
	if !ok {
		return 0, nil, false, nil
	}

	targetPath := strings.TrimSpace(path)
	if targetPath == "" {
		targetPath = "/"
	}
	if !strings.HasPrefix(targetPath, "/") {
		targetPath = "/" + targetPath
	}

	reqURL := normalizeGuardBaseURL(baseURL) + targetPath
	req, reqErr := http.NewRequestWithContext(ctxOrBackground(ctx), strings.TrimSpace(method), reqURL, bytes.NewReader(body))
	if reqErr != nil {
		return 0, nil, true, reqErr
	}
	if strings.TrimSpace(contentType) != "" {
		req.Header.Set("Content-Type", contentType)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return resp.StatusCode, nil, true, readErr
	}
	return resp.StatusCode, raw, true, nil
}

func normalizeGuardBaseURL(baseURL string) string {
	u := strings.TrimSpace(baseURL)
	if u == "" {
		u = DefaultGuardBaseURL
	}
	return strings.TrimRight(u, "/")
}
