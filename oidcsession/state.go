package oidcsession

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type statePayload struct {
	IssuedAt             time.Time `json:"iat"`
	ExpiresAt            time.Time `json:"exp"`
	Nonce                string    `json:"nonce"`
	PKCEVerifier         string    `json:"pkce"`
	LoginMethodID        string    `json:"method,omitempty"`
	ReturnPath           string    `json:"return"`
	BrowserBindingDigest string    `json:"binding"`
}

func (service *sessionService) encodeState(ctx context.Context, state statePayload) (string, error) {
	plaintext, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	ciphertext, err := service.stateEncryptor.Encrypt(ctx, plaintext)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func (service *sessionService) decodeState(ctx context.Context, raw string, binding string) (statePayload, error) {
	if raw == "" || binding == "" {
		return statePayload{}, errors.New("state or browser binding is missing")
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return statePayload{}, errors.New("state is malformed")
	}
	plaintext, err := service.stateEncryptor.Decrypt(ctx, ciphertext)
	if err != nil {
		return statePayload{}, errors.New("state cannot be authenticated")
	}
	var state statePayload
	if err := json.Unmarshal(plaintext, &state); err != nil {
		return statePayload{}, errors.New("state is malformed")
	}
	now := time.Now()
	if state.IssuedAt.After(now) || !state.ExpiresAt.After(now) || state.ExpiresAt.Sub(state.IssuedAt) > service.config.stateTTL ||
		state.Nonce == "" || state.PKCEVerifier == "" || !validReturnPath(state.ReturnPath) ||
		subtle.ConstantTimeCompare([]byte(state.BrowserBindingDigest), []byte(bindingDigest(binding))) != 1 {
		return statePayload{}, errors.New("state is invalid or expired")
	}
	if state.LoginMethodID != "" {
		if _, found := service.config.methods[state.LoginMethodID]; !found {
			return statePayload{}, errors.New("state login method is unknown")
		}
	}
	return state, nil
}

func cloneLoginMethods(methods []LoginMethod) (map[string]url.Values, error) {
	result := make(map[string]url.Values, len(methods)+1)
	result[""] = nil
	for _, method := range methods {
		if strings.TrimSpace(method.ID) == "" {
			return nil, errors.New("login method ID is required")
		}
		if _, duplicate := result[method.ID]; duplicate {
			return nil, fmt.Errorf("duplicate login method %q", method.ID)
		}
		parameters := make(url.Values, len(method.AuthorizationParameters))
		for key, values := range method.AuthorizationParameters {
			if reservedAuthorizationParameter(key) || len(values) != 1 || values[0] == "" {
				return nil, fmt.Errorf("login method %q has invalid authorization parameter %q", method.ID, key)
			}
			parameters[key] = []string{values[0]}
		}
		result[method.ID] = parameters
	}
	return result, nil
}

func reservedAuthorizationParameter(key string) bool {
	switch strings.ToLower(key) {
	case "state", "nonce", "code_challenge", "code_challenge_method", "redirect_uri", "client_id", "response_type", "scope":
		return true
	default:
		return false
	}
}

func validReturnPath(value string) bool {
	if value == "" || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return false
	}
	parsed, err := url.Parse(value)
	return err == nil && !parsed.IsAbs() && parsed.Host == "" && parsed.RawQuery == "" && parsed.Fragment == ""
}

func normalizedReturnPath(value string, fallback string) string {
	if validReturnPath(value) {
		return value
	}
	return fallback
}

func randomValue(bytes int) (string, error) {
	value := make([]byte, bytes)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func bindingDigest(binding string) string {
	digest := sha256.Sum256([]byte(binding))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}
