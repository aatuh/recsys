Below is a **profit-first, low-hassle** way to proceed *and* a **developer-ready backlog** (with decisions already made) for turning your 4-module suite into something a real company can adopt, evaluate, and pay for.

---

## Decisions locked in (so the backlog doesn’t stall)

### Product / monetization model

**Decision:** Sell **self-hosted “Recsys Suite Pro”** (customer runs it), plus optional paid support.
Why: fastest path to revenue without you operating production. It matches your “wysiwyg: they get what they see in code/docs” goal.

**Packaging decision:** ship as:

* Docker images (service + pipelines + eval)
* Helm chart for Kubernetes
* docker-compose for local + small pilots

### Licensing (clear wedge + paywall without SaaS)

**Decision:** Open-source the “evaluation + pipelines” as the adoption wedge; keep the “serving + algorithms” source-available until paid.

* `recsys-eval`: **Apache-2.0**
* `recsys-pipelines`: **Apache-2.0**
* `recsys-svc (api/)` + `recsys-algo (api/recsys-algo/)`: **BUSL-1.1** with:

  * **Additional Use Grant:** free for non-production + internal PoCs; production requires paid license.
  * **Change Date:** 3 years after each release
  * **Change License:** **GPL-2.0-or-later** (BUSL requires GPLv2-compatible change license) ([akka.io][1])

This is a common “show the code, pay for production” model. (Enforcement is legal, not technical.)

### Documentation architecture

**Decision:** Keep Diátaxis structure (you already have it) and enforce it in CI. ([Diátaxis][2])

### Versioning & releases

**Decision:** Use **Semantic Versioning** and treat each module as independently versioned *if you keep multiple go.mod files*. ([Semantic Versioning][3])

* Git tags for multi-module Go repos must be prefixed with the module subdir (e.g., `recsys-eval/v0.3.0`) when module is not in repo root. ([go.dev][4])

### Supply-chain / enterprise trust baseline

**Decision:** target:

* SBOM generation (Syft) on every release ([GitHub][5])
* SLSA provenance attestation for releases ([SLSA][6])
* OpenSSF Scorecard run on every merge to main ([undefined][7])

---

## Quick reality check from your uploaded zip (important blockers)

These are concrete issues I saw that will bite real users immediately:

1. **`docker-compose.yml` references a `./web` build context but `web/` is missing** → `docker compose up` can’t be your “golden path” right now.
2. **Workflows located inside `recsys-pipelines/.github/workflows/` don’t run on GitHub** (only root `.github/workflows/` is used). So you effectively have less CI/release automation than it looks.
3. **Go module paths are inconsistent** (`module recsys-eval`, `github.com/pureapi/recsys-pipelines`, `github.com/aatuh/...`), which will confuse customers and break reproducible builds/imports in a monorepo context.
4. API spec appears duplicated/misaligned (OpenAPI 3.1 in docs vs a swagger/ folder pattern). You need one canonical spec. ([OpenAPI Initiative Publications][8])

Those become P0 tickets below.

---

# Comprehensive backlog (decisions already made)

## [ ] Epic A — Make a “golden path” that works on a clean machine (P0)

### [ ] RECSYS-001 — Fix broken docker-compose golden path

**Decision:** remove `web` from default compose; ship it later as optional add-on.
**What to do**

* Edit `docker-compose.yml`: remove the `web:` service (or move it to `docker-compose.web.yml`).
* Ensure `docker compose up` brings up: `db`, `minio`, `minio-init`, `proxy`, `api`.
* Add healthchecks for `api`, `db`, `minio`.
  **Acceptance**
* New user can run:

  ```bash
  git clone ...
  make dev
  ```

  and reach the API through proxy without editing files.

### [ ] RECSYS-002 — Add a one-command demo script

**Decision:** provide `./scripts/demo.sh` as the canonical “does it work?” entrypoint.
**What to do**

* Create `scripts/demo.sh` that:

  1. boots compose
  2. runs pipelines to produce an artifact
  3. loads the artifact into the service
  4. calls a “get recommendations” endpoint and prints output
     **Acceptance**
* `./scripts/demo.sh` exits 0 and prints a sample recommendation list.

### [ ] RECSYS-003 — Provide a minimal demo dataset

**Decision:** store a small, permissively-licensed synthetic dataset in-repo.
**What to do**

* Add `examples/data/` with 1k–10k interactions (CSV/Parquet).
* Document schema and generation method.
  **Acceptance**
* Pipelines can run end-to-end without external data.

---

## [ ] Epic B — Standardize repo identity & Go module structure (P0)

