# Python OTLP Example

Run with the Grafana stack:

```bash
make grafana-stack-up
make example-python-otel
```

This example creates a span, writes a correlated JSON log to `logs/python-otel.log`, and exports traces/metrics to the local OpenTelemetry Collector.
