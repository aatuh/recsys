GO ?= go

GOWORK ?= off
export GOWORK

SHELL := /bin/bash
BIN_DIR := bin

GOLANGCI_LINT_VERSION ?= latest
GOSEC_VERSION ?= latest
GOVULNCHECK_VERSION ?= latest

.PHONY: help tools fmt lint vuln gosec tidy test test-race fuzz clean finalize build smoke

help: ## Show help
	@awk 'BEGIN {FS=":.*## "}; \
		/^[a-zA-Z0-9_.-]+:.*## / { \
			if (match($$0, /## .*## /)) { \
				printf "error: multiple ## in help comment for target %s\n", $$1; exit 1; \
			} \
			printf "  %-14s %s\n", $$1, $$2 \
		}' $(MAKEFILE_LIST)

tools: ## Install lint/vuln tools
	@$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@$(GO) install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	@$(GO) install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

fmt: ## Run gofmt
	$(GO) fmt ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

vuln: ## Run govulncheck
	govulncheck ./...

gosec: ## Run gosec
	gosec ./...

tidy: ## Run go mod tidy
	$(GO) mod tidy

test: ## Run unit tests
	$(GO) test ./...

test-race: ## Run unit tests with race detector
	$(GO) test ./... -race -count=1

fuzz: ## Run fuzz smoke tests
	$(GO) test ./... -run=^$ -fuzz=Fuzz -fuzztime=10s

clean: ## Clean test cache
	$(GO) clean -testcache

finalize: ## Run every quality assurance tool
	$(MAKE) tools
	$(MAKE) fmt
	$(MAKE) lint
	$(MAKE) vuln
	$(MAKE) gosec
	$(MAKE) tidy
	$(MAKE) test
	$(MAKE) test-race
	$(MAKE) fuzz
	$(MAKE) clean

build: ## Build all the binaries.
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/recsys-pipelines ./cmd/recsys-pipelines
	go build -o $(BIN_DIR)/job_ingest ./cmd/job_ingest
	go build -o $(BIN_DIR)/job_popularity ./cmd/job_popularity
	go build -o $(BIN_DIR)/job_cooc ./cmd/job_cooc
	go build -o $(BIN_DIR)/job_db_signals ./cmd/job_db_signals
	go build -o $(BIN_DIR)/job_catalog ./cmd/job_catalog
	go build -o $(BIN_DIR)/job_publish ./cmd/job_publish
	go build -o $(BIN_DIR)/job_validate ./cmd/job_validate

smoke: build ## Build and run smoke test.
	./scripts/dev_smoke.sh
