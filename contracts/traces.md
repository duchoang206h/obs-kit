# Traces Contract

SDKs should initialize OpenTelemetry tracing early in application startup and preserve context across supported framework boundaries.

Required resource attributes:

- `service.name`
- `service.version`
- `deployment.environment`

HTTP framework adapters should attach route, method, status code, and error information when available. Health-check and metrics endpoints should be configurable as ignored paths.

Database client spans should follow OpenTelemetry semantic conventions where the runtime instrumentation supports them. At minimum, helper spans should attach:

- `db.system`: database system such as `postgresql`, `mysql`, `sqlite`, or `mongodb`.
- `db.operation.name`: operation name such as `SELECT`, `INSERT`, or `find`.
- `db.namespace`: database name, schema, or logical namespace when known.
- `db.query.text`: query text only when it is safe to record and does not contain secrets or high-cardinality values.

Redis client spans should attach:

- `db.system`: `redis`.
- `db.operation.name`: command name such as `GET`, `SET`, or `HGETALL`.
- `db.namespace`: database index or logical namespace when known.

SDKs should export spans to the configured OTLP collector endpoint. In the Grafana stack, the collector forwards traces to Tempo.
