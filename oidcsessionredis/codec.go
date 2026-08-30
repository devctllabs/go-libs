package oidcsessionredis

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

type tokenPayload struct {
	AccessToken     string    `json:"access_token"`
	RefreshToken    string    `json:"refresh_token"`
	AccessExpiresAt time.Time `json:"access_expires_at"`
}

func (backend *Backend) encryptTokens(ctx context.Context, payload tokenPayload) (string, error) {
	plaintext, err := json.Marshal(payload) //nolint:gosec // The secret-bearing JSON is immediately encrypted and never persisted as plaintext.
	if err != nil {
		return "", err
	}
	ciphertext, err := backend.encryptor.Encrypt(ctx, plaintext)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

func (backend *Backend) decryptTokens(ctx context.Context, encoded string) (tokenPayload, error) {
	ciphertext, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return tokenPayload{}, errors.New("stored token payload is malformed")
	}
	plaintext, err := backend.encryptor.Decrypt(ctx, ciphertext)
	if err != nil {
		return tokenPayload{}, err
	}
	var payload tokenPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return tokenPayload{}, errors.New("stored token payload is malformed")
	}
	return payload, nil
}
