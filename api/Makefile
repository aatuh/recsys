SHELL := /bin/bash

BIN ?= bin/recsys-service
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS ?= -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

API_SVC ?= 1
API_SVC_PATH ?= ../api-svc
export API_SVC

COMPOSE_FILES := -f ../docker-compose.yml
DOCKER_COMPOSE := docker compose $(COMPOSE_FILES)

.PHONY: help env test-env dev down logs codegen migrate-build migrate-up migrate-down migrate-status migrate-preflight migrate-new tidy test test-cached test-fast test-clean health build-bin

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

env: ## Create local API env file if missing
	@if [ -f .env ]; then \
		echo "api/.env already exists; leaving it unchanged"; \
	elif [ -f .env.example ]; then \
		cp .env.example .env; \
		echo "Created api/.env from api/.env.example"; \
	else \
		echo "missing required example env file: api/.env.example" >&2; \
		exit 1; \
	fi

test-env: ## Create local API test env file if missing
	@if [ -f .env.test ]; then \
		echo "api/.env.test already exists; leaving it unchanged"; \
	elif [ -f .env.test.example ]; then \
		cp .env.test.example .env.test; \
		echo "Created api/.env.test from api/.env.test.example"; \
	else \
		echo "missing required example env file: api/.env.test.example" >&2; \
		exit 1; \
	fi

dev: env ## Start development environment
	@$(DOCKER_COMPOSE) --env-file .env up api --build -d

down: ## Clean up containers and volumes
	@$(DOCKER_COMPOSE) down api -v

cycle: ## Cycle the containers down and up
	@make down
	@make dev

logs: ## View the logs of the service
	@$(DOCKER_COMPOSE) logs -f api

OPENAPI_SOURCE ?= ../docs/reference/api/openapi.yaml
OPENAPI_OUT ?= ./docs

codegen: ## Generate OpenAPI artifacts from the canonical spec
	@go run ./cmd/openapi-sync -input $(OPENAPI_SOURCE) -out-dir $(OPENAPI_OUT)

migrate-up: ## Apply migrations
	@$(DOCKER_COMPOSE) exec api sh -c 'dir_arg=""; \
		if [ -n "$$MIGRATIONS_DIR" ] && [ "$$MIGRATIONS_DIR" != "-" ]; then \
			dir_arg="-dir $$MIGRATIONS_DIR"; \
		fi; \
		GOTMPDIR=/app/tmp TMPDIR=/app/tmp go run ./cmd/migrate $$dir_arg up'

migrate-down: ## Rollback migrations
	@$(DOCKER_COMPOSE) exec api sh -c 'dir_arg=""; \
		if [ -n "$$MIGRATIONS_DIR" ] && [ "$$MIGRATIONS_DIR" != "-" ]; then \
			dir_arg="-dir $$MIGRATIONS_DIR"; \
		fi; \
		GOTMPDIR=/app/tmp TMPDIR=/app/tmp go run ./cmd/migrate $$dir_arg --allow-down down'

migrate-status: ## Show status of the migrations
	@$(DOCKER_COMPOSE) exec api sh -c 'dir_arg=""; \
		if [ -n "$$MIGRATIONS_DIR" ] && [ "$$MIGRATIONS_DIR" != "-" ]; then \
			dir_arg="-dir $$MIGRATIONS_DIR"; \
		fi; \
		GOTMPDIR=/app/tmp TMPDIR=/app/tmp go run ./cmd/migrate $$dir_arg status'

migrate-preflight: ## Run migration preflight checks
	@$(DOCKER_COMPOSE) exec api sh -c 'dir_arg=""; \
		if [ -n "$$MIGRATIONS_DIR" ] && [ "$$MIGRATIONS_DIR" != "-" ]; then \
			dir_arg="-dir $$MIGRATIONS_DIR"; \
		fi; \
		GOTMPDIR=/app/tmp TMPDIR=/app/tmp go run ./cmd/migrate $$dir_arg preflight'

tidy: ## Tidy go modules
	@$(DOCKER_COMPOSE) exec api go mod tidy

build-bin: ## Build local API binary with version metadata
	@echo "🏗️ Building API binary '$(BIN)'"
	@mkdir -p $(dir $(BIN))
	@go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/recsys-service
	@echo "✅ Build complete"

TEST_PATTERN ?=
FLAGS ?=
PKG ?= ./...
FAST ?= false
SKIP_API_WAIT ?= false
CACHE_TTL ?= 3600
CACHE_ENABLED ?= false
TEST_ENV_FILE ?= ./api/.env.test

