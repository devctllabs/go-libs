package telemetry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// Config identifies the service and controls telemetry initialization.
// A zero Config disables telemetry and does not read OpenTelemetry environment variables.
type Config struct {
	Enabled               bool
	ServiceName           string
	ServiceVersion        string
	DeploymentEnvironment string
}

// Runtime owns the providers and lifecycle of one telemetry instance.
// It does not install process-global providers or propagators.
type Runtime struct {
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
	propagator     propagation.TextMapPropagator

	metricForceFlush func(context.Context) error
	traceForceFlush  func(context.Context) error
	metricShutdown   func(context.Context) error
	traceShutdown    func(context.Context) error

	shutdownOnce sync.Once
	shutdownErr  error
}

// Open constructs explicitly injectable trace and metric providers from Config and standard
// OpenTelemetry environment variables. Only OTLP push export and the none exporter are supported.
func Open(ctx context.Context, cfg Config) (*Runtime, error) {
	propagator := newPropagator()
	if !cfg.Enabled {
		return newRuntime(
			tracenoop.NewTracerProvider(),
			metricnoop.NewMeterProvider(),
			propagator,
			nil, nil, nil, nil,
		), nil
	}
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	signals, err := readSignalConfig()
	if err != nil {
		return nil, err
	}
	res, err := buildResource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("build telemetry resource: %w", err)
	}
	providers, err := openTelemetryProviders(ctx, signals, res)
	if err != nil {
		return nil, err
	}
	return providers.runtime(propagator), nil
}

type telemetryProviders struct {
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
	traceSDK       *sdktrace.TracerProvider
	metricSDK      *sdkmetric.MeterProvider
}

func openTelemetryProviders(ctx context.Context, signals signalsConfig, res *resource.Resource) (*telemetryProviders, error) {
	traceExporter, err := openTraceExporter(ctx, signals.traces)
	if err != nil {
		return nil, err
	}
	metricExporter, err := openMetricExporter(ctx, signals.metrics)
	if err != nil {
		return nil, errors.Join(err, shutdownTraceExporter(ctx, traceExporter))
	}
	tracerProvider, traceSDK := buildTracerProvider(res, traceExporter)
	meterProvider, metricSDK, err := buildMeterProvider(res, metricExporter)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("start Go runtime metrics: %w", err),
			metricSDK.Shutdown(ctx),
			shutdownTracerProvider(ctx, traceSDK),
		)
	}
	return &telemetryProviders{
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
		traceSDK:       traceSDK,
		metricSDK:      metricSDK,
	}, nil
}

func buildTracerProvider(res *resource.Resource, exporter sdktrace.SpanExporter) (trace.TracerProvider, *sdktrace.TracerProvider) {
	if exporter == nil {
		return tracenoop.NewTracerProvider(), nil
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)
	return provider, provider
}

func buildMeterProvider(res *resource.Resource, exporter sdkmetric.Exporter) (metric.MeterProvider, *sdkmetric.MeterProvider, error) {
	if exporter == nil {
		return metricnoop.NewMeterProvider(), nil, nil
	}
	reader := sdkmetric.NewPeriodicReader(exporter)
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)
	if err := startRuntimeMetrics(provider); err != nil {
		return nil, provider, err
	}
	return provider, provider, nil
}

func (p *telemetryProviders) runtime(propagator propagation.TextMapPropagator) *Runtime {
	return newRuntime(
		p.tracerProvider,
		p.meterProvider,
		propagator,
		forceFlushMeterProvider(p.metricSDK),
		forceFlushTracerProvider(p.traceSDK),
		shutdownMeterProvider(p.metricSDK),
		shutdownTracerProviderFunc(p.traceSDK),
	)
}

// TracerProvider returns the instance-owned provider for explicit injection.
func (r *Runtime) TracerProvider() trace.TracerProvider {
	return r.tracerProvider
}

// MeterProvider returns the instance-owned provider for explicit injection.
func (r *Runtime) MeterProvider() metric.MeterProvider {
	return r.meterProvider
}

// Propagator returns the W3C Trace Context and Baggage propagator for explicit injection.
func (r *Runtime) Propagator() propagation.TextMapPropagator {
	return r.propagator
}

// ForceFlush immediately flushes metrics, then traces, and joins signal errors.
func (r *Runtime) ForceFlush(ctx context.Context) error {
	return errors.Join(callLifecycle(ctx, r.metricForceFlush), callLifecycle(ctx, r.traceForceFlush))
}

// Shutdown stops metrics, then traces. It is safe for concurrent use and returns the first
// shutdown result to every caller.
func (r *Runtime) Shutdown(ctx context.Context) error {
	r.shutdownOnce.Do(func() {
		r.shutdownErr = errors.Join(
			callLifecycle(ctx, r.metricShutdown),
			callLifecycle(ctx, r.traceShutdown),
		)
	})
	return r.shutdownErr
}

func validateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.ServiceName) == "" {
		return errors.New("telemetry Config.ServiceName must not be blank")
	}
	if strings.TrimSpace(cfg.ServiceVersion) == "" {
		return errors.New("telemetry Config.ServiceVersion must not be blank")
	}
	if strings.TrimSpace(cfg.DeploymentEnvironment) == "" {
		return errors.New("telemetry Config.DeploymentEnvironment must not be blank")
	}
	return nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newRuntime(
	tracerProvider trace.TracerProvider,
	meterProvider metric.MeterProvider,
	propagator propagation.TextMapPropagator,
	metricForceFlush func(context.Context) error,
	traceForceFlush func(context.Context) error,
	metricShutdown func(context.Context) error,
	traceShutdown func(context.Context) error,
) *Runtime {
	return &Runtime{
		tracerProvider:   tracerProvider,
		meterProvider:    meterProvider,
		propagator:       propagator,
		metricForceFlush: metricForceFlush,
		traceForceFlush:  traceForceFlush,
		metricShutdown:   metricShutdown,
		traceShutdown:    traceShutdown,
	}
}

func startRuntimeMetrics(provider metric.MeterProvider) error {
	return otelruntime.Start(otelruntime.WithMeterProvider(provider))
}

func callLifecycle(ctx context.Context, call func(context.Context) error) error {
	if call == nil {
		return nil
	}
	return call(ctx)
}

func forceFlushMeterProvider(provider *sdkmetric.MeterProvider) func(context.Context) error {
	if provider == nil {
		return nil
	}
	return provider.ForceFlush
}

func forceFlushTracerProvider(provider *sdktrace.TracerProvider) func(context.Context) error {
	if provider == nil {
		return nil
	}
	return provider.ForceFlush
}

func shutdownMeterProvider(provider *sdkmetric.MeterProvider) func(context.Context) error {
	if provider == nil {
		return nil
	}
	return provider.Shutdown
}

func shutdownTracerProviderFunc(provider *sdktrace.TracerProvider) func(context.Context) error {
	if provider == nil {
		return nil
	}
	return provider.Shutdown
}

func shutdownTracerProvider(ctx context.Context, provider *sdktrace.TracerProvider) error {
	if provider == nil {
		return nil
	}
	return provider.Shutdown(ctx)
}
