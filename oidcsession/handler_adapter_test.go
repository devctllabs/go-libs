package oidcsession

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLoginMapsHTTPRequestAndServiceResult(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	application := NewMocksessionApplication(controller)
	expiresAt := time.Now().Add(time.Minute)
	application.EXPECT().BeginLogin(gomock.Any(), beginLoginCommand{
		methodID: "sso", returnPath: "/projects/", browserBinding: "",
	}).Return(beginLoginResult{
		location: "https://issuer.example/authorize", bindingToStore: &browserBinding{value: "binding", expiresAt: expiresAt},
	}, nil)
	handlers := testHTTPAdapter(application)
	request := httptest.NewRequest(http.MethodGet, "https://api.example/login?method=sso&return=/projects/", nil)
	response := httptest.NewRecorder()

	handlers.Login(response, request)

	require.Equal(t, http.StatusFound, response.Code)
	require.Equal(t, "https://issuer.example/authorize", response.Header().Get("Location"))
	cookies := response.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "test-login", cookies[0].Name)
	require.Equal(t, "binding", cookies[0].Value)
	require.WithinDuration(t, expiresAt, cookies[0].Expires, time.Second)
}

func TestLoginDoesNotReplaceExistingBrowserBinding(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	application := NewMocksessionApplication(controller)
	application.EXPECT().BeginLogin(gomock.Any(), beginLoginCommand{
		methodID: "", returnPath: "", browserBinding: "existing-binding",
	}).Return(beginLoginResult{location: "https://issuer.example/authorize"}, nil)
	handlers := testHTTPAdapter(application)
	request := httptest.NewRequest(http.MethodGet, "https://api.example/login", nil)
	request.AddCookie(&http.Cookie{Name: "test-login", Value: "existing-binding"})
	response := httptest.NewRecorder()

	handlers.Login(response, request)

	require.Equal(t, http.StatusFound, response.Code)
	require.Empty(t, response.Result().Cookies())
}

func TestCallbackRejectsMissingBrowserBinding(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	handlers := testHTTPAdapter(NewMocksessionApplication(controller))
	request := httptest.NewRequest(http.MethodGet, "https://api.example/callback?code=code&state=state", nil)
	response := httptest.NewRecorder()

	handlers.Callback(response, request)

	require.Equal(t, http.StatusSeeOther, response.Code)
	require.Equal(t, "https://ui.example/app/error?error=login_failed", response.Header().Get("Location"))
}

func TestUILocationJoinsBaseAndPreservesTrailingSlash(t *testing.T) {
	t.Parallel()
	handlers := testHTTPAdapter(nil)
	tests := []struct {
		name       string
		returnPath string
		errorCode  string
		expected   string
	}{
		{name: "root", returnPath: "/", expected: "https://ui.example/app/"},
		{name: "path", returnPath: "/projects", expected: "https://ui.example/app/projects"},
		{name: "trailing slash", returnPath: "/projects/", expected: "https://ui.example/app/projects/"},
		{name: "error", returnPath: "/error", errorCode: "login_failed", expected: "https://ui.example/app/error?error=login_failed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, test.expected, handlers.uiLocation(test.returnPath, test.errorCode))
		})
	}
}

func TestRefreshMapsInvalidSessionAndClearsCredentialCookies(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	application := NewMocksessionApplication(controller)
	application.EXPECT().Refresh(gomock.Any(), SessionCredential("credential")).Return(RefreshSessionResult{}, ErrInvalidGrant)
	handlers := testHTTPAdapter(application)
	request := httptest.NewRequest(http.MethodPost, "https://api.example/refresh", nil)
	request.Header.Set(defaultCSRFHeaderName, "1")
	request.AddCookie(&http.Cookie{Name: "test-session", Value: "credential"})
	response := httptest.NewRecorder()

	handlers.Refresh(response, request)

	require.Equal(t, http.StatusUnauthorized, response.Code)
	cookies := response.Result().Cookies()
	require.Len(t, cookies, 2)
	require.Equal(t, "test-access", cookies[0].Name)
	require.Equal(t, -1, cookies[0].MaxAge)
	require.Equal(t, "test-session", cookies[1].Name)
	require.Equal(t, -1, cookies[1].MaxAge)
}

func TestValidConfirmationHeader(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		header   string
		expected bool
	}{
		{name: "custom token", header: "X-CSRF-Protection", expected: true},
		{name: "invalid token character", header: "Bad Header", expected: false},
		{name: "browser controlled", header: "Sec-Fetch-Site", expected: false},
		{name: "credential header", header: "Authorization", expected: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, test.expected, validConfirmationHeader(test.header))
		})
	}
}

func testHTTPAdapter(application sessionApplication) *Handlers {
	base, err := url.Parse("https://ui.example/app/")
	if err != nil {
		panic(err)
	}
	return &Handlers{
		application: application,
		uiBaseURL:   base,
		errorPath:   "/error",
		cookies:     cookieNames{access: "test-access", session: "test-session", login: "test-login"},
		crossOrigin: http.NewCrossOriginProtection(),
		csrfHeader:  defaultCSRFHeaderName,
		insecure:    true,
	}
}
