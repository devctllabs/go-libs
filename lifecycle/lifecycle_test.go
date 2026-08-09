package lifecycle_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/lifecycle"
	"github.com/stretchr/testify/require"
)

func TestRunValidatesConfigBeforeStartingAnything(t *testing.T) {
	t.Parallel()

	validTask := lifecycle.Task{Name: "api", Run: func(context.Context) error { return nil }}
	validShutdown := func(context.Context) error { return nil }
	tests := []struct {
		name string
		ctx  context.Context
		cfg  lifecycle.Config
	}{
		{name: "nil context", cfg: lifecycle.Config{ShutdownTimeout: time.Second, Shutdown: validShutdown, Tasks: []lifecycle.Task{validTask}}},
		{name: "zero timeout", ctx: context.Background(), cfg: lifecycle.Config{Shutdown: validShutdown, Tasks: []lifecycle.Task{validTask}}},
		{name: "negative timeout", ctx: context.Background(), cfg: lifecycle.Config{ShutdownTimeout: -time.Second, Shutdown: validShutdown, Tasks: []lifecycle.Task{validTask}}},
		{name: "nil shutdown", ctx: context.Background(), cfg: lifecycle.Config{ShutdownTimeout: time.Second, Tasks: []lifecycle.Task{validTask}}},
		{name: "no tasks", ctx: context.Background(), cfg: lifecycle.Config{ShutdownTimeout: time.Second, Shutdown: validShutdown}},
		{name: "blank task name", ctx: context.Background(), cfg: lifecycle.Config{ShutdownTimeout: time.Second, Shutdown: validShutdown, Tasks: []lifecycle.Task{{Name: " ", Run: validTask.Run}}}},
		{name: "duplicate task name", ctx: context.Background(), cfg: lifecycle.Config{ShutdownTimeout: time.Second, Shutdown: validShutdown, Tasks: []lifecycle.Task{validTask, validTask}}},
		{name: "nil task run", ctx: context.Background(), cfg: lifecycle.Config{ShutdownTimeout: time.Second, Shutdown: validShutdown, Tasks: []lifecycle.Task{{Name: "api"}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var taskCalls atomic.Int32
			var shutdownCalls atomic.Int32
			for index := range tt.cfg.Tasks {
				if tt.cfg.Tasks[index].Run != nil {
					tt.cfg.Tasks[index].Run = func(context.Context) error {
						taskCalls.Add(1)
						return nil
					}
				}
			}
			if tt.cfg.Shutdown != nil {
				tt.cfg.Shutdown = func(context.Context) error {
					shutdownCalls.Add(1)
					return nil
				}
			}

			err := lifecycle.Run(tt.ctx, tt.cfg)

			require.Error(t, err)
			require.Zero(t, taskCalls.Load())
			require.Zero(t, shutdownCalls.Load())
		})
	}
}

func TestRunTreatsTaskReturningNilAsUnexpectedStop(t *testing.T) {
	t.Parallel()

	var shutdownCalls atomic.Int32
	err := lifecycle.Run(context.Background(), lifecycle.Config{
		ShutdownTimeout: time.Second,
		Shutdown: func(context.Context) error {
			shutdownCalls.Add(1)
			return nil
		},
		Tasks: []lifecycle.Task{{Name: "api", Run: func(context.Context) error { return nil }}},
	})

	require.ErrorIs(t, err, lifecycle.ErrTaskStopped)
	require.ErrorContains(t, err, "api")
	require.Equal(t, int32(1), shutdownCalls.Load())
}

func TestRunCancelsSiblingsAndShutsDownWithFreshContext(t *testing.T) {
	t.Parallel()

	type contextKey string
	const key contextKey = "request"
	taskErr := errors.New("serve API")
	parent := context.WithValue(context.Background(), key, "value")
	siblingStarted := make(chan struct{})
	var siblingStopped atomic.Bool
	var shutdownCalls atomic.Int32

	err := lifecycle.Run(parent, lifecycle.Config{
		ShutdownTimeout: time.Second,
		Shutdown: func(ctx context.Context) error {
			shutdownCalls.Add(1)
			require.NoError(t, ctx.Err())
			require.Equal(t, "value", ctx.Value(key))
			_, hasDeadline := ctx.Deadline()
			require.True(t, hasDeadline)
			return nil
		},
		Tasks: []lifecycle.Task{
			{
				Name: "api",
				Run: func(context.Context) error {
					<-siblingStarted
					return taskErr
				},
			},
			{
				Name: "consumer",
				Run: func(ctx context.Context) error {
					close(siblingStarted)
					<-ctx.Done()
					siblingStopped.Store(true)
					return ctx.Err()
				},
			},
		},
	})

	require.ErrorIs(t, err, taskErr)
	require.ErrorContains(t, err, "api")
	require.NotContains(t, err.Error(), "consumer")
	require.Equal(t, int32(1), shutdownCalls.Load())
	require.True(t, siblingStopped.Load())
}

func TestRunCallsShutdownBeforeWaitingForTasks(t *testing.T) {
	t.Parallel()

	serverStarted := make(chan struct{})
	stopServer := make(chan struct{})
	errSentinel := errors.New("consumer failed")
	err := lifecycle.Run(context.Background(), lifecycle.Config{
		ShutdownTimeout: time.Second,
		Shutdown: func(context.Context) error {
			close(stopServer)
			return nil
		},
		Tasks: []lifecycle.Task{
			{Name: "server", Run: func(context.Context) error {
				close(serverStarted)
				<-stopServer
				return nil
			}},
			{Name: "consumer", Run: func(context.Context) error {
				<-serverStarted
				return errSentinel
			}},
		},
	})

	require.ErrorIs(t, err, errSentinel)
}

func TestRunTreatsParentCancellationAsCleanStop(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- lifecycle.Run(ctx, lifecycle.Config{
			ShutdownTimeout: time.Second,
			Shutdown:        func(context.Context) error { return nil },
			Tasks: []lifecycle.Task{{Name: "api", Run: func(ctx context.Context) error {
				close(started)
				<-ctx.Done()
				return ctx.Err()
			}}},
		})
	}()
	<-started
	cancel()

	require.NoError(t, <-done)
}

