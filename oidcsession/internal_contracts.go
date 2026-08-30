package oidcsession

import (
	"context"
	"net/url"
)

//go:generate go tool mockgen -source=internal_contracts.go -destination=internal_contracts.gen_test.go -package=oidcsession -typed
//go:generate go tool mockgen -destination=internal_dependencies.gen_test.go -package=oidcsession -typed -self_package=github.com/devctllabs/go-libs/oidcsession . SessionBackend,Encryptor

type sessionApplication interface {
	// BeginLogin prepares the provider redirect and browser-binding state for a login attempt.
	BeginLogin(ctx context.Context, command beginLoginCommand) (beginLoginResult, error)
	// CompleteLogin verifies the callback and creates a server-side session.
	CompleteLogin(ctx context.Context, command completeLoginCommand) (completeLoginResult, error)
	// Refresh returns current credentials for an existing server-side session.
	Refresh(ctx context.Context, credential SessionCredential) (RefreshSessionResult, error)
	// Logout invalidates credential and records non-terminal provider failures.
	Logout(ctx context.Context, credential SessionCredential)
	// Session returns safe status metadata for credential.
	Session(ctx context.Context, credential SessionCredential) (SessionStatus, error)
}

type loginProvider interface {
	// authorizationURL builds a provider authorization URL from verified flow state.
	authorizationURL(state string, nonce string, verifier string, parameters url.Values) (string, error)
	// exchange verifies an authorization response and returns its provider tokens and identity.
	exchange(ctx context.Context, code string, nonce string, verifier string, methodID string) (ProviderTokens, VerifiedIdentity, error)
}
