package oidcsessionredis

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	"github.com/devctllabs/go-libs/oidcsession"
)

type sessionKey string

func parseCredential(credential oidcsession.SessionCredential) (sessionKey, error) {
	secret, err := base64.RawURLEncoding.DecodeString(string(credential))
	if err != nil || len(secret) != 32 {
		return "", oidcsession.ErrInvalidSession
	}
	return sessionKey(encodedDigest(secret)), nil
}

func encodedDigest(value []byte) string {
	digest := sha256.Sum256(value)
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func randomEncoded(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
