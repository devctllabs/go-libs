package oidcsession

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

var errUnknownLoginMethod = errors.New("unknown login method")

type sessionServiceConfig struct {
	defaultReturnPath string
	stateTTL          time.Duration
	loginAuthorizer   LoginAuthorizer
	observer          Observer
	methods           map[string]url.Values
}

type sessionService struct {
	config         sessionServiceConfig
	provider       loginProvider
	sessions       SessionBackend
	stateEncryptor Encryptor
}

type beginLoginCommand struct {
	methodID       string
	returnPath     string
	browserBinding string
}

// browserBinding ties an OIDC login transaction to the browser that initiated it.
type browserBinding struct {
	value     string
	expiresAt time.Time
}

type beginLoginResult struct {
	location       string
	bindingToStore *browserBinding
}

type completeLoginCommand struct {
	code           string
	state          string
	browserBinding string
}

type completeLoginResult struct {
	tokens           ProviderTokens
	credential       SessionCredential
	sessionExpiresAt time.Time
	returnPath       string
}

func newSessionService(config sessionServiceConfig, provider loginProvider, sessions SessionBackend, stateEncryptor Encryptor) *sessionService {
	return &sessionService{config: config, provider: provider, sessions: sessions, stateEncryptor: stateEncryptor}
}

func (service *sessionService) BeginLogin(ctx context.Context, command beginLoginCommand) (beginLoginResult, error) {
	started := time.Now()
	parameters, found := service.config.methods[command.methodID]
	if command.methodID != "" && !found {
		return beginLoginResult{}, errUnknownLoginMethod
	}
	result := beginLoginResult{}
	binding := command.browserBinding
	if binding == "" {
		generated, err := randomValue(32)
		if err != nil {
			return result, err
		}
		binding = generated
		result.bindingToStore = &browserBinding{value: binding, expiresAt: time.Now().Add(service.config.stateTTL)}
	}
	nonce, err := randomValue(32)
	if err != nil {
		service.observeFailure(ctx, OperationLogin, started, err)
		return result, err
	}
	now := time.Now()
	verifier := oauth2.GenerateVerifier()
	state, err := service.encodeState(ctx, statePayload{
		IssuedAt: now, ExpiresAt: now.Add(service.config.stateTTL), Nonce: nonce,
		PKCEVerifier: verifier, LoginMethodID: command.methodID,
		ReturnPath:           normalizedReturnPath(command.returnPath, service.config.defaultReturnPath),
		BrowserBindingDigest: bindingDigest(binding),
	})
	if err != nil {
		service.observeFailure(ctx, OperationLogin, started, err)
		return result, err
	}
	result.location, err = service.provider.authorizationURL(state, nonce, verifier, parameters)
	if err != nil {
		service.observeFailure(ctx, OperationLogin, started, err)
		return result, err
	}
	return result, nil
}

func (service *sessionService) CompleteLogin(ctx context.Context, command completeLoginCommand) (completeLoginResult, error) {
	started := time.Now()
	state, err := service.decodeState(ctx, command.state, command.browserBinding)
	if err != nil || command.code == "" {
		return completeLoginResult{}, errors.New("invalid login callback")
	}
	tokens, identity, err := service.provider.exchange(ctx, command.code, state.Nonce, state.PKCEVerifier, state.LoginMethodID)
	if err != nil {
		service.observeFailure(ctx, OperationCallback, started, err)
		return completeLoginResult{}, err
	}
	if service.config.loginAuthorizer != nil {
		if err := service.config.loginAuthorizer(ctx, identity); err != nil {
			if !errors.Is(err, ErrLoginDenied) {
				service.observeFailure(ctx, OperationCallback, started, err)
			}
			return completeLoginResult{}, fmt.Errorf("%w: %v", ErrLoginDenied, err)
		}
	}
	created, err := service.sessions.Create(ctx, CreateSessionParams(tokens))
	if err != nil {
		service.observeFailure(ctx, OperationCallback, started, err)
		return completeLoginResult{}, err
	}
	return completeLoginResult{
		tokens: tokens, credential: created.Credential, sessionExpiresAt: created.SessionExpiresAt, returnPath: state.ReturnPath,
	}, nil
}

func (service *sessionService) Refresh(ctx context.Context, credential SessionCredential) (RefreshSessionResult, error) {
	started := time.Now()
	refreshed, err := service.sessions.Refresh(ctx, credential)
	if err == nil {
		return refreshed, nil
	}
	if errors.Is(err, ErrInvalidGrant) || errors.Is(err, ErrInvalidSession) {
		_ = service.sessions.Revoke(ctx, credential)
		return RefreshSessionResult{}, err
	}
	service.observeFailure(ctx, OperationRefresh, started, err)
	return RefreshSessionResult{}, err
}

func (service *sessionService) Logout(ctx context.Context, credential SessionCredential) {
	started := time.Now()
	if err := service.sessions.Revoke(ctx, credential); err != nil && !errors.Is(err, ErrInvalidSession) {
		service.observeFailure(ctx, OperationLogout, started, err)
	}
}

func (service *sessionService) Session(ctx context.Context, credential SessionCredential) (SessionStatus, error) {
	started := time.Now()
	status, err := service.sessions.Status(ctx, credential)
	if err != nil && !errors.Is(err, ErrInvalidSession) {
		service.observeFailure(ctx, OperationSession, started, err)
	}
	return status, err
}

func (service *sessionService) observeFailure(ctx context.Context, operation Operation, started time.Time, err error) {
	observe(ctx, service.config.observer, Observation{Operation: operation, Err: err, Duration: time.Since(started)})
}

var _ sessionApplication = (*sessionService)(nil)
