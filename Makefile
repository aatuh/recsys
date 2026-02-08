.PHONY: help codegen dev down build test fmt lint health reuse mdlint codespell mkdocs-serve

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
	$(MAKE) down
	$(MAKE) dev

logs: ## View the logs of the service
	@$(DOCKER_COMPOSE) $(COMPOSE_FILES) logs -f

codegen: ## Generate code
	@echo "🔄 Generating code..."
	cd api && $(MAKE) codegen

build: ## Build all Docker Compose services
	@echo "🔨 Building all services..."
	$(MAKE) down
	@$(DOCKER_COMPOSE) $(COMPOSE_FILES) build

test: ## Run tests
	@echo "🧪 Running tests..."
	cd api && $(MAKE) test
	cd recsys-eval && $(MAKE) test
	cd recsys-pipelines && $(MAKE) test
	cd recsys-algo && $(MAKE) test

fmt: ## Format code
	@echo "🎨 Formatting code..."
	cd api && $(MAKE) fmt
	cd recsys-eval && $(MAKE) fmt
	cd recsys-pipelines && $(MAKE) fmt
	cd recsys-algo && $(MAKE) fmt

lint: ## Lint code
	@echo "🔍 Linting code..."
	cd api && $(MAKE) lint
	cd recsys-eval && $(MAKE) lint
	cd recsys-pipelines && $(MAKE) lint
	cd recsys-algo && $(MAKE) lint

mdlint: ## Lint Markdown files
	@echo "🧾 Linting Markdown..."
	npx --yes markdownlint-cli2@0.20.0

codespell: ## Spell check docs (codespell)
	@if [ ! -x .venv/bin/codespell ]; then \
		echo "Installing codespell into .venv..."; \
		python -m venv .venv; \
		. .venv/bin/activate && python -m pip install --disable-pip-version-check codespell; \
	fi
	@. .venv/bin/activate && .venv/bin/codespell --config .codespellrc.ini docs mkdocs.yml README.md AGENTS.md

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
	cd api && $(MAKE) health

finalize: ## Thorough validity check and generation
	@echo "✅ Finalizing code..."
	$(MAKE) fmt
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) codegen
	$(MAKE) mdlint
	$(MAKE) docs-check

.PHONY: docs-serve docs-build docs-check

DOCS_PIP_PACKAGES := mkdocs mkdocs-swagger-ui-tag mkdocs-material pymdown-extensions \
	mkdocs-git-revision-date-localized-plugin mkdocs-git-authors-plugin \
	mkdocs-redirects mkdocs-minify-plugin mkdocs-glightbox \
	pillow cairosvg

.PHONY: docs-deps

docs-deps: ## Install MkDocs and required plugins into .venv
	@if [ ! -d .venv ]; then \
		echo "Installing docs dependencies into .venv..."; \
		python -m venv .venv; \
	fi
	@. .venv/bin/activate && python -m pip install --disable-pip-version-check $(DOCS_PIP_PACKAGES)

docs-serve: ## Serve MkDocs site locally
	$(MAKE) mdlint
	$(MAKE) docs-check
	@echo "📚 Serving MkDocs at http://localhost:8001 ..."
	@command -v mkdocs >/dev/null 2>&1 || { \
		if [ ! -x .venv/bin/mkdocs ]; then \
			echo "Installing mkdocs into .venv..."; \
			python -m venv .venv; \
			. .venv/bin/activate && python -m pip install --disable-pip-version-check mkdocs mkdocs-swagger-ui-tag mkdocs-material pymdown-extensions; \
		fi; \
	}
	@. .venv/bin/activate 2>/dev/null || true; \
	if command -v mkdocs >/dev/null 2>&1; then \
		mkdocs serve -f mkdocs.yml -a 0.0.0.0:8001; \
	else \
		. .venv/bin/activate && .venv/bin/mkdocs serve -f mkdocs.yml -a 0.0.0.0:8001; \
	fi


docs-build: ## Build MkDocs site (strict)
	$(MAKE) docs-deps
	@command -v mkdocs >/dev/null 2>&1 || { \
		if [ ! -x .venv/bin/mkdocs ]; then \
			echo "Installing mkdocs into .venv..."; \
			python -m venv .venv; \
			. .venv/bin/activate && python -m pip install --disable-pip-version-check $(DOCS_PIP_PACKAGES); \
		fi; \
	}
	@. .venv/bin/activate 2>/dev/null || true; \
	if command -v mkdocs >/dev/null 2>&1; then \
		mkdocs build --strict -f mkdocs.yml -d .site; \
	else \
		. .venv/bin/activate && .venv/bin/mkdocs build --strict -f mkdocs.yml -d .site; \
	fi


docs-check: ## Check docs internal links + strict MkDocs build
	python3 scripts/docs_linkcheck.py
	python3 scripts/docs_external_linkcheck.py
	$(MAKE) codespell
	$(MAKE) docs-build
