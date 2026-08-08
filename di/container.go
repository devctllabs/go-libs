package di

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"

	do "github.com/samber/do/v2"
)

// Provider lazily constructs one dependency. It must resolve dependencies
// synchronously through the provided Resolver so their graph edges can be
// tracked, and it must not retain Resolver after returning. Returning a nil
// dependency without an error causes resolution to fail with ErrNilService.
type Provider[T any] func(Resolver) (T, error)

// Cleanup releases a resource owned by a Container. It is called at most once,
// only after the resource was successfully constructed, with the context from
// the first Shutdown call. A cleanup panic is recovered and returned as a
// shutdown error.
type Cleanup[T any] func(context.Context, T) error

// Resolver is the restricted dependency lookup context passed to providers.
// It is valid only while that provider is running. Container also implements
// Resolver for top-level resolution.
type Resolver interface {
	resolutionContext() resolveContext
}

type phase uint8

const (
	phaseRegistering phase = iota
	phaseSealed
	phaseClosing
	phaseClosed
)

type registrationKind uint8

const (
	registrationSingleton registrationKind = iota
	registrationResource
)

type registration struct {
	description string
	kind        registrationKind
}

// Container stores dependency providers and owns explicitly registered
// resources. Create one with New; a zero Container is not usable. Container
// coordinates concurrent resolution and shutdown.
type Container struct {
	engine *do.RootScope

	mu            sync.Mutex
	idle          *sync.Cond
	phase         phase
	activeResolve int
	registrations map[string]registration
	shutdownDone  chan struct{}
	shutdownErr   error
}

// New creates an empty dependency container in its registration phase.
func New() *Container {
	container := &Container{
		engine:        do.New(),
		phase:         phaseRegistering,
		registrations: make(map[string]registration),
		shutdownDone:  make(chan struct{}),
	}
	container.idle = sync.NewCond(&container.mu)
	return container
}

type resolveContext struct {
	container *Container
	injector  do.Injector
	valid     *atomic.Bool
}

func (c *Container) resolutionContext() resolveContext {
	if !validContainer(c) {
		return resolveContext{}
	}
	return resolveContext{container: c, injector: c.engine}
}

type providerResolver struct {
	context resolveContext
}

func (r *providerResolver) resolutionContext() resolveContext {
	if r == nil {
		return resolveContext{}
	}
	return r.context
}

type singleton[T any] struct {
	value T
}

type resource[T any] struct {
	value   T
	cleanup Cleanup[T]
}

func (r *resource[T]) Shutdown(ctx context.Context) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("cleanup panic: %v", recovered)
		}
	}()
	return r.cleanup(ctx, r.value)
}

// Provide registers a lazy singleton by its type. Registration errors match
// ErrInvalidArgument, ErrDuplicate, ErrSealed, or ErrClosed. Provider errors
// are returned when the dependency is resolved.
func Provide[T any](container *Container, provider Provider[T]) error {
	return provide(container, serviceKey[T](""), serviceDescription[T](""), provider, nil)
}

// ProvideNamed registers a named lazy singleton. Names are scoped by type and
// must not be empty. Registration errors match ErrInvalidArgument,
// ErrDuplicate, ErrSealed, or ErrClosed.
func ProvideNamed[T any](container *Container, name string, provider Provider[T]) error {
	if name == "" {
		return fmt.Errorf("%w: dependency name is empty", ErrInvalidArgument)
	}
	return provide(container, serviceKey[T](name), serviceDescription[T](name), provider, nil)
}

// ProvideResource registers a lazy singleton and transfers its cleanup
// ownership to the container. Provider and cleanup must be non-nil.
// Registration errors match ErrInvalidArgument, ErrDuplicate, ErrSealed, or
// ErrClosed.
func ProvideResource[T any](container *Container, provider Provider[T], cleanup Cleanup[T]) error {
	if cleanup == nil {
		return fmt.Errorf("%w: cleanup is nil", ErrInvalidArgument)
	}
	return provide(container, serviceKey[T](""), serviceDescription[T](""), provider, cleanup)
}

// ProvideNamedResource registers a named lazy singleton and transfers its
// cleanup ownership to the container. Names are scoped by type and must not be
// empty; provider and cleanup must be non-nil. Registration errors match
// ErrInvalidArgument, ErrDuplicate, ErrSealed, or ErrClosed.
func ProvideNamedResource[T any](
	container *Container,
	name string,
	provider Provider[T],
	cleanup Cleanup[T],
) error {
	if name == "" {
		return fmt.Errorf("%w: dependency name is empty", ErrInvalidArgument)
	}
	if cleanup == nil {
		return fmt.Errorf("%w: cleanup is nil", ErrInvalidArgument)
	}
	return provide(container, serviceKey[T](name), serviceDescription[T](name), provider, cleanup)
}

