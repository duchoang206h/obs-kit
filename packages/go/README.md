# Obs Kit Go

Go SDK for the shared observability contracts.

## Development

```bash
go test ./...
go vet ./...
gofmt -w .
```

The package provides contract-aligned configuration, structured JSON logging, OpenTelemetry OTLP trace/metric export, and `net/http` middleware.

## OpenTelemetry

Use `observability.Init(ctx, config)` to configure OTLP trace and metric export. Wrap `net/http` handlers with `observability.HTTPMiddleware(config, handler)` for request spans and trace context propagation.

The Go SDK exports telemetry to the configured collector with OTLP over HTTP:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
```

The configured metrics export interval is applied to the OpenTelemetry periodic reader.

Production `net/http` shape:

```go
obs, err := observability.Init(ctx, &observability.Config{
	ServiceName:    "orders-api",
	ServiceVersion: "1.4.2",
	Environment:    "production",
	Tracing: observability.TracingConfig{
		SampleRate: observability.Float64(0.1),
	},
})
if err != nil {
	return err
}
defer obs.Shutdown(ctx)

handler := observability.HTTPMiddleware(&obs.Config, mux)
```

Use `observability.Float64(0)` when a service should intentionally disable trace sampling. The HTTP middleware extracts incoming W3C trace context and preserves optional `http.ResponseWriter` interfaces such as `http.Flusher`, `http.Hijacker`, and `http.Pusher` when the underlying writer supports them.
