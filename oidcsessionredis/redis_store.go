package oidcsessionredis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/devctllabs/go-libs/oidcsession"
	"github.com/redis/go-redis/v9"
)

const recordFormat = 1

type redisSessionStore struct {
	client    redis.UniversalClient
	keyPrefix string
	gate      *redis.Script
	commit    *redis.Script
	release   *redis.Script
}

func newRedisSessionStore(client redis.UniversalClient, keyPrefix string) *redisSessionStore {
	return &redisSessionStore{
		client: client, keyPrefix: keyPrefix,
		gate: redis.NewScript(gateScript), commit: redis.NewScript(commitScript),
		release: redis.NewScript(releaseScript),
	}
}

func (store *redisSessionStore) Create(ctx context.Context, key sessionKey, record storedRecord, expiresAt time.Time) (bool, error) {
	encoded, err := json.Marshal(record)
	if err != nil {
		return false, err
	}
	status, err := store.client.Do(ctx, "SET", store.key(key), encoded, "NX", "PXAT", expiresAt.UnixMilli()).Text()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("create Redis session: %w", err)
	}
	return status == "OK", nil
}

func (store *redisSessionStore) Status(ctx context.Context, key sessionKey) (storedRecord, error) {
	raw, err := store.client.Get(ctx, store.key(key)).Result()
	if errors.Is(err, redis.Nil) {
		return storedRecord{}, oidcsession.ErrInvalidSession
	}
	if err != nil {
		return storedRecord{}, fmt.Errorf("read Redis session: %w", err)
	}
	return decodeRecord(raw)
}

func (store *redisSessionStore) GateRefresh(ctx context.Context, params refreshGateParams) (refreshState, storedRecord, error) {
	result, err := store.gate.Run(ctx, store.client, []string{store.key(params.key)},
		params.now.UnixMilli(), params.refreshWindow.Milliseconds(), params.idleTimeout.Milliseconds(),
		params.owner, params.leaseDuration.Milliseconds()).Slice()
	if errors.Is(err, redis.Nil) {
		return 0, storedRecord{}, oidcsession.ErrInvalidSession
	}
	if err != nil {
		return 0, storedRecord{}, fmt.Errorf("coordinate Redis session refresh: %w", err)
	}
	return decodeGateResult(result)
}

func (store *redisSessionStore) CommitRefresh(ctx context.Context, params refreshCommitParams) error {
	committed, err := store.commit.Run(ctx, store.client, []string{store.key(params.key)}, params.owner, params.payload,
		params.accessExpiresAt.UnixMilli(), params.now.UnixMilli(), params.idleTimeout.Milliseconds()).Int64()
	if err != nil {
		return fmt.Errorf("commit Redis session refresh: %w", err)
	}
	if committed != 1 {
		return oidcsession.ErrInvalidSession
	}
	return nil
}

func (store *redisSessionStore) ReleaseRefresh(ctx context.Context, key sessionKey, owner string) error {
	_, err := store.release.Run(ctx, store.client, []string{store.key(key)}, owner).Result()
	return err
}

func (store *redisSessionStore) Revoke(ctx context.Context, key sessionKey) (string, error) {
	raw, err := store.client.GetDel(ctx, store.key(key)).Result()
	if errors.Is(err, redis.Nil) {
		return "", oidcsession.ErrInvalidSession
	}
	if err != nil {
		return "", fmt.Errorf("revoke Redis session: %w", err)
	}
	record, err := decodeRecord(raw)
	if err != nil {
		return "", err
	}
	return record.Payload, nil
}

func (store *redisSessionStore) key(key sessionKey) string {
	return store.keyPrefix + string(key)
}

func decodeRecord(raw string) (storedRecord, error) {
	var record storedRecord
	if err := json.Unmarshal([]byte(raw), &record); err != nil || record.Format != recordFormat || record.Payload == "" {
		return storedRecord{}, errors.New("stored session record is malformed")
	}
	return record, nil
}

func decodeGateResult(result []any) (refreshState, storedRecord, error) {
	if len(result) == 0 {
		return 0, storedRecord{}, oidcsession.ErrInvalidSession
	}
	rawState, ok := result[0].(int64)
	if !ok {
		return 0, storedRecord{}, errors.New("redis returned a malformed refresh state")
	}
	state := refreshState(rawState)
	if state == refreshWaiting {
		return state, storedRecord{}, nil
	}
	if len(result) != 2 {
		return 0, storedRecord{}, errors.New("redis returned a malformed refresh record")
	}
	raw, ok := result[1].(string)
	if !ok {
		return 0, storedRecord{}, errors.New("redis returned a malformed refresh record")
	}
	record, err := decodeRecord(raw)
	return state, record, err
}

var _ sessionStore = (*redisSessionStore)(nil)
