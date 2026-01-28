# RecSys: A Recommendation Service

RecSys is a domain-agnostic recommendation platform. You send opaque item/user/event IDs and receive top-K recommendations or “similar items” tuned per namespace. The system favors safe defaults, multi-tenant isolation, and clear guardrails (automatic thresholds that block risky changes before they ship; see `docs/concepts_and_metrics.md` for terminology) so you can ship quickly without sacrificing control.

> **Where this fits:** Architecture & mental model.

---

## New here? Start with this path

If you want a two-minute, non-technical explanation first, read [`docs/recsys_in_plain_language.md`](docs/recsys_in_plain_language.md).

1. **Understand the product** – Read [`docs/business_overview.md`](docs/business_overview.md) to see what RecSys does, why guardrails matter, and how it fits into your business workflows.
2. **Follow the narrative tour** – Walk through [`docs/zero_to_first_recommendation.md`](docs/zero_to_first_recommendation.md) to ingest a tiny catalog and fetch your first recommendations with Acme Outfitters.
3. **Integrate via HTTP** – Use [`docs/quickstart_http.md`](docs/quickstart_http.md) for the full hosted API quickstart (ingestion, troubleshooting, error handling). If you only want to prove the loop works end to end, start with the “Hello RecSys in 3 calls” section in that doc.
4. **Dive into the reference** – Keep [`docs/api_reference.md`](docs/api_reference.md) handy for every endpoint, limits, and behavioral guarantees.

Prefer persona/lifecycle views or a phase-based checklist? After finishing the four steps above, skim [`docs/overview.md`](docs/overview.md) (personas) and [`docs/onboarding_checklist.md`](docs/onboarding_checklist.md) (week-by-week plan).

---

## What you can build

- **Personalized feeds** for storefronts, news, or OTT apps.
- **“Similar items” widgets** on PDPs that favor diversity and cold-start safety.
- **Cart/checkout upsells** that respect availability, caps, and exclusions.
- **Triggered experiences** (email, push) using dedicated namespaces and overrides.
- **Merchandising tools** that preview rule changes before they go live.

---

## Quickstart summary

> Local-only note: the steps below are for developers running RecSys from source. Hosted API consumers can skip to [`docs/quickstart_http.md`](docs/quickstart_http.md).

1. **Prep env files:** `make env PROFILE=dev`
2. **Start the stack:** `make dev` (Go API, Postgres, proxy, UI)
3. **Seed data:** `python3 analysis/scripts/seed_dataset.py --base-url http://localhost:8000 --org-id 00000000-0000-0000-0000-000000000001 --namespace demo`
4. **Call the API:** `curl -H "X-Org-ID: …" -d '{"namespace":"demo","user_id":"user_0001","k":8}' http://localhost:8000/v1/recommendations`

Need the detailed walkthrough (logs, troubleshooting, screenshots)? Head to [`GETTING_STARTED.md`](GETTING_STARTED.md). For hosted integrations that only need HTTP calls, use [`docs/quickstart_http.md`](docs/quickstart_http.md).

---

### Hosted API vs local stack

- **Hosted integration** – Follow [`docs/quickstart_http.md`](docs/quickstart_http.md) (plus [`docs/zero_to_first_recommendation.md`](docs/zero_to_first_recommendation.md)) if you only need the managed HTTP API.
- **Local stack / contributors** – Follow [`GETTING_STARTED.md`](GETTING_STARTED.md) if you want to run the repo locally, inspect internals, or contribute code.

---

## Persona map – where to go next

- **Business / Product** – focus on value narrative, rollout, and guardrails. Read: [`docs/business_overview.md`](docs/business_overview.md), [`docs/overview.md`](docs/overview.md), [`docs/concepts_and_metrics.md`](docs/concepts_and_metrics.md), [`docs/rules_runbook.md`](docs/rules_runbook.md).
- **Integration Engineer** – need ingestion, API usage, namespace config. Read: [`GETTING_STARTED.md`](GETTING_STARTED.md), [`docs/quickstart_http.md`](docs/quickstart_http.md), [`docs/api_reference.md`](docs/api_reference.md), [`docs/env_reference.md`](docs/env_reference.md), [`docs/database_schema.md`](docs/database_schema.md).
- **Developer / Ops** – tuning, guardrails, on-call. Read: [`docs/tuning_playbook.md`](docs/tuning_playbook.md), [`docs/simulations_and_guardrails.md`](docs/simulations_and_guardrails.md), [`docs/rules_runbook.md`](docs/rules_runbook.md), [`docs/overview.md`](docs/overview.md).

