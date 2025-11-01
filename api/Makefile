SHELL := /bin/bash

.PHONY: dev down logs swag migrate-build migrate-up migrate-down migrate-status migrate-new tidy test

help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  dev             - Start the development environment"
	@echo "  down            - Stop the development environment"
	@echo "  logs            - View the logs of the development environment"
	@echo "  swag            - Generate the swagger files"
	@echo "  migrate-up      - Apply the migrations"
	@echo "  migrate-down    - Rollback the migrations"
	@echo "  migrate-status  - Show the status of the migrations"
	@echo "  tidy            - Tidy the go modules"
	@echo "  test            - Run the tests"
	@echo "  health          - Check service healthiness"

dev:
	docker compose --env-file .env up --build

down:
	docker compose down -v

logs:
	docker compose logs -f api

swag:
	# generate swagger files inside the api container
	docker compose exec api swag init -g cmd/api/main.go -o ./swagger

migrate-up:
	docker compose exec api sh -c "go run ./cmd/migrate -dir \$$MIGRATIONS_DIR up"

migrate-down:
	docker compose exec api sh -c "go run ./cmd/migrate -dir \$$MIGRATIONS_DIR --allow-down down"

migrate-status:
	docker compose exec api sh -c "go run ./cmd/migrate -dir \$$MIGRATIONS_DIR status"

tidy:
	docker compose exec api go mod tidy

FLAGS   ?=
TEST_PATTERN ?=
PKG          ?= ./...
test:
	@echo "🧪 Running tests in Docker. Pattern: '$(TEST_PATTERN)'  Flags: '$(FLAGS)'  Pkg: '$(PKG)'"
	@API_ENV_FILE=./api/.env.test docker compose -f ../docker-compose.yml up -d db api
	@API_ENV_FILE=./api/.env.test docker compose -f ../docker-compose.yml run --no-deps --rm \
		-e TEST_PATTERN='$(TEST_PATTERN)' \
		-e FLAGS='$(FLAGS)' \
		-e PKG='$(PKG)' \
		-e FAST='$(FAST)' \
		-e SKIP_API_WAIT='$(SKIP_API_WAIT)' \
		test

health:
	@echo "🏥 Checking service healthiness..."
	@echo "📋 Docker Compose logs (last 20 lines):"
	@docker compose logs --tail=20 api || true
	@echo ""
	@echo "🌐 Health endpoint check:"
	@set -e; \
	PORT=$${API_PORT:-8000}; \
	for i in {1..5}; do \
		if curl -sf http://localhost:$$PORT/health >/dev/null; then \
			echo "✅ Health check completed"; \
			exit 0; \
		fi; \
		sleep 2; \
	done; \
	echo "❌ Health check failed"; \
	exit 1
