package oidcsession

import (
	"context"
	"encoding/json"
	"net/url"
	"time"
)

//go:generate go tool mockgen -destination mocks/interfaces.gen.go -package mocks . SessionBackend,TokenService,Encryptor

// SessionCredential is an opaque server-side session credential.
type SessionCredential string

// CreateSessionParams contains the complete provider token set for a new session.
type CreateSessionParams struct {
	AccessToken     string
	RefreshToken    string
	AccessExpiresAt time.Time
}

// CreateSessionResult contains the opaque credential and its current idle/absolute expiry.
type CreateSessionResult struct {
	Credential       SessionCredential
	SessionExpiresAt time.Time
}

// SessionStatus contains only safe session expiry metadata.
type SessionStatus struct {
	AccessExpiresAt  time.Time
	SessionExpiresAt time.Time
}

// RefreshSessionResult contains the current access token and renewed session expiry.
type RefreshSessionResult struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	SessionExpiresAt time.Time
}

// SessionBackend owns opaque refresh credentials and provider token persistence.
type SessionBackend interface {
	// Create persists params and returns a newly generated opaque credential.
	Create(ctx context.Context, params CreateSessionParams) (result CreateSessionResult, err error)
	// Status returns safe expiry metadata without extending idle lifetime.
	Status(ctx context.Context, credential SessionCredential) (status SessionStatus, err error)
	// Refresh returns a usable access token and extends idle lifetime on success.
	Refresh(ctx context.Context, credential SessionCredential) (result RefreshSessionResult, err error)
	// Revoke invalidates credential locally and attempts provider revocation.
	Revoke(ctx context.Context, credential SessionCredential) error
}

// LoginMethod adds validated broker-specific authorization parameters.
type LoginMethod struct {
	ID                      string
	AuthorizationParameters url.Values
}

// VerifiedIdentity is a fully verified identity presented to local login admission policy.
type VerifiedIdentity struct {
	Issuer        string
	Subject       string
	LoginMethodID string
	Claims        json.RawMessage
}

// LoginAuthorizer performs a quick local allow/deny decision after OIDC verification.
type LoginAuthorizer func(ctx context.Context, identity VerifiedIdentity) error
