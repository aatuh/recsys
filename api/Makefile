SHELL := /bin/bash

.PHONY: dev down logs swag migrate-apply migrate-status tidy test

help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  dev             - Start the development environment"
	@echo "  down            - Stop the development environment"
	@echo "  logs            - View the logs of the development environment"
	@echo "  swag            - Generate the swagger files"
	@echo "  migrate-diff    - Generate the migration files"
	@echo "  migrate-apply   - Apply the migration files"
	@echo "  migrate-status  - Show the status of the migrations"
	@echo "  tidy            - Tidy the go modules"
	@echo "  test            - Run the tests"

dev:
	docker compose --env-file .env up --build

down:
	docker compose down -v

logs:
	docker compose logs -f api

swag:
	# generate swagger files inside the api container
	docker compose exec api swag init -g cmd/api/main.go -o ./swagger

migrate-diff:
	docker compose run --rm atlas migrate diff \
	--env dev \
	--dev-url "postgres://recsys:recsys@db:5432/recsys?sslmode=disable" \
	--dir "file://migrations" \
	--to  "file://schema.hcl"

migrate-apply:
	docker compose run --rm atlas \
	  migrate apply --env dev

migrate-status:
	docker compose run --rm atlas \
	  migrate status --env dev

tidy:
	docker compose exec api go mod tidy

FLAGS   ?=
TEST_PATTERN ?=
PKG          ?= ./...
test:
	@echo "🧪 Running tests in Docker. Pattern: '$(TEST_PATTERN)'  Flags: '$(FLAGS)'  Pkg: '$(PKG)'"
	@docker compose up -d db api
	@docker compose run --no-deps --rm \
		-e TEST_PATTERN='$(TEST_PATTERN)' \
		-e FLAGS='$(FLAGS)' \
		-e PKG='$(PKG)' \
		-e FAST='$(FAST)' \
		-e SKIP_API_WAIT='$(SKIP_API_WAIT)' \
		test
