package oapivalidator

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
)

// FailureKind identifies a stable category of request validation failure.
type FailureKind string

const (
	FailureNotFound             FailureKind = "not_found"
	FailureMethodNotAllowed     FailureKind = "method_not_allowed"
	FailureMalformedRequest     FailureKind = "malformed_request"
	FailureInvalidRequest       FailureKind = "invalid_request"
	FailureUnsupportedMediaType FailureKind = "unsupported_media_type"
	FailureUnauthenticated      FailureKind = "unauthenticated"
	FailureForbidden            FailureKind = "forbidden"
	FailureInternal             FailureKind = "internal"
)

const (
	ProblemTypeNotFound             = "urn:devctl:oapivalidator:problem:not-found"
	ProblemTypeMethodNotAllowed     = "urn:devctl:oapivalidator:problem:method-not-allowed"
	ProblemTypeMalformedRequest     = "urn:devctl:oapivalidator:problem:malformed-request"
	ProblemTypeInvalidRequest       = "urn:devctl:oapivalidator:problem:invalid-request"
	ProblemTypeUnsupportedMediaType = "urn:devctl:oapivalidator:problem:unsupported-media-type"
	ProblemTypeUnauthenticated      = "urn:devctl:oapivalidator:problem:unauthenticated"
	ProblemTypeForbidden            = "urn:devctl:oapivalidator:problem:forbidden"
	ProblemTypeInternal             = "urn:devctl:oapivalidator:problem:internal"
)

// Location identifies the part of the request containing an invalid value.
type Location string

const (
	LocationBody   Location = "body"
	LocationPath   Location = "path"
	LocationQuery  Location = "query"
	LocationHeader Location = "header"
	LocationCookie Location = "cookie"
)

// FieldError is a safe, normalized request validation error.
type FieldError struct {
	Code      string   `json:"code"`
	Detail    string   `json:"detail"`
	In        Location `json:"in,omitempty"`
	Pointer   string   `json:"pointer,omitempty"`
	Parameter string   `json:"parameter,omitempty"`
}

// Problem is an RFC 9457 problem details response.
type Problem struct {
	Type      string       `json:"type"`
	Title     string       `json:"title"`
	Status    int          `json:"status"`
	Detail    string       `json:"detail,omitempty"`
	Instance  string       `json:"instance,omitempty"`
	Errors    []FieldError `json:"errors,omitempty"`
	Truncated bool         `json:"truncated,omitempty"`
}

// Failure retains the private cause while exposing safe client diagnostics.
type Failure struct {
	Kind        FailureKind
	Status      int
	OperationID string
	Errors      []FieldError
	Truncated   bool
	Cause       error

	wwwAuthenticate string
}

// Error implements error without exposing the underlying validator message.
func (failure *Failure) Error() string {
	if failure == nil {
		return "<nil>"
	}
	return fmt.Sprintf("openapi request validation failed (%s)", failure.Kind)
}

// Unwrap exposes the original failure to server-side errors.Is/errors.As.
func (failure *Failure) Unwrap() error {
	if failure == nil {
		return nil
	}
	return failure.Cause
}

// Problem returns the safe RFC 9457 representation of failure.
func (failure *Failure) Problem() Problem {
	return problemForFailure(failure)
}

// FailureHandler handles a normalized validation failure.
type FailureHandler interface {
	// Handle writes or returns the response for failure.
	Handle(c *echo.Context, failure *Failure) error
}

// FailureHandlerFunc adapts a function to FailureHandler.
type FailureHandlerFunc func(c *echo.Context, failure *Failure) error

// Handle implements FailureHandler.
func (fn FailureHandlerFunc) Handle(c *echo.Context, failure *Failure) error {
	return fn(c, failure)
}

func problemForFailure(failure *Failure) Problem {
	problemType, title, detail := problemMetadata(failure.Kind)
	return Problem{
		Type:      problemType,
		Title:     title,
		Status:    failure.Status,
		Detail:    detail,
		Errors:    failure.Errors,
		Truncated: failure.Truncated,
	}
}

func problemMetadata(kind FailureKind) (problemType, title, detail string) {
	switch kind {
	case FailureNotFound:
		return ProblemTypeNotFound, "Operation not found", "No OpenAPI operation matches this request."
	case FailureMethodNotAllowed:
		return ProblemTypeMethodNotAllowed, "Method not allowed", "The path does not support this HTTP method."
	case FailureMalformedRequest:
		return ProblemTypeMalformedRequest, "Malformed request", "The request could not be decoded."
	case FailureInvalidRequest:
		return ProblemTypeInvalidRequest, "Invalid request", "The request does not satisfy the API contract."
	case FailureUnsupportedMediaType:
		return ProblemTypeUnsupportedMediaType, "Unsupported media type", "The request content type is not supported."
	case FailureUnauthenticated:
		return ProblemTypeUnauthenticated, "Authentication required", "Valid credentials are required."
	case FailureForbidden:
		return ProblemTypeForbidden, "Forbidden", "The credentials do not grant the required access."
	default:
		return ProblemTypeInternal, "Internal server error", "The request could not be validated."
	}
}

type defaultFailureHandler struct{}

func (defaultFailureHandler) Handle(c *echo.Context, failure *Failure) error {
	c.Response().Header().Set("Content-Type", "application/problem+json")
	return c.JSON(failure.Status, failure.Problem())
}

func statusForKind(kind FailureKind) int {
	switch kind {
	case FailureNotFound:
		return http.StatusNotFound
	case FailureMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case FailureMalformedRequest:
		return http.StatusBadRequest
	case FailureInvalidRequest:
		return http.StatusUnprocessableEntity
	case FailureUnsupportedMediaType:
		return http.StatusUnsupportedMediaType
	case FailureUnauthenticated:
		return http.StatusUnauthorized
	case FailureForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