func provide[T any](
	container *Container,
	key string,
	description string,
	provider Provider[T],
	cleanup Cleanup[T],
) error {
	if !validContainer(container) {
		return fmt.Errorf("%w: container was not created with di.New", ErrInvalidArgument)
	}
	if provider == nil {
		return fmt.Errorf("%w: provider for %s is nil", ErrInvalidArgument, description)
	}

	kind := registrationSingleton
	if cleanup != nil {
		kind = registrationResource
	}

	if cleanup == nil {
		return container.register(key, registration{description: description, kind: kind}, func() {
			do.ProvideNamed(container.engine, key, func(injector do.Injector) (*singleton[T], error) {
				value, err := invokeProvider(container, injector, provider)
				if err != nil {
					return nil, err
				}
				return &singleton[T]{value: value}, nil
			})
		})
	}

	return container.register(key, registration{description: description, kind: kind}, func() {
		do.ProvideNamed(container.engine, key, func(injector do.Injector) (*resource[T], error) {
			value, err := invokeProvider(container, injector, provider)
			if err != nil {
				return nil, err
			}
			return &resource[T]{value: value, cleanup: cleanup}, nil
		})
	})
}

// ProvideValue registers an immediately available value by its type. The
// container does not infer cleanup ownership from the value's methods. A nil
// value fails with ErrNilService; other registration errors match
// ErrInvalidArgument, ErrDuplicate, ErrSealed, or ErrClosed.
func ProvideValue[T any](container *Container, value T) error {
	return provideValue(container, serviceKey[T](""), serviceDescription[T](""), value)
}

// ProvideNamedValue registers an immediately available named value. Names are
// scoped by type and must not be empty. The container does not infer cleanup
// ownership. A nil value fails with ErrNilService; other registration errors
// match ErrInvalidArgument, ErrDuplicate, ErrSealed, or ErrClosed.
func ProvideNamedValue[T any](container *Container, name string, value T) error {
	if name == "" {
		return fmt.Errorf("%w: dependency name is empty", ErrInvalidArgument)
	}
	return provideValue(container, serviceKey[T](name), serviceDescription[T](name), value)
}

func provideValue[T any](container *Container, key, description string, value T) error {
	if !validContainer(container) {
		return fmt.Errorf("%w: container was not created with di.New", ErrInvalidArgument)
	}
	if isNil(value) {
		return fmt.Errorf("%w: %s", ErrNilService, description)
	}
	return container.register(key, registration{
		description: description,
		kind:        registrationSingleton,
	}, func() {
		do.ProvideNamedValue(container.engine, key, &singleton[T]{value: value})
	})
}

func (c *Container) register(key string, candidate registration, register func()) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.phase {
	case phaseRegistering:
	case phaseSealed:
		return fmt.Errorf("%w: register %s", ErrSealed, candidate.description)
	case phaseClosing, phaseClosed:
		return fmt.Errorf("%w: register %s", ErrClosed, candidate.description)
	}
	if _, exists := c.registrations[key]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicate, candidate.description)
	}
	register()
	c.registrations[key] = candidate
	return nil
}

func invokeProvider[T any](container *Container, injector do.Injector, provider Provider[T]) (T, error) {
	valid := &atomic.Bool{}
	valid.Store(true)
	defer valid.Store(false)
	resolver := &providerResolver{context: resolveContext{
		container: container,
		injector:  injector,
		valid:     valid,
	}}
	value, err := provider(resolver)
	if err != nil {
		var zero T
		return zero, err
	}
	if isNil(value) {
		var zero T
		return zero, ErrNilService
	}
	return value, nil
}

// Resolve returns the singleton registered for T, constructing it lazily when
// needed. Concurrent successful resolutions return the same singleton. The
// first top-level resolution seals the container against further registration.
// Resolution errors may match ErrInvalidArgument, ErrClosed, ErrNotFound,
// ErrCircularDependency, or ErrNilService; provider errors remain wrapped.
func Resolve[T any](resolver Resolver) (T, error) {
	return resolve[T](resolver, "")
}

// ResolveNamed returns the named singleton registered for T. Names are scoped
// by type and name must not be empty. The first top-level resolution seals the
// container. Its errors have the same categories as Resolve.
func ResolveNamed[T any](resolver Resolver, name string) (T, error) {
	if name == "" {
		var zero T
		return zero, fmt.Errorf("%w: dependency name is empty", ErrInvalidArgument)
	}
	return resolve[T](resolver, name)
}

func resolve[T any](resolver Resolver, name string) (T, error) {
	var zero T
	context, finish, err := beginResolution(resolver)
	if err != nil {
		return zero, err
	}
	defer finish()

	key := serviceKey[T](name)
	description := serviceDescription[T](name)
	registration, ok := context.container.lookup(key)
	if !ok {
		return zero, fmt.Errorf("resolve %s: %w", description, ErrNotFound)
	}

	switch registration.kind {
	case registrationSingleton:
		wrapped, invokeErr := do.InvokeNamed[*singleton[T]](context.injector, key)
		if invokeErr != nil {
			return zero, resolveError(description, invokeErr)
		}
		return wrapped.value, nil
	case registrationResource:
		wrapped, invokeErr := do.InvokeNamed[*resource[T]](context.injector, key)
		if invokeErr != nil {
			return zero, resolveError(description, invokeErr)
		}
		return wrapped.value, nil
	default:
		return zero, fmt.Errorf("resolve %s: %w", description, ErrNotFound)
	}
}

