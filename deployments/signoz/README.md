# SigNoz Deployment

This local-development deployment vendors the SigNoz Docker Compose services and required config files under `deployments/signoz/`. It also includes a separate obs-kit bridge collector compose file for examples that write JSON logs to `logs/*.log`.

## Start SigNoz

```bash
make signoz-up
```

By default this starts `deployments/signoz/docker-compose.yml` with SigNoz, ClickHouse, ZooKeeper, and the SigNoz OpenTelemetry Collector. It also starts `deployments/signoz/bridge.docker-compose.yml` for the obs-kit bridge collector.

To pin a SigNoz release:

```bash
make signoz-up SIGNOZ_VERSION=v0.124.0 SIGNOZ_OTELCOL_VERSION=v0.144.4
```

For any shared environment, set a real JWT secret:

```bash
SIGNOZ_TOKENIZER_JWT_SECRET="$(openssl rand -hex 32)" make signoz-up
```

SigNoz UI is available at `http://localhost:8080`. Its OpenTelemetry collector listens on:

- OTLP/gRPC: `http://localhost:4317`
- OTLP/HTTP: `http://localhost:4318`

## Local File Log Bridge

The current examples write JSON logs into `logs/*.log` so local collectors can tail them. Start the bridge collector when you want those file logs in SigNoz too:

```bash
make signoz-bridge-up
```

`make signoz-up` already starts the bridge. Use `make signoz-bridge-up` only when SigNoz is already running and you want to start or restart just the bridge collector. The bridge compose file has no dependency on the local SigNoz compose, so it can point at a remote SigNoz collector.

The bridge collector receives telemetry on alternate ports and forwards it to SigNoz:

- OTLP/gRPC: `http://localhost:14317`
- OTLP/HTTP: `http://localhost:14318`

Use this endpoint for examples when the bridge is running:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:14318
```

If SigNoz is running on another host, point the bridge at its collector:

```bash
SIGNOZ_OTLP_GRPC_ENDPOINT=signoz.example.com:4317 make signoz-bridge-up
```

## Stop

```bash
make signoz-bridge-down
make signoz-down
```
