package oapivalidator_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/devctllabs/go-libs/oapivalidator"
	"github.com/devctllabs/go-libs/oapivalidator/mocks"
)

type contextKey string

func TestNewRequiresAuthenticatorOnlyForMandatorySecurity(t *testing.T) {
	t.Parallel()
	_, err := oapivalidator.New(loadDocument(t, securedDocument))
	require.Error(t, err)

	middleware, err := oapivalidator.New(loadDocument(t, optionalSecurityDocument))
	require.NoError(t, err)
	recorder := serve(t, middleware, http.MethodGet, "/optional", "", "", nil)
	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestAuthenticatorReceivesSchemeScopesAndEnrichesHandlerContext(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	authenticator := mocks.NewMockAuthenticator(controller)
	first := authenticator.EXPECT().Authenticate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, input oapivalidator.AuthenticationInput) (context.Context, error) {
			require.Equal(t, "andSecurity", input.OperationID)
			require.Equal(t, "apiKey", input.SecuritySchemeName)
			require.Empty(t, input.Scopes)
			return context.WithValue(ctx, contextKey("api-key"), "accepted"), nil
		},
	)
	authenticator.EXPECT().Authenticate(gomock.Any(), gomock.Any()).After(first).DoAndReturn(
		func(ctx context.Context, input oapivalidator.AuthenticationInput) (context.Context, error) {
			require.Equal(t, "bearerAuth", input.SecuritySchemeName)
			require.Equal(t, []string{"widgets:read"}, input.Scopes)
			require.Equal(t, "accepted", ctx.Value(contextKey("api-key")))
			return context.WithValue(ctx, contextKey("subject"), "user-1"), nil
		},
	)
	middleware, err := oapivalidator.New(loadDocument(t, securedDocument), oapivalidator.WithAuthenticator(authenticator))
	require.NoError(t, err)

	recorder := serve(t, middleware, http.MethodGet, "/and", "", "", func(c *echo.Context) error {
		require.Equal(t, "accepted", c.Request().Context().Value(contextKey("api-key")))
		require.Equal(t, "user-1", c.Request().Context().Value(contextKey("subject")))
		return c.NoContent(http.StatusNoContent)
	})

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestAuthenticatorAccumulatesContextAcrossFailedSecurityAlternative(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	authenticator := mocks.NewMockAuthenticator(controller)
	apiKey := authenticator.EXPECT().Authenticate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, input oapivalidator.AuthenticationInput) (context.Context, error) {
			require.Equal(t, "apiKey", input.SecuritySchemeName)
			return context.WithValue(ctx, contextKey("first-alternative"), "enriched"), nil
		},
	)
	bearer := authenticator.EXPECT().Authenticate(gomock.Any(), gomock.Any()).After(apiKey).DoAndReturn(
		func(ctx context.Context, input oapivalidator.AuthenticationInput) (context.Context, error) {
			require.Equal(t, "bearerAuth", input.SecuritySchemeName)
			require.Equal(t, "enriched", ctx.Value(contextKey("first-alternative")))
			return ctx, oapivalidator.ErrUnauthenticated
		},
	)
	authenticator.EXPECT().Authenticate(gomock.Any(), gomock.Any()).After(bearer).DoAndReturn(
		func(ctx context.Context, input oapivalidator.AuthenticationInput) (context.Context, error) {
			require.Equal(t, "cookieKey", input.SecuritySchemeName)
			require.Equal(t, "enriched", ctx.Value(contextKey("first-alternative")))
			return context.WithValue(ctx, contextKey("subject"), "fallback-user"), nil
		},
	)
	middleware, err := oapivalidator.New(loadDocument(t, securedDocument), oapivalidator.WithAuthenticator(authenticator))
	require.NoError(t, err)

	recorder := serve(t, middleware, http.MethodGet, "/fallback", "", "", func(c *echo.Context) error {
		require.Equal(t, "enriched", c.Request().Context().Value(contextKey("first-alternative")))
		require.Equal(t, "fallback-user", c.Request().Context().Value(contextKey("subject")))
		return c.NoContent(http.StatusNoContent)
	})

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestAuthenticationFailureMapping(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		authError error
		status    int
		challenge string
	}{
		{name: "unauthenticated", authError: fmt.Errorf("wrapped: %w", oapivalidator.ErrUnauthenticated), status: http.StatusUnauthorized, challenge: "Bearer"},
		{name: "forbidden", authError: fmt.Errorf("wrapped: %w", oapivalidator.ErrForbidden), status: http.StatusForbidden},
		{name: "authentication unavailable", authError: fmt.Errorf("wrapped: %w", oapivalidator.ErrAuthenticationUnavailable), status: http.StatusServiceUnavailable},
		{name: "backend failure", authError: errors.New("identity provider unavailable"), status: http.StatusInternalServerError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			controller := gomock.NewController(t)
			authenticator := mocks.NewMockAuthenticator(controller)
			authenticator.EXPECT().Authenticate(gomock.Any(), gomock.Any()).Return(context.Background(), test.authError)
			middleware, err := oapivalidator.New(loadDocument(t, securedDocument), oapivalidator.WithAuthenticator(authenticator))
			require.NoError(t, err)

			recorder := serve(t, middleware, http.MethodGet, "/secure", "", "", nil)

			require.Equal(t, test.status, recorder.Code)
			require.Equal(t, test.challenge, recorder.Header().Get("WWW-Authenticate"))
			require.NotContains(t, recorder.Body.String(), "identity provider")
		})
	}
}

