package oapivalidatorjwt

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/devctllabs/go-libs/oapivalidator"
	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultMaxTokenBytes = 16 * 1024
	defaultMaxKeyIDBytes = 256
)

// Config defines the accepted issuer, audiences, signing algorithms, and input limits.
type Config struct {
	Issuer            string
	Audiences         []string
	AllowedAlgorithms []string
	Leeway            time.Duration
	MaxTokenBytes     int
	MaxKeyIDBytes     int
	CookieProtection  *CookieProtectionConfig
}

// CookieProtectionConfig configures CSRF checks for JWTs transported in cookies.
type CookieProtectionConfig struct {
	TrustedOrigins []string
	CSRFHeaderName string
}

// KeyfuncProvider returns the signing-key resolver for the current request context.
type KeyfuncProvider func(ctx context.Context) jwt.Keyfunc

// ClaimsMapper converts a completely validated JWT into application request context.
type ClaimsMapper func(ctx context.Context, token *jwt.Token) (nextCtx context.Context, err error)

// Authenticator validates JWT credentials for supported OpenAPI security schemes.
type Authenticator struct {
	issuer           string
	audiences        []string
	algorithms       []string
	algorithmSet     map[string]struct{}
	leeway           time.Duration
	maxTokenBytes    int
	maxKeyIDBytes    int
	keyfuncProvider  KeyfuncProvider
	claimsMapper     ClaimsMapper
	cookieProtection *cookieProtection
}

// New constructs an Authenticator without performing network I/O.
func New(config Config, keyFuncProvider KeyfuncProvider, claimsMapper ClaimsMapper) (*Authenticator, error) {
	if strings.TrimSpace(config.Issuer) == "" {
		return nil, errors.New("issuer is required")
	}
	if len(config.Audiences) == 0 || slices.ContainsFunc(config.Audiences, func(value string) bool { return strings.TrimSpace(value) == "" }) {
		return nil, errors.New("at least one non-empty audience is required")
	}
	if len(config.AllowedAlgorithms) == 0 || slices.ContainsFunc(config.AllowedAlgorithms, func(value string) bool { return strings.TrimSpace(value) == "" }) {
		return nil, errors.New("at least one non-empty allowed algorithm is required")
	}
	if config.Leeway < 0 {
		return nil, errors.New("leeway must not be negative")
	}
	if keyFuncProvider == nil {
		return nil, errors.New("keyfunc provider is required")
	}
	if claimsMapper == nil {
		return nil, errors.New("claims mapper is required")
	}
	if config.MaxTokenBytes < 0 || config.MaxKeyIDBytes < 0 {
		return nil, errors.New("JWT input limits must not be negative")
	}
	maxTokenBytes := config.MaxTokenBytes
	if maxTokenBytes == 0 {
		maxTokenBytes = defaultMaxTokenBytes
	}
	maxKeyIDBytes := config.MaxKeyIDBytes
	if maxKeyIDBytes == 0 {
		maxKeyIDBytes = defaultMaxKeyIDBytes
	}
	protection, err := newCookieProtection(config.CookieProtection)
	if err != nil {
		return nil, fmt.Errorf("configure cookie protection: %w", err)
	}
	algorithmSet := make(map[string]struct{}, len(config.AllowedAlgorithms))
	for _, algorithm := range config.AllowedAlgorithms {
		if _, duplicate := algorithmSet[algorithm]; duplicate {
			return nil, fmt.Errorf("allowed algorithm %q is duplicated", algorithm)
		}
		algorithmSet[algorithm] = struct{}{}
	}
	return &Authenticator{
		issuer:           config.Issuer,
		audiences:        slices.Clone(config.Audiences),
		algorithms:       slices.Clone(config.AllowedAlgorithms),
		algorithmSet:     algorithmSet,
		leeway:           config.Leeway,
		maxTokenBytes:    maxTokenBytes,
		maxKeyIDBytes:    maxKeyIDBytes,
		keyfuncProvider:  keyFuncProvider,
		claimsMapper:     claimsMapper,
		cookieProtection: protection,
	}, nil
}

