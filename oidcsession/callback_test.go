package oidcsession_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/oidcsession"
	"github.com/devctllabs/go-libs/oidcsession/mocks"
	"github.com/devctllabs/go-libs/retry"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/oauth2"
)

func TestCallbackVerifiesOIDCTokensAndCreatesSession(t *testing.T) {
	t.Parallel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	var issuer string
	var expectedNonce string
	var expectedChallenge string
	providerErrors := make(chan error, 3)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"issuer": issuer, "authorization_endpoint": issuer + "/authorize", "token_endpoint": issuer + "/token", "jwks_uri": issuer + "/keys",
			})
		case "/keys":
			_ = json.NewEncoder(writer).Encode(map[string]any{"keys": []any{rsaJWK(&key.PublicKey)}})
		case "/token":
			if err := request.ParseForm(); err != nil {
				providerErrors <- err
				http.Error(writer, "invalid form", http.StatusBadRequest)
				return
			}
			if actual := oauth2.S256ChallengeFromVerifier(request.Form.Get("code_verifier")); actual != expectedChallenge {
				providerErrors <- fmt.Errorf("PKCE challenge = %q, want %q", actual, expectedChallenge)
				http.Error(writer, "invalid verifier", http.StatusBadRequest)
				return
			}
			accessToken := "verified-access"
			digest := sha256.Sum256([]byte(accessToken))
			idToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
				"iss": issuer, "aud": "client", "sub": "user-42", "exp": time.Now().Add(time.Minute).Unix(),
				"iat": time.Now().Unix(), "nonce": expectedNonce, "at_hash": base64.RawURLEncoding.EncodeToString(digest[:len(digest)/2]),
			})
			idToken.Header["kid"] = "test-key"
			rawIDToken, signErr := idToken.SignedString(key)
			if signErr != nil {
				providerErrors <- signErr
				http.Error(writer, "signing failed", http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"access_token": accessToken, "refresh_token": "verified-refresh", "token_type": "Bearer", "expires_in": 300, "id_token": rawIDToken,
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()
	issuer = server.URL
	policy, err := retry.NewExponential(retry.ExponentialConfig{InitialDelay: time.Millisecond, MaxDelay: time.Second, Multiplier: 2})
	require.NoError(t, err)
	provider, err := oidcsession.NewProvider(oidcsession.ProviderConfig{
		IssuerURL: issuer, ClientID: "client", ClientSecret: "secret", RedirectURL: "https://api.example/auth/callback",
		Scopes: []string{"openid"}, HTTPClient: &http.Client{Timeout: time.Second}, DiscoveryRetry: policy,
	})
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- provider.Run(ctx) }()
	require.Eventually(t, func() bool { return provider.Check(context.Background()) == nil }, time.Second, time.Millisecond)
	defer func() { cancel(); require.NoError(t, receiveWithin(t, done, time.Second)) }()

	controller := gomock.NewController(t)
	sessions := mocks.NewMockSessionBackend(controller)
	sessionExpiry := time.Now().Add(time.Hour)
	sessions.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, params oidcsession.CreateSessionParams) (oidcsession.CreateSessionResult, error) {
		require.Equal(t, "verified-access", params.AccessToken)
		require.Equal(t, "verified-refresh", params.RefreshToken)
		return oidcsession.CreateSessionResult{Credential: "opaque-session", SessionExpiresAt: sessionExpiry}, nil
	})
	uiURL, err := url.Parse("https://ui.example")
	require.NoError(t, err)
	handlers, err := oidcsession.NewHandlers(oidcsession.HTTPConfig{
		UIBaseURL: uiURL, DefaultReturnPath: "/", ErrorPath: "/auth/error", CookiePrefix: "example",
	}, provider, sessions, oidcsession.InsecureNoopEncryptor())
	require.NoError(t, err)

	loginRequest := httptest.NewRequest(http.MethodGet, "https://api.example/auth/login?return=/onboarding", nil)
	loginRecorder := httptest.NewRecorder()
	handlers.Login(loginRecorder, loginRequest)
	require.Equal(t, http.StatusFound, loginRecorder.Code)
	authorizationURL, err := url.Parse(loginRecorder.Header().Get("Location"))
	require.NoError(t, err)
	expectedNonce = authorizationURL.Query().Get("nonce")
	expectedChallenge = authorizationURL.Query().Get("code_challenge")
	require.NotEmpty(t, expectedNonce)
	require.NotEmpty(t, expectedChallenge)

	callbackRequest := httptest.NewRequest(http.MethodGet, "https://api.example/auth/callback?code=one-time-code&state="+url.QueryEscape(authorizationURL.Query().Get("state")), nil)
	callbackRequest.AddCookie(loginRecorder.Result().Cookies()[0])
	callbackRecorder := httptest.NewRecorder()
	handlers.Callback(callbackRecorder, callbackRequest)
	close(providerErrors)
	for providerErr := range providerErrors {
		require.NoError(t, providerErr)
	}

	require.Equal(t, http.StatusSeeOther, callbackRecorder.Code)
	require.Equal(t, "https://ui.example/onboarding", callbackRecorder.Header().Get("Location"))
	cookies := callbackRecorder.Result().Cookies()
	require.Len(t, cookies, 2)
	require.Equal(t, "__Host-example-access", cookies[0].Name)
	require.Equal(t, "verified-access", cookies[0].Value)
	require.Equal(t, "__Host-example-session", cookies[1].Name)
	require.Equal(t, "opaque-session", cookies[1].Value)
}

func rsaJWK(key *rsa.PublicKey) map[string]any {
	exponent := big.NewInt(int64(key.E)).Bytes()
	return map[string]any{
		"kty": "RSA", "use": "sig", "alg": "RS256", "kid": "test-key",
		"n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()), "e": base64.RawURLEncoding.EncodeToString(exponent),
	}
}
