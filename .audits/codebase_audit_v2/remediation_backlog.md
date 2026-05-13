# Backlog

Project: RecSys suite

Status legend:

- [ ] not done
- [x] done

## Epic E1 - File access and static security closure [x]

Description: Establish one clear file trust model and close the remaining filesystem-related scanner failures without weakening operator workflows.

### Ticket E1-T1 - Define the repository file trust policy [x]

Description: Document which file paths are untrusted, root-confined, symlink-sensitive, or operator-trusted across API, eval, pipelines, secret files, and generated docs/tools.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Include the exact quality gates that prove future path changes are safe.
- Keep operator-trusted CLI paths explicit rather than implied.

### Ticket E1-T2 - Root-confine API artifact filesystem reads [x]

Description: Add a configured root, trust guard, or explicit production disablement for API `file://` artifact reads so manifests cannot point at arbitrary local files by default.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Add happy path, traversal, and outside-root tests.
- Keep public error responses generic and free of filesystem paths.

### Ticket E1-T3 - Classify API trusted file reads and writes [x]

Description: Review API secret-file, license-file, exposure-log, migration preflight, and OpenAPI sync findings; fix root confinement or add narrow documented suppressions where the path is operator-trusted.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Do not suppress findings without a short rationale beside the code.
- Preserve Docker secret-file support.

### Ticket E1-T4 - Close eval and pipeline filesystem findings [x]

Description: Apply the file trust policy to eval report output and pipeline filesystem adapters, including permissions, symlink assumptions, and remaining `gosec` findings.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Add tests for any root-confined paths that are not already covered.
- Use narrower permissions where shared read access is not required.

### Ticket E1-T5 - Make gosec pass or fail only on accepted findings [x]

Description: Ensure `gosec ./...` passes in API, eval, and pipelines, or fails only on intentionally accepted findings with explicit suppressions and documentation.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Minimum verification: `cd api && GOWORK=off gosec ./...`, `cd recsys-eval && GOWORK=off gosec ./...`, and `cd recsys-pipelines && GOWORK=off gosec ./...`.

## Epic E2 - Admin audit and tenant-isolation hardening [x]

Description: Reduce sensitive control-plane exposure and add stronger tenant-isolation guardrails.

### Ticket E2-T1 - Define admin audit detail access policy [x]

Description: Decide which roles can see audit metadata, raw `before`/`after` payloads, and extra details for config, rules, cache invalidation, and future admin events.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Preserve operator debuggability while preventing viewer-level overexposure.

### Ticket E2-T2 - Implement audit redaction or stricter role gating [x]

Description: Change admin audit responses so viewer-level reads do not expose raw state unless the policy explicitly allows it.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Add tests for viewer, operator, and admin roles.
- Update OpenAPI and admin docs if response shape changes.

### Ticket E2-T3 - Add a tenant-isolation database guardrail [x]

Description: Choose and implement a production guardrail for DB tenant isolation, such as enforced RLS policies or a startup/preflight check that makes the non-RLS posture explicit.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- If RLS is not enabled by default, document the accepted risk and operational controls clearly.
- Include migration/preflight tests where feasible.

## Epic E3 - Repository quality gate alignment [x]

Description: Make root, module, and CI validation produce consistent safety signals for maintainers.

### Ticket E3-T1 - Add a root security quality target [x]

Description: Add a root command that runs module `govulncheck` and `gosec` consistently across API, algo, eval, and pipelines.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- The target may be separate from Docker-backed `make finalize` if that keeps local checks usable.

### Ticket E3-T2 - Align CI with module security and build gates [x]

Description: Update GitHub Actions so changed modules run the same meaningful checks as local module gates, including security scans and builds where appropriate.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Avoid adding flaky Docker dependencies to jobs that do not need Docker.

### Ticket E3-T3 - Normalize recsys-algo plugin build expectations [x]

Description: Make the custom algorithm plugin example explicit in build tooling so raw `go build ./...` behavior does not surprise developers.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Consider a dedicated `make build` or `make plugin-example` target.
- Keep the documented `-buildmode=plugin` flow working.

### Ticket E3-T4 - Document root versus module finalize workflows [x]

Description: Update developer docs so maintainers know when to run root `make finalize`, module finalize targets, security scans, Docker-backed API tests, and docs checks.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Include expected high-level success output for each main command.

## Epic E4 - API boundary cleanup [ ]

Description: Reduce outward dependencies from the API service layer after the remaining security gates are stable.

### Ticket E4-T1 - Move the recsys-algo adapter out of the service package [ ]

Description: Create a clearly named adapter package for recsys-algo integration and keep `recsysvc` focused on the recommendation service port.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Avoid changing algorithm behavior while moving packages.

### Ticket E4-T2 - Introduce an inward-facing tenant scope abstraction [ ]

Description: Stop reading toolkit authorization context directly inside `recsysvc`; pass tenant scope through a service-owned request field or small port.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Preserve tenant mismatch behavior already tested in HTTP middleware.

### Ticket E4-T3 - Split API dependency wiring into smaller builders [ ]

Description: Break `buildAppDeps` into cohesive builders for artifact stores, algorithm engine, admin service, exposure logging, experiment assignment, and licensing.

Implementation rules:

- implement the ticket in the smallest sensible step
- run `make finalize` after completing the ticket, or an equivalent quality toolkit if `make finalize` is unavailable
- ensure the quality check covers testing, formatting, linting, and other relevant validation for the repository
- create a git commit immediately after the ticket is complete
- use Conventional Commits style for the commit message
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Add focused tests only around behavior that changes or becomes easier to exercise.
