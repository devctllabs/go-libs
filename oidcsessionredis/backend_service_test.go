package oidcsessionredis

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/oidcsession"
	"github.com/devctllabs/go-libs/oidcsession/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestBackendStatusUsesStoreContract(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	store := NewMocksessionStore(controller)
	now := time.Now()
	key, credential := validCredential()
	store.EXPECT().Status(gomock.Any(), key).Return(storedRecord{
		Format: recordFormat, Payload: "payload", AccessExpiresAt: now.Add(time.Minute).UnixMilli(),
		LastRefreshAt: now.UnixMilli(), AbsoluteExpiresAt: now.Add(time.Hour).UnixMilli(),
	}, nil)
	backend := newBackend(BackendConfig{IdleTimeout: 10 * time.Minute, AbsoluteLifetime: time.Hour},
		mocks.NewMockTokenService(controller), mocks.NewMockEncryptor(controller), store)

	status, err := backend.Status(context.Background(), credential)

	require.NoError(t, err)
	require.WithinDuration(t, now.Add(time.Minute), status.AccessExpiresAt, time.Millisecond)
	require.WithinDuration(t, now.Add(10*time.Minute), status.SessionExpiresAt, time.Millisecond)
}

func TestBackendRefreshReturnsUsableStoredTokenWithoutProviderRefresh(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	store := NewMocksessionStore(controller)
	tokens := mocks.NewMockTokenService(controller)
	encryptor := mocks.NewMockEncryptor(controller)
	key, credential := validCredential()
	now := time.Now()
	payload := tokenPayload{AccessToken: "access", RefreshToken: "refresh", AccessExpiresAt: now.Add(time.Minute)}
	plaintext, err := json.Marshal(payload)
	require.NoError(t, err)
	ciphertext := []byte("encrypted")
	record := storedRecord{
		Format: recordFormat, Payload: base64.RawStdEncoding.EncodeToString(ciphertext), LastRefreshAt: now.UnixMilli(),
		AccessExpiresAt: payload.AccessExpiresAt.UnixMilli(), AbsoluteExpiresAt: now.Add(time.Hour).UnixMilli(),
	}
	var gate refreshGateParams
	store.EXPECT().GateRefresh(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, params refreshGateParams) (refreshState, storedRecord, error) {
		gate = params
		return refreshReady, record, nil
	})
	encryptor.EXPECT().Decrypt(gomock.Any(), ciphertext).Return(plaintext, nil)
	backend := newBackend(BackendConfig{IdleTimeout: 10 * time.Minute, AbsoluteLifetime: time.Hour, RefreshWindow: time.Minute}, tokens, encryptor, store)

	result, err := backend.Refresh(context.Background(), credential)

	require.NoError(t, err)
	require.Equal(t, "access", result.AccessToken)
	require.Equal(t, key, gate.key)
	require.NotEmpty(t, gate.owner)
	require.Equal(t, time.Minute, gate.refreshWindow)
}

func TestBackendRefreshReleasesLeaseAfterInvalidGrant(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	store := NewMocksessionStore(controller)
	tokens := mocks.NewMockTokenService(controller)
	encryptor := mocks.NewMockEncryptor(controller)
	key, credential := validCredential()
	payload := tokenPayload{AccessToken: "access", RefreshToken: "refresh", AccessExpiresAt: time.Now().Add(time.Minute)}
	plaintext, err := json.Marshal(payload)
	require.NoError(t, err)
	ciphertext := []byte("encrypted")
	record := storedRecord{Format: recordFormat, Payload: base64.RawStdEncoding.EncodeToString(ciphertext)}
	var leaseOwner string
	store.EXPECT().GateRefresh(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, params refreshGateParams) (refreshState, storedRecord, error) {
		leaseOwner = params.owner
		return refreshOwned, record, nil
	})
	encryptor.EXPECT().Decrypt(gomock.Any(), ciphertext).Return(plaintext, nil)
	tokens.EXPECT().Refresh(gomock.Any(), "refresh").Return(oidcsession.ProviderTokens{}, oidcsession.ErrInvalidGrant)
	var releasedOwner string
	store.EXPECT().ReleaseRefresh(gomock.Any(), key, gomock.Any()).DoAndReturn(func(_ context.Context, _ sessionKey, owner string) error {
		releasedOwner = owner
		return nil
	})
	backend := newBackend(BackendConfig{IdleTimeout: 10 * time.Minute, AbsoluteLifetime: time.Hour, RefreshWindow: time.Minute}, tokens, encryptor, store)

	_, err = backend.Refresh(context.Background(), credential)

	require.ErrorIs(t, err, oidcsession.ErrInvalidGrant)
	require.NotEmpty(t, leaseOwner)
	require.Equal(t, leaseOwner, releasedOwner)
}

func validCredential() (sessionKey, oidcsession.SessionCredential) {
	secret := make([]byte, 32)
	credential := oidcsession.SessionCredential(base64.RawURLEncoding.EncodeToString(secret))
	return sessionKey(encodedDigest(secret)), credential
}
