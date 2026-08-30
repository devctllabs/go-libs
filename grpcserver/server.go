package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"buf.build/go/protovalidate"
	grpcprotovalidate "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Config contains the server settings that are stable across environments.
type Config struct {
	Address                  string
	TLS                      TLSConfig
	DisableDefaultReflection bool
	DisableDefaultValidation bool
	DisableDefaultRecovery   bool
}

// TLSConfig configures file-backed server transport credentials.
type TLSConfig struct {
	Enabled         bool
	CertificateFile string
	PrivateKeyFile  string
	ClientCAFile    string
}

// Option configures a Server during construction.
type Option interface {
	apply(*serverConfig) error
}

type optionFunc func(*serverConfig) error

func (f optionFunc) apply(cfg *serverConfig) error { return f(cfg) }

type serverConfig struct {
	panicObservers     []PanicObserver
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	serverOptions      []grpc.ServerOption
	telemetry          *Telemetry
}

// WithServerOptions appends native options. Callers must not duplicate
// transport credentials, stats handlers, or interceptor chains owned here.
func WithServerOptions(options ...grpc.ServerOption) Option {
	return optionFunc(func(cfg *serverConfig) error {
		cfg.serverOptions = append(cfg.serverOptions, options...)
		return nil
	})
}

// WithStreamInterceptors appends interceptors in execution order.
func WithStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) Option {
	return optionFunc(func(cfg *serverConfig) error {
		for _, interceptor := range interceptors {
			if interceptor == nil {
				return errors.New("grpcserver: stream interceptor must not be nil")
			}
			cfg.streamInterceptors = append(cfg.streamInterceptors, interceptor)
		}
		return nil
	})
}

// WithUnaryInterceptors appends interceptors in execution order.
func WithUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) Option {
	return optionFunc(func(cfg *serverConfig) error {
		for _, interceptor := range interceptors {
			if interceptor == nil {
				return errors.New("grpcserver: unary interceptor must not be nil")
			}
			cfg.unaryInterceptors = append(cfg.unaryInterceptors, interceptor)
		}
		return nil
	})
}

// Server owns a gRPC server and its lifecycle.
type Server struct {
	address string
	grpc    *grpc.Server
}

// New constructs a server without opening a listener.
func New(config Config, options ...Option) (*Server, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	var cfg serverConfig
	if err := applyOptions(&cfg, options); err != nil {
		return nil, err
	}
	grpcOptions, err := buildGRPCOptions(config, &cfg)
	if err != nil {
		return nil, err
	}
	server := &Server{address: config.Address, grpc: grpc.NewServer(grpcOptions...)}
	if !config.DisableDefaultReflection {
		reflection.Register(server)
	}
	return server, nil
}

func validateConfig(config Config) error {
	if strings.TrimSpace(config.Address) == "" {
		return errors.New("grpcserver: address must not be blank")
	}
	if config.TLS.Enabled && strings.TrimSpace(config.TLS.CertificateFile) == "" {
		return errors.New("grpcserver: TLS certificate file must not be blank")
	}
	if config.TLS.Enabled && strings.TrimSpace(config.TLS.PrivateKeyFile) == "" {
		return errors.New("grpcserver: TLS private key file must not be blank")
	}
	return nil
}

func applyOptions(cfg *serverConfig, options []Option) error {
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option.apply(cfg); err != nil {
			return err
		}
	}
	return nil
}

func buildGRPCOptions(config Config, cfg *serverConfig) ([]grpc.ServerOption, error) {
	if cfg.telemetry != nil && cfg.telemetry.MeterProvider != nil {
		observer, err := newPanicMetricObserver(cfg.telemetry.MeterProvider)
		if err != nil {
			return nil, fmt.Errorf("grpcserver: create panic counter: %w", err)
		}
		cfg.panicObservers = append(cfg.panicObservers, observer)
	}

	unaryInterceptors := make([]grpc.UnaryServerInterceptor, 0, len(cfg.unaryInterceptors)+2)
	streamInterceptors := make([]grpc.StreamServerInterceptor, 0, len(cfg.streamInterceptors)+2)
	if !config.DisableDefaultRecovery {
		unaryInterceptors = append(unaryInterceptors, recoveryUnaryInterceptor(cfg.panicObservers))
		streamInterceptors = append(streamInterceptors, recoveryStreamInterceptor(cfg.panicObservers))
	}
	unaryInterceptors = append(unaryInterceptors, cfg.unaryInterceptors...)
	streamInterceptors = append(streamInterceptors, cfg.streamInterceptors...)
	if !config.DisableDefaultValidation {
		validator, err := protovalidate.New()
		if err != nil {
			return nil, err
		}
		unaryInterceptors = append(unaryInterceptors, grpcprotovalidate.UnaryServerInterceptor(validator))
		streamInterceptors = append(streamInterceptors, grpcprotovalidate.StreamServerInterceptor(validator))
	}
	grpcOptions := append([]grpc.ServerOption(nil), cfg.serverOptions...)
	if cfg.telemetry != nil {
		grpcOptions = append(grpcOptions, grpc.StatsHandler(newServerStatsHandler(*cfg.telemetry)))
	}
	if len(unaryInterceptors) > 0 {
		grpcOptions = append(grpcOptions, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	}
	if len(streamInterceptors) > 0 {
		grpcOptions = append(grpcOptions, grpc.ChainStreamInterceptor(streamInterceptors...))
	}
	if config.TLS.Enabled {
		tlsOption, err := serverTLSOption(config.TLS)
		if err != nil {
			return nil, err
		}
		grpcOptions = append(grpcOptions, tlsOption)
	}
	return grpcOptions, nil
}

// Address returns the configured listen address.
func (s *Server) Address() string { return s.address }

// ListenAndServe listens on Address and serves until stopped or an error occurs.
func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("grpcserver: listen: %w", err)
	}
	return s.Serve(listener)
}

// RegisterService implements grpc.ServiceRegistrar.
func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl any) {
	s.grpc.RegisterService(desc, impl)
}

// GetServiceInfo returns metadata for registered services.
func (s *Server) GetServiceInfo() map[string]grpc.ServiceInfo {
	return s.grpc.GetServiceInfo()
}

// Serve accepts connections from listener until stopped or an error occurs.
func (s *Server) Serve(listener net.Listener) error {
	if listener == nil {
		return errors.New("grpcserver: listener must not be nil")
	}
	err := s.grpc.Serve(listener)
	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}

// Shutdown gracefully stops the server. If ctx expires, active RPCs are
// forcefully stopped and ctx.Err is returned.
func (s *Server) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.grpc.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.grpc.Stop()
		<-done
		return ctx.Err()
	}
}
