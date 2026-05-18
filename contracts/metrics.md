# Metrics Contract

SDKs should expose a small API for business metrics while preserving OpenTelemetry compatibility.

Recommended instruments:

- Counter: monotonically increasing event count.
- Histogram: distribution of durations, sizes, or business values.
- Gauge: current value when supported by the language SDK.

Metric names should be lowercase, dot-separated, and stable, for example `orders.created` or `http.server.duration`. Attribute keys should use lowercase dot notation such as `service.name` and `deployment.environment`.

Recommended client metrics:

- `db.client.operation.duration`: histogram of database operation duration in milliseconds.
- `redis.client.operation.duration`: histogram of Redis command duration in milliseconds.

Recommended attributes:

- `db.system`: database system such as `postgresql`, `mysql`, `sqlite`, `mongodb`, or `redis`.
- `db.operation.name`: operation or command name.
- `db.namespace`: database name, schema, logical namespace, or Redis database index when known.
- `error.type`: error type when the operation fails.

SDKs should export metrics to the configured OTLP collector endpoint. In the Grafana stack, the collector exposes metrics through a Prometheus scrape endpoint.
