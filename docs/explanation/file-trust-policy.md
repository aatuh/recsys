---
diataxis: explanation
tags:
  - explanation
  - security
  - ops
---
# File trust policy

This page defines how RecSys treats local filesystem paths. The goal is to keep service/runtime file reads confined while
still allowing operator-owned CLI workflows.

## Trust classes

| Area | Trust model | Guardrail | Quality gate |
| --- | --- | --- | --- |
| `recsys-service` artifact `file://` and bare paths | Manifest and artifact paths are data read by the service and must not be able to read arbitrary host files. | Set `RECSYS_ARTIFACT_FILE_ROOT`; local artifact reads are disabled when the root is unset. Reads use `os.Root` and reject traversal and symlink escapes. | `cd api && GOWORK=off go test ./internal/objectstore ./internal/store ./cmd/recsys-service && GOWORK=off gosec ./...` |
| `*_FILE` secret env vars | Operator-provisioned Docker/Kubernetes secret mounts. | Values are read at startup only; errors name the env var, not the secret value. | `cd api && GOWORK=off gosec ./...` |
| License files | Operator-configured local files. | Paths remain explicit config and are not derived from request data. | `cd api && GOWORK=off go test ./internal/license && GOWORK=off gosec ./...` |
| Exposure log path | Operator-configured output file or directory. | Files are created `0600`; directory mode uses service-generated UTC filenames and `0700` directories. | `cd api && GOWORK=off go test ./internal/exposure && GOWORK=off gosec ./...` |
| `recsys-eval` report outputs | Trusted CLI output destinations. | Report writers create exactly the operator-requested destination; callers choose the output location. | `cd recsys-eval && GOWORK=off go test ./... && GOWORK=off gosec ./...` |
| `recsys-pipelines` filesystem adapters | Configured roots for object store, staging, checkpoint, canonical, and registry storage. | Logical tenant/surface/window/object-key segments are validated and paths are confined under adapter roots. Private directories use `0750` or narrower file modes. | `cd recsys-pipelines && GOWORK=off go test ./... && GOWORK=off gosec ./...` |
| Developer generators such as OpenAPI sync | Trusted local developer tooling. | Inputs/outputs are explicit flags; generated output directories use `0750`. | `cd api && GOWORK=off go test ./cmd/openapi-sync ./cmd/migrate && GOWORK=off gosec ./...` |

## Operational rules

- Prefer S3/object storage for production artifact mode. If local artifacts are used, set `RECSYS_ARTIFACT_FILE_ROOT` to
  the artifact registry root and ensure manifests reference only files below that root.
- Do not point service config at broad host paths such as `/`, `/etc`, or a shared home directory.
- Keep CLI report/config paths out of customer-controlled input. They are local operator choices, not API fields.
- Treat `#nosec` comments as evidence: each accepted finding must explain why the path is operator-trusted or already
  confined before the filesystem call.
