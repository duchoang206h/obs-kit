# Metrics Contract

SDKs should expose a small API for business metrics while preserving OpenTelemetry compatibility.

Recommended instruments:

- Counter: monotonically increasing event count.
- Histogram: distribution of durations, sizes, or business values.
- Gauge: current value when supported by the language SDK.

Metric names should be lowercase, dot-separated, and stable, for example `orders.created` or `http.server.duration`. Attribute keys should use lowercase dot notation such as `service.name` and `deployment.environment`.

SDKs should export metrics to the configured OTLP collector endpoint. In the local stack, the collector exposes metrics through a Prometheus scrape endpoint.
