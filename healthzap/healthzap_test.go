package healthzap_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/health"
	"github.com/devctllabs/go-libs/healthzap"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestObserverLogsEveryFailureAndOneRecovery(t *testing.T) {
	t.Parallel()
	core, observed := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	logObserver, err := healthzap.New(logger)
	require.NoError(t, err)

	criticalFailure := health.Observation{
		Name:     "postgres",
		Critical: true,
		Outcome:  health.OutcomeTimeout,
		Duration: time.Second,
		Err:      errors.New("database unavailable"),
	}
	logObserver.Observe(context.Background(), criticalFailure)
	logObserver.Observe(context.Background(), criticalFailure)
	logObserver.Observe(context.Background(), health.Observation{
		Name: "postgres", Outcome: health.OutcomeOK, Duration: 10 * time.Millisecond,
	})
	logObserver.Observe(context.Background(), health.Observation{
		Name: "postgres", Outcome: health.OutcomeOK, Duration: 10 * time.Millisecond,
	})
	logObserver.Observe(context.Background(), health.Observation{
		Name: "cache", Critical: false, Outcome: health.OutcomeError,
		Duration: 25 * time.Millisecond, Err: errors.New("cache unavailable"),
	})

	entries := observed.All()
	require.Len(t, entries, 4)
	require.Equal(t, zapcore.ErrorLevel, entries[0].Level)
	require.Equal(t, zapcore.ErrorLevel, entries[1].Level)
	require.Equal(t, zapcore.InfoLevel, entries[2].Level)
	require.Equal(t, "health check recovered", entries[2].Message)
	require.Equal(t, zapcore.WarnLevel, entries[3].Level)
	require.Equal(t, "health check failed", entries[3].Message)
	require.Equal(t, "cache", entries[3].ContextMap()["check.name"])
	require.Equal(t, false, entries[3].ContextMap()["check.critical"])
	require.Equal(t, "error", entries[3].ContextMap()["check.outcome"])
	require.Equal(t, "cache unavailable", entries[3].ContextMap()["error"])
}

func TestObserverDoesNotLogInitialSuccess(t *testing.T) {
	t.Parallel()
	core, observed := observer.New(zapcore.DebugLevel)
	logObserver, err := healthzap.New(zap.New(core))
	require.NoError(t, err)

	logObserver.Observe(context.Background(), health.Observation{
		Name: "postgres", Critical: true, Outcome: health.OutcomeOK,
	})
	require.Empty(t, observed.All())
}

func TestNewRejectsNilLogger(t *testing.T) {
	t.Parallel()
	_, err := healthzap.New(nil)
	require.ErrorContains(t, err, "must not be nil")
}

func TestObserverIsSafeForConcurrentFailures(t *testing.T) {
	t.Parallel()
	core, observed := observer.New(zapcore.DebugLevel)
	logObserver, err := healthzap.New(zap.New(core))
	require.NoError(t, err)
	failure := health.Observation{
		Name: "postgres", Critical: true, Outcome: health.OutcomeError,
		Err: errors.New("database unavailable"),
	}

	var wait sync.WaitGroup
	for range 100 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			logObserver.Observe(context.Background(), failure)
		}()
	}
	wait.Wait()
	logObserver.Observe(context.Background(), health.Observation{
		Name: "postgres", Critical: true, Outcome: health.OutcomeOK,
	})

	require.Len(t, observed.FilterLevelExact(zapcore.ErrorLevel).All(), 100)
	require.Len(t, observed.FilterLevelExact(zapcore.InfoLevel).All(), 1)
}
