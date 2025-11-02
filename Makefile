.PHONY: help codegen dev build test lint fmt typecheck uilint uitypecheck prod-build prod-run env

PROFILE ?= dev
SINCE ?= 24h

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

codegen: ## Generate code
	@echo "🔄 Generating code..."
	@cd api && make swag
	@cp api/swagger/swagger.json swagger-service/public/swagger.json
	@cp api/swagger/swagger.yaml swagger-service/public/swagger.yaml
	@cd web && pnpm run codegen:api:local && pnpm run sync:readme
	@cd shop && pnpm run codegen:api:local
	@cd demo && pnpm run codegen:api:local
	@echo "🔄 Code generated successfully"

dev: ## Start development environment
	@echo "🚀 Starting development environment..."
	@docker-compose up

down: ## Clean up containers and volumes
	@echo "🧹 Cleaning up..."
	@docker-compose down -v

build: ## Build all services
	@echo "🔨 Building all services..."
	@docker-compose build

env: ## Generate .env/.env.test from profile (PROFILE=dev|test|prod)
	@profile=$(PROFILE); \
	case "$$profile" in \
		dev) src="api/env/dev.env"; dst="api/.env";; \
		test) src="api/env/test.env"; dst="api/.env.test";; \
		prod) src="api/env/prod.env"; dst="api/.env";; \
		*) echo "Unknown PROFILE '$$profile'. Use dev, test, or prod."; exit 1;; \
	esac; \
	cp "$$src" "$$dst"; \
	echo "Generated $$dst from $$src"

test: ## Run tests
	@echo "🧪 Running tests..."
	@cd api && make test

lint: ## Lint API and UI
	@echo "🔍 Linting UI..."
	@cd web && pnpm run lint

fmt: ## Format UI (eslint --fix)
	@echo "🧼 Formatting UI..."
	@cd web && pnpm run lint:fix || true

typecheck: ## Typecheck UI
	@echo "🔎 Typechecking UI..."
	@cd web && pnpm run typecheck

prod-build: ## Build production Docker image
	@echo "🏗️  Building production image..."
	@cd web && docker build -f Dockerfile.prod -t recsys-web:prod .
	@echo "✅ Production image built: recsys-web:prod"

prod-run: ## Run production environment
	@echo "🚀 Starting production environment..."
	@docker-compose -f docker-compose.yml -f web/docker-compose.prod.yml up

health:
	@echo "🏥 Checking service healthiness..."
	@cd api && make health

catalog-backfill: ## Backfill catalog metadata and refresh embeddings
	@echo "📦 Running catalog backfill..."
	@cd api && go run ./cmd/catalog_backfill --mode backfill --batch 200

catalog-refresh: ## Refresh catalog metadata since duration (use SINCE=24h or RFC3339)
	@echo "♻️ Refreshing catalog metadata..."
	@cd api && go run ./cmd/catalog_backfill --mode refresh --batch 200 --since $(SINCE)
