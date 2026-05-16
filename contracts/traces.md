# Traces Contract

SDKs should initialize OpenTelemetry tracing early in application startup and preserve context across supported framework boundaries.

Required resource attributes:

- `service.name`
- `service.version`
- `deployment.environment`

HTTP framework adapters should attach route, method, status code, and error information when available. Health-check and metrics endpoints should be configurable as ignored paths.

SDKs should export spans to the configured OTLP collector endpoint. In the local stack, the collector forwards traces to Tempo.
