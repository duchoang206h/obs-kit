.PHONY: test test-python test-go lint-python lint-go fmt-go example-python example-python-otel example-python-fastapi example-python-microservices-gateway example-python-microservices-inventory example-go example-go-http example-go-echo grafana-stack-up grafana-stack-down grafana-stack-logs observability-up observability-down observability-logs signoz-up signoz-down signoz-logs signoz-bridge-up signoz-bridge-down signoz-bridge-logs

GO_CACHE := /private/tmp/obs-kit-go-cache
PYTHON_ENV := /private/tmp/obs-kit-python-venv
UV_CACHE := /private/tmp/obs-kit-uv-cache
SIGNOZ_VERSION ?= v0.124.0
SIGNOZ_OTELCOL_VERSION ?= v0.144.4

test: test-python test-go

test-python:
	cd packages/python && UV_CACHE_DIR=$(UV_CACHE) UV_PROJECT_ENVIRONMENT=$(PYTHON_ENV) uv run pytest

lint-python:
	cd packages/python && UV_CACHE_DIR=$(UV_CACHE) UV_PROJECT_ENVIRONMENT=$(PYTHON_ENV) uv run ruff check . && UV_CACHE_DIR=$(UV_CACHE) UV_PROJECT_ENVIRONMENT=$(PYTHON_ENV) uv run mypy src tests

test-go:
	mkdir -p $(GO_CACHE)
	cd packages/go && GOCACHE=$(GO_CACHE) go test ./...

lint-go:
	cd packages/go && GOCACHE=$(GO_CACHE) go vet ./...

fmt-go:
	cd packages/go && gofmt -w observability

example-python:
	cd examples/python-basic && UV_CACHE_DIR=$(UV_CACHE) UV_PROJECT_ENVIRONMENT=$(PYTHON_ENV) uv run --project ../../packages/python python main.py

example-python-otel:
	cd examples/python-otel && UV_CACHE_DIR=$(UV_CACHE) UV_PROJECT_ENVIRONMENT=$(PYTHON_ENV) uv run --project ../../packages/python python main.py

example-python-fastapi:
	cd examples/python-fastapi && UV_CACHE_DIR=$(UV_CACHE) UV_PROJECT_ENVIRONMENT=$(PYTHON_ENV) uv run --project ../../packages/python --extra auto --with fastapi --with uvicorn python main.py

example-python-microservices-gateway:
	cd examples/python-microservices && UV_CACHE_DIR=$(UV_CACHE) UV_PROJECT_ENVIRONMENT=$(PYTHON_ENV) uv run --project ../../packages/python --extra auto --with fastapi --with uvicorn python gateway.py

example-python-microservices-inventory:
	cd examples/python-microservices && UV_CACHE_DIR=$(UV_CACHE) UV_PROJECT_ENVIRONMENT=$(PYTHON_ENV) uv run --project ../../packages/python --extra auto --with fastapi --with uvicorn python inventory.py

example-go:
	mkdir -p $(GO_CACHE)
	cd examples/go-basic && GOCACHE=$(GO_CACHE) go run .

example-go-http:
	mkdir -p $(GO_CACHE)
	cd examples/go-http && GOCACHE=$(GO_CACHE) go run .

example-go-echo:
	mkdir -p $(GO_CACHE)
	cd examples/go-echo && GOCACHE=$(GO_CACHE) go run .

grafana-stack-up:
	mkdir -p logs
	touch logs/python-micro-gateway.log logs/python-micro-inventory.log
	docker compose -f deployments/grafana-stack/docker-compose.yml up -d

grafana-stack-down:
	docker compose -f deployments/grafana-stack/docker-compose.yml down

grafana-stack-logs:
	docker compose -f deployments/grafana-stack/docker-compose.yml logs -f

observability-up: grafana-stack-up

observability-down: grafana-stack-down

observability-logs: grafana-stack-logs

signoz-up:
	SIGNOZ_VERSION="$(SIGNOZ_VERSION)" SIGNOZ_OTELCOL_VERSION="$(SIGNOZ_OTELCOL_VERSION)" docker compose -f deployments/signoz/docker-compose.yml up -d --remove-orphans
	$(MAKE) signoz-bridge-up

signoz-down:
	$(MAKE) signoz-bridge-down
	docker compose -f deployments/signoz/docker-compose.yml down

signoz-logs:
	docker compose -f deployments/signoz/docker-compose.yml logs -f

signoz-bridge-up:
	mkdir -p logs
	docker compose -f deployments/signoz/bridge.docker-compose.yml up -d

signoz-bridge-down:
	docker compose -f deployments/signoz/bridge.docker-compose.yml down

signoz-bridge-logs:
	docker compose -f deployments/signoz/bridge.docker-compose.yml logs -f
