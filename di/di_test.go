package di_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/di"
	"github.com/stretchr/testify/require"
)

type testDependency struct{}
type testDependent struct{}
type testUnused struct{}
type testOther struct{}

func TestResolveBuildsSingletonOnceConcurrently(t *testing.T) {
	t.Parallel()

	container := di.New()
	var builds atomic.Int32
	require.NoError(t, di.Provide(container, func(di.Resolver) (*testDependency, error) {
		builds.Add(1)
		return &testDependency{}, nil
	}))

	const goroutines = 32
	values := make(chan *testDependency, goroutines)
	errors := make(chan error, goroutines)
	var group sync.WaitGroup
	group.Add(goroutines)
	for range goroutines {
		go func() {
			defer group.Done()
			value, err := di.Resolve[*testDependency](container)
			values <- value
			errors <- err
		}()
	}
	group.Wait()
	close(values)
	close(errors)

	for err := range errors {
		require.NoError(t, err)
	}
	var first *testDependency
	for value := range values {
		if first == nil {
			first = value
		}
		require.Same(t, first, value)
	}
	require.Equal(t, int32(1), builds.Load())
}

func TestNamedServicesAreIsolatedByType(t *testing.T) {
	t.Parallel()

	container := di.New()
	require.NoError(t, di.ProvideNamedValue(container, "primary", 42))
	require.NoError(t, di.ProvideNamedValue(container, "primary", "value"))

	integer, err := di.ResolveNamed[int](container, "primary")
	require.NoError(t, err)
	require.Equal(t, 42, integer)

	text, err := di.ResolveNamed[string](container, "primary")
	require.NoError(t, err)
	require.Equal(t, "value", text)
}

func TestRegistrationValidationAndAutoSeal(t *testing.T) {
	t.Parallel()

	container := di.New()
	require.ErrorIs(t, di.ProvideNamedValue(container, "", 1), di.ErrInvalidArgument)

	var nilProvider di.Provider[int]
	require.ErrorIs(t, di.Provide(container, nilProvider), di.ErrInvalidArgument)

	require.NoError(t, di.ProvideValue(container, 1))
	require.ErrorIs(t, di.ProvideValue(container, 2), di.ErrDuplicate)

	value, err := di.Resolve[int](container)
	require.NoError(t, err)
	require.Equal(t, 1, value)
	require.ErrorIs(t, di.ProvideValue(container, "late"), di.ErrSealed)
}

func TestRegistrationIsAtomicWithFirstResolve(t *testing.T) {
	t.Parallel()

	for range 500 {
		container := di.New()
		start := make(chan struct{})
		provideResult := make(chan error, 1)
		resolveResult := make(chan error, 1)

		go func() {
			<-start
			provideResult <- di.ProvideValue(container, 42)
		}()
		go func() {
			<-start
			_, err := di.Resolve[int](container)
			resolveResult <- err
		}()
		close(start)

		provideErr := <-provideResult
		resolveErr := <-resolveResult
		if provideErr == nil {
			require.NoError(t, resolveErr)
			continue
		}
		require.ErrorIs(t, provideErr, di.ErrSealed)
		require.ErrorIs(t, resolveErr, di.ErrNotFound)
	}
}

func TestZeroContainerIsRejected(t *testing.T) {
	t.Parallel()

	container := &di.Container{}
	require.ErrorIs(t, di.ProvideValue(container, 1), di.ErrInvalidArgument)
	_, err := di.Resolve[int](container)
	require.ErrorIs(t, err, di.ErrInvalidArgument)
	require.ErrorIs(t, container.Shutdown(context.Background()), di.ErrInvalidArgument)
}

