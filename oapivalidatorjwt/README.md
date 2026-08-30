# oapivalidatorjwt

`oapivalidatorjwt` is an `oapivalidator.Authenticator` for JWTs carried either as a strict
`Authorization: Bearer` credential or by an OpenAPI `apiKey` cookie security scheme. The
application still installs only `oapivalidator.New(...)` as Echo middleware; this module is the
authenticator supplied through `oapivalidator.WithAuthenticator`.

The recommended JWKS owner is [`keyfunc/v3`](https://pkg.go.dev/github.com/MicahParks/keyfunc/v3).
Construct it in the application lifecycle with a bounded HTTP client, rate-limit unknown-`kid`
refreshes, and pass its request-aware key function:

```go
keys, err := keyfunc.NewDefaultOverrideCtx(runCtx, []string{jwksURL}, keyfunc.Override{
	Client:            &http.Client{Timeout: 3 * time.Second},
	RateLimitWaitMax:  250 * time.Millisecond,
	RefreshInterval:   time.Hour,
	RefreshUnknownKID: rate.NewLimiter(rate.Every(5*time.Minute), 1),
})
if err != nil {
	return err
}

authenticator, err := oapivalidatorjwt.New(config, func(ctx context.Context) jwt.Keyfunc {
	resolve := keys.KeyfuncCtx(ctx)
	return func(token *jwt.Token) (any, error) {
		key, err := resolve(token)
		if err != nil {
			return nil, errors.Join(oapivalidator.ErrAuthenticationUnavailable, err)
		}
		return key, nil
	}
}, claimsMapper)
```

Do not log raw tokens, raw claims, or raw `kid` values. Pass the same trusted origins and CSRF
header name to this module and to `oidcsession` when browser cookies are enabled. Bearer requests
do not require CSRF confirmation.

