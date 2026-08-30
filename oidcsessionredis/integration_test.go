//go:build integration

package oidcsessionredis

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/oidcsession"
	"github.com/devctllabs/go-libs/oidcsession/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcwait "github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/mock/gomock"
)

func TestRedisBackendEncryptsTokensAndCoordinatesRefreshAcrossReplicas(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "redis:7.4-alpine", ExposedPorts: []string{"6379/tcp"}, WaitingFor: tcwait.ForListeningPort("6379/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(container)) })
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "6379/tcp")
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%s", host, port.Port())})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	assertRedisStorePrimitives(ctx, t, client)

	controller := gomock.NewController(t)
	tokens := mocks.NewMockTokenService(controller)
	key := make([]byte, 32)
	_, err = rand.Read(key)
	require.NoError(t, err)
	encryptor, err := oidcsession.NewAESGCMEncryptor(key)
	require.NoError(t, err)
	config := BackendConfig{
		KeyPrefix: "test:sessions:", IdleTimeout: 2 * time.Second, AbsoluteLifetime: time.Minute, RefreshWindow: time.Second,
	}
	first, err := NewBackend(client, config, tokens, encryptor)
	require.NoError(t, err)
	second, err := NewBackend(client, config, tokens, encryptor)
	require.NoError(t, err)
	created, err := first.Create(ctx, oidcsession.CreateSessionParams{
		AccessToken: "sensitive-access", RefreshToken: "sensitive-refresh", AccessExpiresAt: time.Now().Add(50 * time.Millisecond),
	})
	require.NoError(t, err)
	credentialBytes, err := base64.RawURLEncoding.DecodeString(string(created.Credential))
	require.NoError(t, err)
	require.Len(t, credentialBytes, 32)

	keys, err := client.Keys(ctx, "test:sessions:*").Result()
	require.NoError(t, err)
	digest := sha256.Sum256(credentialBytes)
	require.Equal(t, []string{"test:sessions:" + base64.RawURLEncoding.EncodeToString(digest[:])}, keys)
	stored, err := client.Get(ctx, keys[0]).Result()
	require.NoError(t, err)
	require.NotContains(t, fmt.Sprint(stored), "sensitive-access")
	require.NotContains(t, fmt.Sprint(stored), "sensitive-refresh")

	time.Sleep(60 * time.Millisecond)
	newExpiry := time.Now().Add(time.Minute)
	tokens.EXPECT().Refresh(gomock.Any(), "sensitive-refresh").DoAndReturn(func(context.Context, string) (oidcsession.ProviderTokens, error) {
		time.Sleep(75 * time.Millisecond)
		return oidcsession.ProviderTokens{AccessToken: "new-access", RefreshToken: "new-refresh", AccessExpiresAt: newExpiry}, nil
	}).Times(1)

	results := make([]oidcsession.RefreshSessionResult, 2)
	errorsByCall := make([]error, 2)
	type refreshCall struct {
		index  int
		result oidcsession.RefreshSessionResult
		err    error
	}
	refreshCtx, cancelRefresh := context.WithTimeout(ctx, 5*time.Second)
	defer cancelRefresh()
	calls := make(chan refreshCall, 2)
	for index, backend := range []*Backend{first, second} {
		go func(index int, backend *Backend) {
			result, err := backend.Refresh(refreshCtx, created.Credential)
			calls <- refreshCall{index: index, result: result, err: err}
		}(index, backend)
	}
	for range 2 {
		select {
		case call := <-calls:
			results[call.index] = call.result
			errorsByCall[call.index] = call.err
		case <-refreshCtx.Done():
			require.FailNow(t, "timed out waiting for Redis refresh workers", "error: %v", refreshCtx.Err())
		}
	}
	for index := range results {
		require.NoError(t, errorsByCall[index])
		require.Equal(t, "new-access", results[index].AccessToken)
		require.WithinDuration(t, newExpiry, results[index].AccessExpiresAt, time.Millisecond)
		require.WithinDuration(t, time.Now().Add(2*time.Second), results[index].SessionExpiresAt, 250*time.Millisecond)
	}

	tokens.EXPECT().Revoke(gomock.Any(), "new-refresh").Return(nil)
	require.NoError(t, second.Revoke(ctx, created.Credential))
	_, err = first.Status(ctx, created.Credential)
	require.ErrorIs(t, err, oidcsession.ErrInvalidSession)
}

