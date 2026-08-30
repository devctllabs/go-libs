package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"time"
)

// Policy determines how long a retry loop waits after each failed attempt.
type Policy interface {
	// Delay returns the delay after failures consecutive failures. Delay(0) must return zero.
	Delay(failures uint) time.Duration
}

// ExponentialConfig configures capped exponential backoff with optional proportional jitter.
type ExponentialConfig struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       float64
}

type exponential struct {
	config ExponentialConfig
	random func() float64
}

// NewExponential constructs a concurrency-safe exponential retry policy.
func NewExponential(config ExponentialConfig) (Policy, error) {
	return newExponential(config, rand.Float64)
}

func newExponential(config ExponentialConfig, random func() float64) (*exponential, error) {
	if config.InitialDelay <= 0 {
		return nil, errors.New("initial delay must be positive")
	}
	if config.MaxDelay <= 0 {
		return nil, errors.New("max delay must be positive")
	}
	if config.MaxDelay < config.InitialDelay {
		return nil, errors.New("max delay must not be shorter than initial delay")
	}
	if config.Multiplier <= 1 || math.IsNaN(config.Multiplier) || math.IsInf(config.Multiplier, 0) {
		return nil, errors.New("multiplier must be finite and greater than one")
	}
	if config.Jitter < 0 || config.Jitter > 1 || math.IsNaN(config.Jitter) || math.IsInf(config.Jitter, 0) {
		return nil, errors.New("jitter must be between zero and one")
	}
	if random == nil {
		return nil, errors.New("random source is required")
	}
	return &exponential{config: config, random: random}, nil
}

func (policy *exponential) Delay(failures uint) time.Duration {
	if failures == 0 {
		return 0
	}
	delay := float64(policy.config.InitialDelay) * math.Pow(policy.config.Multiplier, float64(failures-1))
	delay = math.Min(delay, float64(policy.config.MaxDelay))
	if policy.config.Jitter != 0 {
		delay *= 1 - policy.config.Jitter + 2*policy.config.Jitter*policy.random()
	}
	return time.Duration(math.Min(delay, float64(policy.config.MaxDelay)))
}

// Operation is called once per attempt until it succeeds or retrying stops.
type Operation func(ctx context.Context) error

// Notify observes a failed attempt immediately before its retry delay.
type Notify func(attempt uint, err error, nextDelay time.Duration)

// Option configures a retry loop.
type Option func(*doConfig) error

type doConfig struct {
	maxAttempts uint
	maxElapsed  time.Duration
	notify      Notify
}

// WithMaxAttempts limits the total operation calls, including the initial attempt.
func WithMaxAttempts(maxAttempts uint) Option {
	return func(config *doConfig) error {
		if maxAttempts == 0 {
			return errors.New("max attempts must be positive")
		}
		config.maxAttempts = maxAttempts
		return nil
	}
}

// WithMaxElapsedTime limits total wall-clock time spent by Do.
func WithMaxElapsedTime(maxElapsed time.Duration) Option {
	return func(config *doConfig) error {
		if maxElapsed <= 0 {
			return errors.New("max elapsed time must be positive")
		}
		config.maxElapsed = maxElapsed
		return nil
	}
}

// WithNotify observes retryable failures. notify runs synchronously in the caller's goroutine.
func WithNotify(notify Notify) Option {
	return func(config *doConfig) error {
		if notify == nil {
			return errors.New("notify callback is required")
		}
		config.notify = notify
		return nil
	}
}

// Do calls operation until success, cancellation, a configured limit, or a Permanent error.
// With no limits, it retries until operation succeeds or ctx ends.
func Do(ctx context.Context, policy Policy, operation Operation, options ...Option) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	if policy == nil {
		return errors.New("retry policy is required")
	}
	if operation == nil {
		return errors.New("retry operation is required")
	}
	config := doConfig{}
	for index, option := range options {
		if option == nil {
			return fmt.Errorf("apply option %d: option is nil", index)
		}
		if err := option(&config); err != nil {
			return fmt.Errorf("apply option %d: %w", index, err)
		}
	}

	started := time.Now()
	for attempt := uint(1); ; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := operation(ctx)
		if err == nil {
			return nil
		}
		var permanent *permanentError
		if errors.As(err, &permanent) {
			return permanent.err
		}
		if config.maxAttempts != 0 && attempt >= config.maxAttempts {
			return err
		}

		delay := policy.Delay(attempt)
		if delay < 0 {
			return errors.New("retry policy returned a negative delay")
		}
		if config.maxElapsed != 0 {
			remaining := config.maxElapsed - time.Since(started)
			if remaining <= 0 || delay > remaining {
				return err
			}
		}
		if config.notify != nil {
			config.notify(attempt, err, delay)
		}
		if err := wait(ctx, delay); err != nil {
			return err
		}
	}
}

func wait(ctx context.Context, delay time.Duration) error {
	if delay == 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type permanentError struct{ err error }

func (err *permanentError) Error() string { return err.err.Error() }
func (err *permanentError) Unwrap() error { return err.err }

// Permanent marks err as non-retryable. A nil error remains nil.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return &permanentError{err: err}
}
