# Business Overview

This narrative explains the RecSys platform for product managers, merchandisers, and business stakeholders. It focuses on value, rollout expectations, and safety controls instead of scripts or infrastructure.

---

## 1. What the system does

- Serves **domain-agnostic recommendations** (products, content, listings) through a stable HTTP API.
- Supports multiple surfaces: personalized feeds, PDP “similar items,” search/rerank, cart/checkout upsells, and triggered emails.
- Keeps **business guardrails** front and center—caps, overrides, curated starter experiences—so experimentation cannot silently damage key KPIs (see `docs/concepts_and_metrics.md` for definitions).
- Provides an **audit trail** for every decision (who/what/why) and plugs into Prometheus for fleet health monitoring.

Think of it as a recommender **control plane**: ingestion, ranking, and guardrails that your apps and merchandising tools sit on top of.

---

## 2. Example use cases

1. **E-commerce storefront:** Personalized “New for you” carousel on the homepage, “Frequently bought with” widgets on PDPs, and cart recommendations tied to current basket contents.
2. **Marketplace or classifieds:** Show similar listings, cross-promote inventory from underexposed sellers, ensure rules prevent repeats from the same merchant.
3. **Content feed / OTT:** Recommend shows/articles using embeddings plus guardrails that guarantee genre diversity and highlight editorial picks when required.
4. **Email / CRM:** Send weekly digests by calling `/v1/recommendations` with a dedicated namespace tuned for long-tail exposure, ensuring repeat content is throttled.
5. **Internal catalog tooling:** Merchandisers preview rule changes via `/v1/admin/rules/dry-run` before promoting curated collections to customers.

- **Acme Retail rollout (fictional case study):**
  1. **Pain:** Acme’s PDP widgets were static “top sellers” picked manually. Merchandisers spent hours curating lists, and cold-start shoppers saw irrelevant items.
  2. **Implementation:** Over three weeks they ingested their catalog/users/events, tuned blend weights for home/PDP/cart surfaces, and enforced guardrails via the simulation suite. They piloted on PDP with a 50/50 bandit experiment before rolling out globally.
  3. **Outcome:** PDP CTR rose from 4.6% → 5.3%, add-to-cart rate +2%, and long-tail exposure improved from 12% → 22%. Merchandisers now use the rule runbook to launch seasonal campaigns safely.
  4. **How:** The story followed the same phases outlined below (Foundations → Tuning → Operationalization → Continuous improvement); the data and guardrail evidence live alongside their PRs for future audits.

---

## 3. Rollout story

- **Week 1 – Foundations** — Spin up the stack (`GETTING_STARTED.md`), mirror a sample catalog, wire ingestion (items/users/events), and verify `/v1/recommendations` output for sample users.
- **Week 2–4 – Tuning & guardrails** — Adjust blend/MMR/cap knobs per surface, run the tuning harness on representative segments, and enforce guardrails via simulations before onboarding traffic.
- **Week 5–8 – Operationalization** — Enable bandits or multi-surface namespaces, expose the rule engine to merchandising, integrate dashboards/alerts, and document rollout plans per region.
- **8+ weeks – Continuous improvement** — Experiment with custom signals, expand guardrails (fairness, supply balancing), and plug auditing endpoints into governance/compliance tooling.

This cadence gives business stakeholders confidence that quality and safety checks accompany every change, not just the initial launch.

---

## 4. Safety and guardrails

- **Automated simulations** replay scripted user journeys (cold start, power users, long-tail shoppers) before config changes ship. They produce metrics + screenshots that PMs can review.
- **Guardrail policies** (YAML) encode coverage floors, diversity requirements, and “starter experience” checks. CI fails fast if any metric drops below thresholds.
- **Rules runbook** defines how overrides interact with guardrails, ensuring manual boosts cannot starve critical cohorts.
- **Observation hooks** (Prometheus metrics + structured traces) make regressions visible in dashboards within minutes.

Together they create a “trust but verify” loop: experimentation is easy, but nothing rolls out silently.

---

## 5. Evidence & auditability

- **Decision traces** – Every `/v1/recommendations` call can return the scoring trace (`include_reasons=true`), showing policy name, signal contributions, and applied rules.
- **Metrics pipelines** – Blend/coverage/diversity stats are exported to Prometheus and can feed Grafana or Datadog.
- **Simulation bundles** – Each tuning run stores fixtures, payloads, and quality reports under `analysis/results/<namespace_timestamp>/` so PMs can review side-by-side.
- **Rule history** – Admin endpoints persist who created/edited rules and provide dry-run payloads for compliance teams.

---

## 6. Where to dig deeper

- `README.md` – Front-door summary plus persona map.
- `docs/overview.md` – Lifecycle checklist for business, integration, and ops personas.
- `docs/concepts_and_metrics.md` – Definitions of key metrics (NDCG/MRR), diversity, guardrails, and other terminology.
- `docs/rules_runbook.md` – Operational steps for overrides and incident response.
- `docs/simulations_and_guardrails.md` – How automated protection works.
- `docs/quickstart_http.md` – Copy-paste API examples your engineers will use.
- `docs/api_errors_and_limits.md` – Error codes, rate limits, and retry guidance for engineers.
- `docs/security_and_data_handling.md` – Transport security, auth model, and data retention guidance for stakeholders who care about compliance.

---

## Advanced concepts (optional)

Curious about the algorithms behind the scenes (embedding signals, blend weights, exploration policies)? `docs/concepts_and_metrics.md` explains the terms in more detail, and the analytics plans under `analytics/` walk through tuning experiments. These are helpful for ML-minded readers but not required to understand the product value.

---

## Product & PM FAQ

**How do we measure success?**

- Track CTR/add-to-cart/revenue per surface, plus guardrail metrics (NDCG, MRR, coverage, long-tail share). Each tuning run and simulation bundle includes these values so you can compare before/after.

**Who owns algorithm knobs vs rules?**

- Engineering/ML teams own blend weights, MMR, personalization, and retriever tuning (see `docs/tuning_playbook.md`). Merchandising owns rules/overrides via `/v1/admin/rules` and `/v1/admin/manual_overrides` following `docs/rules_runbook.md`. Guardrails ensure neither side accidentally hurts key cohorts.

**How long before we see impact after launch?**

- Expect 1–2 weeks to seed data, tune, and pass guardrails, then another 2–4 weeks of monitored ramp. Bandit experiments and simulations keep risk low during this period.

**What happens if recommendations look wrong?**

- Capture `trace_id` from the client and inspect `/v1/audit/decisions/{trace_id}`. Check guardrail dashboards for failures and review recent rule changes. Re-run the simulation suite if necessary.

**How do we roll out safely?**

- Use the simulation harness + guardrails before promoting changes, then ramp with bandits or feature flags (see `analytics/bandit_exploration_plan.md`). Guardrail alerts (Prometheus) and decision traces catch regressions early.

**What if a stakeholder wants specific items promoted or removed?**

- Use rules/overrides plus `/v1/admin/rules/dry-run` to preview effects. The runbook explains how to keep guardrails satisfied while honoring campaigns.
