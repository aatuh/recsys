.PHONY: help codegen dev down build test fmt lint health

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
export VERSION
export COMMIT
export DATE

TOOLKIT ?= 1
API_TOOLKIT_PATH ?= ../api-toolkit
export API_TOOLKIT_PATH

API_SVC ?= 1
API_SVC_PATH ?= ../api-svc
export API_SVC

COMPOSE_FILES := -f docker-compose.yml
ifneq (,$(or $(filter 1,$(TOOLKIT)),$(filter 1,$(API_SVC))))
COMPOSE_FILES += -f docker-compose.toolkit.yml
endif

COMPOSE_ANSI ?= auto
COMPOSE_PROGRESS ?= plain
DOCKER_COMPOSE := docker compose --ansi $(COMPOSE_ANSI) --progress $(COMPOSE_PROGRESS)

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Start development environment
	@echo "🚀 Starting development environment (VERSION=$(VERSION), COMMIT=$(COMMIT), TOOLKIT=$(TOOLKIT))..."
	@if [ "$(TOOLKIT)" = "1" ]; then echo "🔧 Using local api-toolkit from '$(API_TOOLKIT_PATH)'"; fi
	@$(DOCKER_COMPOSE) $(COMPOSE_FILES) up -d

down: ## Clean up containers and volumes
	@echo "🧹 Cleaning up..."
	@$(DOCKER_COMPOSE) $(COMPOSE_FILES) down -v

cycle: ## Cycle the containers down and up
	@echo "♻️ Cycling development environment..."
	@$(MAKE) down
	@$(MAKE) dev

logs: ## View the logs of the service
	@$(DOCKER_COMPOSE) $(COMPOSE_FILES) logs -f

codegen: ## Generate code
	@echo "🔄 Generating code..."
	cd api && make codegen
	cd web && make codegen

build: ## Build all Docker Compose services
	@echo "🔨 Building all services..."
	@$(MAKE) down
	@$(DOCKER_COMPOSE) $(COMPOSE_FILES) build

test: ## Run tests
	@echo "🧪 Running tests..."
	cd api && make test
	cd web && make test

fmt: ## Format code
	@echo "🎨 Formatting code..."
	cd api && make fmt
	cd web && make fmt

lint: ## Lint code
	@echo "🔍 Linting code..."
	cd api && make lint
	cd web && make lint

health: ## Check service healthiness
	@echo "🏥 Checking service healthiness..."
	cd api && make health

finalize: ## Thorough validity check and generation
	@echo "✅ Finalizing code..."
	make fmt
	make lint
	make test
	make codegen
