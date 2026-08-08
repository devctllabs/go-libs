package oapivalidator

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
	"github.com/labstack/echo/v5"
)

//go:generate go tool mockgen -destination mocks/interfaces.gen.go -package mocks . Authenticator,FailureHandler

// New constructs request validation middleware for document.
func New(document *openapi3.T, options ...Option) (echo.MiddlewareFunc, error) {
	if document == nil {
		return nil, errors.New("openapi document must not be nil")
	}
	if err := document.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("validate OpenAPI document: %w", err)
	}

	configuration := defaultConfig()
	for index, option := range options {
		if option == nil {
			return nil, fmt.Errorf("option %d must not be nil", index)
		}
		if err := option.apply(&configuration); err != nil {
			return nil, fmt.Errorf("apply option %d: %w", index, err)
		}
	}
	if configuration.failureHandler == nil {
		configuration.failureHandler = defaultFailureHandler{}
	}
	if configuration.authenticator == nil && documentRequiresAuthentication(document) {
		return nil, errors.New("openapi document has operations with mandatory security but no authenticator is configured")
	}

	routingDocument := *document
	routingDocument.Servers = nil
	if routingDocument.Paths == nil {
		routingDocument.Paths = openapi3.NewPaths()
	}
	router, err := legacyrouter.NewRouter(&routingDocument)
	if err != nil {
		return nil, fmt.Errorf("construct OpenAPI router: %w", err)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			return validate(c, next, router, configuration)
		}
	}, nil
}

func validate(c *echo.Context, next echo.HandlerFunc, router routers.Router, configuration config) error {
	validationRequest, ok := requestForValidation(c.Request(), configuration.baseURL)
	if !ok {
		failure := newFailure(FailureNotFound, "", nil, false, routers.ErrPathNotFound)
		return handleFailure(c, configuration.failureHandler, failure)
	}

	route, pathParameters, err := router.FindRoute(validationRequest)
	if err != nil {
		kind := FailureNotFound
		allow := allowedMethods(router, validationRequest)
		if isRouteError(err, routers.ErrMethodNotAllowed) || allow != "" {
			kind = FailureMethodNotAllowed
			c.Response().Header().Set("Allow", allow)
		}
		return handleFailure(c, configuration.failureHandler, newFailure(kind, "", nil, false, err))
	}

	validationOptions := &openapi3filter.Options{MultiError: true}
	validationInput := &openapi3filter.RequestValidationInput{
		Request:    validationRequest,
		PathParams: pathParameters,
		Route:      route,
		Options:    validationOptions,
	}
	validationOptions.AuthenticationFunc = authenticationFunc(c, configuration.authenticator)
	if err := openapi3filter.ValidateRequest(validationRequest.Context(), validationInput); err != nil {
		synchronizeRequest(c, validationInput.Request)
		failure := normalizeFailure(err, route.Operation.OperationID, configuration.maxReportedErrors)
		return handleFailure(c, configuration.failureHandler, failure)
	}

	synchronizeRequest(c, validationInput.Request)
	return next(c)
}

func requestForValidation(request *http.Request, baseURL string) (*http.Request, bool) {
	validationRequest := request.Clone(request.Context())
	urlCopy := *request.URL
	validationRequest.URL = &urlCopy
	if baseURL == "" {
		return validationRequest, true
	}
	path := validationRequest.URL.Path
	if path != baseURL && !strings.HasPrefix(path, baseURL+"/") {
		return nil, false
	}
	path = strings.TrimPrefix(path, baseURL)
	if path == "" {
		path = "/"
	}
	validationRequest.URL.Path = path
	validationRequest.URL.RawPath = ""
	return validationRequest, true
}

func synchronizeRequest(c *echo.Context, validated *http.Request) {
	if validated == nil {
		return
	}
	current := c.Request().WithContext(validated.Context())
	current.Body = validated.Body
	current.GetBody = validated.GetBody
	current.ContentLength = validated.ContentLength
	c.SetRequest(current)
}

func authenticationFunc(c *echo.Context, authenticator Authenticator) openapi3filter.AuthenticationFunc {
	return func(_ context.Context, input *openapi3filter.AuthenticationInput) error {
		if authenticator == nil {
			return &authenticationError{cause: ErrUnauthenticated, scheme: input.SecurityScheme}
		}
		currentContext := input.RequestValidationInput.Request.Context()
		nextContext, err := authenticator.Authenticate(currentContext, AuthenticationInput{
			Request:            input.RequestValidationInput.Request,
			OperationID:        input.RequestValidationInput.Route.Operation.OperationID,
			SecuritySchemeName: input.SecuritySchemeName,
			SecurityScheme:     input.SecurityScheme,
			Scopes:             slices.Clone(input.Scopes),
		})
		if err != nil {
			return &authenticationError{cause: err, scheme: input.SecurityScheme}
		}
		if nextContext == nil {
			return &authenticationError{cause: errNilAuthenticationContext, scheme: input.SecurityScheme}
		}
		updatedValidationRequest := input.RequestValidationInput.Request.WithContext(nextContext)
		input.RequestValidationInput.Request = updatedValidationRequest
		c.SetRequest(c.Request().WithContext(nextContext))
		return nil
	}
}

func documentRequiresAuthentication(document *openapi3.T) bool {
	if document.Paths == nil {
		return false
	}
	for _, pathItem := range document.Paths.Map() {
		for _, operation := range pathItem.Operations() {
			security := operation.Security
			if security == nil {
				security = &document.Security
			}
			if securityIsMandatory(security) {
				return true
			}
		}
	}
	return false
}

func handleFailure(c *echo.Context, handler FailureHandler, failure *Failure) error {
	if failure.wwwAuthenticate != "" {
		c.Response().Header().Set("WWW-Authenticate", failure.wwwAuthenticate)
	}
	return handler.Handle(c, failure)
}

func securityIsMandatory(requirements *openapi3.SecurityRequirements) bool {
	if requirements == nil || len(*requirements) == 0 {
		return false
	}
	for _, requirement := range *requirements {
		if len(requirement) == 0 {
			return false
		}
	}
	return true
}

func isRouteError(actual, target error) bool {
	return actual != nil && target != nil && actual.Error() == target.Error()
}

func allowedMethods(router routers.Router, request *http.Request) string {
	methods := []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodConnect, http.MethodTrace}
	allowed := make([]string, 0, len(methods))
	for _, method := range methods {
		candidate := request.Clone(request.Context())
		candidate.Method = method
		if _, _, err := router.FindRoute(candidate); err == nil {
			allowed = append(allowed, method)
		}
	}
	return strings.Join(allowed, ", ")
}
