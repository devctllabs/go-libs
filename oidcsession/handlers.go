package oidcsession

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	defaultStateTTL       = 10 * time.Minute
	defaultCSRFHeaderName = "X-CSRF-Protection"
	errorCodeLoginFailed  = "login_failed"
)

// HTTPConfig configures the framework-neutral OIDC HTTP handlers.
type HTTPConfig struct {
	UIBaseURL          *url.URL
	DefaultReturnPath  string
	ErrorPath          string
	LoginMethods       []LoginMethod
	TrustedOrigins     []string
	CSRFHeaderName     string
	CookiePrefix       string
	InsecureDevCookies bool
	StateTTL           time.Duration
	LoginAuthorizer    LoginAuthorizer
	Observer           Observer
}

type cookieNames struct {
	access  string
	session string
	login   string
}

// Handlers exposes login, callback, refresh, logout, and session endpoints for caller-owned routing.
type Handlers struct {
	application sessionApplication
	uiBaseURL   *url.URL
	errorPath   string
	cookies     cookieNames
	crossOrigin *http.CrossOriginProtection
	csrfHeader  string
	insecure    bool
}

// NewHandlers validates HTTP policy and constructs handlers without registering routes.
func NewHandlers(config HTTPConfig, provider *Provider, sessions SessionBackend, stateEncryptor Encryptor) (*Handlers, error) {
	if provider == nil || sessions == nil || stateEncryptor == nil {
		return nil, errors.New("provider, session backend, and state encryptor are required")
	}
	if config.UIBaseURL == nil || !config.UIBaseURL.IsAbs() || config.UIBaseURL.Host == "" || config.UIBaseURL.RawQuery != "" || config.UIBaseURL.Fragment != "" {
		return nil, errors.New("UI base URL must be absolute and contain no query or fragment")
	}
	if !validReturnPath(config.DefaultReturnPath) || !validReturnPath(config.ErrorPath) {
		return nil, errors.New("default return path and error path must be local absolute paths")
	}
	if !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*$`).MatchString(config.CookiePrefix) {
		return nil, errors.New("cookie prefix must contain only letters, digits, and hyphens")
	}
	if config.StateTTL < 0 {
		return nil, errors.New("state TTL must not be negative")
	}
	if config.StateTTL == 0 {
		config.StateTTL = defaultStateTTL
	}
	header := config.CSRFHeaderName
	if header == "" {
		header = defaultCSRFHeaderName
	}
	if !validConfirmationHeader(header) {
		return nil, errors.New("CSRF header name is invalid or unsafe")
	}
	protection := http.NewCrossOriginProtection()
	for _, origin := range config.TrustedOrigins {
		if err := protection.AddTrustedOrigin(origin); err != nil {
			return nil, fmt.Errorf("add trusted origin: %w", err)
		}
	}
	methods, err := cloneLoginMethods(config.LoginMethods)
	if err != nil {
		return nil, err
	}
	prefix := config.CookiePrefix + "-"
	if !config.InsecureDevCookies {
		prefix = "__Host-" + prefix
	}
	application := newSessionService(sessionServiceConfig{
		defaultReturnPath: config.DefaultReturnPath,
		stateTTL:          config.StateTTL,
		loginAuthorizer:   config.LoginAuthorizer,
		observer:          config.Observer,
		methods:           methods,
	}, provider, sessions, stateEncryptor)
	return &Handlers{
		application: application,
		uiBaseURL:   cloneURL(config.UIBaseURL),
		errorPath:   config.ErrorPath,
		cookies:     cookieNames{access: prefix + "access", session: prefix + "session", login: prefix + "login"},
		crossOrigin: protection,
		csrfHeader:  http.CanonicalHeaderKey(header),
		insecure:    config.InsecureDevCookies,
	}, nil
}

// Login starts an OIDC authorization-code flow with PKCE, nonce, and encrypted state.
func (handlers *Handlers) Login(writer http.ResponseWriter, request *http.Request) {
	if !requireMethod(writer, request, http.MethodGet) {
		return
	}
	binding, err := optionalCookie(request, handlers.cookies.login)
	if err != nil {
		handlers.redirectError(writer, request, errorCodeLoginFailed)
		return
	}
	result, err := handlers.application.BeginLogin(request.Context(), beginLoginCommand{
		methodID: request.URL.Query().Get("method"), returnPath: request.URL.Query().Get("return"), browserBinding: binding,
	})
	if result.bindingToStore != nil {
		http.SetCookie(writer, handlers.cookie(handlers.cookies.login, result.bindingToStore.value, result.bindingToStore.expiresAt))
	}
	if err != nil {
		if errors.Is(err, errUnknownLoginMethod) {
			http.Error(writer, "unknown login method", http.StatusBadRequest)
			return
		}
		code := errorCodeLoginFailed
		if errors.Is(err, ErrProviderUnavailable) {
			code = "provider_unavailable"
		}
		handlers.redirectError(writer, request, code)
		return
	}
	http.Redirect(writer, request, result.location, http.StatusFound)
}

// Callback completes OIDC verification and creates a server-side session.
func (handlers *Handlers) Callback(writer http.ResponseWriter, request *http.Request) {
	if !requireMethod(writer, request, http.MethodGet) {
		return
	}
	if request.URL.Query().Get("error") != "" {
		handlers.redirectError(writer, request, "login_cancelled")
		return
	}
	binding, err := strictCookie(request, handlers.cookies.login)
	if err != nil {
		handlers.redirectError(writer, request, errorCodeLoginFailed)
		return
	}
	result, err := handlers.application.CompleteLogin(request.Context(), completeLoginCommand{
		code: request.URL.Query().Get("code"), state: request.URL.Query().Get("state"), browserBinding: binding,
	})
	if err != nil {
		code := errorCodeLoginFailed
		switch {
		case errors.Is(err, ErrProviderUnavailable):
			code = "provider_unavailable"
		case errors.Is(err, ErrLoginDenied):
			code = "access_denied"
		}
		handlers.redirectError(writer, request, code)
		return
	}
	handlers.setCredentialCookies(writer, result.tokens.AccessToken, result.tokens.AccessExpiresAt, result.credential, result.sessionExpiresAt)
	http.Redirect(writer, request, handlers.uiLocation(result.returnPath, ""), http.StatusSeeOther)
}

// Refresh renews or reuses an access token through the server-side session.
func (handlers *Handlers) Refresh(writer http.ResponseWriter, request *http.Request) {
	if !requireMethod(writer, request, http.MethodPost) || !handlers.checkUnsafe(writer, request) {
		return
	}
	credential, err := strictCookie(request, handlers.cookies.session)
	if err != nil {
		http.Error(writer, "authentication required", http.StatusUnauthorized)
		return
	}
	refreshed, err := handlers.application.Refresh(request.Context(), SessionCredential(credential))
	if err != nil {
		if errors.Is(err, ErrInvalidSession) || errors.Is(err, ErrInvalidGrant) {
			handlers.clearCredentialCookies(writer)
			http.Error(writer, "authentication required", http.StatusUnauthorized)
			return
		}
		http.Error(writer, "authentication unavailable", http.StatusServiceUnavailable)
		return
	}
	handlers.setCredentialCookies(writer, refreshed.AccessToken, refreshed.AccessExpiresAt, SessionCredential(credential), refreshed.SessionExpiresAt)
	writeJSON(writer, refreshResponse{AccessExpiresAt: refreshed.AccessExpiresAt, SessionExpiresAt: refreshed.SessionExpiresAt})
}

// Logout invalidates the local session and clears browser credentials. Broker SSO is unchanged.
func (handlers *Handlers) Logout(writer http.ResponseWriter, request *http.Request) {
	if !requireMethod(writer, request, http.MethodPost) || !handlers.checkUnsafe(writer, request) {
		return
	}
	credential, err := strictCookie(request, handlers.cookies.session)
	handlers.clearCredentialCookies(writer)
	if err == nil {
		handlers.application.Logout(request.Context(), SessionCredential(credential))
	}
	writer.WriteHeader(http.StatusNoContent)
}

// Session returns safe session expiry metadata without exposing credentials or extending idle TTL.
func (handlers *Handlers) Session(writer http.ResponseWriter, request *http.Request) {
	if !requireMethod(writer, request, http.MethodGet) {
		return
	}
	credential, err := strictCookie(request, handlers.cookies.session)
	if err != nil {
		writeJSON(writer, sessionResponse{Authenticated: false})
		return
	}
	status, err := handlers.application.Session(request.Context(), SessionCredential(credential))
	if err != nil {
		if errors.Is(err, ErrInvalidSession) {
			handlers.clearCredentialCookies(writer)
			writeJSON(writer, sessionResponse{Authenticated: false})
			return
		}
		http.Error(writer, "authentication unavailable", http.StatusServiceUnavailable)
		return
	}
	writeJSON(writer, sessionResponse{
		Authenticated: true, AccessExpiresAt: &status.AccessExpiresAt, SessionExpiresAt: &status.SessionExpiresAt,
	})
}

type refreshResponse struct {
	AccessExpiresAt  time.Time `json:"accessExpiresAt"`
	SessionExpiresAt time.Time `json:"sessionExpiresAt"`
}

type sessionResponse struct {
	Authenticated    bool       `json:"authenticated"`
	AccessExpiresAt  *time.Time `json:"accessExpiresAt,omitempty"`
	SessionExpiresAt *time.Time `json:"sessionExpiresAt,omitempty"`
}

func (handlers *Handlers) setCredentialCookies(writer http.ResponseWriter, accessToken string, accessExpiry time.Time, credential SessionCredential, sessionExpiry time.Time) {
	http.SetCookie(writer, handlers.cookie(handlers.cookies.access, accessToken, accessExpiry))
	http.SetCookie(writer, handlers.cookie(handlers.cookies.session, string(credential), sessionExpiry))
}

func (handlers *Handlers) clearCredentialCookies(writer http.ResponseWriter) {
	for _, name := range []string{handlers.cookies.access, handlers.cookies.session} {
		//nolint:gosec // The helper applies HttpOnly, SameSite=Lax, and Secure unless explicit dev mode is configured.
		cookie := handlers.cookie(name, "", time.Unix(1, 0))
		cookie.MaxAge = -1
		http.SetCookie(writer, cookie)
	}
}

func (handlers *Handlers) cookie(name, value string, expires time.Time) *http.Cookie {
	//nolint:gosec // InsecureDevCookies is an explicit opt-in for local HTTP development only.
	return &http.Cookie{Name: name, Value: value, Path: "/", Expires: expires, Secure: !handlers.insecure, HttpOnly: true, SameSite: http.SameSiteLaxMode}
}

func (handlers *Handlers) checkUnsafe(writer http.ResponseWriter, request *http.Request) bool {
	if err := handlers.crossOrigin.Check(request); err != nil {
		http.Error(writer, "forbidden", http.StatusForbidden)
		return false
	}
	values := request.Header.Values(handlers.csrfHeader)
	if len(values) != 1 || values[0] != "1" {
		http.Error(writer, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func (handlers *Handlers) redirectError(writer http.ResponseWriter, request *http.Request, code string) {
	http.Redirect(writer, request, handlers.uiLocation(handlers.errorPath, code), http.StatusSeeOther)
}

func (handlers *Handlers) uiLocation(returnPath, errorCode string) string {
	destination := handlers.uiBaseURL.JoinPath(returnPath)
	if errorCode != "" {
		destination.RawQuery = url.Values{"error": {errorCode}}.Encode()
	}
	return destination.String()
}

func requireMethod(writer http.ResponseWriter, request *http.Request, method string) bool {
	if request.Method == method {
		return true
	}
	writer.Header().Set("Allow", method)
	http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
	return false
}

func writeJSON(writer http.ResponseWriter, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(value)
}

func optionalCookie(request *http.Request, name string) (string, error) {
	cookies := request.CookiesNamed(name)
	if len(cookies) > 1 {
		return "", errors.New("duplicate cookies")
	}
	if len(cookies) == 0 {
		return "", nil
	}
	return cookies[0].Value, nil
}

func strictCookie(request *http.Request, name string) (string, error) {
	value, err := optionalCookie(request, name)
	if err != nil || value == "" {
		return "", ErrInvalidSession
	}
	return value, nil
}

func cloneURL(value *url.URL) *url.URL {
	copy := *value
	return &copy
}

func validConfirmationHeader(name string) bool {
	if name == "" {
		return false
	}
	for index := range len(name) {
		if !isHeaderTokenCharacter(name[index]) {
			return false
		}
	}
	lower := strings.ToLower(name)
	if strings.HasPrefix(lower, "sec-") || strings.HasPrefix(lower, "proxy-") {
		return false
	}
	return !slices.Contains([]string{"accept", "accept-language", "content-language", "content-type", "range", "authorization", "cookie", "host", "origin", "referer", "user-agent"}, lower)
}

func isHeaderTokenCharacter(character byte) bool {
	return character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' ||
		strings.ContainsRune("!#$%&'*+-.^_`|~", rune(character))
}
