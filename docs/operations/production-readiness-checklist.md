---
tags:
  - ops
  - checklist
  - security
---

# Production readiness checklist (RecSys suite)

## Who this is for

Lead developers, platform/SRE, and security reviewers preparing to run `recsys-service` in production.

## What you will get

A practical checklist to catch the most common “we went live and it broke” gaps: auth/tenancy, data modes, privacy,
observability, runbooks, backups, and safe rollout/rollback.

## 0) Decide your serving mode (DB-only vs artifact/manifest)

- [ ] Pick a serving mode per tenant/environment:
  - [ ] **DB-only mode** (fastest pilot; signals live in Postgres)
  - [ ] **Artifact/manifest mode** (production-like; pipelines publish artifacts + manifest)
- [ ] Read: [Data modes](../explanation/data-modes.md)
- [ ] Document the choice for your deployment (values/env + runbooks).

## 1) Tenancy and authentication

- [ ] Decide the production auth mechanism:
  - [ ] JWT (recommended for production)
  - [ ] API keys (for server-to-server integrations)
- [ ] Ensure dev headers are **disabled** in production (`DEV_AUTH_ENABLED=false`).
- [ ] If using JWT:
  - [ ] Configure `JWT_JWKS_URL`, `JWT_ISSUER`, `JWT_AUDIENCE`
  - [ ] Configure `AUTH_JWKS_ALLOWED_HOSTS` (and keep `AUTH_ALLOW_INSECURE_JWKS=false`)
  - [ ] Confirm readiness check passes from the cluster network (`/readyz`)
- [ ] If using API keys:
  - [ ] Set `API_KEY_ENABLED=true`, store `API_KEY_HASH_SECRET` securely
  - [ ] Rotate keys and document the process
- [ ] Decide and document tenancy source:
  - [ ] tenant claim(s) in JWT (`AUTH_TENANT_CLAIMS`) and/or header (`TENANT_HEADER_NAME`)
- [ ] Validate RBAC rules for admin endpoints:
  - [ ] `AUTH_VIEWER_ROLE`, `AUTH_OPERATOR_ROLE`, `AUTH_ADMIN_ROLE`

See: [Admin API + local bootstrap](../reference/api/admin.md)

## 2) Data contracts, logging, and privacy

- [ ] Decide what identifiers are allowed (user_id / anonymous_id / session_id).
- [ ] Confirm exposure/outcome logging design:
  - [ ] Required IDs + correlation strategy (request_id, item_id, subject id)
  - [ ] Retention policy and access control
- [ ] If using exposure hashing, set and rotate `EXPOSURE_HASH_SALT` as a secret.
- [ ] Document your PII stance (what fields are considered PII in your org).

See:

- [Exposure logging & attribution](../explanation/exposure-logging-and-attribution.md)
- [Eval events](../reference/data-contracts/eval-events.md)

## 3) Pipelines readiness (artifact/manifest mode only)

- [ ] Confirm artifact publishing is automated (scheduler) and has an owner/on-call.
- [ ] Confirm you can roll back the manifest safely.
- [ ] Define freshness SLOs and alerting.

See:

- [SLOs and freshness (pipelines)](../recsys-pipelines/docs/operations/slos-and-freshness.md)
- [Roll back the manifest](../recsys-pipelines/docs/how-to/rollback-manifest.md)

## 4) Database and migrations

- [ ] Ensure Postgres is provisioned with backups and a tested restore procedure.
- [ ] Confirm migrations are applied safely:
  - [ ] preflight checks in CI
  - [ ] an explicit migration job in production (not “hope MIGRATE_ON_START is fine”)
- [ ] Document your rollback strategy for schema changes.

See: [Database migrations](../reference/database/migrations.md)

## 5) Observability and runbooks

- [ ] Liveness/readiness probes are wired:
  - [ ] `/healthz` (liveness)
  - [ ] `/readyz` (readiness)
  - [ ] `/health/detailed` (debugging)
- [ ] Metrics are scraped and dashboards exist for:
  - [ ] request rate, error rate, latency (p50/p95/p99)
  - [ ] empty-recs rate
  - [ ] cache hit/miss (if enabled)
- [ ] Tracing/logging is configured per your standards (and does not leak secrets).
- [ ] Runbooks exist and have been exercised at least once:
  - [ ] [Service not ready](runbooks/service-not-ready.md)
  - [ ] [Empty recs](runbooks/empty-recs.md)
  - [ ] [Roll back config/rules](runbooks/rollback-config-rules.md)

## 6) Performance and capacity

- [ ] Run a load test against production-like data and record results.
- [ ] Configure caching and backpressure based on observed behavior:
  - [ ] `RECSYS_CONFIG_CACHE_TTL`, `RECSYS_RULES_CACHE_TTL`
  - [ ] `RECSYS_BACKPRESSURE_MAX_INFLIGHT`, `RECSYS_BACKPRESSURE_MAX_QUEUE`
- [ ] Decide if and how you will enable profiling endpoints (keep `PPROF_ENABLED=false` by default).

See: [Performance and capacity](performance-and-capacity.md)

## 7) Safe rollout and rollback

- [ ] Define “ship” and “rollback” procedures for:
  - [ ] config and rules (admin API, audit log)
  - [ ] artifact manifests (pipelines)
  - [ ] algorithm version changes (deployments)
- [ ] Confirm you can answer: “Which config/rules/algo version served this request?”
  - [ ] `meta.config_version`, `meta.rules_version`, `meta.algo_version` in responses
- [ ] Document gates and criteria for shipping.

See: [Run eval and ship](../how-to/run-eval-and-ship.md)
