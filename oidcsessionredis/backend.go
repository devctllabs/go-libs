package oidcsessionredis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/devctllabs/go-libs/oidcsession"
	"github.com/redis/go-redis/v9"
)

const (
	defaultRefreshWindow = time.Minute
	refreshLeaseDuration = 15 * time.Second
	refreshPollInterval  = 20 * time.Millisecond
)

// BackendConfig controls Redis key ownership and idle/absolute session policy.
type BackendConfig struct {
	KeyPrefix        string
	IdleTimeout      time.Duration
	AbsoluteLifetime time.Duration
	RefreshWindow    time.Duration
	Observer         oidcsession.Observer
}

// Backend persists encrypted provider tokens and coordinates refresh across replicas.
// The Redis client remains caller-owned and is never closed by Backend.
type Backend struct {
	config    BackendConfig
	tokens    oidcsession.TokenService
	encryptor oidcsession.Encryptor
	store     sessionStore
}

// NewBackend validates policy and constructs a Redis session backend without performing I/O.
func NewBackend(client redis.UniversalClient, config BackendConfig, tokens oidcsession.TokenService, encryptor oidcsession.Encryptor) (*Backend, error) {
	if client == nil || tokens == nil || encryptor == nil {
		return nil, errors.New("redis client, token service, and encryptor are required")
	}
	if err := validateConfig(&config); err != nil {
		return nil, err
	}
	return newBackend(config, tokens, encryptor, newRedisSessionStore(client, config.KeyPrefix)), nil
}

func newBackend(config BackendConfig, tokens oidcsession.TokenService, encryptor oidcsession.Encryptor, store sessionStore) *Backend {
	return &Backend{config: config, tokens: tokens, encryptor: encryptor, store: store}
}

func validateConfig(config *BackendConfig) error {
	if strings.TrimSpace(config.KeyPrefix) == "" {
		return errors.New("key prefix is required")
	}
	if config.IdleTimeout <= 0 || config.AbsoluteLifetime <= 0 {
		return errors.New("idle timeout and absolute lifetime must be positive")
	}
	if config.IdleTimeout > config.AbsoluteLifetime {
		return errors.New("idle timeout must not exceed absolute lifetime")
	}
	if config.RefreshWindow < 0 {
		return errors.New("refresh window must not be negative")
	}
	if config.RefreshWindow == 0 {
		config.RefreshWindow = defaultRefreshWindow
	}
	return nil
}

// Create implements oidcsession.SessionBackend.
func (backend *Backend) Create(ctx context.Context, params oidcsession.CreateSessionParams) (oidcsession.CreateSessionResult, error) {
	if err := validateContext(ctx); err != nil {
		return oidcsession.CreateSessionResult{}, err
	}
	if params.AccessToken == "" || params.RefreshToken == "" || params.AccessExpiresAt.IsZero() || !params.AccessExpiresAt.After(time.Now()) {
		return oidcsession.CreateSessionResult{}, errors.New("complete future-dated provider tokens are required")
	}
	payload, err := backend.encryptTokens(ctx, tokenPayload(params))
	if err != nil {
		return oidcsession.CreateSessionResult{}, fmt.Errorf("encrypt provider tokens: %w", err)
	}
	for range 3 {
		secret, err := randomEncoded(32)
		if err != nil {
			return oidcsession.CreateSessionResult{}, err
		}
		credential := oidcsession.SessionCredential(secret)
		key, err := parseCredential(credential)
		if err != nil {
			return oidcsession.CreateSessionResult{}, err
		}
		now := time.Now()
		absoluteExpiry := now.Add(backend.config.AbsoluteLifetime)
		sessionExpiry := minTime(now.Add(backend.config.IdleTimeout), absoluteExpiry)
		record := storedRecord{
			Format: recordFormat, Payload: payload, LastRefreshAt: now.UnixMilli(),
			AccessExpiresAt: params.AccessExpiresAt.UnixMilli(), AbsoluteExpiresAt: absoluteExpiry.UnixMilli(),
		}
		created, err := backend.store.Create(ctx, key, record, sessionExpiry)
		if err != nil {
			return oidcsession.CreateSessionResult{}, err
		}
		if created {
			return oidcsession.CreateSessionResult{Credential: credential, SessionExpiresAt: sessionExpiry}, nil
		}
	}
	return oidcsession.CreateSessionResult{}, errors.New("generate a unique session credential")
}

// Status implements oidcsession.SessionBackend without extending the idle timeout.
func (backend *Backend) Status(ctx context.Context, credential oidcsession.SessionCredential) (oidcsession.SessionStatus, error) {
	key, err := parseCredential(credential)
	if err != nil {
		return oidcsession.SessionStatus{}, oidcsession.ErrInvalidSession
	}
	record, err := backend.store.Status(ctx, key)
	if err != nil {
		return oidcsession.SessionStatus{}, err
	}
	return oidcsession.SessionStatus{
		AccessExpiresAt: time.UnixMilli(record.AccessExpiresAt), SessionExpiresAt: backend.sessionExpiry(record),
	}, nil
}

