// Package di provides a small, type-safe dependency container for application
// composition roots.
//
// # Application integration
//
// The package intentionally exposes only lazy singletons, values, named
// registrations, explicit resource ownership, resolution, and shutdown. It
// does not expose its underlying DI engine. Application code should keep a
// Container inside its composition package (commonly internal/deps) and expose
// application-specific typed getters to entrypoints.
//
// Register the complete graph before resolving anything. A scenario constructor
// should eagerly resolve and cache the validated config and only the runtime
// roots that scenario needs. Fixed getters can then be error-free because
// construction already proved the graph. If construction fails after creating
// resources, the application owns a bounded rollback context and should join
// the build and shutdown errors.
//
// # Providers and resolution
//
// Registration and resolution are separate phases. The first call to Resolve
// or ResolveNamed seals the container; later registrations fail with
// ErrSealed. Providers must resolve their dependencies synchronously through
// the Resolver passed to them:
//
//	di.Provide(container, func(r di.Resolver) (*Service, error) {
//		repository, err := di.Resolve[Repository](r)
//		if err != nil {
//			return nil, err
//		}
//		return NewService(repository), nil
//	})
//
// A provider must not retain its Resolver or resolve through a captured root
// Container. Doing so either loses the dependency edge or uses an expired
// resolution scope.
//
// # Resource ownership
//
// ProvideResource and ProvideNamedResource explicitly transfer cleanup
// ownership to the container. Provide and ProvideValue never infer ownership
// from methods such as Close, Shutdown, Stop, or Sync. During Shutdown,
// dependents are stopped before their dependencies, while independent branches
// may be stopped concurrently. Only resources that were successfully built are
// stopped.
//
// The runtime owns the shutdown deadline and calls Container.Shutdown with a
// fresh context. Cleanup failures remain compatible with errors.Is through the
// joined shutdown result.
//
// # Testing
//
// Use a real Container in composition-root smoke tests to prove registration,
// selected implementations, eager roots, and explicit ownership. Do not mock
// the DI container in component unit tests; construct the component directly
// with mocks of its own consumer-side dependencies.
package di