### [ ] RECSYS-010 — Decide canonical repo + module paths and migrate

**Decision:** monorepo named `github.com/pureapi/recsys-suite` with module paths:

* `github.com/pureapi/recsys-suite/recsys-eval`
* `github.com/pureapi/recsys-suite/recsys-pipelines`
* `github.com/pureapi/recsys-suite/api`
* `github.com/pureapi/recsys-suite/api/recsys-algo`
  **What to do**
* Update each `go.mod` `module ...` line accordingly.
* Update imports across the codebase.
* Update any `replace` directives.
  **Acceptance**
* `go test ./...` succeeds inside each module.
* `go mod tidy` produces clean diffs.

### [ ] RECSYS-011 — Add a Go workspace for local dev

**Decision:** use `go.work` at repo root for multi-module local development.
**What to do**

* Create `go.work` referencing all 4 modules.
* Update docs: “Use Go workspaces for dev; releases are tagged per module.”
  **Acceptance**
* From repo root: `go work sync` + `go test ./...` works.

### [ ] RECSYS-012 — Make version tags correct for multi-module Go

**Decision:** tag releases using module prefixes (e.g., `recsys-eval/v0.2.0`). ([go.dev][4])
**What to do**

* Update release docs for each module.
* Update CI release workflows to trigger on `recsys-eval/v*`, `recsys-pipelines/v*`, `api/v*`.
  **Acceptance**
* `go list -m -versions <module>` shows released versions correctly.

---

## [ ] Epic C — Licensing & compliance that enterprises trust (P0)

### [ ] RECSYS-020 — Add licenses and make them unambiguous

**Decision:** Apache-2.0 for eval/pipelines; BUSL-1.1 for service/algo with GPL-2.0-or-later change license. ([akka.io][1])
**What to do**

* Add `LICENSE` files per module (or root `LICENSES/` structure).
* Add `COMMERCIAL.md` explaining BUSL usage & how to buy a production license.
  **Acceptance**
* A lawyer/developer can answer “Can I use this in production?” in <2 minutes.

### [ ] RECSYS-021 — Adopt REUSE + SPDX headers

**Decision:** enforce REUSE compliance and SPDX file tags. ([reuse.software][9])
**What to do**

* Add `LICENSES/` with SPDX-named license texts.
* Add SPDX headers to source files (or `.reuse/` metadata).
* Add `reuse lint` to CI.
  **Acceptance**
* `reuse lint` passes on main.

---

## [ ] Epic D — CI/CD that actually runs + release artifacts (P0)

### [ ] RECSYS-030 — Consolidate GitHub workflows to root

**Decision:** only root `.github/workflows/*` exists; delete inert subdir workflows.
**What to do**

* Move relevant pipeline release logic into root workflows.
* Create matrix workflows that run per module.
  **Acceptance**
* CI runs on PR and main merges for all modules.

### [ ] RECSYS-031 — Add “quality gates” per module

**Decision:** gates = fmt + lint + tests + vulnerability scan + SBOM on release.
**What to do**

* `golangci-lint` (config committed)
* unit tests
* integration tests (service against Postgres/MinIO)
  **Acceptance**
* PR fails if any gate fails.

### [ ] RECSYS-032 — Release binaries + container images

**Decision:** use GoReleaser for CLIs; Docker build for service; publish to GHCR.
**What to do**

* Add GoReleaser configs for `recsys-eval` and `recsys-pipelines`.
* Publish Docker images: `ghcr.io/pureapi/recsys-svc:<version>`, etc.
  **Acceptance**
* Git tag → GitHub Release with binaries + images.

### [ ] RECSYS-033 — SBOM on release (Syft)

**Decision:** generate SPDX SBOM and attach to releases. ([GitHub][5])
**What to do**

* Add `anchore/sbom-action` workflow step.
  **Acceptance**
* Release assets include SBOM.

### [ ] RECSYS-034 — SLSA provenance on release

**Decision:** generate provenance attestation (Build L1/L2 baseline). ([SLSA][6])
**What to do**

* Add provenance generation using GitHub OIDC-based tooling.
  **Acceptance**
* Each release has provenance attached and verifiable.

### [ ] RECSYS-035 — Security posture: Scorecard + pinned actions

**Decision:** run OpenSSF Scorecard regularly + pin GitHub Actions by SHA. ([undefined][7])
**What to do**

* Add scorecard-action.
* Pin all actions.
  **Acceptance**
* Scorecard runs and produces results; actions are SHA-pinned.

---

## [ ] Epic E — Service API: make it “company-usable” (P0 → P1)

