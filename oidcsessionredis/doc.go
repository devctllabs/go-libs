// Package oidcsessionredis stores encrypted OIDC token sets behind opaque session credentials.
//
// Refresh uses a bounded Redis lease so replicas normally perform one provider refresh. A process
// crash after provider-side refresh-token rotation but before Redis commit fails the session closed
// and requires login again.
package oidcsessionredis
