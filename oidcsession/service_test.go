package oidcsession

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSessionServiceRefreshRevokesInvalidSession(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	provider := NewMockloginProvider(controller)
	sessions := NewMockSessionBackend(controller)
	encryptor := NewMockEncryptor(controller)
	credential := SessionCredential("credential")
	sessions.EXPECT().Refresh(gomock.Any(), credential).Return(RefreshSessionResult{}, ErrInvalidGrant)
	sessions.EXPECT().Revoke(gomock.Any(), credential).Return(nil)
	service := newSessionService(sessionServiceConfig{}, provider, sessions, encryptor)

	_, err := service.Refresh(context.Background(), credential)

	require.ErrorIs(t, err, ErrInvalidGrant)
}

func TestSessionServiceRejectsUnknownLoginMethodBeforeCreatingState(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	service := newSessionService(sessionServiceConfig{methods: map[string]url.Values{"": nil}},
		NewMockloginProvider(controller), NewMockSessionBackend(controller), NewMockEncryptor(controller))

	_, err := service.BeginLogin(context.Background(), beginLoginCommand{methodID: "missing"})

	require.ErrorIs(t, err, errUnknownLoginMethod)
}

func TestSessionServiceRejectsStateFromAnotherBrowser(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	provider := NewMockloginProvider(controller)
	var state string
	provider.EXPECT().authorizationURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(encodedState string, _ string, _ string, _ url.Values) (string, error) {
			state = encodedState
			return "https://issuer.example/authorize", nil
		},
	)
	service := newSessionService(sessionServiceConfig{
		defaultReturnPath: "/", stateTTL: time.Minute, methods: map[string]url.Values{"": nil},
	}, provider, NewMockSessionBackend(controller), InsecureNoopEncryptor())

	started, err := service.BeginLogin(context.Background(), beginLoginCommand{browserBinding: "browser-a"})
	require.NoError(t, err)
	require.Nil(t, started.bindingToStore)
	require.NotEmpty(t, state)

	_, err = service.CompleteLogin(context.Background(), completeLoginCommand{
		code: "authorization-code", state: state, browserBinding: "browser-b",
	})

	require.Error(t, err)
}

func TestSessionServiceObservesSessionBackendFailure(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	sessions := NewMockSessionBackend(controller)
	backendErr := errors.New("backend unavailable")
	sessions.EXPECT().Status(gomock.Any(), SessionCredential("credential")).Return(SessionStatus{}, backendErr)
	var observation Observation
	service := newSessionService(sessionServiceConfig{observer: ObserverFunc(func(_ context.Context, observed Observation) {
		observation = observed
	})}, NewMockloginProvider(controller), sessions, NewMockEncryptor(controller))

	_, err := service.Session(context.Background(), SessionCredential("credential"))

	require.ErrorIs(t, err, backendErr)
	require.Equal(t, OperationSession, observation.Operation)
	require.ErrorIs(t, observation.Err, backendErr)
}