func TestResolveReportsMissingCycleProviderAndNilErrors(t *testing.T) {
	t.Parallel()

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		container := di.New()
		_, err := di.Resolve[*testDependency](container)
		require.ErrorIs(t, err, di.ErrNotFound)
		require.ErrorContains(t, err, "*di_test.testDependency")
	})

	t.Run("cycle", func(t *testing.T) {
		t.Parallel()
		container := di.New()
		require.NoError(t, di.Provide(container, func(resolver di.Resolver) (*testDependency, error) {
			_, err := di.Resolve[*testDependent](resolver)
			return &testDependency{}, err
		}))
		require.NoError(t, di.Provide(container, func(resolver di.Resolver) (*testDependent, error) {
			_, err := di.Resolve[*testDependency](resolver)
			return &testDependent{}, err
		}))

		_, err := di.Resolve[*testDependency](container)
		require.ErrorIs(t, err, di.ErrCircularDependency)
	})

	t.Run("provider", func(t *testing.T) {
		t.Parallel()
		sentinel := errors.New("provider failed")
		container := di.New()
		require.NoError(t, di.Provide(container, func(di.Resolver) (int, error) {
			return 0, sentinel
		}))

		_, err := di.Resolve[int](container)
		require.ErrorIs(t, err, sentinel)
	})

	t.Run("nil service", func(t *testing.T) {
		t.Parallel()
		container := di.New()
		require.NoError(t, di.Provide(container, func(di.Resolver) (*testDependency, error) {
			return nil, nil
		}))

		_, err := di.Resolve[*testDependency](container)
		require.ErrorIs(t, err, di.ErrNilService)
	})
}

type hasAutomaticShutdown struct {
	calls *atomic.Int32
}

func (s *hasAutomaticShutdown) Shutdown(context.Context) error {
	s.calls.Add(1)
	return nil
}

func TestOwnershipIsAlwaysExplicit(t *testing.T) {
	t.Parallel()

	container := di.New()
	var automaticCalls atomic.Int32
	var explicitCalls atomic.Int32
	require.NoError(t, di.Provide(container, func(di.Resolver) (*hasAutomaticShutdown, error) {
		return &hasAutomaticShutdown{calls: &automaticCalls}, nil
	}))
	require.NoError(t, di.ProvideResource(container,
		func(di.Resolver) (*testOther, error) { return &testOther{}, nil },
		func(context.Context, *testOther) error {
			explicitCalls.Add(1)
			return nil
		},
	))

	_, err := di.Resolve[*hasAutomaticShutdown](container)
	require.NoError(t, err)
	_, err = di.Resolve[*testOther](container)
	require.NoError(t, err)
	require.NoError(t, container.Shutdown(context.Background()))

	require.Zero(t, automaticCalls.Load())
	require.Equal(t, int32(1), explicitCalls.Load())
}

func TestShutdownUsesReverseDependencyOrderAndSkipsUnusedResources(t *testing.T) {
	t.Parallel()

	container := di.New()
	var dependentClosed atomic.Bool
	var dependencyObservedOrder atomic.Bool
	var unusedCalls atomic.Int32
	dependencyStarted := make(chan struct{})

	require.NoError(t, di.ProvideResource(container,
		func(di.Resolver) (*testDependency, error) { return &testDependency{}, nil },
		func(context.Context, *testDependency) error {
			close(dependencyStarted)
			dependencyObservedOrder.Store(dependentClosed.Load())
			return nil
		},
	))
	require.NoError(t, di.ProvideResource(container,
		func(resolver di.Resolver) (*testDependent, error) {
			_, err := di.Resolve[*testDependency](resolver)
			return &testDependent{}, err
		},
		func(context.Context, *testDependent) error {
			dependentClosed.Store(true)
			return nil
		},
	))
	require.NoError(t, di.ProvideResource(container,
		func(di.Resolver) (*testUnused, error) { return &testUnused{}, nil },
		func(context.Context, *testUnused) error {
			unusedCalls.Add(1)
			return nil
		},
	))

	_, err := di.Resolve[*testDependent](container)
	require.NoError(t, err)
	require.NoError(t, container.Shutdown(context.Background()))

	select {
	case <-dependencyStarted:
	default:
		require.Fail(t, "dependency cleanup was not called")
	}
	require.True(t, dependencyObservedOrder.Load())
	require.Zero(t, unusedCalls.Load())
}

func TestShutdownRunsIndependentResourcesConcurrently(t *testing.T) {
	t.Parallel()

	container := di.New()
	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})
	release := make(chan struct{})

	require.NoError(t, di.ProvideResource(container,
		func(di.Resolver) (*testDependency, error) { return &testDependency{}, nil },
		func(context.Context, *testDependency) error {
			close(firstStarted)
			<-release
			return nil
		},
	))
	require.NoError(t, di.ProvideResource(container,
		func(di.Resolver) (*testOther, error) { return &testOther{}, nil },
		func(context.Context, *testOther) error {
			close(secondStarted)
			<-release
			return nil
		},
	))
	_, err := di.Resolve[*testDependency](container)
	require.NoError(t, err)
	_, err = di.Resolve[*testOther](container)
	require.NoError(t, err)

	shutdownDone := make(chan error, 1)
	go func() { shutdownDone <- container.Shutdown(context.Background()) }()

	requireStarted(t, firstStarted)
	requireStarted(t, secondStarted)
	close(release)
	require.NoError(t, <-shutdownDone)
}