// Refresh implements oidcsession.SessionBackend and extends idle lifetime after every success.
func (backend *Backend) Refresh(ctx context.Context, credential oidcsession.SessionCredential) (oidcsession.RefreshSessionResult, error) {
	key, err := parseCredential(credential)
	if err != nil {
		return oidcsession.RefreshSessionResult{}, oidcsession.ErrInvalidSession
	}
	owner, err := randomEncoded(16)
	if err != nil {
		return oidcsession.RefreshSessionResult{}, err
	}
	for {
		state, record, err := backend.store.GateRefresh(ctx, refreshGateParams{
			key: key, now: time.Now(), refreshWindow: backend.config.RefreshWindow,
			idleTimeout: backend.config.IdleTimeout, owner: owner, leaseDuration: refreshLeaseDuration,
		})
		if err != nil {
			return oidcsession.RefreshSessionResult{}, err
		}
		switch state {
		case refreshReady:
			return backend.refreshResult(ctx, record)
		case refreshWaiting:
			if err := wait(ctx, refreshPollInterval); err != nil {
				return oidcsession.RefreshSessionResult{}, err
			}
		case refreshOwned:
			return backend.refreshOwned(ctx, key, owner, record)
		default:
			return oidcsession.RefreshSessionResult{}, errors.New("redis returned an unknown refresh state")
		}
	}
}

func (backend *Backend) refreshOwned(ctx context.Context, key sessionKey, owner string, record storedRecord) (oidcsession.RefreshSessionResult, error) {
	payload, err := backend.decryptTokens(ctx, record.Payload)
	if err != nil {
		backend.releaseLease(ctx, key, owner)
		return oidcsession.RefreshSessionResult{}, fmt.Errorf("decrypt provider tokens: %w", err)
	}
	started := time.Now()
	tokens, err := backend.tokens.Refresh(ctx, payload.RefreshToken)
	if err != nil {
		backend.releaseLease(ctx, key, owner)
		if !errors.Is(err, oidcsession.ErrInvalidGrant) {
			backend.observe(ctx, oidcsession.OperationRefresh, started, err)
		}
		return oidcsession.RefreshSessionResult{}, err
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" || tokens.AccessExpiresAt.IsZero() || !tokens.AccessExpiresAt.After(time.Now()) {
		backend.releaseLease(ctx, key, owner)
		return oidcsession.RefreshSessionResult{}, errors.New("token service returned an incomplete token set")
	}
	encrypted, err := backend.encryptTokens(ctx, tokenPayload(tokens))
	if err != nil {
		backend.releaseLease(ctx, key, owner)
		return oidcsession.RefreshSessionResult{}, err
	}
	committedAt := time.Now()
	if err := backend.store.CommitRefresh(ctx, refreshCommitParams{
		key: key, owner: owner, payload: encrypted, accessExpiresAt: tokens.AccessExpiresAt,
		now: committedAt, idleTimeout: backend.config.IdleTimeout,
	}); err != nil {
		return oidcsession.RefreshSessionResult{}, err
	}
	record.LastRefreshAt = committedAt.UnixMilli()
	return oidcsession.RefreshSessionResult{
		AccessToken: tokens.AccessToken, AccessExpiresAt: tokens.AccessExpiresAt, SessionExpiresAt: backend.sessionExpiry(record),
	}, nil
}

func (backend *Backend) refreshResult(ctx context.Context, record storedRecord) (oidcsession.RefreshSessionResult, error) {
	payload, err := backend.decryptTokens(ctx, record.Payload)
	if err != nil {
		return oidcsession.RefreshSessionResult{}, fmt.Errorf("decrypt provider tokens: %w", err)
	}
	return oidcsession.RefreshSessionResult{
		AccessToken: payload.AccessToken, AccessExpiresAt: payload.AccessExpiresAt, SessionExpiresAt: backend.sessionExpiry(record),
	}, nil
}

// Revoke implements oidcsession.SessionBackend. Provider revocation is best effort after atomic local deletion.
func (backend *Backend) Revoke(ctx context.Context, credential oidcsession.SessionCredential) error {
	key, err := parseCredential(credential)
	if err != nil {
		return oidcsession.ErrInvalidSession
	}
	rawPayload, err := backend.store.Revoke(ctx, key)
	if err != nil {
		return err
	}
	payload, err := backend.decryptTokens(ctx, rawPayload)
	if err != nil {
		return fmt.Errorf("decrypt revoked provider tokens: %w", err)
	}
	started := time.Now()
	if err := backend.tokens.Revoke(ctx, payload.RefreshToken); err != nil {
		backend.observe(ctx, oidcsession.OperationLogout, started, err)
	}
	return nil
}

func (backend *Backend) sessionExpiry(record storedRecord) time.Time {
	return minTime(time.UnixMilli(record.LastRefreshAt).Add(backend.config.IdleTimeout), time.UnixMilli(record.AbsoluteExpiresAt))
}

func (backend *Backend) releaseLease(ctx context.Context, key sessionKey, owner string) {
	_ = backend.store.ReleaseRefresh(ctx, key, owner)
}

func (backend *Backend) observe(ctx context.Context, operation oidcsession.Operation, started time.Time, err error) {
	if backend.config.Observer != nil {
		backend.config.Observer.Observe(ctx, oidcsession.Observation{Operation: operation, Err: err, Duration: time.Since(started)})
	}
}

func validateContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	return ctx.Err()
}

func wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func minTime(first time.Time, second time.Time) time.Time {
	if first.Before(second) {
		return first
	}
	return second
}

var _ oidcsession.SessionBackend = (*Backend)(nil)
