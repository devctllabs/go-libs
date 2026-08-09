package health_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/health"
	"github.com/devctllabs/go-libs/health/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestReadinessRunsChecksImmediately(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	checker := mocks.NewMockChecker(ctrl)
	checker.EXPECT().Check(gomock.Any()).Return(nil)
	probes, err := health.New(health.Critical("postgres", checker))
	require.NoError(t, err)

	require.Equal(t, health.StatusOK, probes.Readiness(context.Background()).Status)
}

func TestReadinessRunsChecksConcurrentlyAndPreservesRegistrationOrder(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	first := mocks.NewMockChecker(ctrl)
	second := mocks.NewMockChecker(ctrl)
	started := make(chan string, 2)
	release := make(chan struct{})
	first.EXPECT().Check(gomock.Any()).DoAndReturn(func(context.Context) error {
		started <- "first"
		<-release
		return nil
	})
	second.EXPECT().Check(gomock.Any()).DoAndReturn(func(context.Context) error {
		started <- "second"
		<-release
		return errors.New("optional unavailable")
	})

	probes, err := health.New(
		health.Critical("first", first),
		health.NonCritical("second", second),
	)
	require.NoError(t, err)

	reports := make(chan health.Report, 1)
	go func() {
		reports <- probes.Readiness(context.Background())
	}()

	seen := make([]string, 0, 2)
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	for len(seen) < 2 {
		select {
		case name := <-started:
			seen = append(seen, name)
		case <-timer.C:
			close(release)
			require.FailNow(t, "checks did not start concurrently", "started: %v", seen)
		}
	}
	close(release)

	report := <-reports
	require.ElementsMatch(t, []string{"first", "second"}, seen)
	require.Equal(t, health.StatusOK, report.Status)
	require.Equal(t, []health.CheckResult{
		{Name: "first", Status: health.StatusOK, Critical: true},
		{Name: "second", Status: health.StatusFail, Critical: false},
	}, report.Checks)
}

func TestReadinessAppliesOneTimeoutAndObservesIt(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	checker := mocks.NewMockChecker(ctrl)
	checker.EXPECT().Check(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	observer := mocks.NewMockObserver(ctrl)
	var observation health.Observation
	observer.EXPECT().Observe(gomock.Any(), gomock.Any()).Do(
		func(_ context.Context, got health.Observation) { observation = got },
	)

	probes, err := health.New(
		health.WithCheckTimeout(20*time.Millisecond),
		health.WithObserver(observer),
		health.Critical("postgres", checker),
	)
	require.NoError(t, err)

	report := probes.Readiness(context.Background())
	require.Equal(t, health.StatusFail, report.Status)
	require.Equal(t, health.OutcomeTimeout, observation.Outcome)
	require.ErrorIs(t, observation.Err, context.DeadlineExceeded)
	require.Equal(t, "postgres", observation.Name)
	require.True(t, observation.Critical)
}

func TestReadinessRecoversCheckerPanicAsObservedFailure(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	checker := mocks.NewMockChecker(ctrl)
	checker.EXPECT().Check(gomock.Any()).DoAndReturn(func(context.Context) error {
		panic("broken checker")
	})
	observer := mocks.NewMockObserver(ctrl)
	var observation health.Observation
	observer.EXPECT().Observe(gomock.Any(), gomock.Any()).Do(
		func(_ context.Context, got health.Observation) { observation = got },
	)

	probes, err := health.New(
		health.WithObserver(observer),
		health.Critical("postgres", checker),
	)
	require.NoError(t, err)

	report := probes.Readiness(context.Background())
	require.Equal(t, health.StatusFail, report.Status)
	require.Equal(t, health.OutcomeError, observation.Outcome)
	require.Error(t, observation.Err)
	require.Contains(t, observation.Err.Error(), "broken checker")
	require.Contains(t, observation.Err.Error(), "goroutine")
}

func TestReadinessTreatsParentCancellationAsError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	checker := mocks.NewMockChecker(ctrl)
	observer := mocks.NewMockObserver(ctrl)
	var observation health.Observation
	observer.EXPECT().Observe(gomock.Any(), gomock.Any()).Do(
		func(_ context.Context, got health.Observation) { observation = got },
	)

	probes, err := health.New(
		health.WithObserver(observer),
		health.Critical("postgres", checker),
	)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.Equal(t, health.StatusFail, probes.Readiness(ctx).Status)
	require.Equal(t, health.OutcomeError, observation.Outcome)
	require.ErrorIs(t, observation.Err, context.Canceled)
}

func TestNewValidatesOptions(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	checker := mocks.NewMockChecker(ctrl)

	_, err := health.New(health.Critical(" ", checker))
	require.ErrorContains(t, err, "must not be blank")

	_, err = health.New(health.Critical("postgres", nil))
	require.ErrorContains(t, err, "must not be nil")

	_, err = health.New(
		health.Critical("postgres", checker),
		health.NonCritical("postgres", checker),
	)
	require.ErrorContains(t, err, "duplicate")

	_, err = health.New(health.WithCheckTimeout(0))
	require.ErrorContains(t, err, "must be positive")

	_, err = health.New(health.WithObserver(nil))
	require.ErrorContains(t, err, "must not be nil")
}

func TestNewUsesDefaultTimeoutAndNotifiesEveryObserver(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	checker := mocks.NewMockChecker(ctrl)
	checker.EXPECT().Check(gomock.Any()).Return(nil)
	first := mocks.NewMockObserver(ctrl)
	second := mocks.NewMockObserver(ctrl)
	first.EXPECT().Observe(gomock.Any(), gomock.Any())
	second.EXPECT().Observe(gomock.Any(), gomock.Any())

	probes, err := health.New(
		health.Critical("postgres", checker),
		health.WithObserver(first),
		health.WithObserver(second),
	)
	require.NoError(t, err)
	require.Equal(t, time.Second, probes.CheckTimeout())
	require.Equal(t, health.StatusOK, probes.Readiness(context.Background()).Status)
}
