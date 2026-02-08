---
diataxis: how-to
tags:
  - how-to
  - integration
  - troubleshooting
  - developer
---
# Troubleshooting for integrators
This guide shows how to troubleshooting for integrators in a reliable, repeatable way.


## Who this is for

- Developers integrating `recsys-service` into an application.
- On-call engineers debugging “empty recs” or unexpected behavior.

## What you will get

- A symptom → cause → fix checklist
- Links to the canonical runbooks and reference specs

## Before you debug: collect these facts

- Tenant ID (`X-Org-Id`) and surface name (`surface`)
- Request ID (`X-Request-Id`) used for the call
- Whether you are in **DB-only** or **artifact/manifest** mode
- Whether exposure logging is enabled and where logs are written

Reference:

- Tenant + auth model: [Auth and tenancy reference](../reference/auth-and-tenancy.md)
- Data modes: [Data modes: DB-only vs artifact/manifest](../explanation/data-modes.md)

## Symptom checklist

### Symptom: `items` is empty

Most common causes:

- No candidates exist for this tenant + surface
- Rules/config exclude everything
- Artifact/manifest points to missing or stale data (artifact mode)

Fix path:

1. Verify tenant scope and surface spelling (must match your data namespaces)  
   See: [Surface namespaces](../explanation/surface-namespaces.md)
2. Check the “Empty recs” runbook  
   See: [Runbook: Empty recs](../operations/runbooks/empty-recs.md)
3. If artifact mode: check stale manifest runbook  
   See: [Runbook: Stale manifest (artifact mode)](../operations/runbooks/stale-manifest.md)

### Symptom: 401 / 403 / tenant seems “wrong”

Most common causes:

- Missing tenant header/claim
- Dev headers used in an environment that requires auth
- Tenant ID mismatch between config/rules and request

Fix path:

- Reference: [Auth and tenancy reference](../reference/auth-and-tenancy.md)
- Admin endpoints (config/rules): [Admin API + local bootstrap (recsys-service)](../reference/api/admin.md)

### Symptom: results change unexpectedly between calls

Most common causes:

- Inputs are not actually identical (request ID, experiment metadata, exclude list)
- Non-deterministic candidate source ordering (ties without stable ordering)
- You switched data without realizing (artifact refresh / DB update)

Fix path:

- Determinism definition: [How it works: architecture and data flow](../explanation/how-it-works.md)
- Verify determinism tutorial: [Verify determinism](../tutorials/verify-determinism.md)
- Ranking determinism pitfalls: [Ranking & constraints reference](../recsys-algo/ranking-reference.md)

### Symptom: exposure logs are missing or not joinable

Most common causes:

- Exposure logging is disabled
- Log path is wrong or not persisted
- Missing/unstable request IDs or user/session IDs across platforms

Fix path:

- Minimum instrumentation spec: [Minimum instrumentation spec (for credible evaluation)](../reference/minimum-instrumentation.md)
- Join logic: [Event join logic (exposures ↔ outcomes ↔ assignments)](../reference/data-contracts/join-logic.md)
- Verify joinability tutorial: [Verify joinability (request IDs → outcomes)](../tutorials/verify-joinability.md)

### Symptom: service never becomes healthy

Fix path:

- Runbook: [Runbook: Service not ready](../operations/runbooks/service-not-ready.md)
- If migrations fail: [Runbook: Database migration issues](../operations/runbooks/db-migration-issues.md)

## Read next

- Integration checklist: [How-to: Integration checklist (one surface)](integration-checklist.md)
- API reference: [API Reference](../reference/api/api-reference.md)
- Production readiness checklist: [Production readiness checklist (RecSys suite)](../operations/production-readiness-checklist.md)
