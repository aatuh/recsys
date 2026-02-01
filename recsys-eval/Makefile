GO ?= go

GOWORK ?= off
export GOWORK

BIN_DIR := bin

GOLANGCI_LINT_VERSION ?= latest
GOSEC_VERSION ?= latest
GOVULNCHECK_VERSION ?= latest

.PHONY: help tools fmt lint vuln gosec tidy test test-race fuzz clean schema-check finalize

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

schema-check: ## Validate reports against JSON schema
	$(GO) test ./internal/app/usecase -run TestReportSchemaValidation -count=1

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
	$(MAKE) schema-check
	$(MAKE) test-race
	$(MAKE) fuzz
	$(MAKE) clean

build: ## Build all the binaries.
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/recsys-eval ./cmd/recsys-eval
