package oapivalidator_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/devctllabs/go-libs/oapivalidator"
	"github.com/devctllabs/go-libs/oapivalidator/mocks"
)

func TestMiddlewareAcceptsValidOpenAPI31Request(t *testing.T) {
	t.Parallel()
	middleware, err := oapivalidator.New(loadDocument(t, publicDocument))
	require.NoError(t, err)

	recorder := serve(t, middleware, http.MethodGet, "/widgets/widget-1", "", "", nil)

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestMiddlewareUsesOpenAPI31NullableTypeSemantics(t *testing.T) {
	t.Parallel()
	middleware, err := oapivalidator.New(loadDocument(t, publicDocument))
	require.NoError(t, err)

	nullResponse := serve(t, middleware, http.MethodPost, "/nullable", `null`, "application/json", nil)
	stringResponse := serve(t, middleware, http.MethodPost, "/nullable", `"value"`, "application/json", nil)
	invalidResponse := serve(t, middleware, http.MethodPost, "/nullable", `42`, "application/json", nil)

	require.Equal(t, http.StatusNoContent, nullResponse.Code)
	require.Equal(t, http.StatusNoContent, stringResponse.Code)
	require.Equal(t, http.StatusUnprocessableEntity, invalidResponse.Code)
}

func TestMiddlewareIgnoresDocumentServers(t *testing.T) {
	t.Parallel()
	document := loadDocument(t, strings.Replace(publicDocument, "paths:", "servers:\n  - url: https://example.com/service\npaths:", 1))
	middleware, err := oapivalidator.New(document)
	require.NoError(t, err)

	recorder := serve(t, middleware, http.MethodGet, "/widgets/widget-1", "", "", nil)

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestMiddlewareSupportsGeneratedBaseURL(t *testing.T) {
	t.Parallel()
	middleware, err := oapivalidator.New(loadDocument(t, publicDocument), oapivalidator.WithBaseURL("/api/v1"))
	require.NoError(t, err)

	recorder := serve(t, middleware, http.MethodGet, "/api/v1/widgets/widget-1", "", "", nil)

	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestMiddlewareWritesSafeRoutingProblems(t *testing.T) {
	t.Parallel()
	middleware, err := oapivalidator.New(loadDocument(t, publicDocument))
	require.NoError(t, err)

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		recorder := serve(t, middleware, http.MethodGet, "/missing", "", "", nil)
		problem := decodeProblem(t, recorder)
		require.Equal(t, http.StatusNotFound, recorder.Code)
		require.Equal(t, "application/problem+json", recorder.Header().Get("Content-Type"))
		require.Equal(t, oapivalidator.ProblemTypeNotFound, problem.Type)
	})

	t.Run("method not allowed", func(t *testing.T) {
		t.Parallel()
		recorder := serve(t, middleware, http.MethodDelete, "/widgets/widget-1", "", "", nil)
		problem := decodeProblem(t, recorder)
		require.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
		require.Equal(t, "GET", recorder.Header().Get("Allow"))
		require.Equal(t, oapivalidator.ProblemTypeMethodNotAllowed, problem.Type)
	})
}

func TestMiddlewareClassifiesAndNormalizesRequestFailures(t *testing.T) {
	t.Parallel()
	middleware, err := oapivalidator.New(loadDocument(t, publicDocument))
	require.NoError(t, err)

	tests := []struct {
		name        string
		method      string
		target      string
		body        string
		contentType string
		status      int
		kindType    string
		location    oapivalidator.Location
	}{
		{
			name: "path schema", method: http.MethodGet, target: "/widgets/private-value",
			status: http.StatusUnprocessableEntity, kindType: oapivalidator.ProblemTypeInvalidRequest,
			location: oapivalidator.LocationPath,
		},
		{
			name: "malformed query parameter", method: http.MethodGet, target: "/widgets/widget-1?limit=not-an-integer",
			status: http.StatusBadRequest, kindType: oapivalidator.ProblemTypeMalformedRequest,
			location: oapivalidator.LocationQuery,
		},
		{
			name: "malformed json", method: http.MethodPost, target: "/widgets", body: `{"name":`, contentType: "application/json",
			status: http.StatusBadRequest, kindType: oapivalidator.ProblemTypeMalformedRequest,
			location: oapivalidator.LocationBody,
		},
		{
			name: "unsupported media type", method: http.MethodPost, target: "/widgets", body: `name=value`, contentType: "text/plain",
			status: http.StatusUnsupportedMediaType, kindType: oapivalidator.ProblemTypeUnsupportedMediaType,
			location: oapivalidator.LocationBody,
		},
		{
			name: "body schema", method: http.MethodPost, target: "/widgets", body: `{"name":"private-value","count":0}`, contentType: "application/json",
			status: http.StatusUnprocessableEntity, kindType: oapivalidator.ProblemTypeInvalidRequest,
			location: oapivalidator.LocationBody,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := serve(t, middleware, test.method, test.target, test.body, test.contentType, nil)
			problem := decodeProblem(t, recorder)
			require.Equal(t, test.status, recorder.Code)
			require.Equal(t, test.kindType, problem.Type)
			require.NotEmpty(t, problem.Errors)
			require.Equal(t, test.location, problem.Errors[0].In)
			require.NotContains(t, recorder.Body.String(), "private-value")
			require.NotContains(t, recorder.Body.String(), "Schema:")
		})
	}
}

