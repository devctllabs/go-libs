package debugserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"
)

type serverConfig struct {
	address string
}

// Option configures a Server during construction.
type Option interface {
	apply(*serverConfig) error
}

type serverOptionFunc func(*serverConfig) error

func (f serverOptionFunc) apply(cfg *serverConfig) error {
	return f(cfg)
}

// WithAddress replaces the default loopback address 127.0.0.1:6060.
func WithAddress(address string) Option {
	return serverOptionFunc(func(cfg *serverConfig) error {
		if strings.TrimSpace(address) == "" {
			return errors.New("debugserver: address must not be blank")
		}
		cfg.address = address
		return nil
	})
}

// Server owns the standalone pprof HTTP server lifecycle.
type Server struct {
	http *http.Server
}

// NewServer creates a standalone pprof server without starting it.
func NewServer(options ...Option) (*Server, error) {
	cfg := serverConfig{address: "127.0.0.1:6060"}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option.apply(&cfg); err != nil {
			return nil, err
		}
	}

	return &Server{http: &http.Server{
		Addr:              cfg.address,
		Handler:           pprofMux(),
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       30 * time.Second,
	}}, nil
}

// Address returns the configured listen address.
func (s *Server) Address() string {
	return s.http.Addr
}

// ListenAndServe starts the configured TCP listener and blocks until shutdown or failure.
func (s *Server) ListenAndServe() error {
	return normalizeServerError(s.http.ListenAndServe())
}

// Serve accepts HTTP connections from listener and blocks until shutdown or failure.
func (s *Server) Serve(listener net.Listener) error {
	if listener == nil {
		return errors.New("debugserver: listener must not be nil")
	}
	return normalizeServerError(s.http.Serve(listener))
}

// Shutdown gracefully stops the server using ctx as its deadline.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func pprofMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /debug/pprof/", pprof.Index)
	mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
	return mux
}

func normalizeServerError(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
