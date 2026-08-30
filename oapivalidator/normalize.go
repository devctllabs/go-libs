package oapivalidator

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
)

func newFailure(kind FailureKind, operationID string, fieldErrors []FieldError, truncated bool, cause error) *Failure {
	return &Failure{
		Kind:        kind,
		Status:      statusForKind(kind),
		OperationID: operationID,
		Errors:      fieldErrors,
		Truncated:   truncated,
		Cause:       cause,
	}
}

func normalizeFailure(cause error, operationID string, limit int) *Failure {
	if kind, challenge, ok := authenticationFailure(cause); ok {
		failure := newFailure(kind, operationID, nil, false, cause)
		failure.wwwAuthenticate = challenge
		return failure
	}

	kind := validationFailureKind(cause)
	fieldErrors := make([]FieldError, 0)
	appendFieldErrors(cause, fieldContext{}, &fieldErrors)
	fieldErrors = sortAndDedupe(fieldErrors)
	truncated := len(fieldErrors) > limit
	if truncated {
		fieldErrors = fieldErrors[:limit]
	}
	return newFailure(kind, operationID, fieldErrors, truncated, cause)
}

func authenticationFailure(cause error) (FailureKind, string, bool) {
	authenticationErrors := make([]*authenticationError, 0)
	walkErrors(cause, func(err error) {
		if authenticationFailure, ok := err.(*authenticationError); ok {
			authenticationErrors = append(authenticationErrors, authenticationFailure)
		}
	})
	if len(authenticationErrors) == 0 {
		return "", "", false
	}

	for _, failure := range authenticationErrors {
		if !errors.Is(failure.cause, ErrUnauthenticated) && !errors.Is(failure.cause, ErrForbidden) && !errors.Is(failure.cause, ErrAuthenticationUnavailable) {
			return FailureInternal, "", true
		}
	}
	for _, failure := range authenticationErrors {
		if errors.Is(failure.cause, ErrAuthenticationUnavailable) {
			return FailureAuthenticationUnavailable, "", true
		}
	}
	for _, failure := range authenticationErrors {
		if errors.Is(failure.cause, ErrForbidden) {
			return FailureForbidden, "", true
		}
	}
	for _, failure := range authenticationErrors {
		if errors.Is(failure.cause, ErrUnauthenticated) {
			return FailureUnauthenticated, authenticationChallenge(failure.scheme), true
		}
	}
	return FailureInternal, "", true
}

func authenticationChallenge(scheme *openapi3.SecurityScheme) string {
	if scheme == nil {
		return ""
	}
	switch strings.ToLower(scheme.Type) {
	case "oauth2", "openidconnect":
		return "Bearer"
	case "http":
		switch strings.ToLower(scheme.Scheme) {
		case "bearer":
			return "Bearer"
		case "basic":
			return "Basic"
		}
	}
	return ""
}

func validationFailureKind(cause error) FailureKind {
	hasRequestFailure := false
	kind := FailureInvalidRequest
	walkErrors(cause, func(err error) {
		requestFailure, ok := err.(*openapi3filter.RequestError)
		if !ok {
			return
		}
		hasRequestFailure = true
		if isUnsupportedMediaType(requestFailure) {
			kind = FailureUnsupportedMediaType
			return
		}
		if kind != FailureUnsupportedMediaType && isMalformed(requestFailure) {
			kind = FailureMalformedRequest
		}
	})
	if !hasRequestFailure {
		return FailureInternal
	}
	return kind
}

func isUnsupportedMediaType(failure *openapi3filter.RequestError) bool {
	return strings.HasPrefix(failure.Reason, "header Content-Type has unexpected value") ||
		strings.Contains(errorText(failure.Err), "unsupported content type")
}

