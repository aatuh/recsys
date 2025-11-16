# Simulations & Guardrails

Simulations let you replay realistic catalogs and user journeys against RecSys before shipping any configuration or rule changes. Guardrails (automatic checks defined in `docs/concepts_and_metrics.md`) enforce the business rules around coverage (how much of the catalog appears over time), diversity (how varied the recommendations are), and starter experience (what new or low-data users see). Treat them as one workflow: simulations create evidence; guardrails encode the pass/fail criteria.

> ⚠️ **Advanced topic**
>
> Read this after you have a basic integration running and want to make changes safely.
>
> Who should read this? Developers and ops engineers running RecSys from source (local Docker or managed deployments). Hosted API users who only need HTTP payloads can stop at `docs/quickstart_http.md`.
> Need a quick reference for every script mentioned here? Check `docs/analysis_scripts_reference.md`.
>
> **Where this fits:** Guardrails & safety.

### TL;DR

- **Purpose:** Teach you how to build fixtures, run the simulation harness, and enforce guardrails before shipping config or rule changes.
- **Use this when:** You need pre-deployment evidence, want to replay a customer catalog, or must diagnose a guardrail failure from CI.
- **Outcome:** Bundled reports (quality metrics, scenario summaries, guardrail verdicts) tied to specific env profiles and fixtures.
- **Not for:** Simple QA of API payloads or hosted integrations—stick to `docs/quickstart_http.md` for those scenarios.

---

## What guardrails are (plain language)

- Guardrails are **automatic checks** that block bad changes (for example, drops in key metrics or loss of starter experiences) before they reach real users.
- They encode **minimum expectations** for coverage, diversity, and quality so experiments cannot silently erode your KPIs.
- They run the same way on every change, which means stakeholders can trust that configuration and rule updates meet an agreed safety bar.

### How simulations fit in

- Simulations **replay traffic or scripted scenarios** against proposed configurations to produce metrics and evidence bundles.
- Guardrails read those metrics and decide whether a change passes or fails.
- Together, they let you review a change with data (“show me the impact on coverage and starter profiles”) before pushing it live.

---

## 1. Why run simulations?

- **Onboard safely:** Validate a customer’s dataset and configuration before exposing real traffic.
- **Regression testing:** Catch coverage or diversity regressions caused by tuning/rules changes.
- **Evidence for stakeholders:** Produce shareable bundles (metrics, traces, manifests) proving the change works.
- **CI/CD:** Automate scenario checks so PRs cannot merge when guardrails would fail.

---

## 2. Build fixtures

1. Start from `analysis/fixtures/templates/*` or `analysis/fixtures/sample_customer.json`.
2. Populate `items`, `users`, and `events` to mirror the cohort you care about (cold start, long-tail shoppers, etc.). Keep namespaces consistent with the env profile you plan to use.
3. Validate locally:

```bash
python analysis/scripts/seed_dataset.py \
  --base-url http://localhost:8000 \
  --namespace customer_a \
  --org-id "$RECSYS_ORG_ID" \
  --fixture-path analysis/fixtures/customers/customer_a.json
```

Review `analysis/evidence/seed_segments.json` to confirm segment counts before committing the fixture.

---

## 3. Run simulations

### 3.1 Single customer / namespace

> Local-only note: the commands in this section assume Docker, Make, and the repo scripts are available. Hosted API consumers can skip them.

```bash
python analysis/scripts/run_simulation.py \
  --customer customer-a \
  --base-url http://localhost:8000 \
  --namespace customer_a \
  --org-id "$RECSYS_ORG_ID" \
  --env-profile customer_a \
  --fixture-path analysis/fixtures/customers/customer_a.json
```

What happens:

1. Loads the requested env profile and guardrail config.
2. Resets the namespace, seeds the fixture, and executes `run_quality_eval.py`.
3. Runs `run_scenarios.py` (the full policy/regression scenario suite) plus optional determinism/load tests.
4. Stores everything under `analysis/reports/customer-a/<timestamp>/` (metrics, traces, Markdown summary).

