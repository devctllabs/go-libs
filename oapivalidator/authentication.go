package oapivalidator

import (
	"context"
	"errors"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
)

// ErrUnauthenticated reports missing or invalid credentials.
var ErrUnauthenticated = errors.New("unauthenticated")

// ErrForbidden reports valid credentials without the required access.
var ErrForbidden = errors.New("forbidden")

// ErrAuthenticationUnavailable reports a temporary authentication dependency failure.
var ErrAuthenticationUnavailable = errors.New("authentication unavailable")

// AuthenticationInput describes one security scheme in the current OpenAPI
// security requirement.
type AuthenticationInput struct {
	Request            *http.Request
	OperationID        string
	SecuritySchemeName string
	SecurityScheme     *openapi3.SecurityScheme
	Scopes             []string
}

// Authenticator validates one OpenAPI security scheme at a time.
type Authenticator interface {
	// Authenticate validates the requested scheme and returns the context that
	// subsequent schemes and the endpoint handler receive.
	Authenticate(ctx context.Context, input AuthenticationInput) (nextCtx context.Context, err error)
}

// AuthenticatorFunc adapts a function to Authenticator.
type AuthenticatorFunc func(ctx context.Context, input AuthenticationInput) (nextCtx context.Context, err error)

// Authenticate implements Authenticator.
func (fn AuthenticatorFunc) Authenticate(ctx context.Context, input AuthenticationInput) (context.Context, error) {
	return fn(ctx, input)
}

type authenticationError struct {
	cause  error
	scheme *openapi3.SecurityScheme
}

func (*authenticationError) Error() string { return "openapi authentication failed" }

func (failure *authenticationError) Unwrap() error { return failure.cause }

var errNilAuthenticationContext = errors.New("authenticator returned a nil context without an error")
