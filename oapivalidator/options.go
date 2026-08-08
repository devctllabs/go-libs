package oapivalidator

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Option configures request validation middleware.
type Option interface {
	apply(config *config) error
}

type optionFunc func(config *config) error

func (fn optionFunc) apply(config *config) error { return fn(config) }

type config struct {
	authenticator     Authenticator
	failureHandler    FailureHandler
	baseURL           string
	maxReportedErrors int
}

const defaultMaxReportedErrors = 20

func defaultConfig() config { return config{maxReportedErrors: defaultMaxReportedErrors} }

// WithAuthenticator configures OpenAPI security-scheme authentication.
func WithAuthenticator(authenticator Authenticator) Option {
	return optionFunc(func(config *config) error {
		if authenticator == nil {
			return errors.New("authenticator must not be nil")
		}
		config.authenticator = authenticator
		return nil
	})
}

// WithFailureHandler replaces the default RFC 9457 problem writer.
func WithFailureHandler(handler FailureHandler) Option {
	return optionFunc(func(config *config) error {
		if handler == nil {
			return errors.New("failure handler must not be nil")
		}
		config.failureHandler = handler
		return nil
	})
}

// WithBaseURL configures the path prefix used when generated handlers are
// registered with oapi-codegen RegisterHandlersOptions.BaseURL.
func WithBaseURL(baseURL string) Option {
	return optionFunc(func(config *config) error {
		normalized, err := normalizeBaseURL(baseURL)
		if err != nil {
			return err
		}
		config.baseURL = normalized
		return nil
	})
}

// WithMaxReportedErrors limits the normalized validation errors returned to a
// client. Validation itself still examines the complete request.
func WithMaxReportedErrors(maxErrors int) Option {
	return optionFunc(func(config *config) error {
		if maxErrors <= 0 {
			return fmt.Errorf("max reported errors must be positive: %d", maxErrors)
		}
		config.maxReportedErrors = maxErrors
		return nil
	})
}

func normalizeBaseURL(baseURL string) (string, error) {
	if baseURL == "" {
		return "", nil
	}
	if strings.HasSuffix(baseURL, "/") {
		return "", fmt.Errorf("base URL must not have a trailing slash: %q", baseURL)
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base URL: %w", err)
	}
	if parsed.IsAbs() || parsed.Host != "" || parsed.RawQuery != "" || parsed.Fragment != "" ||
		!strings.HasPrefix(parsed.Path, "/") || parsed.Path != baseURL {
		return "", fmt.Errorf("base URL must be an absolute path without query or fragment: %q", baseURL)
	}
	return parsed.Path, nil
}
