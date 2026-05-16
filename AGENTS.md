# Repository Guidelines

## Project Structure & Module Organization

This repository contains Python and Go observability SDKs. Runtime implementations live under `packages/`; shared behavior is documented in `contracts/`.

- `packages/python/`: Python SDK.
- `packages/go/`: Go SDK.
- `contracts/`: cross-language rules for env vars, config, logs, metrics, and traces.
- `deployments/local/`: local OpenTelemetry Collector, Loki, and Grafana stack.
- `examples/`: runnable framework examples when added.

Do not share implementation code across languages. Share contracts and keep each SDK idiomatic for its runtime.

## Build, Test, and Development Commands

- `make test-python`: run Python tests with `uv run pytest`.
- `make lint-python`: run Ruff and MyPy for the Python package.
- `make test-go`: run Go tests with an in-repo writable build cache.
- `make lint-go`: run `go vet ./...`.
- `make fmt-go`: format Go files with `gofmt`.
- `make observability-up`: start Collector, Loki, Tempo, Prometheus, and Grafana.
- `make example-python` / `make example-go`: emit JSON logs into `logs/`.
- `make example-python-otel`: emit a trace-correlated Python log and OTLP span.
- `make example-python-fastapi` / `make example-go-echo`: run framework examples.
- `make test`: run Python and Go tests.

Package-level commands are also valid, for example `cd packages/go && go test ./...`.

## Coding Style & Naming Conventions

Python uses Python 3.11+, Ruff, MyPy, 4-space indentation, `snake_case` functions, and `PascalCase` classes. Keep package code under `src/obs_kit/`.

Go uses `gofmt`, short lowercase package names, `PascalCase` for exported identifiers, and `camelCase` for private identifiers. Keep tests colocated as `*_test.go`.

## Testing Guidelines

Python tests live in `packages/python/tests/` and should follow `test_*.py`. Go tests should use the standard `testing` package and table-driven tests where useful. Add or update tests whenever config resolution, emitted telemetry fields, or public APIs change.

## Commit & Pull Request Guidelines

Use concise, imperative commit messages, for example `Add Go config resolver` or `Document Python log fields`. Pull requests should include a summary, test results, affected packages, and any contract changes. When behavior changes across languages, update `contracts/` before implementation.

## Security & Configuration Tips

Do not commit collector credentials, webhooks, tokens, or real API keys. Document configuration with environment variables such as `OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `ENVIRONMENT`, and `LOG_LEVEL`.
