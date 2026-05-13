# Backlog

Project: RecSys suite

Status legend:

- [ ] not done
- [x] done

## Epic E1 - Security dependency and debug exposure hardening [x]

Description: Remove the highest-risk API security findings from the audit before broader refactors.

### Ticket E1-T1 - Upgrade vulnerable API OpenTelemetry dependency [x]

Description: Upgrade the API module's OpenTelemetry packages from `v1.39.0` to a version fixed for `GO-2026-4394`, then verify `govulncheck` is clean for reachable vulnerabilities.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Minimum verification: `cd api && GOWORK=off govulncheck ./...`.
- Also run API tests and lint because the dependency is reachable from service startup/tracing.

### Ticket E1-T2 - Protect pprof debug endpoints [x]

Description: Ensure pprof cannot be exposed publicly when `PPROF_ENABLED=true`; require an explicit safe mount strategy such as admin-only auth, separate listener, or local-only binding.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Add tests proving `/debug/pprof` is not mounted by default and is protected when enabled.
- Update docs/config reference if operational behavior changes.

### Ticket E1-T3 - Require production hashing secrets [x]

Description: Fail startup in production when exposure logging, experiment assignment, or API-key auth are enabled without the required salt/secret.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Cover happy path, missing-secret failure mode, and non-production/local behavior.
- Do not log secret values in errors.

## Epic E2 - Path and tenant-surface input hardening [ ]

Description: Close filesystem traversal and path inclusion risks across API artifact mode and pipeline filesystem adapters.

### Ticket E2-T1 - Add a shared path-segment validation helper [x]

Description: Define a small reusable validator for logical IDs used as filesystem path segments, including tenant, surface, segment, artifact type/key, and artifact version.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Reject empty, `.`, `..`, slash, backslash, and control/whitespace-only path segments.
- Keep the helper dependency direction inward; domain packages should not import adapters.

### Ticket E2-T2 - Apply path validation to API artifact mode [x]

Description: Validate surface and manifest path substitutions before loading filesystem-backed artifacts in the API service.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Add regression tests for malicious `surface` values in recommend and similar requests.
- Ensure public error responses stay generic and do not leak filesystem paths.

### Ticket E2-T3 - Apply path validation to pipeline filesystem adapters [x]

Description: Harden `recsys-pipelines` staging, artifact registry, checkpoint, object store, and file datasource paths against traversal through tenant, surface, segment, and object key inputs.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Extend the existing traversal-version test to tenant, surface, segment, and object key cases.
- Preserve valid existing artifact paths and documented examples.

### Ticket E2-T4 - Constrain filesystem readers to configured roots [ ]

Description: Where filesystem readers accept file URIs or paths, ensure reads and writes stay under a configured root or are explicitly documented as operator-trusted inputs.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Prefer root-scoped file access where feasible.
- Document intentionally trusted operator paths if any remain.

## Epic E3 - Config and data-quality correctness [ ]

Description: Make misconfiguration and data-contract violations fail visibly instead of being silently accepted.

### Ticket E3-T1 - Make API config parsing fail fast [x]

Description: Replace silent defaults for invalid numeric and list environment values with startup validation errors.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Cover invalid floats, invalid int16 CSV values, and valid defaults.
- Keep error messages actionable and free of secret values.

### Ticket E3-T2 - Enforce safe production S3 defaults [x]

Description: Require TLS for production S3 artifact mode unless an explicit local-development override is used.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Cover production failure mode and local MinIO/dev behavior.
- Update config docs with the exact override semantics.

### Ticket E3-T3 - Detect duplicate eval request IDs [x]

Description: Add duplicate exposure `request_id` detection to `recsys-eval` joins and surface it in data-quality output.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Add a regression test that would currently fail because duplicates are overwritten.
- Preserve existing behavior for clean datasets.

### Ticket E3-T4 - Review admin audit response detail [ ]

Description: Decide whether viewer-level audit reads should include raw before/after state; implement redaction or stricter role gating if needed.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Preserve operator debuggability while reducing accidental secret or personal-data exposure.
- Update API docs if response shape or authorization changes.

## Epic E4 - Test and quality-gate reliability [ ]

Description: Make local and CI validation clearer, more complete, and less environment-fragile.

### Ticket E4-T1 - Split or self-gate API integration tests [ ]

Description: Ensure raw `go test ./...` in `api/` either skips integration tests with a clear message or moves them behind an explicit integration command.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Keep the existing Docker Compose integration workflow available.
- Avoid panics for missing `API_HOST` or `DATABASE_URL`.

### Ticket E4-T2 - Add security regression tests for audit findings [ ]

Description: Add focused tests for pprof protection, production secrets, artifact path traversal, and duplicate eval request IDs.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Prefer small unit or adapter tests over broad E2E tests.
- Include happy path, edge case, and failure mode assertions.

### Ticket E4-T3 - Align root quality gates with module gates [ ]

Description: Update root quality-gate documentation and/or Makefile targets so security scans and builds are not easy to miss.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Decide whether root `make finalize` should call module `finalize` targets or document a separate security/deep-finalize target.
- Keep the default workflow practical for contributors.

## Epic E5 - Architecture boundary cleanup [ ]

Description: Reduce API application-layer coupling after the security and quality-gate risks are addressed.

### Ticket E5-T1 - Move recsys-algo adapter out of the API service package [ ]

Description: Relocate the API's recsys-algo adapter behind a clearer adapter boundary while preserving the `recsysvc.Engine` port.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Avoid a broad rewrite; keep behavior stable.
- Use existing service tests as regression coverage and add adapter tests only where behavior moves.

### Ticket E5-T2 - Remove framework auth leakage from recommendation service logic [ ]

Description: Replace direct toolkit authorization context usage in `recsysvc` with a small internal tenant-context port/helper owned by the API application boundary.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Keep middleware behavior unchanged.
- Preserve tenant mismatch/error behavior and tests.

### Ticket E5-T3 - Remove duplicate explain enforcement [ ]

Description: Remove the duplicate `enforceExplainControls` call in the recommend handler and add a small regression test around explain authorization.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- This is intentionally last because it is low risk and does not materially improve the security score.