func TestMiddlewareCapsSortedDeduplicatedErrors(t *testing.T) {
	t.Parallel()
	middleware, err := oapivalidator.New(
		loadDocument(t, publicDocument),
		oapivalidator.WithMaxReportedErrors(1),
	)
	require.NoError(t, err)

	recorder := serve(t, middleware, http.MethodPost, "/widgets", `{"name":"x","count":0}`, "application/json", nil)
	problem := decodeProblem(t, recorder)

	require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	require.Len(t, problem.Errors, 1)
	require.True(t, problem.Truncated, "%+v", problem)
}

func TestNewRejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		document *openapi3.T
		option   oapivalidator.Option
	}{
		{name: "nil document"},
		{name: "zero max errors", document: loadDocument(t, publicDocument), option: oapivalidator.WithMaxReportedErrors(0)},
		{name: "relative base URL", document: loadDocument(t, publicDocument), option: oapivalidator.WithBaseURL("api")},
		{name: "base URL query", document: loadDocument(t, publicDocument), option: oapivalidator.WithBaseURL("/api?debug=true")},
		{name: "base URL trailing slash", document: loadDocument(t, publicDocument), option: oapivalidator.WithBaseURL("/api/")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			options := []oapivalidator.Option{}
			if test.option != nil {
				options = append(options, test.option)
			}
			_, err := oapivalidator.New(test.document, options...)
			require.Error(t, err)
		})
	}
}

func TestMiddlewareDelegatesNormalizedFailure(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	handler := mocks.NewMockFailureHandler(controller)
	wantErr := errors.New("handled")
	handler.EXPECT().Handle(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ *echo.Context, failure *oapivalidator.Failure) error {
			require.Equal(t, oapivalidator.FailureNotFound, failure.Kind)
			require.Error(t, failure.Cause)
			return wantErr
		},
	)
	middleware, err := oapivalidator.New(loadDocument(t, publicDocument), oapivalidator.WithFailureHandler(handler))
	require.NoError(t, err)

	actualErr := invoke(t, middleware, http.MethodGet, "/missing", "", "", nil)

	require.ErrorIs(t, actualErr, wantErr)
}

func loadDocument(t *testing.T, source string) *openapi3.T {
	t.Helper()
	document, err := openapi3.NewLoader().LoadFromData([]byte(source))
	require.NoError(t, err)
	return document
}

func serve(t *testing.T, middleware echo.MiddlewareFunc, method, target, body, contentType string, handler echo.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	recorder, err := invokeWithRecorder(t, middleware, method, target, body, contentType, handler)
	require.NoError(t, err)
	return recorder
}

func invoke(t *testing.T, middleware echo.MiddlewareFunc, method, target, body, contentType string, handler echo.HandlerFunc) error {
	t.Helper()
	_, err := invokeWithRecorder(t, middleware, method, target, body, contentType, handler)
	return err
}

func invokeWithRecorder(t *testing.T, middleware echo.MiddlewareFunc, method, target, body, contentType string, handler echo.HandlerFunc) (*httptest.ResponseRecorder, error) {
	t.Helper()
	e := echo.New()
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	if handler == nil {
		handler = func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) }
	}
	return recorder, middleware(handler)(context)
}

func decodeProblem(t *testing.T, recorder *httptest.ResponseRecorder) oapivalidator.Problem {
	t.Helper()
	var problem oapivalidator.Problem
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &problem))
	return problem
}

const publicDocument = `
openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /widgets/{id}:
    get:
      operationId: getWidget
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            const: widget-1
        - name: limit
          in: query
          schema:
            type: integer
            minimum: 1
      responses:
        "204":
          description: Found
  /widgets:
    post:
      operationId: createWidget
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name, count]
              properties:
                name:
                  type: string
                  minLength: 3
                count:
                  type: integer
                  minimum: 1
      responses:
        "204":
          description: Created
  /nullable:
    post:
      operationId: acceptNullable
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: [string, "null"]
      responses:
        "204":
          description: Accepted
`
