package lhttp

import (
	"errors"
	"net"
	"net/http"
	"time"
)

type Route struct {
	Path    string
	Handler http.HandlerFunc
}

func NewServeMux(routes ...Route) *http.ServeMux {
	mux := http.NewServeMux()
	for _, route := range routes {
		if route.Handler == nil {
			continue
		}
		mux.HandleFunc(route.Path, route.Handler)
	}
	return mux
}

type ServerOptions struct {
	ListenAddr        string
	Handler           http.Handler
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type StartedServer struct {
	Server   *http.Server
	Listener net.Listener
}

func StartServer(opts ServerOptions) (*StartedServer, error) {
	srv := &http.Server{
		Addr:              opts.ListenAddr,
		Handler:           opts.Handler,
		ReadTimeout:       withDefaultDuration(opts.ReadTimeout, 10*time.Second),
		ReadHeaderTimeout: withDefaultDuration(opts.ReadHeaderTimeout, 5*time.Second),
		WriteTimeout:      withDefaultDuration(opts.WriteTimeout, 30*time.Second),
		IdleTimeout:       withDefaultDuration(opts.IdleTimeout, 60*time.Second),
	}
	ln, err := net.Listen("tcp", opts.ListenAddr)
	if err != nil {
		return nil, err
	}
	return &StartedServer{
		Server:   srv,
		Listener: ln,
	}, nil
}

func ServeInBackground(started *StartedServer, onError func(error)) {
	if started == nil || started.Server == nil || started.Listener == nil {
		return
	}
	go func() {
		if err := started.Server.Serve(started.Listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			if onError != nil {
				onError(err)
			}
		}
	}()
}

func withDefaultDuration(value time.Duration, fallback time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return fallback
}
