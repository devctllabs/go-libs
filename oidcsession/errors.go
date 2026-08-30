package oidcsession

import "errors"

var (
	// ErrProviderUnavailable reports that OIDC metadata or a provider operation is temporarily unavailable.
	ErrProviderUnavailable = errors.New("OIDC provider unavailable")
	// ErrInvalidSession reports a missing, expired, revoked, or malformed session credential.
	ErrInvalidSession = errors.New("invalid session")
	// ErrInvalidGrant reports that the provider rejected a refresh grant permanently.
	ErrInvalidGrant = errors.New("invalid grant")
	// ErrLoginDenied reports that local admission policy rejected an otherwise verified identity.
	ErrLoginDenied = errors.New("login denied")
)
