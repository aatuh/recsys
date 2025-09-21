.PHONY: help codegen dev build test lint fmt typecheck uilint uitypecheck

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

codegen: ## Generate code
	@echo "🔄 Generating code..."
	@cd api && make swag
	@cp api/swagger/swagger.json swagger-service/public/swagger.json
	@cp api/swagger/swagger.yaml swagger-service/public/swagger.yaml
	@cd demo-ui && pnpm run codegen:api && pnpm run sync:readme
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
	@cd demo-ui && pnpm run lint

fmt: ## Format UI (eslint --fix)
	@echo "🧼 Formatting UI..."
	@cd demo-ui && pnpm run lint:fix || true

typecheck: ## Typecheck UI
	@echo "🔎 Typechecking UI..."
	@cd demo-ui && pnpm run typecheck
