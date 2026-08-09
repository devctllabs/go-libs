package postgresdb

import (
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// Telemetry configures PostgreSQL tracing and pool metrics without consulting
// OpenTelemetry global providers. Query text is excluded by default and query
// parameters are never recorded.
type Telemetry struct {
	TracerProvider   trace.TracerProvider
	MeterProvider    metric.MeterProvider
	IncludeQueryText bool
}

type telemetryConfig struct {
	tracerProvider   trace.TracerProvider
	meterProvider    metric.MeterProvider
	includeQueryText bool
}

func newTelemetryConfig(config Telemetry) telemetryConfig {
	tracerProvider := config.TracerProvider
	if tracerProvider == nil {
		tracerProvider = tracenoop.NewTracerProvider()
	}
	meterProvider := config.MeterProvider
	if meterProvider == nil {
		meterProvider = metricnoop.NewMeterProvider()
	}
	return telemetryConfig{
		tracerProvider:   tracerProvider,
		meterProvider:    meterProvider,
		includeQueryText: config.IncludeQueryText,
	}
}

func (config telemetryConfig) tracer(role string) *otelpgx.Tracer {
	attributes := databaseAttributes(role)
	options := []otelpgx.Option{
		otelpgx.WithTracerProvider(config.tracerProvider),
		otelpgx.WithMeterProvider(config.meterProvider),
		otelpgx.WithTracerAttributes(attributes...),
		otelpgx.WithMeterAttributes(attributes...),
		otelpgx.WithTrimSQLInSpanName(),
	}
	if !config.includeQueryText {
		options = append(options, otelpgx.WithDisableSQLStatementInAttributes())
	}
	return otelpgx.NewTracer(options...)
}

func (config telemetryConfig) recordPoolStats(pool *pgxpool.Pool, role string) error {
	return otelpgx.RecordStats(
		pool,
		otelpgx.WithStatsMeterProvider(config.meterProvider),
		otelpgx.WithStatsAttributes(databaseAttributes(role)...),
	)
}

func databaseAttributes(role string) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.DBSystemNamePostgreSQL,
		semconv.DBClientConnectionPoolName(role),
		attribute.String("db.role", role),
	}
}
