package kafkaoutbox

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
)

type persistedTraceContext struct {
	traceparent *string
	tracestate  *string
}

func traceContextFrom(ctx context.Context) persistedTraceContext {
	carrier := propagation.MapCarrier{}
	propagation.TraceContext{}.Inject(ctx, carrier)
	traceparent := carrier.Get("traceparent")
	if traceparent == "" {
		return persistedTraceContext{}
	}
	result := persistedTraceContext{traceparent: &traceparent}
	if tracestate := carrier.Get("tracestate"); tracestate != "" {
		result.tracestate = &tracestate
	}
	return result
}

func (traceContext persistedTraceContext) debeziumProperties() *string {
	if traceContext.traceparent == nil {
		return nil
	}
	properties := "traceparent=" + *traceContext.traceparent + "\n"
	if traceContext.tracestate != nil {
		properties += "tracestate=" + *traceContext.tracestate + "\n"
	}
	return &properties
}
