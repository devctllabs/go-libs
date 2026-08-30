package oapivalidatorjwt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const defaultCSRFHeaderName = "X-CSRF-Protection"

type cookieProtection struct {
	crossOrigin *http.CrossOriginProtection
	headerName  string
}

func newCookieProtection(config *CookieProtectionConfig) (*cookieProtection, error) {
	if config == nil {
		return nil, nil
	}
	headerName := config.CSRFHeaderName
	if headerName == "" {
		headerName = defaultCSRFHeaderName
	}
	if !validCSRFHeaderName(headerName) {
		return nil, fmt.Errorf("invalid or unsafe CSRF header name %q", headerName)
	}
	protection := http.NewCrossOriginProtection()
	for _, origin := range config.TrustedOrigins {
		if err := protection.AddTrustedOrigin(origin); err != nil {
			return nil, fmt.Errorf("add trusted origin: %w", err)
		}
	}
	return &cookieProtection{crossOrigin: protection, headerName: http.CanonicalHeaderKey(headerName)}, nil
}

func (protection *cookieProtection) check(request *http.Request) error {
	if isSafeMethod(request.Method) {
		return nil
	}
	if err := protection.crossOrigin.Check(request); err != nil {
		return err
	}
	values := request.Header.Values(protection.headerName)
	if len(values) != 1 || values[0] != "1" {
		return errors.New("CSRF confirmation header is missing or invalid")
	}
	return nil
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func validCSRFHeaderName(name string) bool {
	if name == "" {
		return false
	}
	for index := 0; index < len(name); index++ {
		if !isTokenCharacter(name[index]) {
			return false
		}
	}
	lower := strings.ToLower(name)
	if strings.HasPrefix(lower, "sec-") || strings.HasPrefix(lower, "proxy-") {
		return false
	}
	switch lower {
	case "accept", "accept-language", "content-language", "content-type", "range",
		"authorization", "cookie", "host", "origin", "referer", "user-agent":
		return false
	default:
		return true
	}
}

func isTokenCharacter(character byte) bool {
	if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' {
		return true
	}
	return strings.ContainsRune("!#$%&'*+-.^_`|~", rune(character))
}