func TestRunWithAlreadyCanceledParentSkipsTasksAndShutsDown(t *testing.T) {
	t.Parallel()

	type contextKey string
	const key contextKey = "request"
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), key, "value"))
	cancel()
	var taskCalls atomic.Int32
	var shutdownCalls atomic.Int32

	err := lifecycle.Run(ctx, lifecycle.Config{
		ShutdownTimeout: time.Second,
		Shutdown: func(shutdownCtx context.Context) error {
			shutdownCalls.Add(1)
			require.NoError(t, shutdownCtx.Err())
			require.Equal(t, "value", shutdownCtx.Value(key))
			return nil
		},
		Tasks: []lifecycle.Task{{Name: "api", Run: func(context.Context) error {
			taskCalls.Add(1)
			return nil
		}}},
	})

	require.NoError(t, err)
	require.Zero(t, taskCalls.Load())
	require.Equal(t, int32(1), shutdownCalls.Load())
}

func TestRunPreservesNonCancellationParentCauses(t *testing.T) {
	t.Parallel()

	t.Run("deadline", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()
		err := lifecycle.Run(ctx, waitingConfig())
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("custom cause", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("terminate deployment")
		ctx, cancel := context.WithCancelCause(context.Background())
		cancel(cause)
		var taskCalls atomic.Int32
		cfg := waitingConfig()
		cfg.Tasks[0].Run = func(context.Context) error {
			taskCalls.Add(1)
			return nil
		}
		err := lifecycle.Run(ctx, cfg)
		require.ErrorIs(t, err, cause)
		require.Zero(t, taskCalls.Load())
	})
}

func TestRunJoinsTaskAndShutdownErrors(t *testing.T) {
	t.Parallel()

	taskErr := errors.New("serve")
	shutdownErr := errors.New("shutdown")
	err := lifecycle.Run(context.Background(), lifecycle.Config{
		ShutdownTimeout: time.Second,
		Shutdown:        func(context.Context) error { return shutdownErr },
		Tasks: []lifecycle.Task{{Name: "api", Run: func(context.Context) error {
			return taskErr
		}}},
	})

	require.ErrorIs(t, err, taskErr)
	require.ErrorIs(t, err, shutdownErr)
}

func TestRunSuppressesTaskErrorsReturnedAfterCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	taskErr := errors.New("listener closed")
	done := make(chan error, 1)
	go func() {
		done <- lifecycle.Run(ctx, lifecycle.Config{
			ShutdownTimeout: time.Second,
			Shutdown:        func(context.Context) error { return nil },
			Tasks: []lifecycle.Task{{Name: "api", Run: func(ctx context.Context) error {
				close(started)
				<-ctx.Done()
				return taskErr
			}}},
		})
	}()
	<-started
	cancel()

	require.NoError(t, <-done)
}

func waitingConfig() lifecycle.Config {
	return lifecycle.Config{
		ShutdownTimeout: time.Second,
		Shutdown:        func(context.Context) error { return nil },
		Tasks: []lifecycle.Task{{Name: "worker", Run: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		}}},
	}
}
