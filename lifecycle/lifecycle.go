package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

// ErrTaskStopped reports that a long-running task returned nil before cancellation.
var ErrTaskStopped = errors.New("lifecycle: task stopped unexpectedly")

// Task is one named long-running workload.
type Task struct {
	Name string
	Run  func(ctx context.Context) error
}

// Config describes the workloads and common application shutdown owned by Run.
type Config struct {
	ShutdownTimeout time.Duration
	Shutdown        func(ctx context.Context) error
	Tasks           []Task
}

// Run coordinates tasks until the parent is canceled or one task stops.
//
// Run invokes Shutdown exactly once with a fresh, bounded context, then waits for every task.
// A plain parent context cancellation is a clean stop. Parent deadlines, custom cancellation
// causes, task failures, and shutdown failures are returned and remain compatible with errors.Is.
func Run(ctx context.Context, cfg Config) error {
	if err := validate(ctx, cfg); err != nil {
		return err
	}

	if ctx.Err() != nil {
		return errors.Join(parentError(ctx), shutdown(ctx, cfg))
	}

	group, runCtx := errgroup.WithContext(ctx)
	for _, configuredTask := range cfg.Tasks {
		task := configuredTask
		task.Name = strings.TrimSpace(task.Name)
		group.Go(func() error {
			err := task.Run(runCtx)
			if runCtx.Err() != nil {
				return nil
			}
			if err == nil {
				return fmt.Errorf("lifecycle: task %q: %w", task.Name, ErrTaskStopped)
			}
			return fmt.Errorf("lifecycle: task %q: %w", task.Name, err)
		})
	}

	<-runCtx.Done()
	shutdownErr := shutdown(ctx, cfg)
	runErr := group.Wait()
	return errors.Join(runErr, parentError(ctx), shutdownErr)
}

func validate(ctx context.Context, cfg Config) error {
	if ctx == nil {
		return errors.New("lifecycle: context must not be nil")
	}
	if cfg.ShutdownTimeout <= 0 {
		return errors.New("lifecycle: shutdown timeout must be positive")
	}
	if cfg.Shutdown == nil {
		return errors.New("lifecycle: shutdown must not be nil")
	}
	if len(cfg.Tasks) == 0 {
		return errors.New("lifecycle: at least one task is required")
	}

	names := make(map[string]struct{}, len(cfg.Tasks))
	for _, task := range cfg.Tasks {
		name := strings.TrimSpace(task.Name)
		if name == "" {
			return errors.New("lifecycle: task name must not be blank")
		}
		if _, exists := names[name]; exists {
			return fmt.Errorf("lifecycle: duplicate task name %q", name)
		}
		if task.Run == nil {
			return fmt.Errorf("lifecycle: task %q run function must not be nil", name)
		}
		names[name] = struct{}{}
	}
	return nil
}

func shutdown(ctx context.Context, cfg Config) error {
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
	defer cancel()
	return cfg.Shutdown(shutdownCtx)
}

func parentError(ctx context.Context) error {
	cause := context.Cause(ctx)
	if cause == context.Canceled {
		return nil
	}
	return cause
}
