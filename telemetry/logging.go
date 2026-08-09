package telemetry

import (
	"context"
	"errors"

	"github.com/go-logr/zapr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// SetGlobalLogger routes OpenTelemetry internal diagnostics to a named zap logger.
// It changes process-global OpenTelemetry logger state and does not restore the previous logger.
func SetGlobalLogger(logger *zap.Logger) error {
	if logger == nil {
		return errors.New("telemetry: global logger must not be nil")
	}
	otel.SetLogger(zapr.NewLogger(logger.Named("opentelemetry")))
	return nil
}

// WithTraceContext returns logger enriched with trace_id and span_id when ctx contains a valid
// span context. It returns logger unchanged when no valid span context exists.
func WithTraceContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	spanContext := trace.SpanContextFromContext(ctx)
	if logger == nil || !spanContext.IsValid() {
		return logger
	}
	return logger.With(
		zap.String("trace_id", spanContext.TraceID().String()),
		zap.String("span_id", spanContext.SpanID().String()),
	)
}