Each doc starts with “Who should read this?” so you can skim with confidence.

New teammate? Follow [`docs/onboarding_checklist.md`](docs/onboarding_checklist.md) for a suggested week-one ramp-up path tailored to engineers and PMs.

---

## Core capabilities at a glance

- **Opinionated ingestion pipeline** – `/v1/items:upsert`, `/v1/users:upsert`, `/v1/events:batch` plus tooling to seed fixtures and reset namespaces safely.
- **Blend + personalization engine** – popularity, co-visitation, embeddings, and tag profiles combined with Maximal Marginal Relevance (MMR, a diversity-aware re-ranking method; see [`docs/concepts_and_metrics.md`](docs/concepts_and_metrics.md)) and caps for diversity.
- **Guardrails & simulations** – repeatable scenario harness, quality metrics (Normalized Discounted Cumulative Gain (NDCG, a ranking quality score), Mean Reciprocal Rank (MRR, “how early do relevant items appear?”), coverage (how much of the catalog we actually surface)), and YAML guardrails that block regressions automatically. See [`docs/concepts_and_metrics.md`](docs/concepts_and_metrics.md) for full definitions.
- **Rule engine & overrides** – boost/pin/block APIs, dry-run support, and telemetry for exposure caps.
- **Observability** – structured traces, Prometheus metrics, and decision evidence stored with every tuning run.

---

## Advanced workflows

- [`docs/tuning_playbook.md`](docs/tuning_playbook.md) – reset → seed → tune → promote workflow plus AI-assisted sweeps.
- [`docs/simulations_and_guardrails.md`](docs/simulations_and_guardrails.md) – how bespoke fixtures and guardrails cooperate in CI/CD.
- [`docs/env_reference.md`](docs/env_reference.md) – canonical environment/config knob list with override guidance.
- [`docs/commands.md`](docs/commands.md) – offline Go command guide for catalog backfills, collaborative factors, and blend evaluations.
- [`docs/api_reference.md`](docs/api_reference.md) – endpoint catalog with payload notes, error codes, and common patterns.
- [`docs/rules_runbook.md`](docs/rules_runbook.md) – operational runbook for overrides, telemetry, and incident response.
- [`docs/configuration.md`](docs/configuration.md) – conceptual explanation of how ingestion, signals, blending, personalization, and rules fit together.
- [`docs/security_and_data_handling.md`](docs/security_and_data_handling.md) – transport security, auth, retention expectations.
- [`docs/doc_ci.md`](docs/doc_ci.md) – how to run the documentation link checker and client example tests locally or in CI.
- [`docs/doc_style.md`](docs/doc_style.md) – shared style and terminology guide for writing/maintaining these docs.

---

## Repository layout & key commands

- `api/` – Go service, migrations, REST handlers. Run `make test` here for API tests.
- `web/` – Vite + React admin/demo UI. Use `pnpm dev`, `pnpm lint`, `pnpm typecheck`.
- `docs/` – Conceptual, runbook, and reference docs (see persona map).
- `analysis/` – Scripts for seeding, tuning, guardrails, load tests, and evidence artifacts.
- `db/`, `proxy/`, `shop/`, `demo/` – Supporting services for local development.

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

- Metrics & jargon primer – [`docs/concepts_and_metrics.md`](docs/concepts_and_metrics.md)
- Business storytelling – [`docs/business_overview.md`](docs/business_overview.md)
- Database schema & SQL tips – [`docs/database_schema.md`](docs/database_schema.md)
- Configuration mindsets – [`docs/configuration.md`](docs/configuration.md)
- System architecture overview – architecture overview lives in [`docs/configuration.md`](docs/configuration.md), [`docs/overview.md`](docs/overview.md), and [`docs/analysis_scripts_reference.md`](docs/analysis_scripts_reference.md).
- Doc map by role/task – [`docs/doc_map.md`](docs/doc_map.md)
- FAQ & troubleshooting – [`docs/faq_and_troubleshooting.md`](docs/faq_and_troubleshooting.md)
