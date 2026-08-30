# oidcsession

`oidcsession` provides an explicit OIDC authorization-code flow for same-site browser applications:

- `NewProvider` performs local validation only; run `Provider.Run(ctx)` as a lifecycle task so the
  service may start degraded and recover when discovery succeeds.
- `NewHandlers` returns framework-neutral `net/http` handlers. The application owns route paths,
  CORS, request deadlines, and server lifecycle.
- Login state uses an AES-256-GCM envelope containing PKCE, nonce, return path, login method, expiry,
  and a digest of a stable browser-binding cookie. Use a dedicated state encryption key.
- The access cookie contains the provider JWT. The session cookie contains only an opaque
  `SessionCredential`; a `SessionBackend` owns refresh tokens.

Register `Login`, `Callback`, `Refresh`, `Logout`, and `Session` at application-chosen paths.
Refresh and logout require the configured CSRF header with value `1` and pass Go's
`http.CrossOriginProtection`. Production cookies are host-only `__Host-...`, Secure, HttpOnly, and
SameSite=Lax. `InsecureDevCookies` is an explicit local-HTTP opt-in.

The browser-binding cookie ties each encrypted login state to the browser that initiated the flow.
This prevents a valid state created in one browser from being replayed as login CSRF in another;
the cookie is short-lived and contains neither provider tokens nor a session credential.

`LoginAuthorizer` is only a quick post-verification login admission decision. Provisioning and
onboarding remain application flows after redirect, and current authorization/block status must
still be checked for every API request. Local logout does not end the upstream broker SSO session;
without an access-token blacklist, an already issued short-lived JWT remains valid until expiry.

For encryption-key rotation, deploy `[new, old]` so new envelopes use the new key while old ones
remain readable, wait out the maximum envelope/session lifetime, then remove the old key. Use
separate keyrings for login state and persisted provider tokens.
