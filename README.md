# RecSys: A Recommendation Service

RecSys is a domain-agnostic recommendation platform. You send opaque item/user/event IDs and receive top-K recommendations or “similar items” tuned per namespace. The system favors safe defaults, multi-tenant isolation, and clear guardrails (see `docs/concepts_and_metrics.md` for terminology) so you can ship quickly without sacrificing control.

---

## What you can build

- **Personalized feeds** for storefronts, news, or OTT apps.
- **“Similar items” widgets** on PDPs that favor diversity and cold-start safety.
- **Cart/checkout upsells** that respect availability, caps, and exclusions.
- **Triggered experiences** (email, push) using dedicated namespaces and overrides.
- **Merchandising tools** that preview rule changes before they go live.

---

## Quickstart summary

> Local-only note: the steps below are for developers running RecSys from source. Hosted API consumers can skip to `docs/quickstart_http.md`.

1. **Prep env files:** `make env PROFILE=dev`
2. **Start the stack:** `make dev` (Go API, Postgres, proxy, UI)
3. **Seed data:** `python3 analysis/scripts/seed_dataset.py --base-url http://localhost:8000 --org-id 00000000-0000-0000-0000-000000000001 --namespace demo`
4. **Call the API:** `curl -H "X-Org-ID: …" -d '{"namespace":"demo","user_id":"user_0001","k":8}' http://localhost:8000/v1/recommendations`

Need the detailed walkthrough (logs, troubleshooting, screenshots)? Head to `GETTING_STARTED.md`. For hosted integrations that only need HTTP calls, use `docs/quickstart_http.md`.

---

## Persona map – where to go next

| Persona / Goal | Read this |
|----------------|-----------|
| **Business / Product** – understand the value, rollout story, guardrails | `docs/business_overview.md`, `docs/overview.md`, `docs/concepts_and_metrics.md`, `docs/rules-runbook.md` |
| **Integration Engineer** – ingest data, hit APIs, configure namespaces | `GETTING_STARTED.md`, `docs/quickstart_http.md`, `docs/api_reference.md`, `docs/env_reference.md`, `docs/database_schema.md` |
| **Developer / Ops** – tune policies, run guardrails, support on-call | `docs/tuning_playbook.md`, `docs/simulations_and_guardrails.md`, `docs/rules-runbook.md`, `docs/overview.md` |

Each doc starts with “Who should read this?” so you can skim with confidence.

New teammate? Follow `docs/onboarding_checklist.md` for a suggested week-one ramp-up path tailored to engineers and PMs.

---

## Core capabilities at a glance

- **Opinionated ingestion pipeline** – `/v1/items:upsert`, `/v1/users:upsert`, `/v1/events:batch` plus tooling to seed fixtures and reset namespaces safely.
- **Blend + personalization engine** – popularity, co-visitation, embeddings, and tag profiles combined with Maximal Marginal Relevance (MMR; see `docs/concepts_and_metrics.md`) and caps for diversity.
- **Guardrails & simulations** – repeatable scenario harness, quality metrics (NDCG, MRR, coverage), and YAML guardrails that block regressions automatically.
- **Rule engine & overrides** – boost/pin/block APIs, dry-run support, and telemetry for exposure caps.
- **Observability** – structured traces, Prometheus metrics, and decision evidence stored with every tuning run.

---

## Advanced workflows

- `docs/tuning_playbook.md` – reset → seed → tune → promote workflow plus AI-assisted sweeps.
- `docs/simulations_and_guardrails.md` – how bespoke fixtures and guardrails cooperate in CI/CD.
- `docs/env_reference.md` – canonical environment/config knob list with override guidance.
- `docs/api_reference.md` – endpoint catalog with payload notes, error codes, and common patterns.
- `docs/rules-runbook.md` – operational runbook for overrides, telemetry, and incident response.
- `CONFIGURATION.md` – conceptual explanation of how ingestion, signals, blending, personalization, and rules fit together.

---

## Repository layout & key commands

| Path | What lives here |
|------|-----------------|
| `api/` | Go service, migrations, REST handlers. Run `make test` here for API tests. |
| `web/` | Vite + React admin/demo UI. Use `pnpm dev`, `pnpm lint`, `pnpm typecheck`. |
| `docs/` | Conceptual, runbook, and reference docs (see persona map). |
| `analysis/` | Scripts for seeding, tuning, guardrails, load tests, and evidence artifacts. |
| `db/`, `proxy/`, `shop/`, `demo/` | Supporting services for local development. |

Common commands:

```bash
make dev          # start Docker stack
make down         # stop & reset volumes
make test         # run Go test suite inside api/
make scenario-suite SCENARIO_BASE_URL=http://localhost:8000
pnpm --dir web lint && pnpm --dir web typecheck
```

Run `make help` to list everything else (`load-test`, `reset-namespace`, etc.).

---

## Need deeper context?

- Metrics & jargon primer – `docs/concepts_and_metrics.md`
- Business storytelling – `docs/business_overview.md`
- Database schema & SQL tips – `docs/database_schema.md`
- Configuration mindsets – `CONFIGURATION.md`

File an issue or open a PR with improvement ideas; follow the guidelines in `AGENTS.md` for coding style, testing expectations, and documentation conventions.
