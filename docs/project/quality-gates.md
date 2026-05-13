---
diataxis: reference
tags:
  - reference
  - developer
  - ci
---
# Quality gates

Use the smallest gate that proves your change, then run the broader gate before merging cross-module work.

## Root gates

| Command | When to run | Expected result |
| --- | --- | --- |
| `make fmt` | Any Go change across modules. | All modules format successfully. |
| `make lint` | Any Go change across modules. | `go vet`/`golangci-lint` pass in each module. |
| `make test` | Cross-module behavior changes. | API Docker-backed tests and module tests pass. |
| `make security` | Security-sensitive changes or release readiness. | `govulncheck` and `gosec` pass in `api`, `recsys-eval`, `recsys-pipelines`, and `recsys-algo`. |
| `make finalize` | Final local check when Docker and docs tooling are available. | format, lint, tests, codegen, Markdown lint, and docs checks all pass. |

## Module gates

| Module | Commands | Expected result |
| --- | --- | --- |
| `api` | `GOWORK=off go test ./...`, `GOWORK=off golangci-lint run ./...`, `GOWORK=off govulncheck ./...`, `GOWORK=off gosec ./...`, `GOWORK=off go build ./...` | Unit/integration tests self-gate when external services are unavailable; scans and build succeed. |
| `recsys-eval` | `GOWORK=off go test ./...`, `GOWORK=off golangci-lint run ./...`, `GOWORK=off govulncheck ./...`, `GOWORK=off gosec ./...`, `make build` | Tests, scans, and CLI build succeed. |
| `recsys-pipelines` | `GOWORK=off go test ./...`, `GOWORK=off golangci-lint run ./...`, `GOWORK=off govulncheck ./...`, `GOWORK=off gosec ./...`, `make build` | Tests, scans, and job binaries build. |
| `recsys-algo` | `GOWORK=off go test ./...`, `GOWORK=off golangci-lint run ./...`, `GOWORK=off govulncheck ./...`, `GOWORK=off gosec ./...`, `make build`, `make plugin-example` | Library packages and standard examples build; the custom algorithm example builds with `-buildmode=plugin`. |

Raw `go build ./...` is intentionally not the `recsys-algo` build gate because `examples/custom_algo` is a Go plugin
package, not a standalone binary. Use `make plugin-example` for that package.
