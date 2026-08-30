package oidcsession_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/oidcsession"
	"github.com/devctllabs/go-libs/retry"
	"github.com/stretchr/testify/require"
)

func TestProviderStartsDegradedAndRecoversThroughRun(t *testing.T) {
	t.Parallel()
	var requests atomic.Int64
	available := atomic.Bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if !available.Load() {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		issuer := "http://" + r.Host
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                 issuer,
			"authorization_endpoint": issuer + "/authorize",
			"token_endpoint":         issuer + "/token",
			"jwks_uri":               issuer + "/keys",
		})
	}))
	defer server.Close()
	policy, err := retry.NewExponential(retry.ExponentialConfig{InitialDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond, Multiplier: 2})
	require.NoError(t, err)
	observations := make(chan oidcsession.Observation, 8)
	provider, err := oidcsession.NewProvider(oidcsession.ProviderConfig{
		IssuerURL:      server.URL,
		ClientID:       "client",
		ClientSecret:   "secret",
		RedirectURL:    "https://api.example/auth/callback",
		Scopes:         []string{"openid"},
		HTTPClient:     &http.Client{Timeout: time.Second},
		DiscoveryRetry: policy,
		Observer:       oidcsession.ObserverFunc(func(_ context.Context, observation oidcsession.Observation) { observations <- observation }),
	})
	require.NoError(t, err)
	require.ErrorIs(t, provider.Check(context.Background()), oidcsession.ErrProviderUnavailable)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- provider.Run(ctx) }()
	require.Eventually(t, func() bool { return requests.Load() >= 2 }, time.Second, time.Millisecond)
	available.Store(true)
	require.Eventually(t, func() bool { return provider.Check(context.Background()) == nil }, time.Second, time.Millisecond)
	readyRequests := requests.Load()
	select {
	case runErr := <-done:
		require.Fail(t, "provider Run returned before cancellation", "error: %v", runErr)
	case <-time.After(20 * time.Millisecond):
	}
	require.Equal(t, readyRequests, requests.Load(), "discovery must not repeat after readiness")
	cancel()
	require.NoError(t, receiveWithin(t, done, time.Second))

	select {
	case observation := <-observations:
		require.Equal(t, oidcsession.OperationDiscovery, observation.Operation)
		require.True(t, observation.Retry)
		require.Error(t, observation.Err)
	case <-time.After(time.Second):
		require.Fail(t, "expected discovery retry observation")
	}
}

func TestNewProviderPerformsOnlyLocalValidation(t *testing.T) {
	t.Parallel()
	policy, err := retry.NewExponential(retry.ExponentialConfig{InitialDelay: time.Millisecond, MaxDelay: time.Second, Multiplier: 2})
	require.NoError(t, err)
	provider, err := oidcsession.NewProvider(oidcsession.ProviderConfig{
		IssuerURL:      "https://unreachable.invalid",
		ClientID:       "client",
		ClientSecret:   "secret",
		RedirectURL:    "https://api.example/auth/callback",
		Scopes:         []string{"openid"},
		HTTPClient:     &http.Client{Timeout: time.Second},
		DiscoveryRetry: policy,
	})
	require.NoError(t, err)
	require.NotNil(t, provider)
}
