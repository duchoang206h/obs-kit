# Logs Contract

Structured logs should use consistent field names across SDKs.

Required fields:

- `level`: log severity.
- `message`: human-readable log message.
- `timestamp`: event time in ISO 8601 format when supported by the logger.
- `service.name`: resolved service name.
- `service.version`: resolved service version.
- `deployment.environment`: resolved deployment environment.

Trace correlation fields:

- `trace_id`: active trace identifier.
- `span_id`: active span identifier.

Error logs should serialize errors into fields that preserve the message, type, and stack trace when available. SDKs must not log secrets.

Default SDK log output is JSON on stdout/stderr. Local examples can write newline-delimited JSON into `logs/*.log`; the OpenTelemetry Collector filelog receiver tails those files and forwards records to Loki.

For production, prefer stdout JSON logs and let the runtime platform or collector agent capture container logs.
