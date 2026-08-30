package oidcsessionredis

import (
	"context"
	"time"
)

//go:generate go tool mockgen -source=store.go -destination=store.gen_test.go -package=oidcsessionredis -typed

type refreshState int64

const (
	refreshReady   refreshState = 1
	refreshWaiting refreshState = 2
	refreshOwned   refreshState = 3
)

type refreshGateParams struct {
	key           sessionKey
	now           time.Time
	refreshWindow time.Duration
	idleTimeout   time.Duration
	owner         string
	leaseDuration time.Duration
}

type refreshCommitParams struct {
	key             sessionKey
	owner           string
	payload         string
	accessExpiresAt time.Time
	now             time.Time
	idleTimeout     time.Duration
}

type sessionStore interface {
	// Create atomically stores record under a new session key until expiresAt.
	Create(ctx context.Context, key sessionKey, record storedRecord, expiresAt time.Time) (bool, error)
	// Status loads the session addressed by key.
	Status(ctx context.Context, key sessionKey) (storedRecord, error)
	// GateRefresh returns the current refresh coordination state and associated record.
	GateRefresh(ctx context.Context, params refreshGateParams) (refreshState, storedRecord, error)
	// CommitRefresh atomically replaces provider tokens for the active lease owner.
	CommitRefresh(ctx context.Context, params refreshCommitParams) error
	// ReleaseRefresh releases the refresh lease when owner still owns it.
	ReleaseRefresh(ctx context.Context, key sessionKey, owner string) error
	// Revoke atomically removes a session and returns its encrypted token payload.
	Revoke(ctx context.Context, key sessionKey) (string, error)
}

type storedRecord struct {
	Format            int    `json:"format"`
	Payload           string `json:"payload"`
	LastRefreshAt     int64  `json:"last_refresh_at"`
	AccessExpiresAt   int64  `json:"access_expires_at"`
	AbsoluteExpiresAt int64  `json:"absolute_expires_at"`
	LeaseOwner        string `json:"lease_owner,omitempty"`
	LeaseUntil        int64  `json:"lease_until,omitempty"`
}
