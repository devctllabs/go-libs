package oidcsession_test

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/devctllabs/go-libs/oidcsession"
	"github.com/stretchr/testify/require"
)

func TestAESGCMEncryptorRoundTripAndTamperDetection(t *testing.T) {
	t.Parallel()
	key := randomKey(t)
	encryptor, err := oidcsession.NewAESGCMEncryptor(key)
	require.NoError(t, err)

	encrypted, err := encryptor.Encrypt(context.Background(), []byte("provider tokens"))
	require.NoError(t, err)
	require.NotContains(t, string(encrypted), "provider tokens")

	decrypted, err := encryptor.Decrypt(context.Background(), encrypted)
	require.NoError(t, err)
	require.Equal(t, []byte("provider tokens"), decrypted)

	encrypted[len(encrypted)-1] ^= 1
	_, err = encryptor.Decrypt(context.Background(), encrypted)
	require.Error(t, err)
}

func TestAESGCMEncryptorSupportsOrderedKeyRotation(t *testing.T) {
	t.Parallel()
	oldKey := randomKey(t)
	newKey := randomKey(t)
	oldEncryptor, err := oidcsession.NewAESGCMEncryptor(oldKey)
	require.NoError(t, err)
	rotated, err := oidcsession.NewAESGCMEncryptor(newKey, oldKey)
	require.NoError(t, err)

	oldCiphertext, err := oldEncryptor.Encrypt(context.Background(), []byte("old"))
	require.NoError(t, err)
	plaintext, err := rotated.Decrypt(context.Background(), oldCiphertext)
	require.NoError(t, err)
	require.Equal(t, []byte("old"), plaintext)

	newCiphertext, err := rotated.Encrypt(context.Background(), []byte("new"))
	require.NoError(t, err)
	_, err = oldEncryptor.Decrypt(context.Background(), newCiphertext)
	require.Error(t, err)
}

func TestAESGCMEncryptorCopiesCallerKeys(t *testing.T) {
	t.Parallel()
	key := randomKey(t)
	encryptor, err := oidcsession.NewAESGCMEncryptor(key)
	require.NoError(t, err)
	clear(key)

	encrypted, err := encryptor.Encrypt(context.Background(), []byte("still works"))
	require.NoError(t, err)
	decrypted, err := encryptor.Decrypt(context.Background(), encrypted)
	require.NoError(t, err)
	require.Equal(t, []byte("still works"), decrypted)
}

func randomKey(t *testing.T) []byte {
	t.Helper()
	value := make([]byte, 32)
	_, err := rand.Read(value)
	require.NoError(t, err)
	return value
}
