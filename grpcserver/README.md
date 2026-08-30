# grpcserver

`grpcserver` owns gRPC server construction and lifecycle. It enables reflection,
Protovalidate request validation, and generic panic recovery by default. TLS and
OpenTelemetry are explicit instance configuration; no global logger or telemetry
provider is consulted.

```go
server, err := grpcserver.New(
	grpcserver.Config{Address: ":9000"},
	grpcserver.WithTelemetry(grpcserver.Telemetry{
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
		Propagator:     propagator,
	}),
)
if err != nil {
	return err
}

examplev1.RegisterExampleServiceServer(server, service)
return server.ListenAndServe()
```

This module does not wrap generated protobuf types and does not provide a
protobuf generator or protoc plugin. Contract-owning modules should continue to
use the standard Go protobuf and gRPC plugins (and Buf when already adopted).

Caller interceptors execute in supplied order inside recovery and before the
default request validator. `Shutdown` attempts graceful completion and forces a
stop when its context expires.
