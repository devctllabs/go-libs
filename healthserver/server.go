package healthserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/devctllabs/go-libs/health"
	"github.com/labstack/echo/v5"
)

type serverConfig struct {
	address string
}

// Option configures a standalone Server during construction.
type Option interface {
	apply(*serverConfig) error
}

type serverOptionFunc func(*serverConfig) error

func (f serverOptionFunc) apply(cfg *serverConfig) error {
	return f(cfg)
}

// WithAddress replaces the default standalone address :8081.
func WithAddress(address string) Option {
	return serverOptionFunc(func(cfg *serverConfig) error {
		if strings.TrimSpace(address) == "" {
			return errors.New("healthserver: address must not be blank")
		}
		cfg.address = address
		return nil
	})
}

// Server owns the standalone health HTTP server lifecycle.
type Server struct {
	http *http.Server
}

// NewServer creates a standalone Echo health server without starting it.
func NewServer(probes *health.Probes, options ...Option) (*Server, error) {
	if probes == nil {
		return nil, errors.New("healthserver: Probes must not be nil")
	}

	cfg := serverConfig{address: ":8081"}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option.apply(&cfg); err != nil {
			return nil, err
		}
	}

	e := echo.New()
	if err := Register(e, probes); err != nil {
		return nil, err
	}
	return &Server{http: &http.Server{
		Addr:              cfg.address,
		Handler:           e,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      probes.CheckTimeout() + time.Second,
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
		return errors.New("healthserver: listener must not be nil")
	}
	return normalizeServerError(s.http.Serve(listener))
}

// Shutdown gracefully stops the standalone server using ctx as its deadline.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func normalizeServerError(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
