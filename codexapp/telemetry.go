package codexapp

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

const instrumentationName = "github.com/devctllabs/go-libs/codexapp"

type instrumentation struct {
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
	requests   metric.Int64Counter
	duration   metric.Float64Histogram
}

func newInstrumentation(config Telemetry) (*instrumentation, error) {
	tracerProvider := config.TracerProvider
	if tracerProvider == nil {
		tracerProvider = tracenoop.NewTracerProvider()
	}
	meterProvider := config.MeterProvider
	if meterProvider == nil {
		meterProvider = metricnoop.NewMeterProvider()
	}
	propagator := config.Propagator
	if propagator == nil {
		propagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	}
	meter := meterProvider.Meter(instrumentationName)
	requests, err := meter.Int64Counter("codexapp.rpc.requests")
	if err != nil {
		return nil, err
	}
	duration, err := meter.Float64Histogram("codexapp.rpc.duration", metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}
	return &instrumentation{
		tracer:     tracerProvider.Tracer(instrumentationName),
		propagator: propagator,
		requests:   requests,
		duration:   duration,
	}, nil
}

func (i *instrumentation) startRPC(ctx context.Context, method string) (context.Context, func(error)) {
	attributes := []attribute.KeyValue{
		attribute.String("rpc.system", "jsonrpc"),
		attribute.String("rpc.method", method),
	}
	startedAt := time.Now()
	ctx, span := i.tracer.Start(ctx, "codexapp."+method,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attributes...),
	)
	return ctx, func(err error) {
		status := "ok"
		if err != nil {
			status = "error"
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		metricAttributes := metric.WithAttributes(
			attribute.String("rpc.method", method),
			attribute.String("status", status),
		)
		i.requests.Add(ctx, 1, metricAttributes)
		i.duration.Record(ctx, time.Since(startedAt).Seconds(), metricAttributes)
		span.End()
	}
}

func (i *instrumentation) inject(ctx context.Context) map[string]string {
	carrier := propagation.MapCarrier{}
	i.propagator.Inject(ctx, carrier)
	if len(carrier) == 0 {
		return nil
	}
	return map[string]string(carrier)
}

func (i *instrumentation) extract(ctx context.Context, headers map[string]string) context.Context {
	if len(headers) == 0 {
		return ctx
	}
	return i.propagator.Extract(ctx, propagation.MapCarrier(headers))
}
