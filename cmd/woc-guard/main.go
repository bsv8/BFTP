package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bsv8/BFTP/pkg/woc"
)

func main() {
	var (
		listenAddr      = flag.String("listen", envOr("BSV_CHAIN_GUARD_LISTEN", "127.0.0.1:18222"), "listen address")
		network         = flag.String("network", envOr("BSV_CHAIN_NETWORK", "test"), "chain network: test/main")
		intervalRaw     = flag.String("interval", envOr("BSV_CHAIN_PROTECT_INTERVAL", "1s"), "minimal interval between upstream API calls")
		metricsLogPath  = flag.String("metrics-log", envOr("BSV_CHAIN_GUARD_METRICS_LOG", "logs/woc-guard-metrics.log"), "metrics log file path (empty to disable)")
		metricsEveryRaw = flag.String("metrics-every", envOr("BSV_CHAIN_GUARD_METRICS_EVERY", "10s"), "metrics flush interval")
	)
	flag.Parse()

	interval, err := time.ParseDuration(strings.TrimSpace(*intervalRaw))
	if err != nil || interval <= 0 {
		fmt.Fprintf(os.Stderr, "invalid interval: %q\n", *intervalRaw)
		os.Exit(2)
	}

	metricsEvery, err := time.ParseDuration(strings.TrimSpace(*metricsEveryRaw))
	if err != nil || metricsEvery <= 0 {
		fmt.Fprintf(os.Stderr, "invalid metrics-every: %q\n", *metricsEveryRaw)
		os.Exit(2)
	}

	srv := woc.NewGuardServer(strings.TrimSpace(*network), "", interval)
	stopMetrics, err := startMetricsLogger(srv, strings.TrimSpace(*metricsLogPath), metricsEvery)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start metrics logger failed: %v\n", err)
		os.Exit(1)
	}
	if stopMetrics != nil {
		defer stopMetrics()
	}
	httpSrv := &http.Server{
		Addr:              strings.TrimSpace(*listenAddr),
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Printf("woc-guard listening on %s, network=%s, interval=%s\n", httpSrv.Addr, strings.TrimSpace(*network), interval.String())
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "woc-guard server failed: %v\n", err)
		os.Exit(1)
	}
}

func startMetricsLogger(srv *woc.GuardServer, logPath string, every time.Duration) (func(), error) {
	if srv == nil || strings.TrimSpace(logPath) == "" {
		return nil, nil
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		defer f.Close()
		for {
			select {
			case <-ticker.C:
				_ = writeMetricLine(f, srv.Stats())
			case <-done:
				_ = writeMetricLine(f, srv.Stats())
				return
			}
		}
	}()
	return func() { close(done) }, nil
}

func writeMetricLine(w io.Writer, stats woc.GuardStats) error {
	rec := struct {
		TimestampUTC string         `json:"timestamp_utc"`
		Stats        woc.GuardStats `json:"stats"`
	}{
		TimestampUTC: time.Now().UTC().Format(time.RFC3339Nano),
		Stats:        stats,
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	_, err = w.Write(append(b, '\n'))
	return err
}

func envOr(name, def string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return def
}
