package oidcsession

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/devctllabs/go-libs/retry"
	"golang.org/x/oauth2"
)

// ProviderConfig configures OIDC discovery and OAuth operations.
type ProviderConfig struct {
	IssuerURL      string
	ClientID       string
	ClientSecret   string
	RedirectURL    string
	Scopes         []string
	HTTPClient     *http.Client
	DiscoveryRetry retry.Policy
	Observer       Observer
}

// ProviderTokens is the provider token set persisted by a session backend.
type ProviderTokens struct {
	AccessToken     string
	RefreshToken    string
	AccessExpiresAt time.Time
}

// TokenService refreshes and revokes provider token sets for a session backend.
type TokenService interface {
	// Refresh exchanges refreshToken and returns a complete current token set.
	Refresh(ctx context.Context, refreshToken string) (tokens ProviderTokens, err error)
	// Revoke makes a best-effort provider revocation request for refreshToken.
	Revoke(ctx context.Context, refreshToken string) error
}

type providerRuntime struct {
	provider      *oidc.Provider
	oauth         oauth2.Config
	revocationURL string
}

// Provider owns discovered OIDC metadata and provider token operations.
type Provider struct {
	config  ProviderConfig
	mu      sync.RWMutex
	runtime *providerRuntime
	running bool
}

// NewProvider validates local configuration without contacting the issuer.
func NewProvider(config ProviderConfig) (*Provider, error) {
	if strings.TrimSpace(config.IssuerURL) == "" || strings.TrimSpace(config.ClientID) == "" || strings.TrimSpace(config.ClientSecret) == "" || strings.TrimSpace(config.RedirectURL) == "" {
		return nil, errors.New("issuer URL, client ID, client secret, and redirect URL are required")
	}
	issuer, err := url.Parse(config.IssuerURL)
	if err != nil || issuer.Scheme == "" || issuer.Host == "" {
		return nil, errors.New("issuer URL must be absolute")
	}
	redirect, err := url.Parse(config.RedirectURL)
	if err != nil || redirect.Scheme == "" || redirect.Host == "" {
		return nil, errors.New("redirect URL must be absolute")
	}
	if !slices.Contains(config.Scopes, oidc.ScopeOpenID) {
		return nil, errors.New("OIDC scopes must contain openid")
	}
	if config.HTTPClient == nil || config.HTTPClient.Timeout <= 0 {
		return nil, errors.New("a bounded HTTP client with a positive timeout is required")
	}
	if config.DiscoveryRetry == nil {
		return nil, errors.New("discovery retry policy is required")
	}
	config.Scopes = slices.Clone(config.Scopes)
	return &Provider{config: config}, nil
}

// Run discovers provider metadata with retries and then remains alive until ctx is canceled.
// It is a long-lived lifecycle task: returning after discovery would signal an unexpected stop to
// lifecycle.Run. Signing-key rotation does not require repeated discovery because go-oidc refreshes
// its remote JWKS cache when it encounters an unknown key ID.
func (provider *Provider) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	provider.mu.Lock()
	if provider.running {
		provider.mu.Unlock()
		return errors.New("provider is already running")
	}
	provider.running = true
	provider.mu.Unlock()
	defer func() {
		provider.mu.Lock()
		provider.running = false
		provider.mu.Unlock()
	}()

	err := provider.discoverUntilReady(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		return fmt.Errorf("discover OIDC provider: %w", err)
	}
	<-ctx.Done()
	return nil
}

func (provider *Provider) discoverUntilReady(ctx context.Context) error {
	var duration time.Duration
	return retry.Do(ctx, provider.config.DiscoveryRetry, func(attemptCtx context.Context) error {
		started := time.Now()
		err := provider.discover(attemptCtx)
		duration = time.Since(started)
		return err
	}, retry.WithNotify(func(_ uint, err error, _ time.Duration) {
		observe(ctx, provider.config.Observer, Observation{Operation: OperationDiscovery, Err: err, Duration: duration, Retry: true})
	}))
}

// Check reports local discovery readiness without network I/O.
func (provider *Provider) Check(ctx context.Context) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	provider.mu.RLock()
	ready := provider.runtime != nil
	provider.mu.RUnlock()
	if !ready {
		return ErrProviderUnavailable
	}
	return nil
}

func (provider *Provider) discover(ctx context.Context) error {
	ctx = oidc.ClientContext(ctx, provider.config.HTTPClient)
	discovered, err := oidc.NewProvider(ctx, provider.config.IssuerURL)
	if err != nil {
		return fmt.Errorf("OIDC discovery failed: %w", err)
	}
	var metadata struct {
		RevocationEndpoint string `json:"revocation_endpoint"`
	}
	if err := discovered.Claims(&metadata); err != nil {
		return fmt.Errorf("decode OIDC provider metadata: %w", err)
	}
	runtime := &providerRuntime{
		provider: discovered,
		oauth: oauth2.Config{
			ClientID:     provider.config.ClientID,
			ClientSecret: provider.config.ClientSecret,
			RedirectURL:  provider.config.RedirectURL,
			Endpoint:     discovered.Endpoint(),
			Scopes:       slices.Clone(provider.config.Scopes),
		},
		revocationURL: metadata.RevocationEndpoint,
	}
	provider.mu.Lock()
	provider.runtime = runtime
	provider.mu.Unlock()
	return nil
}

func (provider *Provider) snapshot() (*providerRuntime, error) {
	provider.mu.RLock()
	runtime := provider.runtime
	provider.mu.RUnlock()
	if runtime == nil {
		return nil, ErrProviderUnavailable
	}
	return runtime, nil
}

// Refresh implements TokenService.
func (provider *Provider) Refresh(ctx context.Context, refreshToken string) (ProviderTokens, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return ProviderTokens{}, ErrInvalidGrant
	}
	runtime, err := provider.snapshot()
	if err != nil {
		return ProviderTokens{}, err
	}
	ctx = oidc.ClientContext(ctx, provider.config.HTTPClient)
	token, err := runtime.oauth.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}).Token()
	if err != nil {
		var retrieveError *oauth2.RetrieveError
		if errors.As(err, &retrieveError) && retrieveError.ErrorCode == "invalid_grant" {
			return ProviderTokens{}, ErrInvalidGrant
		}
		return ProviderTokens{}, fmt.Errorf("refresh provider token: %w", ErrProviderUnavailable)
	}
	if token.AccessToken == "" || token.Expiry.IsZero() || !token.Expiry.After(time.Now()) {
		return ProviderTokens{}, fmt.Errorf("provider returned an incomplete token set: %w", ErrProviderUnavailable)
	}
	nextRefreshToken := token.RefreshToken
	if nextRefreshToken == "" {
		nextRefreshToken = refreshToken
	}
	return ProviderTokens{AccessToken: token.AccessToken, RefreshToken: nextRefreshToken, AccessExpiresAt: token.Expiry}, nil
}

// Revoke implements TokenService. Providers without a revocation endpoint are treated as unsupported.
func (provider *Provider) Revoke(ctx context.Context, refreshToken string) error {
	runtime, err := provider.snapshot()
	if err != nil {
		return err
	}
	if runtime.revocationURL == "" || refreshToken == "" {
		return nil
	}
	values := url.Values{"token": {refreshToken}, "token_type_hint": {"refresh_token"}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, runtime.revocationURL, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("build revocation request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(provider.config.ClientID, provider.config.ClientSecret)
	response, err := provider.config.HTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("revoke provider token: %w", ErrProviderUnavailable)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("revoke provider token: %w", ErrProviderUnavailable)
	}
	return nil
}

var _ TokenService = (*Provider)(nil)
