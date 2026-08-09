package oapivalidator_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/devctllabs/go-libs/oapivalidator"
	"github.com/devctllabs/go-libs/oapivalidator/internal/testapi"
	"github.com/devctllabs/go-libs/oapivalidator/mocks"
)

func TestGeneratedEcho5StrictServerReceivesAuthenticatedContext(t *testing.T) {
	t.Parallel()
	document, err := testapi.GetSpec()
	require.NoError(t, err)

	controller := gomock.NewController(t)
	authenticator := mocks.NewMockAuthenticator(controller)
	authenticator.EXPECT().Authenticate(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, input oapivalidator.AuthenticationInput) (context.Context, error) {
			require.Equal(t, "GetFixture", input.OperationID)
			return context.WithValue(ctx, contextKey("subject"), "fixture-user"), nil
		},
	)
	middleware, err := oapivalidator.New(
		document,
		oapivalidator.WithAuthenticator(authenticator),
		oapivalidator.WithBaseURL("/api"),
	)
	require.NoError(t, err)

	seenContext := make(chan context.Context, 1)
	server := fixtureServer{seenContext: seenContext}
	e := echo.New()
	e.Use(middleware)
	testapi.RegisterHandlersWithOptions(e, testapi.NewStrictHandler(server, nil), testapi.RegisterHandlersOptions{BaseURL: "/api"})

	request := httptest.NewRequest(http.MethodGet, "/api/fixture/fixture-1", nil)
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	strictContext := <-seenContext
	require.Equal(t, "fixture-user", strictContext.Value(contextKey("subject")))
}

type fixtureServer struct {
	seenContext chan<- context.Context
}

func (server fixtureServer) GetFixture(ctx context.Context, _ testapi.GetFixtureRequestObject) (testapi.GetFixtureResponseObject, error) {
	server.seenContext <- ctx
	return testapi.GetFixture204Response{}, nil
}
