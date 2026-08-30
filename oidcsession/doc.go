// Package oidcsession implements explicit OIDC login flows and browser session handlers.
//
// Provider discovery has an explicit Run lifecycle. Browser access tokens and opaque refresh
// credentials are stored in separate host-only cookies. Applications remain responsible for
// resource-level authorization and post-login onboarding.
package oidcsession
