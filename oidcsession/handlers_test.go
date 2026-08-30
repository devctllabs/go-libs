package oidcsession_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/oidcsession"
	"github.com/devctllabs/go-libs/oidcsession/mocks"
	"github.com/devctllabs/go-libs/retry"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLoginBuildsOIDCAuthorizationRequestWithPKCENonceAndBrokerHint(t *testing.T) {
	t.Parallel()
	provider, stop := readyProvider(t)
	defer stop()
	controller := gomock.NewController(t)
	sessions := mocks.NewMockSessionBackend(controller)
	uiURL, err := url.Parse("https://ui.example")
	require.NoError(t, err)
	handlers, err := oidcsession.NewHandlers(oidcsession.HTTPConfig{
		UIBaseURL:         uiURL,
		DefaultReturnPath: "/",
		ErrorPath:         "/auth/error",
		CookiePrefix:      "example",
		LoginMethods: []oidcsession.LoginMethod{{
			ID:                      "google",
			AuthorizationParameters: url.Values{"provider_hint": {"google"}},
		}},
	}, provider, sessions, oidcsession.InsecureNoopEncryptor())
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "https://api.example/auth/login?method=google&return=/onboarding", nil)
	recorder := httptest.NewRecorder()

	handlers.Login(recorder, request)

	require.Equal(t, http.StatusFound, recorder.Code)
	location, err := url.Parse(recorder.Header().Get("Location"))
	require.NoError(t, err)
	require.Equal(t, "S256", location.Query().Get("code_challenge_method"))
	require.NotEmpty(t, location.Query().Get("code_challenge"))
	require.NotEmpty(t, location.Query().Get("nonce"))
	require.NotEmpty(t, location.Query().Get("state"))
	require.Equal(t, "google", location.Query().Get("provider_hint"))
	cookies := recorder.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "__Host-example-login", cookies[0].Name)
	require.True(t, cookies[0].Secure)
	require.True(t, cookies[0].HttpOnly)
}

func TestLoginRejectsUnknownMethod(t *testing.T) {
	t.Parallel()
	provider, stop := readyProvider(t)
	defer stop()
	controller := gomock.NewController(t)
	handlers := newTestHandlers(t, provider, mocks.NewMockSessionBackend(controller))
	request := httptest.NewRequest(http.MethodGet, "https://api.example/auth/login?method=unknown", nil)
	recorder := httptest.NewRecorder()

	handlers.Login(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestRefreshRequiresCrossOriginAndConfirmationChecks(t *testing.T) {
	t.Parallel()
	provider, stop := readyProvider(t)
	defer stop()
	controller := gomock.NewController(t)
	handlers := newTestHandlers(t, provider, mocks.NewMockSessionBackend(controller))
	request := httptest.NewRequest(http.MethodPost, "https://api.example/auth/refresh", nil)
	request.AddCookie(&http.Cookie{Name: "__Host-example-session", Value: "credential"})
	recorder := httptest.NewRecorder()

	handlers.Refresh(recorder, request)

	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestSessionWithoutCredentialIsAnonymous(t *testing.T) {
	t.Parallel()
	provider, stop := readyProvider(t)
	defer stop()
	controller := gomock.NewController(t)
	handlers := newTestHandlers(t, provider, mocks.NewMockSessionBackend(controller))
	request := httptest.NewRequest(http.MethodGet, "https://api.example/auth/session", nil)
	recorder := httptest.NewRecorder()

	handlers.Session(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"authenticated":false}`, recorder.Body.String())
}

func newTestHandlers(t *testing.T, provider *oidcsession.Provider, sessions oidcsession.SessionBackend) *oidcsession.Handlers {
	t.Helper()
	uiURL, err := url.Parse("https://ui.example")
	require.NoError(t, err)
	handlers, err := oidcsession.NewHandlers(oidcsession.HTTPConfig{
		UIBaseURL:         uiURL,
		DefaultReturnPath: "/",
		ErrorPath:         "/auth/error",
		CookiePrefix:      "example",
		TrustedOrigins:    []string{"https://ui.example"},
	}, provider, sessions, oidcsession.InsecureNoopEncryptor())
	require.NoError(t, err)
	return handlers
}

func readyProvider(t *testing.T) (*oidcsession.Provider, func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issuer := "http://" + r.Host
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                 issuer,
			"authorization_endpoint": issuer + "/authorize",
			"token_endpoint":         issuer + "/token",
			"jwks_uri":               issuer + "/keys",
		})
	}))
	policy, err := retry.NewExponential(retry.ExponentialConfig{InitialDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond, Multiplier: 2})
	require.NoError(t, err)
	provider, err := oidcsession.NewProvider(oidcsession.ProviderConfig{
		IssuerURL:      server.URL,
		ClientID:       "client",
		ClientSecret:   "secret",
		RedirectURL:    "https://api.example/auth/callback",
		Scopes:         []string{"openid"},
		HTTPClient:     &http.Client{Timeout: time.Second},
		DiscoveryRetry: policy,
	})
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- provider.Run(ctx) }()
	require.Eventually(t, func() bool { return provider.Check(context.Background()) == nil }, time.Second, time.Millisecond)
	return provider, func() {
		cancel()
		require.NoError(t, receiveWithin(t, done, time.Second))
		server.Close()
	}
}
