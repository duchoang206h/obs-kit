# Grafana Stack

Local development stack with OpenTelemetry Collector, Loki, Tempo, Prometheus, and Grafana.

```bash
make grafana-stack-up
make grafana-stack-logs
make grafana-stack-down
```

Compatibility aliases remain available:

```bash
make observability-up
make observability-logs
make observability-down
```

Grafana is available at `http://localhost:3001`. The collector receives OTLP on `localhost:4317` and `localhost:4318`.
