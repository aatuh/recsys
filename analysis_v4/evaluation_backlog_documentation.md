# Recsys Comprehensive Documentation Backlog (AP-301+)

## Problem Statement
The current documentation (README, runbooks, bespoke simulation guide) covers
workflows and guardrails, but lacks a single source of truth for:
- API endpoints (methods, payloads, use cases, sample requests).
- Algorithm/env configuration (what each knob does, interaction guidance, override policy).
- Database schema (tables, columns, relationships) that integrators need when connecting data sources.

To support business stakeholders, integration engineers, and developers, we need
polished documentation that explains the system end-to-end in plain English, with
actionable examples.

---

## Epics & Tickets

### AP-301: API Reference
Publish an endpoint catalog that explains every REST route, its payloads, and typical use cases for operators and integrators.

- [x] **AP-301A – Endpoint survey**  
  *Inventory all public endpoints (`/v1/recommendations`, `/v1/bandit/*`, `/v1/admin/*`, `/v1/items/users/events`, audit/health) and capture method, auth requirements, rate limits, and request/response schemas.*  
  - Added `docs/api_endpoints.md` enumerating ingestion, ranking, bandit, configuration, admin, audit, and health endpoints with descriptions, typical users, and cross-links to Swagger.

- [x] **AP-301B – API reference document**  
  *Create `docs/api_endpoints.md` with sections per surface (ranking, bandit, admin, data-management). Include tables describing parameters, response fields, guardrails, and sample curl requests. Link to Swagger where relevant.*  
  - Expanded `docs/api_endpoints.md` with common request fields, sample curl commands, bandit workflow notes, configuration tips, and audit guidance plus cross-links to other docs.

- [x] **AP-301C – README integration**  
  *Update README to summarize the endpoint categories and link to the API reference so business/integrator personas can find the right routes quickly.*  
  - README now features an “API Surface at a Glance” table that highlights ingestion, ranking, configuration, bandit, and audit endpoints with pointers to `docs/api_endpoints.md` for full details.

### AP-302: Env Var & Algorithm Guide
Explain every algorithm-related env var in plain English, covering interactions and recommended tuning strategies.

- [x] **AP-302A – Env var audit**  
  *Enumerate all relevant knobs (retriever blends, personalization, coverage, rules, bandits). Note dependencies (e.g., `RULE_EXCLUDE_EVENTS` ↔ `PURCHASED_WINDOW_DAYS`) and current override coverage.*  
  - `docs/env_vars.md` now lists every algorithm/env knob with descriptions, impacts, related settings, and override availability—including interaction notes and guardrail considerations.

- [x] **AP-302B – Env var reference doc**  
  *Produce `docs/env_vars.md` (or expand README) with a structured table: variable, default, impact on results, related knobs, on-the-fly override availability, and cautions. Include scenario-based examples (e.g., “boost long-tail coverage” → tweak `POPULARITY_FANOUT`, `BLEND_GAMMA`).*  
  - `docs/env_vars.md` now groups knobs by category (retrievers, personalization, diversity, rules, bandit), lists impacts/interactions, and calls out tips (e.g., blending adjustments, purchase suppression pairs). README links to the doc for quick reference.

- [x] **AP-302C – Override policy section**  
  *Document when to use env profiles vs. per-request overrides vs. guardrail adjustments. Update README and the bespoke simulation guide to link to this policy.*  
  - `docs/env_vars.md` now contains a “Runtime overrides vs. env profiles” policy (per-request overrides vs env profiles vs admin APIs), and README links to it under configuration.

### AP-303: Database Schema Guide
Document the Postgres schema so integrators know how to ingest and query data safely.

- [x] **AP-303A – Schema extraction**  
  *Dump table/column metadata from migrations (items, users, events, rules, bandit tables, audit tables). Capture types, primary keys, and important indexes.*  
  - Summaries added to `docs/database_schema.md` pulling field info from `api/migrations/*.sql`.

- [x] **AP-303B – Schema documentation**  
  *Create `docs/database_schema.md` with per-table sections explaining each column, units, nullable vs. required, and relationships (FKs). Add diagrams or bullets for key workflows (e.g., how events drive recommendations).*  
  - Doc now covers catalog tables, events, segments, rules/manual overrides, bandit policies, decision traces, and embedding factors with descriptions and tips.

- [x] **AP-303C – Developer/integration guidance**  
  *Add a README subsection linking to the schema doc and outlining common tasks (seeding items, exporting audit data, cleaning namespaces). Include tips about retention, migrations, and sample SQL queries.*  
  - `docs/database_schema.md` now includes a “Developer & Integration Guidance” section (seeding verification, guardrail troubleshooting, namespace cleanup, exporting data, schema evolution). README links to the schema doc.

### AP-304: Persona-Focused Overview
Ensure business stakeholders and integrator engineers have a narrative that ties endpoints, env knobs, and schema together.

- [x] **AP-304A – Business overview**  
  *Add a short “How Recsys Works” guide (could be README or `docs/overview.md`) describing the lifecycle: seed data → configure env → run guardrails → monitor telemetry. Highlight key docs for each persona.*  
  - Created `docs/overview.md` outlining workflows for business stakeholders, integration engineers, and developers, with links to deeper docs.

- [x] **AP-304B – Navigation & cross-links**  
  *Update README/bespoke_simulations/runbook to link to the new API/env/schema docs, ensuring there’s a cohesive doc tree.*  
  - README now points to `docs/api_endpoints.md`, `docs/env_vars.md`, `docs/database_schema.md`, and `docs/overview.md`; other docs cross-reference each other.

---

Each ticket should update the relevant documentation files and cross-links so the new references are easy to discover from README, CI outputs, and the bespoke simulation guide. Once these epics are done, Recsys will have business-friendly, integrator-ready, and developer-focused documentation covering APIs, configuration, and data storage. 
- [x] **AP-305A – Documentation map / README entry point**  
  *Add a “Start Here” section (README or docs/README.md) that explains the doc hierarchy and points each persona to the right guide (API reference, env vars, schema, overview).*  
  - README now includes a “Start Here” section with persona-to-doc mappings and quick links.

- [x] **AP-305B – Cross-linking & personas**  
  *Update each major doc (API reference, env vars, database schema, overview, rules runbook) with a “Who should read this” blurb and cross-links to related docs. Ensure business/integrator/developer personas have a clear pathway through the docs.*  
  - Added audience blurbs to `docs/api_endpoints.md`, `docs/env_vars.md`, `docs/database_schema.md`, `docs/overview.md`, and `docs/rules-runbook.md`, reinforcing how the docs interconnect.

- [x] **AP-305C – Lifecycle narrative**  
  *Expand README (or docs/overview.md) with a numbered lifecycle checklist covering seed → configure → simulate/guardrails → deploy → monitor, referencing the relevant commands/docs at each step.*  
  - README + `docs/overview.md` now include lifecycle checklists covering seeding, config, simulations, deployment, and monitoring, with links to the relevant docs/commands.
