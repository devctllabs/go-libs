# grpcclient

`grpcclient` constructs non-blocking gRPC client connections from resolver
targets. It does not add retries, deadlines, or blocking dial behavior.

```go
conn, err := grpcclient.New(grpcclient.Config{
	Target: "dns:///catalog.default.svc.cluster.local:9000",
	TLS: grpcclient.TLSConfig{
		Enabled:    true,
		RootCAFile: "/var/run/secrets/catalog-ca.pem",
		ServerName: "catalog.default.svc.cluster.local",
	},
})
if err != nil {
	return err
}
defer conn.Close()
```

With TLS enabled, an empty root CA file uses operating-system roots. Client
certificate and key files are optional but must be supplied together. The caller
owns the returned connection and all per-RPC deadlines and retry policy.
