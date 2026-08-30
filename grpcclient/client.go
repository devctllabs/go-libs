package grpcclient

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/stats"
)

// Config contains connection settings that are stable across environments.
type Config struct {
	Target string
	TLS    TLSConfig
}

// TLSConfig configures file-backed client transport credentials.
type TLSConfig struct {
	Enabled         bool
	RootCAFile      string
	CertificateFile string
	PrivateKeyFile  string
	ServerName      string
}

// Telemetry contains the OpenTelemetry dependencies used by the client.
// Nil members disable the corresponding signal instead of consulting globals.
type Telemetry struct {
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
	Propagator     propagation.TextMapPropagator
}

// Option configures a connection during construction.
type Option interface{ apply(*clientConfig) error }

type optionFunc func(*clientConfig) error

func (f optionFunc) apply(cfg *clientConfig) error { return f(cfg) }

type clientConfig struct {
	unaryInterceptors  []grpc.UnaryClientInterceptor
	streamInterceptors []grpc.StreamClientInterceptor
	dialOptions        []grpc.DialOption
	telemetry          *Telemetry
}

// WithTelemetry enables explicitly supplied OpenTelemetry providers.
func WithTelemetry(telemetry Telemetry) Option {
	return optionFunc(func(cfg *clientConfig) error {
		cfg.telemetry = &telemetry
		return nil
	})
}

// WithStreamInterceptors appends interceptors in execution order.
func WithStreamInterceptors(interceptors ...grpc.StreamClientInterceptor) Option {
	return optionFunc(func(cfg *clientConfig) error {
		for _, interceptor := range interceptors {
			if interceptor == nil {
				return errors.New("grpcclient: stream interceptor must not be nil")
			}
			cfg.streamInterceptors = append(cfg.streamInterceptors, interceptor)
		}
		return nil
	})
}

// WithUnaryInterceptors appends interceptors in execution order.
func WithUnaryInterceptors(interceptors ...grpc.UnaryClientInterceptor) Option {
	return optionFunc(func(cfg *clientConfig) error {
		for _, interceptor := range interceptors {
			if interceptor == nil {
				return errors.New("grpcclient: unary interceptor must not be nil")
			}
			cfg.unaryInterceptors = append(cfg.unaryInterceptors, interceptor)
		}
		return nil
	})
}

// WithDialOptions appends native options. Callers must not duplicate transport
// credentials, stats handlers, or interceptor chains owned by this package.
func WithDialOptions(options ...grpc.DialOption) Option {
	return optionFunc(func(cfg *clientConfig) error {
		cfg.dialOptions = append(cfg.dialOptions, options...)
		return nil
	})
}

// New creates a non-blocking gRPC client connection. The caller owns Close.
func New(config Config, options ...Option) (*grpc.ClientConn, error) {
	if strings.TrimSpace(config.Target) == "" {
		return nil, errors.New("grpcclient: target must not be blank")
	}
	if config.TLS.Enabled && (config.TLS.CertificateFile == "") != (config.TLS.PrivateKeyFile == "") {
		return nil, errors.New("grpcclient: TLS certificate file and private key file must be provided together")
	}
	var cfg clientConfig
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option.apply(&cfg); err != nil {
			return nil, err
		}
	}
	dialOptions := append([]grpc.DialOption(nil), cfg.dialOptions...)
	transportCredentials, err := clientCredentials(config.TLS)
	if err != nil {
		return nil, err
	}
	dialOptions = append(dialOptions, grpc.WithTransportCredentials(transportCredentials))
	if cfg.telemetry != nil {
		dialOptions = append(dialOptions, grpc.WithStatsHandler(newClientStatsHandler(*cfg.telemetry)))
	}
	if len(cfg.unaryInterceptors) > 0 {
		dialOptions = append(dialOptions, grpc.WithChainUnaryInterceptor(cfg.unaryInterceptors...))
	}
	if len(cfg.streamInterceptors) > 0 {
		dialOptions = append(dialOptions, grpc.WithChainStreamInterceptor(cfg.streamInterceptors...))
	}
	return grpc.NewClient(config.Target, dialOptions...)
}

func newClientStatsHandler(telemetry Telemetry) stats.Handler {
	tracerProvider := telemetry.TracerProvider
	if tracerProvider == nil {
		tracerProvider = tracenoop.NewTracerProvider()
	}
	meterProvider := telemetry.MeterProvider
	if meterProvider == nil {
		meterProvider = metricnoop.NewMeterProvider()
	}
	propagator := telemetry.Propagator
	if propagator == nil {
		propagator = propagation.NewCompositeTextMapPropagator()
	}
	return otelgrpc.NewClientHandler(
		otelgrpc.WithTracerProvider(tracerProvider),
		otelgrpc.WithMeterProvider(meterProvider),
		otelgrpc.WithPropagators(propagator),
	)
}

func clientCredentials(cfg TLSConfig) (credentials.TransportCredentials, error) {
	if !cfg.Enabled {
		return insecure.NewCredentials(), nil
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: cfg.ServerName}
	if cfg.RootCAFile != "" {
		roots, err := x509.SystemCertPool()
		if err != nil || roots == nil {
			roots = x509.NewCertPool()
		}
		pem, err := os.ReadFile(cfg.RootCAFile)
		if err != nil {
			return nil, fmt.Errorf("grpcclient: read root CA: %w", err)
		}
		if !roots.AppendCertsFromPEM(pem) {
			return nil, errors.New("grpcclient: root CA file contains no certificates")
		}
		tlsConfig.RootCAs = roots
	}
	if cfg.CertificateFile != "" {
		certificate, err := tls.LoadX509KeyPair(cfg.CertificateFile, cfg.PrivateKeyFile)
		if err != nil {
			return nil, fmt.Errorf("grpcclient: load TLS key pair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}
	}
	return credentials.NewTLS(tlsConfig), nil
}
