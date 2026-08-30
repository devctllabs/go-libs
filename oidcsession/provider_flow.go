package oidcsession

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func (provider *Provider) authorizationURL(state, nonce, verifier string, parameters url.Values) (string, error) {
	runtime, err := provider.snapshot()
	if err != nil {
		return "", err
	}
	options := []oauth2.AuthCodeOption{oauth2.S256ChallengeOption(verifier), oidc.Nonce(nonce), oauth2.AccessTypeOffline}
	for key, values := range parameters {
		if len(values) == 1 {
			options = append(options, oauth2.SetAuthURLParam(key, values[0]))
		}
	}
	return runtime.oauth.AuthCodeURL(state, options...), nil
}

func (provider *Provider) exchange(ctx context.Context, code, nonce, verifier, methodID string) (ProviderTokens, VerifiedIdentity, error) {
	runtime, err := provider.snapshot()
	if err != nil {
		return ProviderTokens{}, VerifiedIdentity{}, err
	}
	ctx = oidc.ClientContext(ctx, provider.config.HTTPClient)
	token, err := runtime.oauth.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		if providerNetworkError(err) {
			return ProviderTokens{}, VerifiedIdentity{}, fmt.Errorf("exchange authorization code: %w", ErrProviderUnavailable)
		}
		return ProviderTokens{}, VerifiedIdentity{}, errors.New("authorization code exchange failed")
	}
	if token.AccessToken == "" || token.RefreshToken == "" || token.Expiry.IsZero() || !token.Expiry.After(time.Now()) {
		return ProviderTokens{}, VerifiedIdentity{}, errors.New("provider returned an incomplete token set")
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return ProviderTokens{}, VerifiedIdentity{}, errors.New("provider did not return an ID token")
	}
	verified, err := runtime.provider.Verifier(&oidc.Config{ClientID: provider.config.ClientID}).Verify(ctx, rawIDToken)
	if err != nil {
		if providerNetworkError(err) {
			return ProviderTokens{}, VerifiedIdentity{}, fmt.Errorf("verify ID token: %w", ErrProviderUnavailable)
		}
		return ProviderTokens{}, VerifiedIdentity{}, errors.New("ID token verification failed")
	}
	if subtle.ConstantTimeCompare([]byte(verified.Nonce), []byte(nonce)) != 1 {
		return ProviderTokens{}, VerifiedIdentity{}, errors.New("ID token nonce mismatch")
	}
	if verified.AccessTokenHash != "" {
		if err := verified.VerifyAccessToken(token.AccessToken); err != nil {
			return ProviderTokens{}, VerifiedIdentity{}, errors.New("ID token access token hash mismatch")
		}
	}
	var claims json.RawMessage
	if err := verified.Claims(&claims); err != nil {
		return ProviderTokens{}, VerifiedIdentity{}, errors.New("decode verified ID token claims")
	}
	return ProviderTokens{
			AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, AccessExpiresAt: token.Expiry,
		}, VerifiedIdentity{
			Issuer: verified.Issuer, Subject: verified.Subject, LoginMethodID: methodID, Claims: append(json.RawMessage(nil), claims...),
		}, nil
}

func providerNetworkError(err error) bool {
	var networkError net.Error
	return errors.As(err, &networkError) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) ||
		strings.Contains(strings.ToLower(err.Error()), "fetching keys")
}
