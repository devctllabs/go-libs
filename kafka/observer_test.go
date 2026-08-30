package kafka

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultiObserverComposesAttemptLifecycle(t *testing.T) {
	t.Parallel()

	var events []string
	first := observerFuncs{
		start: func(ctx context.Context, _ Attempt) (context.Context, AttemptDone) {
			events = append(events, "start-first")
			return ctx, func(error) { events = append(events, "done-first") }
		},
	}
	second := observerFuncs{
		start: func(ctx context.Context, _ Attempt) (context.Context, AttemptDone) {
			events = append(events, "start-second")
			return ctx, func(error) { events = append(events, "done-second") }
		},
	}
	observer, err := NewMultiObserver(first, second)
	require.NoError(t, err)

	_, done := observer.StartAttempt(context.Background(), Attempt{Phase: AttemptHandler, Attempt: 1})
	done(nil)

	require.Equal(t, []string{
		"start-first", "start-second", "done-second", "done-first",
	}, events)
}

type observerFuncs struct {
	start func(context.Context, Attempt) (context.Context, AttemptDone)
}

func (observer observerFuncs) StartAttempt(ctx context.Context, attempt Attempt) (context.Context, AttemptDone) {
	return observer.start(ctx, attempt)
}
func (observerFuncs) Retry(context.Context, RetryEvent)                    {}
func (observerFuncs) BatchCompleted(context.Context, BatchResult)          {}
func (observerFuncs) RecordDisposition(context.Context, RecordDisposition) {}
