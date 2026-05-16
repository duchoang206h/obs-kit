# Environment Variables Contract

All SDKs should resolve these environment variables consistently.

| Variable | Config Field | Description |
| --- | --- | --- |
| `OTEL_SERVICE_NAME` | `serviceName` | Logical service name reported to telemetry backends. |
| `OTEL_SERVICE_VERSION` | `serviceVersion` | Application or package version. |
| `ENVIRONMENT` | `environment` | Runtime environment such as `development`, `staging`, or `production`. |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `collector.url` | Base OTLP collector endpoint. |
| `OTEL_TRACES_SAMPLER_ARG` | `tracing.sampleRate` | Trace sampling ratio from `0` to `1`. |
| `OTEL_METRICS_EXPORT_INTERVAL` | `metrics.exportInterval` | Metrics export interval in milliseconds. |
| `LOG_LEVEL` | `logging.level` | Minimum log level. |

Configuration is resolved in this order:

1. SDK defaults.
2. Environment variables.
3. Explicit code configuration.

Explicit code configuration only overrides fields that are provided. Omitted fields continue to use environment values or defaults.
