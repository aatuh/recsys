.PHONY: help codegen dev build test lint fmt typecheck uilint uitypecheck prod-build prod-run env

PROFILE ?= dev
SINCE ?= 24h
NAMESPACE ?= default
SCENARIO_BASE_URL ?= http://localhost:8000
SCENARIO_ORG_ID ?= 00000000-0000-0000-0000-000000000001
SCENARIO_NAMESPACE ?= default
RESET_NAMESPACE ?= $(SCENARIO_NAMESPACE)
RESTART_BASE_URL ?= http://localhost:8000
S7_MIN_AVG_MRR ?= 0.2
S7_MIN_AVG_CATEGORIES ?= 4

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
		ci) src="api/env/ci.env"; dst="api/.env";; \
		*) echo "Unknown PROFILE '$$profile'. Use dev, test, prod, or ci."; exit 1;; \
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

.PHONY: scenario-suite
scenario-suite: ## Seed sample data and run S1-S10 scenario regression suite
	@echo "🎯 Running end-to-end scenario suite..."
	@python analysis/scripts/seed_dataset.py --base-url $(SCENARIO_BASE_URL) --org-id $(SCENARIO_ORG_ID) --namespace $(SCENARIO_NAMESPACE)
	@python analysis/scripts/run_scenarios.py --base-url $(SCENARIO_BASE_URL) --org-id $(SCENARIO_ORG_ID) --namespace $(SCENARIO_NAMESPACE) --s7-min-avg-mrr $(S7_MIN_AVG_MRR) --s7-min-avg-categories $(S7_MIN_AVG_CATEGORIES)

.PHONY: reset-namespace
reset-namespace: ## Delete all items/users/events in the configured namespace
	@echo "♻️ Resetting namespace '$(RESET_NAMESPACE)' on $(SCENARIO_BASE_URL)..."
	@python analysis/scripts/reset_namespace.py --base-url $(SCENARIO_BASE_URL) --org-id $(SCENARIO_ORG_ID) --namespace $(RESET_NAMESPACE) --force

.PHONY: restart-api
restart-api: ## Restart the API container and wait for /health
	@python analysis/scripts/restart_api.py --base-url $(RESTART_BASE_URL)

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

collab-factors: ## Generate collaborative filtering factors for items and users
	@echo "🤝 Generating collaborative factors..."
	@cd api && go run ./cmd/collab_factors --namespace $(NAMESPACE) --since $(SINCE)
