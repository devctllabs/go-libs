package oapivalidatorjwt_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/oapivalidator"
	"github.com/devctllabs/go-libs/oapivalidatorjwt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

type contextKey string

func TestBearerAuthenticationMapsValidatedClaims(t *testing.T) {
	t.Parallel()
	privateKey := newKey(t)
	raw := sign(t, privateKey, jwt.MapClaims{
		"iss": "https://issuer.example",
		"aud": []string{"another-api", "orders-api"},
		"sub": "user-42",
		"exp": time.Now().Add(time.Minute).Unix(),
		"iat": time.Now().Add(-time.Minute).Unix(),
	}, "active")
	authenticator := newAuthenticator(t, privateKey, func(ctx context.Context, token *jwt.Token) (context.Context, error) {
		claims, ok := token.Claims.(jwt.MapClaims)
		require.True(t, ok)
		return context.WithValue(ctx, contextKey("subject"), claims["sub"]), nil
	})
	request := httptest.NewRequest(http.MethodGet, "https://api.example/orders", nil)
	request.Header.Set("Authorization", "Bearer "+raw)

	next, err := authenticator.Authenticate(request.Context(), oapivalidator.AuthenticationInput{
		Request:        request,
		SecurityScheme: &openapi3.SecurityScheme{Type: "http", Scheme: "bearer"},
	})

	require.NoError(t, err)
	require.Equal(t, "user-42", next.Value(contextKey("subject")))
}

func TestBearerRejectsAmbiguousOrInvalidCredentialsBeforeKeyLookup(t *testing.T) {
	t.Parallel()
	privateKey := newKey(t)
	valid := sign(t, privateKey, validClaims(), "active")
	tests := []struct {
		name    string
		headers []string
	}{
		{name: "missing"},
		{name: "duplicate", headers: []string{"Bearer " + valid, "Bearer " + valid}},
		{name: "wrong scheme", headers: []string{"Basic " + valid}},
		{name: "extra spacing", headers: []string{"Bearer  " + valid}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			lookups := 0
			authenticator, err := oapivalidatorjwt.New(baseConfig(), func(context.Context) jwt.Keyfunc {
				return func(*jwt.Token) (any, error) { lookups++; return &privateKey.PublicKey, nil }
			}, passthroughMapper)
			require.NoError(t, err)
			request := httptest.NewRequest(http.MethodGet, "https://api.example/orders", nil)
			request.Header["Authorization"] = test.headers

			_, err = authenticator.Authenticate(request.Context(), bearerInput(request))

			require.ErrorIs(t, err, oapivalidator.ErrUnauthenticated)
			require.Zero(t, lookups)
		})
	}
}

func TestJWTRejectsAlgorithmKidAndTokenLimitsBeforeKeyLookup(t *testing.T) {
	t.Parallel()
	privateKey := newKey(t)
	none := jwt.NewWithClaims(jwt.SigningMethodNone, validClaims())
	none.Header["kid"] = "active"
	unsigned, err := none.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	oversizedKid := sign(t, privateKey, validClaims(), "too-long")
	valid := sign(t, privateKey, validClaims(), "active")

	tests := []struct {
		name   string
		raw    string
		config oapivalidatorjwt.Config
	}{
		{name: "algorithm", raw: unsigned, config: baseConfig()},
		{name: "kid", raw: oversizedKid, config: func() oapivalidatorjwt.Config { c := baseConfig(); c.MaxKeyIDBytes = 3; return c }()},
		{name: "token size", raw: valid, config: func() oapivalidatorjwt.Config { c := baseConfig(); c.MaxTokenBytes = 8; return c }()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			lookups := 0
			authenticator, err := oapivalidatorjwt.New(test.config, func(context.Context) jwt.Keyfunc {
				return func(*jwt.Token) (any, error) { lookups++; return &privateKey.PublicKey, nil }
			}, passthroughMapper)
			require.NoError(t, err)
			request := httptest.NewRequest(http.MethodGet, "https://api.example/orders", nil)
			request.Header.Set("Authorization", "Bearer "+test.raw)

			_, err = authenticator.Authenticate(request.Context(), bearerInput(request))

			require.ErrorIs(t, err, oapivalidator.ErrUnauthenticated)
			require.Zero(t, lookups)
		})
	}
}