func TestShutdownJoinsErrorsAndReturnsCachedResult(t *testing.T) {
	t.Parallel()

	firstError := errors.New("first cleanup")
	secondError := errors.New("second cleanup")
	container := di.New()
	var calls atomic.Int32

	require.NoError(t, di.ProvideResource(container,
		func(di.Resolver) (*testDependency, error) { return &testDependency{}, nil },
		func(context.Context, *testDependency) error {
			calls.Add(1)
			return firstError
		},
	))
	require.NoError(t, di.ProvideResource(container,
		func(di.Resolver) (*testOther, error) { return &testOther{}, nil },
		func(context.Context, *testOther) error {
			calls.Add(1)
			return secondError
		},
	))
	_, err := di.Resolve[*testDependency](container)
	require.NoError(t, err)
	_, err = di.Resolve[*testOther](container)
	require.NoError(t, err)

	firstResult := container.Shutdown(context.Background())
	secondResult := container.Shutdown(context.Background())
	require.ErrorIs(t, firstResult, firstError)
	require.ErrorIs(t, firstResult, secondError)
	require.Same(t, firstResult, secondResult)
	require.Equal(t, int32(2), calls.Load())
	require.ErrorIs(t, func() error {
		_, resolveErr := di.Resolve[*testDependency](container)
		return resolveErr
	}(), di.ErrClosed)
	require.ErrorIs(t, di.ProvideValue(container, 1), di.ErrClosed)
}

func TestShutdownConvertsCleanupPanicAndContinues(t *testing.T) {
	t.Parallel()

	container := di.New()
	var completed atomic.Bool
	require.NoError(t, di.ProvideResource(container,
		func(di.Resolver) (*testDependency, error) { return &testDependency{}, nil },
		func(context.Context, *testDependency) error {
			panic("cleanup failed")
		},
	))
	require.NoError(t, di.ProvideResource(container,
		func(di.Resolver) (*testOther, error) { return &testOther{}, nil },
		func(context.Context, *testOther) error {
			completed.Store(true)
			return nil
		},
	))
	_, err := di.Resolve[*testDependency](container)
	require.NoError(t, err)
	_, err = di.Resolve[*testOther](container)
	require.NoError(t, err)

	err = container.Shutdown(context.Background())
	require.ErrorContains(t, err, "cleanup panic: cleanup failed")
	require.True(t, completed.Load())
}

func TestShutdownWaitsForActiveResolutionAndRejectsNewOnes(t *testing.T) {
	t.Parallel()

	container := di.New()
	providerStarted := make(chan struct{})
	releaseProvider := make(chan struct{})
	require.NoError(t, di.Provide(container, func(di.Resolver) (*testDependency, error) {
		close(providerStarted)
		<-releaseProvider
		return &testDependency{}, nil
	}))
	require.NoError(t, di.ProvideValue(container, testOther{}))

	resolveDone := make(chan error, 1)
	go func() {
		_, err := di.Resolve[*testDependency](container)
		resolveDone <- err
	}()
	requireStarted(t, providerStarted)

	shutdownDone := make(chan error, 1)
	go func() { shutdownDone <- container.Shutdown(context.Background()) }()

	require.Eventually(t, func() bool {
		_, err := di.Resolve[testOther](container)
		return errors.Is(err, di.ErrClosed)
	}, time.Second, time.Millisecond)
	close(releaseProvider)

	require.NoError(t, <-resolveDone)
	require.NoError(t, <-shutdownDone)
}

func requireStarted(t *testing.T, started <-chan struct{}) {
	t.Helper()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		require.Fail(t, "operation did not start")
	}
}

func TestErrorsCarryReadableServiceDescriptions(t *testing.T) {
	t.Parallel()

	container := di.New()
	_, err := di.ResolveNamed[fmt.Stringer](container, "primary")
	require.ErrorIs(t, err, di.ErrNotFound)
	require.ErrorContains(t, err, "fmt.Stringer[primary]")
}
