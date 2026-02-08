---
diataxis: tutorial
tags:
  - tutorial
  - determinism
  - developer
  - recsys-algo
---
# Tutorial: Verify determinism
In this tutorial you will follow a guided walkthrough and verify a working result.


## Who this is for

- Developers and recsys engineers who need confidence that ranking output is stable and auditable.

## What you will get

- A practical test that validates the operational definition of “deterministic” in this suite
- A checklist for the most common sources of non-determinism

## Prereqs

- A running `recsys-service` instance (local or staging)
- A known tenant + surface with non-empty candidates

If you don't have one yet, start with: [Quickstart (10 minutes)](quickstart.md)

## Operational definition (what “deterministic” means here)

For a given tenant, `POST /v1/recommend` is deterministic when the following inputs are the same:

- tenant scope + surface
- request payload (including exclude lists, k, etc.)
- request ID and experiment metadata (if enabled)
- underlying candidate data and config/rules
- ranking implementation

See: [How it works: architecture and data flow](../explanation/how-it-works.md)

## Step 1: Pick stable inputs

Choose:

- `tenant = demo` (or your tenant)
- `surface = home`
- `request_id = det-1`
- A fixed payload (same JSON every time)

## Step 2: Call the endpoint repeatedly

Run this 10 times and capture the raw JSON:

```bash
for i in $(seq 1 10); do
  curl -fsS http://localhost:8000/v1/recommend     -H 'Content-Type: application/json'     -H 'X-Request-Id: det-1'     -H 'X-Dev-User-Id: dev-user-1'     -H 'X-Dev-Org-Id: demo'     -H 'X-Org-Id: demo'     -d '{"surface":"home","k":10,"user":{"user_id":"u_1","session_id":"s_1"}}'     > "/tmp/recsys.det.$i.json"
done
```

## Step 3: Compare outputs

If you have `jq`, compare the ordered item IDs:

```bash
for i in $(seq 1 10); do
  jq -r '.items[].item_id' "/tmp/recsys.det.$i.json" | tr '\n' ',' ; echo
done | uniq -c
```

Expected:

- The sequence of item IDs is identical across runs.

## Step 4: If it’s not deterministic, use this checklist

Common causes:

- **Inputs differ** between calls (request ID, exclude list, experiment metadata)
- **Candidate sources are unstable** (ties without stable ordering)
- **Store backends are not deterministic** (DB queries without a stable sort)
- **Rules change** (admin config/rules updated between calls)

Reference:

- Determinism pitfalls: [Ranking & constraints reference](../recsys-algo/ranking-reference.md)

## Verify (Definition of Done)

- [ ] With identical inputs, the ordered `item_id` list is identical across 10 runs.

## Read next

- Candidate vs ranking (mental model): [Candidate generation vs ranking](../explanation/candidate-vs-ranking.md)
- Run a local end-to-end loop: [Tutorial: Local end-to-end (20–30 minutes)](local-end-to-end.md)
- Verify joinability: [Tutorial: verify joinability](verify-joinability.md)
