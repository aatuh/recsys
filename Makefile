.PHONY: help codegen dev build test lint fmt typecheck uilint uitypecheck prod-build prod-run

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

codegen: ## Generate code
	@echo "🔄 Generating code..."
	@cd api && make swag
	@cp api/swagger/swagger.json swagger-service/public/swagger.json
	@cp api/swagger/swagger.yaml swagger-service/public/swagger.yaml
	@cd web && pnpm run codegen:api && pnpm run sync:readme
	@cd shop && pnpm run codegen:api || pnpm run codegen:api:local
	@cd demo && pnpm run codegen:api
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

health: ## Check service healthiness (logs + curl)
	@echo "🏥 Checking service healthiness..."
	@echo "📋 Docker Compose logs (last 20 lines):"
	@docker-compose logs --tail=20 api
	@echo ""
	@echo "🌐 Health endpoint check:"
	@set -e; \
	for i in {1..5}; do \
		if curl -sf http://localhost:8000/health >/dev/null; then \
			echo "✅ Health check completed"; \
			exit 0; \
		fi; \
		sleep 2; \
	done; \
	echo "❌ Health check failed"; \
	exit 1