func assertRedisStorePrimitives(ctx context.Context, t *testing.T, client redis.UniversalClient) {
	t.Helper()
	store := newRedisSessionStore(client, "test:native:")
	record := storedRecord{Format: recordFormat, Payload: "encrypted"}
	key := sessionKey("session")
	expiresAt := time.Now().Add(time.Minute)

	created, err := store.Create(ctx, key, record, expiresAt)
	require.NoError(t, err)
	require.True(t, created)
	created, err = store.Create(ctx, key, record, expiresAt)
	require.NoError(t, err)
	require.False(t, created)
	ttl, err := client.PTTL(ctx, "test:native:session").Result()
	require.NoError(t, err)
	require.Positive(t, ttl)
	require.LessOrEqual(t, ttl, time.Minute+100*time.Millisecond)

	loaded, err := store.Status(ctx, key)
	require.NoError(t, err)
	require.Equal(t, record, loaded)
	_, err = store.Status(ctx, sessionKey("missing"))
	require.ErrorIs(t, err, oidcsession.ErrInvalidSession)

	assertRedisRefreshCoordination(ctx, t, store)

	malformed := sessionKey("malformed")
	malformedKey := store.key(malformed)
	require.NoError(t, client.Set(ctx, malformedKey, "not-json", time.Minute).Err())
	_, err = store.Revoke(ctx, malformed)
	require.Error(t, err)
	require.ErrorIs(t, client.Get(ctx, malformedKey).Err(), redis.Nil)
}

func assertRedisRefreshCoordination(ctx context.Context, t *testing.T, store *redisSessionStore) {
	t.Helper()
	now := time.Now().Truncate(time.Millisecond)
	key := sessionKey("refresh")
	record := storedRecord{
		Format: recordFormat, Payload: "old", LastRefreshAt: now.UnixMilli(),
		AccessExpiresAt: now.Add(-time.Minute).UnixMilli(), AbsoluteExpiresAt: now.Add(time.Hour).UnixMilli(),
	}
	created, err := store.Create(ctx, key, record, now.Add(time.Hour))
	require.NoError(t, err)
	require.True(t, created)

	gate := refreshGateParams{
		key: key, now: now, refreshWindow: time.Minute, idleTimeout: time.Minute,
		owner: "owner-1", leaseDuration: refreshLeaseDuration,
	}
	state, leased, err := store.GateRefresh(ctx, gate)
	require.NoError(t, err)
	require.Equal(t, refreshOwned, state)
	require.Equal(t, "owner-1", leased.LeaseOwner)

	gate.owner = "owner-2"
	state, _, err = store.GateRefresh(ctx, gate)
	require.NoError(t, err)
	require.Equal(t, refreshWaiting, state)
	require.NoError(t, store.ReleaseRefresh(ctx, key, "not-owner"))
	state, _, err = store.GateRefresh(ctx, gate)
	require.NoError(t, err)
	require.Equal(t, refreshWaiting, state)

	require.NoError(t, store.ReleaseRefresh(ctx, key, "owner-1"))
	state, _, err = store.GateRefresh(ctx, gate)
	require.NoError(t, err)
	require.Equal(t, refreshOwned, state)

	gate.now = now.Add(refreshLeaseDuration + time.Millisecond)
	gate.owner = "owner-3"
	state, _, err = store.GateRefresh(ctx, gate)
	require.NoError(t, err)
	require.Equal(t, refreshOwned, state)

	committedAt := gate.now
	accessExpiresAt := committedAt.Add(time.Minute)
	err = store.CommitRefresh(ctx, refreshCommitParams{
		key: key, owner: "owner-2", payload: "stale", accessExpiresAt: accessExpiresAt,
		now: committedAt, idleTimeout: time.Minute,
	})
	require.ErrorIs(t, err, oidcsession.ErrInvalidSession)
	require.NoError(t, store.CommitRefresh(ctx, refreshCommitParams{
		key: key, owner: "owner-3", payload: "updated", accessExpiresAt: accessExpiresAt,
		now: committedAt, idleTimeout: time.Minute,
	}))

	gate.now = committedAt.Add(time.Second)
	gate.refreshWindow = time.Second
	state, ready, err := store.GateRefresh(ctx, gate)
	require.NoError(t, err)
	require.Equal(t, refreshReady, state)
	require.Equal(t, "updated", ready.Payload)
	require.Equal(t, gate.now.UnixMilli(), ready.LastRefreshAt)
	require.Empty(t, ready.LeaseOwner)

	_, err = store.Revoke(ctx, sessionKey("wrong"))
	require.ErrorIs(t, err, oidcsession.ErrInvalidSession)
	revokedPayload, err := store.Revoke(ctx, key)
	require.NoError(t, err)
	require.Equal(t, "updated", revokedPayload)
	_, err = store.Status(ctx, key)
	require.ErrorIs(t, err, oidcsession.ErrInvalidSession)
}