func isMalformed(failure *openapi3filter.RequestError) bool {
	if failure.Reason == "failed to decode request body" || failure.Reason == "reading failed" {
		return true
	}
	var syntaxError *json.SyntaxError
	if errors.As(failure.Err, &syntaxError) {
		return true
	}
	if failure.Parameter == nil || failure.Err == nil ||
		errors.Is(failure.Err, openapi3filter.ErrInvalidRequired) ||
		errors.Is(failure.Err, openapi3filter.ErrInvalidEmptyValue) {
		return false
	}
	var schemaError *openapi3.SchemaError
	return !errors.As(failure.Err, &schemaError)
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

type fieldContext struct {
	location  Location
	parameter string
	body      bool
}

const requiredFieldCode = "required"

func appendFieldErrors(err error, inherited fieldContext, destination *[]FieldError) {
	if err == nil {
		return
	}
	switch failure := err.(type) {
	case *authenticationError:
		return
	case openapi3.MultiError:
		for _, child := range failure {
			appendFieldErrors(child, inherited, destination)
		}
		return
	case *openapi3filter.SecurityRequirementsError:
		return
	case *openapi3filter.RequestError:
		appendRequestFieldErrors(failure, inherited, destination)
	case *openapi3.SchemaError:
		appendSchemaFieldError(failure, inherited, destination)
	default:
		appendWrappedFieldErrors(err, inherited, destination)
	}
}

func appendRequestFieldErrors(failure *openapi3filter.RequestError, inherited fieldContext, destination *[]FieldError) {
	context := inherited
	if failure.Parameter != nil {
		context.location = parameterLocation(failure.Parameter.In)
		context.parameter = failure.Parameter.Name
		context.body = false
	}
	if failure.RequestBody != nil {
		context.location = LocationBody
		context.parameter = ""
		context.body = true
	}
	if failure.Err != nil && hasStructuredChildren(failure.Err) {
		appendFieldErrors(failure.Err, context, destination)
		return
	}
	*destination = append(*destination, requestFieldError(failure, context))
}

func appendSchemaFieldError(failure *openapi3.SchemaError, inherited fieldContext, destination *[]FieldError) {
	if failure.Origin != nil {
		appendFieldErrors(failure.Origin, inherited, destination)
		return
	}
	code := "invalid"
	detail := "The value does not satisfy the schema."
	if failure.SchemaField == requiredFieldCode {
		code = requiredFieldCode
		detail = "A required value is missing."
	}
	fieldError := FieldError{Code: code, Detail: detail, In: inherited.location, Parameter: inherited.parameter}
	if inherited.body {
		fieldError.Pointer = jsonPointer(failure.JSONPointer())
		if fieldError.Pointer == "" {
			fieldError.Pointer = schemaPointerFromReason(failure.Reason)
		}
	}
	*destination = append(*destination, fieldError)
}

func appendWrappedFieldErrors(err error, inherited fieldContext, destination *[]FieldError) {
	type multiUnwrapper interface{ Unwrap() []error }
	if unwrapped, ok := err.(multiUnwrapper); ok {
		for _, child := range unwrapped.Unwrap() {
			appendFieldErrors(child, inherited, destination)
		}
		return
	}
	if child := errors.Unwrap(err); child != nil {
		appendFieldErrors(child, inherited, destination)
	}
}

func schemaPointerFromReason(reason string) string {
	const marker = `error at "`
	start := strings.Index(reason, marker)
	if start < 0 {
		return ""
	}
	start += len(marker)
	end := strings.IndexByte(reason[start:], '"')
	if end < 0 {
		return ""
	}
	pointer := reason[start : start+end]
	if !strings.HasPrefix(pointer, "/") {
		return ""
	}
	return pointer
}

func hasStructuredChildren(err error) bool {
	if err == nil {
		return false
	}
	switch err.(type) {
	case openapi3.MultiError, *openapi3.SchemaError:
		return true
	default:
		return false
	}
}

func requestFieldError(failure *openapi3filter.RequestError, context fieldContext) FieldError {
	code := "invalid"
	detail := "The value is invalid."
	switch {
	case isUnsupportedMediaType(failure):
		code = "unsupported_media_type"
		detail = "The content type is not supported."
	case isMalformed(failure):
		code = "malformed"
		detail = "The value could not be decoded."
	case errors.Is(failure.Err, openapi3filter.ErrInvalidRequired):
		code = "required"
		detail = "A required value is missing."
	}
	return FieldError{Code: code, Detail: detail, In: context.location, Parameter: context.parameter}
}

func parameterLocation(in string) Location {
	switch in {
	case openapi3.ParameterInPath:
		return LocationPath
	case openapi3.ParameterInQuery:
		return LocationQuery
	case openapi3.ParameterInHeader:
		return LocationHeader
	case openapi3.ParameterInCookie:
		return LocationCookie
	default:
		return ""
	}
}

func jsonPointer(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	escaped := make([]string, len(parts))
	for index, part := range parts {
		part = strings.ReplaceAll(part, "~", "~0")
		escaped[index] = strings.ReplaceAll(part, "/", "~1")
	}
	return "/" + strings.Join(escaped, "/")
}

func sortAndDedupe(fieldErrors []FieldError) []FieldError {
	sort.SliceStable(fieldErrors, func(left, right int) bool {
		return fieldErrorKey(fieldErrors[left]) < fieldErrorKey(fieldErrors[right])
	})
	result := fieldErrors[:0]
	var previous string
	for index, fieldError := range fieldErrors {
		key := fieldErrorKey(fieldError)
		if index > 0 && key == previous {
			continue
		}
		result = append(result, fieldError)
		previous = key
	}
	return result
}

func fieldErrorKey(fieldError FieldError) string {
	return strings.Join([]string{string(fieldError.In), fieldError.Pointer, fieldError.Parameter, fieldError.Code, fieldError.Detail}, "\x00")
}

func walkErrors(err error, visit func(error)) {
	if err == nil {
		return
	}
	visit(err)
	if multiError, ok := err.(openapi3.MultiError); ok {
		for _, child := range multiError {
			walkErrors(child, visit)
		}
		return
	}
	type multiUnwrapper interface{ Unwrap() []error }
	if unwrapped, ok := err.(multiUnwrapper); ok {
		for _, child := range unwrapped.Unwrap() {
			walkErrors(child, visit)
		}
		return
	}
	if child := errors.Unwrap(err); child != nil {
		walkErrors(child, visit)
	}
}
