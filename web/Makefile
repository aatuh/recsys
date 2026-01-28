.PHONY: help codegen test fmt lint install dev

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## Install dependencies
	@echo "Installing dependencies..."
	pnpm install

dev: ## Run the web app
	@echo "Starting web dev server..."
	@pnpm dev

codegen: ## Generate code
	@echo "🔄 Generating code..."
	pnpm run generate:api

test: ## Run tests
	@echo "🧪 Running tests..."
	pnpm test

fmt: ## Format code
	@echo "🎨 Formatting code..."
	@pnpm lint --fix

lint: ## Lint code
	@echo "🔍 Linting code..."
	@rm -rf .next
	@pnpm lint
	@pnpm typecheck
	@pnpm build

finalize: ## Thorough validity check and generation
	@echo "✅ Finalizing code..."
	make fmt
	make lint
	make test
	make codegen