func TestJWTValidatesRequiredClaims(t *testing.T) {
	t.Parallel()
	privateKey := newKey(t)
	tests := []struct {
		name   string
		claims jwt.MapClaims
	}{
		{name: "expiration required", claims: jwt.MapClaims{"iss": "https://issuer.example", "aud": "orders-api"}},
		{name: "issuer", claims: jwt.MapClaims{"iss": "other", "aud": "orders-api", "exp": time.Now().Add(time.Minute).Unix()}},
		{name: "audience", claims: jwt.MapClaims{"iss": "https://issuer.example", "aud": "other", "exp": time.Now().Add(time.Minute).Unix()}},
		{name: "future issued at", claims: jwt.MapClaims{"iss": "https://issuer.example", "aud": "orders-api", "exp": time.Now().Add(time.Minute).Unix(), "iat": time.Now().Add(time.Minute).Unix()}},
		{name: "not before", claims: jwt.MapClaims{"iss": "https://issuer.example", "aud": "orders-api", "exp": time.Now().Add(time.Minute).Unix(), "nbf": time.Now().Add(time.Minute).Unix()}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			authenticator := newAuthenticator(t, privateKey, passthroughMapper)
			request := httptest.NewRequest(http.MethodGet, "https://api.example/orders", nil)
			request.Header.Set("Authorization", "Bearer "+sign(t, privateKey, test.claims, "active"))
			_, err := authenticator.Authenticate(request.Context(), bearerInput(request))
			require.ErrorIs(t, err, oapivalidator.ErrUnauthenticated)
		})
	}
}

func TestCookieAuthenticationRequiresCSRFOnUnsafeMethods(t *testing.T) {
	t.Parallel()
	privateKey := newKey(t)
	authenticator := newCookieAuthenticator(t, privateKey)
	raw := sign(t, privateKey, validClaims(), "active")

	request := httptest.NewRequest(http.MethodPost, "https://api.example/orders", nil)
	request.AddCookie(&http.Cookie{Name: "access", Value: raw})
	_, err := authenticator.Authenticate(request.Context(), cookieInput(request))
	require.ErrorIs(t, err, oapivalidator.ErrForbidden)

	request.Header.Set("X-CSRF-Protection", "1")
	next, err := authenticator.Authenticate(request.Context(), cookieInput(request))
	require.NoError(t, err)
	require.NotNil(t, next)
}

func TestCookieAuthenticationRejectsDuplicateCookiesAndCrossOriginRequests(t *testing.T) {
	t.Parallel()
	privateKey := newKey(t)
	authenticator := newCookieAuthenticator(t, privateKey)
	raw := sign(t, privateKey, validClaims(), "active")

	duplicate := httptest.NewRequest(http.MethodGet, "https://api.example/orders", nil)
	duplicate.Header.Add("Cookie", "access="+raw)
	duplicate.Header.Add("Cookie", "access="+raw)
	_, err := authenticator.Authenticate(duplicate.Context(), cookieInput(duplicate))
	require.ErrorIs(t, err, oapivalidator.ErrUnauthenticated)

	crossOrigin := httptest.NewRequest(http.MethodPost, "https://api.example/orders", nil)
	crossOrigin.AddCookie(&http.Cookie{Name: "access", Value: raw})
	crossOrigin.Header.Set("X-CSRF-Protection", "1")
	crossOrigin.Header.Set("Origin", "https://evil.example")
	_, err = authenticator.Authenticate(crossOrigin.Context(), cookieInput(crossOrigin))
	require.ErrorIs(t, err, oapivalidator.ErrForbidden)
}

