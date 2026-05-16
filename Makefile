.PHONY: test test-python test-go lint-python lint-go fmt-go example-python example-python-otel example-python-fastapi example-python-microservices-gateway example-python-microservices-inventory example-go example-go-http example-go-echo observability-up observability-down observability-logs

GO_CACHE := /private/tmp/obs-kit-go-cache
PYTHON_ENV := /private/tmp/obs-kit-python-venv
UV_CACHE := /private/tmp/obs-kit-uv-cache

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

observability-up:
	mkdir -p logs
	touch logs/python-micro-gateway.log logs/python-micro-inventory.log
	docker compose -f deployments/local/docker-compose.yml up -d

observability-down:
	docker compose -f deployments/local/docker-compose.yml down

observability-logs:
	docker compose -f deployments/local/docker-compose.yml logs -f
