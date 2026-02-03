.PHONY: help codegen dev down build test fmt lint health reuse mdlint mkdocs-serve

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
export VERSION
export COMMIT
export DATE

API_SVC ?= 1
API_SVC_PATH ?= ../api-svc
export API_SVC

COMPOSE_FILES := -f docker-compose.yml

COMPOSE_ANSI ?= auto
COMPOSE_PROGRESS ?= plain
DOCKER_COMPOSE := docker compose --ansi $(COMPOSE_ANSI) --progress $(COMPOSE_PROGRESS)

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Start development environment
	@echo "🚀 Starting development environment (VERSION=$(VERSION), COMMIT=$(COMMIT), TOOLKIT=$(TOOLKIT))..."
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

build: ## Build all Docker Compose services
	@echo "🔨 Building all services..."
	@$(MAKE) down
	@$(DOCKER_COMPOSE) $(COMPOSE_FILES) build

test: ## Run tests
	@echo "🧪 Running tests..."
	cd api && make test
	cd recsys-eval && make test
	cd recsys-pipelines && make test
	cd recsys-algo && make test

fmt: ## Format code
	@echo "🎨 Formatting code..."
	cd api && make fmt
	cd recsys-eval && make fmt
	cd recsys-pipelines && make fmt
	cd recsys-algo && make fmt

lint: ## Lint code
	@echo "🔍 Linting code..."
	cd api && make lint
	cd recsys-eval && make lint
	cd recsys-pipelines && make lint
	cd recsys-algo && make lint

mdlint: ## Lint Markdown files
	@echo "🧾 Linting Markdown..."
	npx --yes markdownlint-cli2


mkdocs: ## Serve MkDocs site locally
	@echo "📚 Serving MkDocs at http://localhost:8001 ..."
	@command -v mkdocs >/dev/null 2>&1 || { \
		echo "Installing mkdocs into .venv..."; \
		python -m venv .venv; \
		. .venv/bin/activate && python -m pip install --upgrade pip mkdocs mkdocs-material; \
	}
	@. .venv/bin/activate 2>/dev/null || true; \
	if command -v mkdocs >/dev/null 2>&1; then \
		mkdocs serve -f mkdocs.yml -a 0.0.0.0:8001; \
	else \
		. .venv/bin/activate && .venv/bin/mkdocs serve -f mkdocs.yml -a 0.0.0.0:8001; \
	fi

reuse: ## Run REUSE lint (license compliance)
	@echo "🔎 Running REUSE lint..."
	@command -v reuse >/dev/null 2>&1 || { \
		echo "Installing reuse into .venv..."; \
		python -m venv .venv; \
		. .venv/bin/activate && python -m pip install --upgrade pip reuse; \
	}
	@. .venv/bin/activate 2>/dev/null || true; \
	if command -v reuse >/dev/null 2>&1; then \
		reuse lint; \
	else \
		. .venv/bin/activate && .venv/bin/reuse lint; \
	fi

health: ## Check service healthiness
	@echo "🏥 Checking service healthiness..."
	cd api && make health

finalize: ## Thorough validity check and generation
	@echo "✅ Finalizing code..."
	make fmt
	make lint
	make test
	make codegen
	make mdlint
	make docsync