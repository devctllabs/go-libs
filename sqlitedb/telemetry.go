package sqlitedb

import (
	"context"

	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel/attribute"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

func telemetryOptions(config Telemetry, role string) ([]otelsql.Option, []otelsql.Option) {
	tracerProvider := config.TracerProvider
	if tracerProvider == nil {
		tracerProvider = tracenoop.NewTracerProvider()
	}
	meterProvider := config.MeterProvider
	if meterProvider == nil {
		meterProvider = metricnoop.NewMeterProvider()
	}
	attributes := []attribute.KeyValue{
		attribute.String("db.system.name", "sqlite"),
		attribute.String("db.client.connection.pool.name", role),
		attribute.String("db.role", role),
	}
	common := []otelsql.Option{
		otelsql.WithTracerProvider(tracerProvider),
		otelsql.WithMeterProvider(meterProvider),
		otelsql.WithAttributes(attributes...),
	}
	query := append([]otelsql.Option(nil), common...)
	query = append(query,
		otelsql.WithSpanOptions(otelsql.SpanOptions{DisableQuery: !config.IncludeQueryText}),
		otelsql.WithSpanNameFormatter(func(_ context.Context, method otelsql.Method, _ string) string {
			return string(method)
		}),
	)
	return query, common
}