test: test-env ## Run tests
	@echo "🧪 Running tests. Pattern: '$(TEST_PATTERN)'  Flags: '$(FLAGS)'  Pkg: '$(PKG)' Fast: ${FAST}"
	@RECSYS_API_ENV_FILE=$(TEST_ENV_FILE) RECSYS_TEST_ENV_FILE=$(TEST_ENV_FILE) $(DOCKER_COMPOSE) up -d db api
	@RECSYS_API_ENV_FILE=$(TEST_ENV_FILE) RECSYS_TEST_ENV_FILE=$(TEST_ENV_FILE) $(DOCKER_COMPOSE) exec -T api sh -c 'dir_arg=""; \
		if [ -n "$$MIGRATIONS_DIR" ] && [ "$$MIGRATIONS_DIR" != "-" ]; then \
			dir_arg="-dir $$MIGRATIONS_DIR"; \
		fi; \
		GOTMPDIR=/app/tmp TMPDIR=/app/tmp go run ./cmd/migrate $$dir_arg up'
	@RECSYS_API_ENV_FILE=$(TEST_ENV_FILE) RECSYS_TEST_ENV_FILE=$(TEST_ENV_FILE) $(DOCKER_COMPOSE) run --build --no-deps --rm \
		-e TEST_PATTERN='$(TEST_PATTERN)' \
		-e FLAGS='$(FLAGS)' \
		-e PKG='$(PKG)' \
		-e FAST='$(FAST)' \
		-e SKIP_API_WAIT='$(SKIP_API_WAIT)' \
		-e TEST_CACHE_ENABLED='$(CACHE_ENABLED)' \
		-e TEST_CACHE_TTL='$(CACHE_TTL)' \
		-e TEST_CACHE_DIR='/tmp/test-cache' \
		test

test-cached: test-env ## Run tests with caching enabled
	@echo "🧪 Running cached tests. Pattern: '$(TEST_PATTERN)'  Flags: '$(FLAGS)'  Pkg: '$(PKG)' Fast: ${FAST} Cache TTL: '$(CACHE_TTL)'"
	@RECSYS_API_ENV_FILE=$(TEST_ENV_FILE) RECSYS_TEST_ENV_FILE=$(TEST_ENV_FILE) $(DOCKER_COMPOSE) up -d db api
	@RECSYS_API_ENV_FILE=$(TEST_ENV_FILE) RECSYS_TEST_ENV_FILE=$(TEST_ENV_FILE) $(DOCKER_COMPOSE) exec -T api sh -c 'dir_arg=""; \
		if [ -n "$$MIGRATIONS_DIR" ] && [ "$$MIGRATIONS_DIR" != "-" ]; then \
			dir_arg="-dir $$MIGRATIONS_DIR"; \
		fi; \
		GOTMPDIR=/app/tmp TMPDIR=/app/tmp go run ./cmd/migrate $$dir_arg up'
	@RECSYS_API_ENV_FILE=$(TEST_ENV_FILE) RECSYS_TEST_ENV_FILE=$(TEST_ENV_FILE) $(DOCKER_COMPOSE) run --build --no-deps --rm \
		-e TEST_PATTERN='$(TEST_PATTERN)' \
		-e FLAGS='$(FLAGS)' \
		-e PKG='$(PKG)' \
		-e FAST='$(FAST)' \
		-e SKIP_API_WAIT='$(SKIP_API_WAIT)' \
		-e TEST_CACHE_ENABLED='true' \
		-e TEST_CACHE_TTL='$(CACHE_TTL)' \
		-e TEST_CACHE_DIR='/tmp/test-cache' \
		test

test-clean: test-env ## Clean test cache
	@echo "🧹 Cleaning test cache..."
	@RECSYS_API_ENV_FILE=$(TEST_ENV_FILE) RECSYS_TEST_ENV_FILE=$(TEST_ENV_FILE) $(DOCKER_COMPOSE) run --no-deps --rm test sh -c "rm -rf /tmp/test-cache .test-cache"

fmt: env ## Format Go code
	@echo "🎨 Formatting Go code..."
	@$(DOCKER_COMPOSE) run --no-deps --rm api gofmt -s -w .
	@echo "✅ Code formatted successfully"

lint: env ## Lint Go code
	@echo "🔍 Linting Go code..."
	$(DOCKER_COMPOSE) run --no-deps --rm api sh -c "GOLANGCI_LINT_CACHE=/tmp/golangci-lint go vet ./... && GOLANGCI_LINT_CACHE=/tmp/golangci-lint golangci-lint run"
	@echo "✅ Code linted successfully"

health: ## Check service healthiness (logs + curl)
	@echo "🏥 Checking service healthiness..."
	@echo "📋 Docker Compose logs (last 20 lines):"
	@$(DOCKER_COMPOSE) logs --tail=20 api
	@echo ""
	@echo "🌐 Health endpoint check:"
	@curl -f http://localhost:8000/health || echo "❌ Health check failed"
	@echo ""
	@echo "✅ Health check completed"

finalize: ## Thorough validity check and generation
	@echo "✅ Finalizing code..."
	make fmt
	make lint
	make test
	make codegen