### [ ] RECSYS-040 — Define canonical API contract (OpenAPI 3.1)

**Decision:** `docs/reference/api/openapi.yaml` is the single source of truth. ([OpenAPI Initiative Publications][8])
**What to do**

* Remove/stop generating legacy swagger artifacts if they diverge.
* Add CI check that validates OpenAPI spec.
* Serve spec at `/openapi.yaml` and UI at `/docs`.
  **Acceptance**
* Spec and implementation match for all endpoints in smoke tests.

### [ ] RECSYS-041 — Standardize error format to RFC 7807

**Decision:** all non-2xx errors return RFC 7807 “problem details”. ([IETF Datatracker][10])
**What to do**

* Implement a central error mapper (domain errors → RFC7807 object).
* Update docs examples.
  **Acceptance**
* Any endpoint error matches the same JSON structure and includes stable `type/code`.

### [ ] RECSYS-042 — Auth strategy

**Decision:** support **API keys** (simple) + **OIDC JWT** (enterprise-ready).
**What to do**

* API keys: header `X-API-Key`, stored hashed in DB.
* OIDC: JWT validation via issuer + JWKS caching.
* Document both with examples.
  **Acceptance**
* Requests rejected without credentials; valid credentials allow access.

### [ ] RECSYS-043 — Multi-tenant isolation rules

**Decision:** tenant is explicit and mandatory (header or path), enforced in DB queries.
**What to do**

* Add tenant middleware that injects tenant into context.
* Ensure every query filters by tenant.
  **Acceptance**
* Integration test proves tenant A cannot read tenant B data.

### [ ] RECSYS-044 — Observability baseline

**Decision:** OpenTelemetry traces + Prometheus metrics + JSON logs.
**What to do**

* Add `/metrics`
* Add trace exporting (OTLP)
* Log correlation IDs
  **Acceptance**
* “request → DB → MinIO” shows up in a trace in local demo stack.

---

## [ ] Epic F — Algorithms: credible defaults + extensibility (P1)

### [ ] RECSYS-050 — Establish “baseline algorithms” set

**Decision:** ship 3 baselines:

1. Popularity (global + per-tenant)
2. Item-item cooccurrence
3. Matrix factorization / implicit feedback baseline
   **What to do**

* Implement or ensure present, but make them selectable by config.
* Add reproducible training + deterministic seeds.
  **Acceptance**
* Demo can switch algorithms via config and get different outputs.

### [ ] RECSYS-051 — Plugin contract for custom algorithms

**Decision:** algorithms are loaded as Go plugins **only for dev**; production uses build-time registration.
**What to do**

* Define `Algo` interface contract and version it.
* Provide `examples/custom_algo` showing how to add a new algo.
  **Acceptance**
* “hello world” algorithm can be added in <30 minutes by an external dev.

### [ ] RECSYS-052 — Model artifact format and compatibility

**Decision:** artifacts stored in object storage with:

* `manifest.json` (schema versioned)
* model params
* metadata (data window, feature config, metrics)
  **What to do**
* Define schema
* Validate schema on load
  **Acceptance**
* Service refuses incompatible artifacts with a clear RFC7807 error.

---

## [ ] Epic G — Pipelines: production-ish data engineering (P1)

### [ ] RECSYS-060 — Input connectors roadmap

**Decision:** implement connectors in this order:

1. Postgres (already implied)
2. S3/MinIO batch ingest
3. Kafka (events)
   **What to do**

* Define connector interfaces
* Add docs + examples for each
  **Acceptance**
* Each connector has a runnable example and integration test.

### [ ] RECSYS-061 — Incremental training & backfills

**Decision:** pipelines support:

* full rebuild
* incremental updates
* backfill by date range
  **What to do**
* Track watermarks/checkpoints
* Ensure idempotent runs
  **Acceptance**
* Re-running same job doesn’t duplicate data.

### [ ] RECSYS-062 — Pipeline scheduling story

**Decision:** ship CronJob examples for Kubernetes; don’t build an internal scheduler.
**What to do**

* Helm chart includes optional CronJobs
* Provide docs for “nightly retrain”
  **Acceptance**
* User can deploy nightly retrain in 1 YAML edit.

---

## [ ] Epic H — Eval: make it decision-grade (P1)

### [ ] RECSYS-070 — Standard metric suite

**Decision:** include:

* Precision@K, Recall@K, NDCG@K
* Coverage, novelty/diversity proxy
* Latency + throughput measurements for serving
  **What to do**
* Implement metrics if missing
* Add docs + examples on interpreting results
  **Acceptance**
