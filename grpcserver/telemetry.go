package grpcserver

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc/stats"
)

// Telemetry contains the OpenTelemetry dependencies used by the server.
// Nil members disable the corresponding signal instead of consulting globals.
type Telemetry struct {
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
	Propagator     propagation.TextMapPropagator
}

// WithTelemetry enables explicitly supplied OpenTelemetry providers.
func WithTelemetry(telemetry Telemetry) Option {
	return optionFunc(func(cfg *serverConfig) error {
		cfg.telemetry = &telemetry
		return nil
	})
}

type panicMetricObserver struct{ counter metric.Int64Counter }

func newPanicMetricObserver(provider metric.MeterProvider) (PanicObserver, error) {
	counter, err := provider.Meter("github.com/devctllabs/go-libs/grpcserver").Int64Counter(
		"grpc.server.panics",
		metric.WithUnit("{panic}"),
		metric.WithDescription("Number of recovered gRPC server panics"),
	)
	if err != nil {
		return nil, err
	}
	return panicMetricObserver{counter: counter}, nil
}

func (o panicMetricObserver) ObservePanic(ctx context.Context, event PanicEvent) {
	o.counter.Add(ctx, 1, metric.WithAttributes(attribute.String("rpc.grpc.full_method", event.FullMethod)))
}

func newServerStatsHandler(telemetry Telemetry) stats.Handler {
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
	return otelgrpc.NewServerHandler(
		otelgrpc.WithTracerProvider(tracerProvider),
		otelgrpc.WithMeterProvider(meterProvider),
		otelgrpc.WithPropagators(propagator),
	)
}
