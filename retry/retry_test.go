package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExponentialDelay(t *testing.T) {
	t.Parallel()
	policy, err := newExponential(ExponentialConfig{
		InitialDelay: time.Second,
		MaxDelay:     5 * time.Second,
		Multiplier:   2,
		Jitter:       0.5,
	}, func() float64 { return 0.5 })
	require.NoError(t, err)

	require.Equal(t, time.Duration(0), policy.Delay(0))
	require.Equal(t, time.Second, policy.Delay(1))
	require.Equal(t, 2*time.Second, policy.Delay(2))
	require.Equal(t, 4*time.Second, policy.Delay(3))
	require.Equal(t, 5*time.Second, policy.Delay(4))
}

func TestNewExponentialRejectsIncompleteConfiguration(t *testing.T) {
	t.Parallel()
	tests := []ExponentialConfig{
		{MaxDelay: time.Second, Multiplier: 2},
		{InitialDelay: time.Second, Multiplier: 2},
		{InitialDelay: time.Second, MaxDelay: time.Second},
		{InitialDelay: 2 * time.Second, MaxDelay: time.Second, Multiplier: 2},
		{InitialDelay: time.Second, MaxDelay: time.Second, Multiplier: 1},
		{InitialDelay: time.Second, MaxDelay: time.Second, Multiplier: 2, Jitter: -0.1},
		{InitialDelay: time.Second, MaxDelay: time.Second, Multiplier: 2, Jitter: 1.1},
	}

	for _, config := range tests {
		_, err := NewExponential(config)
		require.Error(t, err)
	}
}

func TestDoRetriesUntilSuccessAndNotifies(t *testing.T) {
	t.Parallel()
	policy := fixedPolicy{delay: time.Millisecond}
	attempts := 0
	notifications := make([]uint, 0)

	err := Do(context.Background(), policy, func(context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary")
		}
		return nil
	}, WithNotify(func(attempt uint, _ error, nextDelay time.Duration) {
		notifications = append(notifications, attempt)
		require.Equal(t, time.Millisecond, nextDelay)
	}))

	require.NoError(t, err)
	require.Equal(t, 3, attempts)
	require.Equal(t, []uint{1, 2}, notifications)
}

func TestDoStopsAtMaxAttempts(t *testing.T) {
	t.Parallel()
	want := errors.New("still unavailable")
	attempts := 0
	err := Do(context.Background(), fixedPolicy{}, func(context.Context) error {
		attempts++
		return want
	}, WithMaxAttempts(3))

	require.ErrorIs(t, err, want)
	require.Equal(t, 3, attempts)
}

func TestDoStopsOnPermanentError(t *testing.T) {
	t.Parallel()
	want := errors.New("invalid configuration")
	attempts := 0
	err := Do(context.Background(), fixedPolicy{}, func(context.Context) error {
		attempts++
		return Permanent(want)
	})

	require.ErrorIs(t, err, want)
	require.Equal(t, 1, attempts)
}

func TestDoHonorsContextWhileWaiting(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	err := Do(ctx, fixedPolicy{delay: time.Hour}, func(context.Context) error {
		attempts++
		cancel()
		return errors.New("temporary")
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, attempts)
}

type fixedPolicy struct{ delay time.Duration }

func (policy fixedPolicy) Delay(uint) time.Duration { return policy.delay }