func TestAuthenticationDependencyAndMapperErrorsKeepTheirCategory(t *testing.T) {
	t.Parallel()
	privateKey := newKey(t)
	raw := sign(t, privateKey, validClaims(), "active")
	request := httptest.NewRequest(http.MethodGet, "https://api.example/orders", nil)
	request.Header.Set("Authorization", "Bearer "+raw)

	unavailable, err := oapivalidatorjwt.New(baseConfig(), func(context.Context) jwt.Keyfunc {
		return func(*jwt.Token) (any, error) { return nil, oapivalidator.ErrAuthenticationUnavailable }
	}, passthroughMapper)
	require.NoError(t, err)
	_, err = unavailable.Authenticate(request.Context(), bearerInput(request))
	require.ErrorIs(t, err, oapivalidator.ErrAuthenticationUnavailable)

	forbidden := newAuthenticator(t, privateKey, func(context.Context, *jwt.Token) (context.Context, error) {
		return nil, oapivalidator.ErrForbidden
	})
	_, err = forbidden.Authenticate(request.Context(), bearerInput(request))
	require.ErrorIs(t, err, oapivalidator.ErrForbidden)

	unexpected := newAuthenticator(t, privateKey, func(context.Context, *jwt.Token) (context.Context, error) {
		return nil, oapivalidator.ErrUnauthenticated
	})
	_, err = unexpected.Authenticate(request.Context(), bearerInput(request))
	require.Error(t, err)
	require.NotErrorIs(t, err, oapivalidator.ErrUnauthenticated)
}

func newAuthenticator(t *testing.T, privateKey *rsa.PrivateKey, mapper oapivalidatorjwt.ClaimsMapper) *oapivalidatorjwt.Authenticator {
	t.Helper()
	authenticator, err := oapivalidatorjwt.New(baseConfig(), func(context.Context) jwt.Keyfunc {
		return func(*jwt.Token) (any, error) { return &privateKey.PublicKey, nil }
	}, mapper)
	require.NoError(t, err)
	return authenticator
}

func newCookieAuthenticator(t *testing.T, privateKey *rsa.PrivateKey) *oapivalidatorjwt.Authenticator {
	t.Helper()
	config := baseConfig()
	config.CookieProtection = &oapivalidatorjwt.CookieProtectionConfig{TrustedOrigins: []string{"https://ui.example"}}
	authenticator, err := oapivalidatorjwt.New(config, func(context.Context) jwt.Keyfunc {
		return func(*jwt.Token) (any, error) { return &privateKey.PublicKey, nil }
	}, passthroughMapper)
	require.NoError(t, err)
	return authenticator
}

func newKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func sign(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims, kid string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	raw, err := token.SignedString(key)
	require.NoError(t, err)
	return raw
}

func validClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"iss": "https://issuer.example",
		"aud": "orders-api",
		"exp": time.Now().Add(time.Minute).Unix(),
		"iat": time.Now().Add(-time.Minute).Unix(),
	}
}

func baseConfig() oapivalidatorjwt.Config {
	return oapivalidatorjwt.Config{
		Issuer:            "https://issuer.example",
		Audiences:         []string{"orders-api"},
		AllowedAlgorithms: []string{"RS256"},
	}
}

func bearerInput(request *http.Request) oapivalidator.AuthenticationInput {
	return oapivalidator.AuthenticationInput{Request: request, SecurityScheme: &openapi3.SecurityScheme{Type: "http", Scheme: "bearer"}}
}

func cookieInput(request *http.Request) oapivalidator.AuthenticationInput {
	return oapivalidator.AuthenticationInput{Request: request, SecurityScheme: &openapi3.SecurityScheme{Type: "apiKey", In: "cookie", Name: "access"}}
}

func passthroughMapper(ctx context.Context, _ *jwt.Token) (context.Context, error) { return ctx, nil }
