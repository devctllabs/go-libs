# healthgrpc

`healthgrpc` exposes a concrete `health.Probes` instance through the standard
gRPC Health protocol. `Check` always evaluates fresh probe state; `Watch` uses
the latest periodically published state and coalesces stale updates for slow
consumers.

```go
healthService, err := healthgrpc.New(healthgrpc.Config{}, probes)
if err != nil {
	return err
}
grpc_health_v1.RegisterHealthServer(grpcServer, healthService)

go healthService.Run(ctx)
```

The empty service name and `readiness` are aggregate readiness. `liveness` is
process liveness. The initial watched state is `NOT_SERVING`; `Run` evaluates
immediately and then every five seconds by default. During shutdown, stop this
health service before stopping the application gRPC server so watchers receive
the terminal `NOT_SERVING` state.
