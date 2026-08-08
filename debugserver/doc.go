// Package debugserver provides a standalone HTTP server for Go pprof endpoints.
//
// The server listens on 127.0.0.1:6060 by default and serves diagnostics from
// a private http.ServeMux. Keep it bound to a loopback or otherwise private
// address and access it through an authenticated operational channel such as
// kubectl port-forward. Do not expose it through a public Service or Ingress.
//
// Importing net/http/pprof registers the same handlers on http.DefaultServeMux
// as a package initialization side effect. Package debugserver does not serve
// that mux, but applications must not expose http.DefaultServeMux publicly.
//
// Mutex and block profiling rates are process-global runtime settings. This
// package does not change them; applications that need those profiles must set
// and own that policy explicitly.
package debugserver
