package di

import "errors"

var (
	// ErrSealed means registration was attempted after resolution started.
	ErrSealed = errors.New("di: container is sealed")
	// ErrClosed means an operation was attempted after shutdown started.
	ErrClosed = errors.New("di: container is closed")
	// ErrDuplicate means the same typed registration already exists.
	ErrDuplicate = errors.New("di: dependency already registered")
	// ErrNotFound means no dependency was registered for the requested type and name.
	ErrNotFound = errors.New("di: dependency not found")
	// ErrCircularDependency means the provider graph contains a cycle.
	ErrCircularDependency = errors.New("di: circular dependency")
	// ErrInvalidArgument means a name, provider, cleanup, resolver, or context is invalid.
	ErrInvalidArgument = errors.New("di: invalid argument")
	// ErrNilService means a provider or value produced a nil dependency without an error.
	ErrNilService = errors.New("di: nil dependency")
)