func beginResolution(resolver Resolver) (resolveContext, func(), error) {
	if resolver == nil {
		return resolveContext{}, func() {}, fmt.Errorf("%w: resolver is nil", ErrInvalidArgument)
	}
	context := resolver.resolutionContext()
	if context.container == nil || isNil(context.injector) {
		return resolveContext{}, func() {}, fmt.Errorf("%w: resolver is invalid", ErrInvalidArgument)
	}
	if context.valid != nil {
		if !context.valid.Load() {
			return resolveContext{}, func() {}, fmt.Errorf("%w: provider resolver expired", ErrInvalidArgument)
		}
		return context, func() {}, nil
	}

	container := context.container
	container.mu.Lock()
	switch container.phase {
	case phaseRegistering:
		container.phase = phaseSealed
	case phaseSealed:
	case phaseClosing, phaseClosed:
		container.mu.Unlock()
		return resolveContext{}, func() {}, ErrClosed
	}
	container.activeResolve++
	container.mu.Unlock()

	return context, func() {
		container.mu.Lock()
		container.activeResolve--
		container.idle.Broadcast()
		container.mu.Unlock()
	}, nil
}

func (c *Container) lookup(key string) (registration, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	registration, ok := c.registrations[key]
	return registration, ok
}

func resolveError(description string, err error) error {
	if errors.Is(err, do.ErrCircularDependency) {
		return fmt.Errorf("resolve %s: %w: %s", description, ErrCircularDependency, err)
	}
	if errors.Is(err, do.ErrServiceNotFound) {
		return fmt.Errorf("resolve %s: %w", description, ErrNotFound)
	}
	return fmt.Errorf("resolve %s: %w", description, err)
}

// Shutdown stops all constructed owned resources. Dependents are stopped
// before dependencies, and independent branches may stop concurrently.
//
// Shutdown waits for resolutions that started before shutdown, rejects new
// operations, and caches the result. The first caller's context controls the
// shutdown; concurrent and later callers receive the cached result. Cleanup
// failures are joined, remain compatible with errors.Is, and do not prevent
// other independent cleanups from running. Cleanup panics are converted to
// errors. An invalid Container or nil context causes ErrInvalidArgument.
func (c *Container) Shutdown(ctx context.Context) error {
	if !validContainer(c) {
		return fmt.Errorf("%w: container was not created with di.New", ErrInvalidArgument)
	}
	if ctx == nil || isNil(ctx) {
		return fmt.Errorf("%w: context is nil", ErrInvalidArgument)
	}

	c.mu.Lock()
	switch c.phase {
	case phaseClosed:
		err := c.shutdownErr
		c.mu.Unlock()
		return err
	case phaseClosing:
		done := c.shutdownDone
		c.mu.Unlock()
		<-done
		c.mu.Lock()
		err := c.shutdownErr
		c.mu.Unlock()
		return err
	case phaseRegistering, phaseSealed:
		c.phase = phaseClosing
	}
	for c.activeResolve > 0 {
		c.idle.Wait()
	}
	c.mu.Unlock()

	report := c.engine.ShutdownWithContext(ctx)
	shutdownErr := c.joinShutdownErrors(report)

	c.mu.Lock()
	c.shutdownErr = shutdownErr
	c.phase = phaseClosed
	close(c.shutdownDone)
	c.mu.Unlock()
	return shutdownErr
}

func (c *Container) joinShutdownErrors(report *do.ShutdownReport) error {
	if report == nil || len(report.Errors) == 0 {
		return nil
	}

	type failure struct {
		description string
		err         error
	}
	failures := make([]failure, 0, len(report.Errors))
	for service, err := range report.Errors {
		if err == nil {
			continue
		}
		description := service.Service
		if registered, ok := c.lookup(service.Service); ok {
			description = registered.description
		}
		failures = append(failures, failure{description: description, err: err})
	}
	sort.Slice(failures, func(i, j int) bool {
		return failures[i].description < failures[j].description
	})

	errorsToJoin := make([]error, 0, len(failures))
	for _, failure := range failures {
		errorsToJoin = append(errorsToJoin, fmt.Errorf("shutdown %s: %w", failure.description, failure.err))
	}
	return errors.Join(errorsToJoin...)
}

func serviceKey[T any](name string) string {
	key := "type:" + do.NameOf[T]()
	if name == "" {
		return key
	}
	return key + ":name:" + name
}

func serviceDescription[T any](name string) string {
	description := reflect.TypeFor[T]().String()
	if name == "" {
		return description
	}
	return description + "[" + name + "]"
}

func isNil[T any](value T) bool {
	reflected := reflect.ValueOf(value)
	if !reflected.IsValid() {
		return true
	}
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func validContainer(container *Container) bool {
	return container != nil &&
		container.engine != nil &&
		container.idle != nil &&
		container.registrations != nil &&
		container.shutdownDone != nil
}