### 3.2 Batch simulations

```bash
python analysis/scripts/run_simulation.py \
  --batch-file analysis/fixtures/batch_simulations.yaml \
  --batch-name pilot-rollout
```

This iterates through each entry (fixture + namespace + env profile) and produces per-customer reports plus a consolidated batch summary under `analysis/reports/batches/`.

---

## 4. Guardrails & `guardrails.yml`

`guardrails.yml` defines thresholds per customer/namespace. Structure:

```yaml
defaults:
  quality:
    min_segment_lift_ndcg: 0.10
    min_segment_lift_mrr: 0.10
    min_catalog_coverage: 0.60
    min_long_tail_share: 0.20
  scenarios:
    s7_min_avg_mrr: 0.20
    s7_min_avg_categories: 4

customers:
  beta-retail:
    namespace: default
    quality:
      min_segment_lift_ndcg: 0.12
      min_segment_lift_mrr: 0.11
      min_catalog_coverage: 0.65
      min_long_tail_share: 0.25
```

- **Quality thresholds** map to metrics emitted by `run_quality_eval.py` (`segment_ndcg_lift`, `catalog_coverage`, etc.).
- **Scenario thresholds** align with the starter-profile guardrail, the boost exposure checks, and other scenarios defined under `analysis/scripts/run_scenarios.py`.
- `analysis/scripts/check_guardrails.py` reads this YAML and fails fast when any threshold is violated. CI runs the same script.
- Additional namespaces can be added under `customers:`. Use descriptive keys so reports remain readable.

> **Tip:** Treat guardrails as contracts. When a threshold fails, adjust configs or data—not the guardrail—unless stakeholders approve the new target.

---

## 5. Interpreting results

- **`quality_metrics.json`** – segment lifts, coverage, long-tail share, determinism stats. Compare against the guardrail thresholds above.
- **`scenario_summary.json` / CSV** – pass/fail per scenario. Pay special attention to:
  - **Starter-profile guardrail**: cold-start experience requires at least a minimum Mean Reciprocal Rank (MRR, “how early do good items appear?”) and category diversity.
  - **Boost exposure checks**: validate promotion knobs and diversity caps.
  - **Rule-heavy flows**: ensure pin/boost/block behavior remains correct.
- **`seed_manifest.json` / `seed_segments.json`** – prove the seeded dataset matches expectations when diagnosing coverage issues.
- **Markdown report** – shareable overview linking to every artifact in the bundle.

---

## 6. Integrating with tuning & deployments

1. Run the tuning harness (`docs/tuning_playbook.md`) to explore configuration options.
2. Use simulations to validate the winning candidates against guardrails.
3. Commit env profiles, guardrail updates, and evidence artifacts in the same PR for traceability.
4. Let CI execute `run_simulation.py` (or at least `check_guardrails.py`) on every change touching configs or rules.
5. Ops teams monitor Prometheus metrics (`policy_guardrail_failures_total`, `policy_rule_blocked_items_total`) to catch drifts after deployment.

---

## 7. Related commands & docs

- `make scenario-suite SCENARIO_BASE_URL=http://localhost:8000` – seeds + runs the full scenario suite quickly (override variables as needed).
- `make determinism`, `make load-test` – specialty guardrails for reproducibility and scale.
- `docs/rules_runbook.md` – operational steps when guardrails trip or rules misbehave.
- `docs/concepts_and_metrics.md` – definitions for metrics mentioned here.
- `docs/env_reference.md` – canonical knob list referenced while tuning for guardrail targets.

---

## Where to go next

- If you’re integrating HTTP calls → see `docs/quickstart_http.md`.
- If you’re a PM → skim `docs/business_overview.md`.
- If you’re tuning quality → read `docs/tuning_playbook.md`.
