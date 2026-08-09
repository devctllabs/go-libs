// Package healthserver exposes health probes through an OpenAPI-generated Echo 5 transport.
//
// Register mounts /livez and /readyz on an existing Echo instance. Existing global middleware
// still applies, so callers must ensure authentication does not block Kubernetes probes. NewServer
// creates a standalone server on :8081 with bounded HTTP timeouts. It does not install signal
// handlers or start goroutines; the application owns Serve or ListenAndServe and calls Shutdown
// explicitly.
//
// Probe clients should use HTTP status codes. Responses contain only {"status":"ok|fail"};
// /readyz?verbose=true additionally returns safe component names, statuses, and criticality.
// Internal checker errors are never serialized. Before the management listener is available, a
// refused connection from /livez is the startup failure signal.
//
// A Kubernetes container can target the standalone server as follows:
//
//	startupProbe:
//	  httpGet: {path: /livez, port: 8081}
//	  timeoutSeconds: 2
//	livenessProbe:
//	  httpGet: {path: /livez, port: 8081}
//	  timeoutSeconds: 2
//	readinessProbe:
//	  httpGet: {path: /readyz, port: 8081}
//	  timeoutSeconds: 2
//
// Keep the management port out of public Services and Ingresses. Tune startup failure thresholds
// to the management listener's initialization budget.
package healthserver