// Authenticate implements oapivalidator.Authenticator.
func (authenticator *Authenticator) Authenticate(ctx context.Context, input oapivalidator.AuthenticationInput) (context.Context, error) {
	if ctx == nil {
		return nil, errors.New("authentication context is required")
	}
	if input.Request == nil || input.SecurityScheme == nil {
		return nil, errors.New("OpenAPI authentication input is incomplete")
	}
	raw, cookieCredential, err := credential(input.Request, input.SecurityScheme.Type, input.SecurityScheme.Scheme, input.SecurityScheme.In, input.SecurityScheme.Name)
	if err != nil {
		return nil, err
	}
	if cookieCredential {
		if authenticator.cookieProtection == nil {
			return nil, errors.New("cookie security scheme requires cookie protection configuration")
		}
		if err := authenticator.cookieProtection.check(input.Request); err != nil {
			return nil, fmt.Errorf("cookie request protection: %w", oapivalidator.ErrForbidden)
		}
	}
	if len(raw) > authenticator.maxTokenBytes {
		return nil, oapivalidator.ErrUnauthenticated
	}
	if err := authenticator.preflight(raw); err != nil {
		return nil, oapivalidator.ErrUnauthenticated
	}

	keyFunc := authenticator.keyfuncProvider(ctx)
	if keyFunc == nil {
		return nil, errors.New("keyfunc provider returned nil")
	}
	parser := jwt.NewParser(
		jwt.WithValidMethods(authenticator.algorithms),
		jwt.WithIssuer(authenticator.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(authenticator.leeway),
	)
	claims := jwt.MapClaims{}
	token, err := parser.ParseWithClaims(raw, claims, keyFunc)
	if err != nil {
		if errors.Is(err, oapivalidator.ErrAuthenticationUnavailable) {
			return nil, fmt.Errorf("resolve JWT signing key: %w", oapivalidator.ErrAuthenticationUnavailable)
		}
		return nil, oapivalidator.ErrUnauthenticated
	}
	if !token.Valid || !authenticator.acceptsAudience(claims) {
		return nil, oapivalidator.ErrUnauthenticated
	}
	nextCtx, err := authenticator.claimsMapper(ctx, token)
	if err != nil {
		if errors.Is(err, oapivalidator.ErrForbidden) || errors.Is(err, oapivalidator.ErrAuthenticationUnavailable) {
			return nil, err
		}
		return nil, errors.New("JWT claims mapping failed")
	}
	if nextCtx == nil {
		return nil, errors.New("claims mapper returned a nil context without an error")
	}
	return nextCtx, nil
}

func credential(request *http.Request, schemeType, scheme, in, name string) (raw string, cookie bool, err error) {
	switch {
	case strings.EqualFold(schemeType, "http") && strings.EqualFold(scheme, "bearer"):
		values := request.Header.Values("Authorization")
		if len(values) != 1 {
			return "", false, oapivalidator.ErrUnauthenticated
		}
		prefix, raw, found := strings.Cut(values[0], " ")
		if !found || !strings.EqualFold(prefix, "Bearer") || raw == "" || strings.ContainsAny(raw, " \t\r\n") {
			return "", false, oapivalidator.ErrUnauthenticated
		}
		return raw, false, nil
	case strings.EqualFold(schemeType, "apiKey") && strings.EqualFold(in, "cookie"):
		if name == "" {
			return "", false, errors.New("OpenAPI cookie security scheme has no cookie name")
		}
		cookies := request.CookiesNamed(name)
		if len(cookies) != 1 || cookies[0].Value == "" {
			return "", true, oapivalidator.ErrUnauthenticated
		}
		return cookies[0].Value, true, nil
	default:
		return "", false, errors.New("unsupported OpenAPI security scheme for JWT authentication")
	}
}

func (authenticator *Authenticator) preflight(raw string) error {
	token, _, err := jwt.NewParser().ParseUnverified(raw, jwt.MapClaims{})
	if err != nil {
		return err
	}
	if _, allowed := authenticator.algorithmSet[token.Method.Alg()]; !allowed {
		return errors.New("JWT signing algorithm is not allowed")
	}
	kid, ok := token.Header["kid"].(string)
	if !ok || kid == "" || len(kid) > authenticator.maxKeyIDBytes {
		return errors.New("JWT key identifier is missing or invalid")
	}
	return nil
}

func (authenticator *Authenticator) acceptsAudience(claims jwt.MapClaims) bool {
	tokenAudiences, err := claims.GetAudience()
	if err != nil || len(tokenAudiences) == 0 {
		return false
	}
	for _, expected := range authenticator.audiences {
		if slices.Contains(tokenAudiences, expected) {
			return true
		}
	}
	return false
}

var _ oapivalidator.Authenticator = (*Authenticator)(nil)
