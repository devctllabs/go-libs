package telemetry

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

func TestRuntimeForceFlushJoinsSignalErrorsInMetricTraceOrder(t *testing.T) {
	t.Parallel()
	metricErr := errors.New("metric flush")
	traceErr := errors.New("trace flush")
	var order []string
	runtime := newRuntime(
		tracenoop.NewTracerProvider(),
		metricnoop.NewMeterProvider(),
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
		func(context.Context) error {
			order = append(order, "metric")
			return metricErr
		},
		func(context.Context) error {
			order = append(order, "trace")
			return traceErr
		},
		nil,
		nil,
	)

	err := runtime.ForceFlush(context.Background())
	require.ErrorIs(t, err, metricErr)
	require.ErrorIs(t, err, traceErr)
	require.Equal(t, []string{"metric", "trace"}, order)
}

func TestRuntimeShutdownIsConcurrentSafeAndIdempotent(t *testing.T) {
	t.Parallel()
	metricErr := errors.New("metric shutdown")
	traceErr := errors.New("trace shutdown")
	var metricCalls atomic.Int32
	var traceCalls atomic.Int32
	var orderMu sync.Mutex
	var order []string
	runtime := newRuntime(
		tracenoop.NewTracerProvider(),
		metricnoop.NewMeterProvider(),
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
		nil,
		nil,
		func(context.Context) error {
			metricCalls.Add(1)
			orderMu.Lock()
			order = append(order, "metric")
			orderMu.Unlock()
			return metricErr
		},
		func(context.Context) error {
			traceCalls.Add(1)
			orderMu.Lock()
			order = append(order, "trace")
			orderMu.Unlock()
			return traceErr
		},
	)

	const callers = 32
	errs := make(chan error, callers)
	var wg sync.WaitGroup
	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- runtime.Shutdown(context.Background())
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		require.ErrorIs(t, err, metricErr)
		require.ErrorIs(t, err, traceErr)
	}
	require.Equal(t, int32(1), metricCalls.Load())
	require.Equal(t, int32(1), traceCalls.Load())
	require.Equal(t, []string{"metric", "trace"}, order)
}
