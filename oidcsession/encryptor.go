package oidcsession

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	envelopeVersion = byte(1)
	keyIDSize       = 8
)

// Encryptor protects opaque session and OAuth state payloads.
type Encryptor interface {
	// Encrypt authenticates and encrypts plaintext. The returned slice is caller-owned.
	Encrypt(ctx context.Context, plaintext []byte) (ciphertext []byte, err error)
	// Decrypt authenticates and decrypts ciphertext. The returned slice is caller-owned.
	Decrypt(ctx context.Context, ciphertext []byte) (plaintext []byte, err error)
}

type aesGCMEncryptor struct {
	primaryID [keyIDSize]byte
	primary   cipher.AEAD
	keyring   map[[keyIDSize]byte]cipher.AEAD
}

// NewAESGCMEncryptor creates an AES-256-GCM ordered keyring. primary encrypts new data;
// fallbacks only decrypt existing envelopes.
func NewAESGCMEncryptor(primary []byte, fallbacks ...[]byte) (Encryptor, error) {
	keys := append([][]byte{primary}, fallbacks...)
	keyring := make(map[[keyIDSize]byte]cipher.AEAD, len(keys))
	var primaryID [keyIDSize]byte
	var primaryAEAD cipher.AEAD
	for index, key := range keys {
		if len(key) != 32 {
			return nil, fmt.Errorf("key %d must contain exactly 32 bytes", index)
		}
		owned := append([]byte(nil), key...)
		block, err := aes.NewCipher(owned)
		clear(owned)
		if err != nil {
			return nil, fmt.Errorf("construct key %d: %w", index, err)
		}
		aead, err := cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("construct GCM key %d: %w", index, err)
		}
		digest := sha256.Sum256(key)
		var id [keyIDSize]byte
		copy(id[:], digest[:keyIDSize])
		if _, duplicate := keyring[id]; duplicate {
			return nil, errors.New("encryption keyring contains duplicate keys")
		}
		keyring[id] = aead
		if index == 0 {
			primaryID = id
			primaryAEAD = aead
		}
	}
	return &aesGCMEncryptor{primaryID: primaryID, primary: primaryAEAD, keyring: keyring}, nil
}

func (encryptor *aesGCMEncryptor) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	nonce := make([]byte, encryptor.primary.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generate encryption nonce: %w", err)
	}
	headerSize := 1 + keyIDSize + 2 + len(nonce)
	envelope := make([]byte, headerSize)
	envelope[0] = envelopeVersion
	copy(envelope[1:1+keyIDSize], encryptor.primaryID[:])
	binary.BigEndian.PutUint16(envelope[1+keyIDSize:], uint16(len(nonce)))
	copy(envelope[1+keyIDSize+2:], nonce)
	return encryptor.primary.Seal(envelope, nonce, plaintext, envelope[:1+keyIDSize]), nil
}

func (encryptor *aesGCMEncryptor) Decrypt(ctx context.Context, envelope []byte) ([]byte, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	if len(envelope) < 1+keyIDSize+2 || envelope[0] != envelopeVersion {
		return nil, errors.New("invalid encrypted envelope")
	}
	var id [keyIDSize]byte
	copy(id[:], envelope[1:1+keyIDSize])
	aead, found := encryptor.keyring[id]
	if !found {
		return nil, errors.New("encrypted envelope uses an unknown key")
	}
	nonceSize := int(binary.BigEndian.Uint16(envelope[1+keyIDSize:]))
	headerSize := 1 + keyIDSize + 2 + nonceSize
	if nonceSize != aead.NonceSize() || len(envelope) < headerSize+aead.Overhead() {
		return nil, errors.New("invalid encrypted envelope")
	}
	plaintext, err := aead.Open(nil, envelope[1+keyIDSize+2:headerSize], envelope[headerSize:], envelope[:1+keyIDSize])
	if err != nil {
		return nil, errors.New("encrypted envelope authentication failed")
	}
	return plaintext, nil
}

type noopEncryptor struct{}

// InsecureNoopEncryptor returns an explicitly insecure plaintext implementation for tests and local development.
func InsecureNoopEncryptor() Encryptor { return noopEncryptor{} }

func (noopEncryptor) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	return append([]byte(nil), plaintext...), nil
}

func (noopEncryptor) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	return append([]byte(nil), ciphertext...), nil
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	return ctx.Err()
}