func TestAuthenticatorRejectsNilSuccessContext(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	authenticator := mocks.NewMockAuthenticator(controller)
	authenticator.EXPECT().Authenticate(gomock.Any(), gomock.Any()).Return(nil, nil)
	middleware, err := oapivalidator.New(loadDocument(t, securedDocument), oapivalidator.WithAuthenticator(authenticator))
	require.NoError(t, err)

	recorder := serve(t, middleware, http.MethodGet, "/secure", "", "", nil)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestAuthenticationFailurePriority(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		errors []error
		status int
	}{
		{name: "unavailable over forbidden and unauthenticated", errors: []error{oapivalidator.ErrUnauthenticated, oapivalidator.ErrForbidden, oapivalidator.ErrAuthenticationUnavailable}, status: http.StatusServiceUnavailable},
		{name: "internal over unavailable", errors: []error{oapivalidator.ErrUnauthenticated, oapivalidator.ErrAuthenticationUnavailable, errors.New("unexpected")}, status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			controller := gomock.NewController(t)
			authenticator := mocks.NewMockAuthenticator(controller)
			var previous *gomock.Call
			for _, authError := range test.errors {
				call := authenticator.EXPECT().Authenticate(gomock.Any(), gomock.Any()).Return(context.Background(), authError)
				if previous != nil {
					call.After(previous)
				}
				previous = call
			}
			middleware, err := oapivalidator.New(loadDocument(t, priorityDocument), oapivalidator.WithAuthenticator(authenticator))
			require.NoError(t, err)

			recorder := serve(t, middleware, http.MethodGet, "/priority", "", "", nil)

			require.Equal(t, test.status, recorder.Code)
		})
	}
}

func TestMiddlewareDoesNotLeakAuthenticationContextBetweenConcurrentRequests(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	authenticator := mocks.NewMockAuthenticator(controller)
	authenticator.EXPECT().Authenticate(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, input oapivalidator.AuthenticationInput) (context.Context, error) {
			requestID := input.Request.Header.Get("X-Request-ID")
			return context.WithValue(ctx, contextKey("request-id"), requestID), nil
		},
	)
	middleware, err := oapivalidator.New(loadDocument(t, securedDocument), oapivalidator.WithAuthenticator(authenticator))
	require.NoError(t, err)

	const requestCount = 32
	errorsChannel := make(chan error, requestCount)
	var waitGroup sync.WaitGroup
	for index := range requestCount {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			requestID := fmt.Sprintf("request-%d", index)
			request := httptest.NewRequest(http.MethodGet, "/secure", nil)
			request.Header.Set("X-Request-ID", requestID)
			recorder := httptest.NewRecorder()
			context := echo.New().NewContext(request, recorder)
			err := middleware(func(c *echo.Context) error {
				actual, _ := c.Request().Context().Value(contextKey("request-id")).(string)
				if actual != requestID {
					return fmt.Errorf("context request ID = %q, want %q", actual, requestID)
				}
				return c.NoContent(http.StatusNoContent)
			})(context)
			if err == nil && recorder.Code != http.StatusNoContent {
				err = fmt.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
			}
			errorsChannel <- err
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)

	for err := range errorsChannel {
		require.NoError(t, err)
	}
}

const securedDocument = `
openapi: 3.1.0
info: {title: Secured API, version: 1.0.0}
components:
  securitySchemes:
    apiKey:
      type: apiKey
      in: header
      name: X-API-Key
    bearerAuth:
      type: http
      scheme: bearer
    cookieKey:
      type: apiKey
      in: cookie
      name: session
security:
  - bearerAuth: []
paths:
  /secure:
    get:
      operationId: secure
      responses:
        "204": {description: OK}
  /and:
    get:
      operationId: andSecurity
      security:
        - apiKey: []
          bearerAuth: [widgets:read]
      responses:
        "204": {description: OK}
  /fallback:
    get:
      operationId: fallbackSecurity
      security:
        - apiKey: []
          bearerAuth: []
        - cookieKey: []
      responses:
        "204": {description: OK}
`

const optionalSecurityDocument = `
openapi: 3.1.0
info: {title: Optional API, version: 1.0.0}
components:
  securitySchemes:
    bearerAuth: {type: http, scheme: bearer}
paths:
  /optional:
    get:
      operationId: optional
      security:
        - bearerAuth: []
        - {}
      responses:
        "204": {description: OK}
`

const priorityDocument = `
openapi: 3.1.0
info: {title: Priority API, version: 1.0.0}
components:
  securitySchemes:
    first: {type: http, scheme: bearer}
    second: {type: apiKey, in: header, name: X-API-Key}
    third: {type: apiKey, in: cookie, name: session}
paths:
  /priority:
    get:
      operationId: priority
      security:
        - first: []
        - second: []
        - third: []
      responses:
        "204": {description: OK}
`
