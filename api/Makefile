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

.PHONY: help dev down logs codegen migrate-build migrate-up migrate-down migrate-status migrate-preflight migrate-new tidy test test-cached test-fast test-clean health build-bin

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Start development environment
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
	@$(DOCKER_COMPOSE) exec api sh -c "go run ./cmd/migrate -dir \$$MIGRATIONS_DIR up"

migrate-down: ## Rollback migrations
	@$(DOCKER_COMPOSE) exec api sh -c "go run ./cmd/migrate -dir \$$MIGRATIONS_DIR --allow-down down"

migrate-status: ## Show status of the migrations
	@$(DOCKER_COMPOSE) exec api sh -c "go run ./cmd/migrate -dir \$$MIGRATIONS_DIR status"

migrate-preflight: ## Run migration preflight checks
	@$(DOCKER_COMPOSE) exec api sh -c "go run ./cmd/migrate -dir \$$MIGRATIONS_DIR preflight"

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

test: ## Run tests
	@echo "🧪 Running tests. Pattern: '$(TEST_PATTERN)'  Flags: '$(FLAGS)'  Pkg: '$(PKG)' Fast: ${FAST}"
	@$(DOCKER_COMPOSE) up -d db api
	@$(DOCKER_COMPOSE) run --no-deps --rm \
		-e TEST_PATTERN='$(TEST_PATTERN)' \
		-e FLAGS='$(FLAGS)' \
		-e PKG='$(PKG)' \
		-e FAST='$(FAST)' \
		-e SKIP_API_WAIT='$(SKIP_API_WAIT)' \
		-e TEST_CACHE_ENABLED='$(CACHE_ENABLED)' \
		-e TEST_CACHE_TTL='$(CACHE_TTL)' \
		-e TEST_CACHE_DIR='/tmp/test-cache' \
		test

test-cached: ## Run tests with caching enabled
	@echo "🧪 Running cached tests. Pattern: '$(TEST_PATTERN)'  Flags: '$(FLAGS)'  Pkg: '$(PKG)' Fast: ${FAST} Cache TTL: '$(CACHE_TTL)'"
	@$(DOCKER_COMPOSE) up -d db api
	@$(DOCKER_COMPOSE) run --no-deps --rm \
		-e TEST_PATTERN='$(TEST_PATTERN)' \
		-e FLAGS='$(FLAGS)' \
		-e PKG='$(PKG)' \
		-e FAST='$(FAST)' \
		-e SKIP_API_WAIT='$(SKIP_API_WAIT)' \
		-e TEST_CACHE_ENABLED='true' \
		-e TEST_CACHE_TTL='$(CACHE_TTL)' \
		-e TEST_CACHE_DIR='/tmp/test-cache' \
		test

test-clean: ## Clean test cache
	@echo "🧹 Cleaning test cache..."
	@$(DOCKER_COMPOSE) run --no-deps --rm test sh -c "rm -rf /tmp/test-cache .test-cache"

fmt: ## Format Go code
	@echo "🎨 Formatting Go code..."
	@$(DOCKER_COMPOSE) run --no-deps --rm api gofmt -s -w .
	@echo "✅ Code formatted successfully"

lint: ## Lint Go code
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