* Demo outputs a markdown/HTML report with key metrics.

### [ ] RECSYS-071 — Regression gating with eval

**Decision:** changes to algo/pipelines must not degrade a baseline metric beyond threshold (configurable).
**What to do**

* Add CI job that runs eval on the demo dataset
* Fail build on regression beyond threshold
  **Acceptance**
* A deliberately degraded change fails CI.

---

## [ ] Epic I — Deployment: real-company install story (P1 → P2)

### [ ] RECSYS-080 — Helm chart for suite

**Decision:** helm is the supported production install method.
**What to do**

* Chart includes: service, optional pipelines CronJobs, Postgres externalized, MinIO optional.
* Values.yaml documents recommended resource requests/limits.
  **Acceptance**
* `helm install recsys ./charts/recsys` works against a real cluster.

### [ ] RECSYS-081 — “Bring your own Postgres/S3” first-class

**Decision:** Postgres and S3 are external dependencies in production; bundled only for local demo.
**What to do**

* Support env vars for external endpoints.
* Document IAM/policies for S3 bucket.
  **Acceptance**
* Helm deploy works with AWS/GCP equivalents.

---

## [ ] Epic J — Enterprise buyer checklist (P2)

### [ ] RECSYS-090 — RBAC + audit log

**Decision:** RBAC roles: `viewer`, `operator`, `admin`; audit log always on for admin actions.
**What to do**

* Add role claims mapping (OIDC)
* Store audit events in DB
  **Acceptance**
* Admin actions generate audit events and can be queried/exported.

### [ ] RECSYS-091 — Upgrade/migration safety

**Decision:** versioned DB migrations; “safe migration” policy documented.
**What to do**

* Add preflight checks
* Document rollback story
  **Acceptance**
* Upgrading from N-1 to N documented and tested.

### [ ] RECSYS-092 — Performance and capacity guide

**Decision:** publish “sizing tiers” with tested benchmarks.
**What to do**

* Add load tests for top endpoints
* Document QPS vs CPU/mem vs cache
  **Acceptance**
* Benchmarks reproducible; published in docs.

---

## [ ] Epic K — “Sell without meetings” assets (P0 → P1, but still dev-doable)

### [ ] RECSYS-100 — Self-serve evaluation pack

**Decision:** ship a downloadable “evaluation pack”:

* docker-compose
* demo dataset
* expected metrics report
  **What to do**
* Bundle scripts + docs into `/dist/eval-pack/`
  **Acceptance**
* A company can evaluate without talking to you.

### [ ] RECSYS-101 — Pricing & licensing mechanics (minimal)

**Decision:** keep it simple:

* Pro license = annual subscription, per-tenant or per-request tier
* No DRM; contract + key file for compliance
  **What to do**
* Add `LICENSE_KEY` support only for “show license status” endpoint (not blocking).
* Provide `pricing.md` and “how to buy” doc.
  **Acceptance**
* Sales flow is clear; no engineering complexity explosion.

---

# Suggested execution order (so you reach revenue fastest)

1. **Epic A + B + C + D (P0)** → “it installs, it runs, it’s licensed, it releases”
2. **Epic E (API contract + errors + auth)** → “it feels professional”
3. **Epic F/G/H (baseline algos + pipelines + eval gating)** → “it’s credible”
4. **Epic I (Helm + BYO infra)** → “real companies can deploy”
5. **Epic J/K** → “enterprise + self-serve sales”

---

## If you want, I can turn this into actual Jira/GitHub Issues text

Meaning: one issue per ticket with copy-paste-ready sections (Description / Steps / Acceptance / Links / Labels).

[1]: https://akka.io/bsl/license "https://akka.io/bsl/license"
[2]: https://diataxis.fr/ "https://diataxis.fr/"
[3]: https://semver.org/ "https://semver.org/"
[4]: https://go.dev/ref/mod "https://go.dev/ref/mod"
[5]: https://github.com/anchore/sbom-action "https://github.com/anchore/sbom-action"
[6]: https://slsa.dev/spec/v1.0/ "https://slsa.dev/spec/v1.0/"
[7]: https://scorecard.dev/ "https://scorecard.dev/"
[8]: https://spec.openapis.org/oas/v3.1.0.html "https://spec.openapis.org/oas/v3.1.0.html"
[9]: https://reuse.software/spec-3.3/ "https://reuse.software/spec-3.3/"
[10]: https://datatracker.ietf.org/doc/html/rfc7807 "https://datatracker.ietf.org/doc/html/rfc7807"
