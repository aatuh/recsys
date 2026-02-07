---
tags:
  - overview
  - deployment
  - developer
  - ops
---

# Deployment shapes (minimal archetypes)

## Who this is for

- Lead developers deciding “what do we actually deploy?” for the first production rollout.
- Platform/SRE deciding how to isolate tenants and control blast radius.

## What you will get

- Three deployment archetypes (local/dev, single-tenant prod, multi-tenant prod).
- The minimal components in each shape and the main tradeoffs.
- Links to the canonical deployment, security, and ops docs.

## Quick decision table

| Shape | Choose this when | Main tradeoff |
| --- | --- | --- |
| Local/dev | You want fastest integration and debugging | Not a production posture |
| Single-tenant prod | One tenant/surface to prove production fit | Less reuse if you later need many tenants |
| Multi-tenant prod | You need shared infra and tenant isolation | Requires stronger auth, limits, and ops discipline |

## 1) Local/dev (single laptop)

**Goal:** get a non-empty response and an eval-compatible exposure log quickly.

Run:

- Postgres
- `recsys-service` (the `api` service in docker compose)
- Optional: `recsys-eval` (to validate schemas and produce a first report)

Recommended mode:

- Start in **DB-only mode** (default). See: [Choose your data mode](choose-data-mode.md).

Start here:

- Tutorial: [Quickstart (10 minutes)](../tutorials/quickstart.md)
- Tutorial: [Local end-to-end](../tutorials/local-end-to-end.md)
- How-to: [First surface end-to-end](../how-to/first-surface-end-to-end.md)

## 2) Single-tenant production

**Goal:** run one tenant with production auth and predictable rollbacks.

Run:

- `recsys-service` (Kubernetes/Helm, or equivalent)
- External Postgres (recommended)
- Optional: S3/MinIO + `recsys-pipelines` (for artifact/manifest mode)

Recommended mode:

- Start DB-only for “first production”.
- Graduate to artifact/manifest mode when pipelines publish versioned artifacts and you want an atomic rollback lever.

Start here:

- How-to: [Deploy with Helm](../how-to/deploy-helm.md)
- Checklist: [Production readiness checklist](../operations/production-readiness-checklist.md)
- Security posture: [Security, privacy, compliance](security-privacy-compliance.md)

## 3) Multi-tenant production

**Goal:** serve multiple tenants while keeping isolation and blast radius under control.

RecSys supports tenant-scoped configuration and rules, but tenant bootstrap is currently DB-only. See:
[Known limitations](known-limitations.md).

Two common patterns:

- **One deployment per tenant (strong isolation)**
  - Pros: clear blast radius, simpler compliance boundaries.
  - Cons: higher ops overhead and duplicated infra.
- **Shared deployment for multiple tenants (shared infra)**
  - Pros: fewer deployments, easier reuse of pipelines and ops.
  - Cons: requires stronger control-plane access control, per-tenant rate limiting, and careful observability.

Minimum requirements for the shared-deployment pattern:

- Production auth (disable dev headers), strict tenancy, and admin RBAC:
  [Auth & tenancy](../reference/auth-and-tenancy.md) and [Admin & bootstrap](../reference/api/admin.md)
- Per-tenant limits and monitoring (at least rate limiting + error/empty-recs guardrails)
- Clear tenant onboarding runbook (including DB bootstrap steps)

Start here:

- How-to: [Deploy with Helm](../how-to/deploy-helm.md)
- Reference: [Auth & tenancy](../reference/auth-and-tenancy.md)
- Checklist: [Production readiness checklist](../operations/production-readiness-checklist.md)

## Read next

- Minimum components by goal: [`minimum-components-by-goal.md`](minimum-components-by-goal.md)
- Data modes (details): [`explanation/data-modes.md`](../explanation/data-modes.md)
- Deploy with Helm: [`how-to/deploy-helm.md`](../how-to/deploy-helm.md)
- Production readiness checklist: [`operations/production-readiness-checklist.md`](../operations/production-readiness-checklist.md)
- Security, privacy, compliance: [`start-here/security-privacy-compliance.md`](security-privacy-compliance.md)
