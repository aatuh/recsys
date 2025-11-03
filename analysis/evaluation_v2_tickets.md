# Evaluation v2 Remediation Plan

## Problem Statement
Recent evaluation (`analysis/report.md`) concludes Recsys fails essential policy and configurability benchmarks. Tag-based filters, overrides, and merchandising rules are non-functional, exposing customers to off-policy recommendations and blocking campaign operations. New-user cohorts underperform baselines, and coverage of high-margin knobs is absent. We must restore deterministic policy enforcement, merchandising control fidelity, and guardrails before the next evaluation cycle.

---

## Epic EPC-001 – Policy & Filter Enforcement
Deliver policy-compliant recommendation pipelines that respect request-time constraints and business exclusion rules.

- [x] **REC-101 – Enforce tag include/exclude constraints**  
  Ensure `constraints.include_tags_any` / `exclude_item_ids` are honored end-to-end. Add integration test mirroring S1 payload (`analysis/evidence/scenario_s1_response.json`) to confirm zero leakage and guard against regressions.

- [x] **REC-102 – Restore tag-based block rules**  
  Diagnose why `POST /v1/admin/rules` with `action=BLOCK` does not remove targeted items (`analysis/evidence/scenario_s2_block_high_margin.json`). Patch rule evaluation to exclude blocked candidates and verify via automated scenario test.

- [x] **REC-103 – Brand/attribute whitelist support**  
  Enforced include tag/brand whitelists directly in the ranking pipeline and added regression coverage mirroring scenario S6 to guarantee compliance across org namespaces and rule scopes.

- [x] **REC-104 – Policy regression alerts**  
  Captured policy enforcement summaries in decision traces and exposed Prometheus counters for include-filter drops, exclude hits, rule actions, and constraint leaks; the handler now logs structured warnings whenever leakage is detected.

---

## Epic EPC-002 – Merchandising Controls Reliability
Re-enable manual overrides and rule actions so operators can boost, pin, and shape exposure predictably.

- [x] **REC-201 – Manual boost effect propagation**  
  Trace manual override path (`analysis/evidence/scenario_s3_boost.json`) to algorithm scoring. Fix score injection and add unit/contract tests that assert rank delta > 0 after applying boosts.

- [x] **REC-202 – Pin rules placement logic**  
  Investigate why PIN actions never surface target items (`analysis/evidence/scenario_s5_pin.json`). Repair evaluator and ensure pinned slots respect max pins and don’t get filtered downstream.

- [x] **REC-203 – BOOST/PIN/Block conflict resolution**  
  Formalize deterministic ordering between conflicting rules. Document precedence, add tests covering rule mixtures, and update explanations to expose active rule IDs.

- [ ] **REC-204 – New-item exposure controls**  
  Implement controllable exploration/boost logic so manual actions measurably change appearance rates (`analysis/evidence/scenario_s8_new_item.json`). Provide monitoring on exploration rates by surface.

- [ ] **REC-205 – Multi-objective trade-off curve**  
  Wire boost value multipliers into scoring to yield monotonic margin shifts (`analysis/evidence/scenario_s9_tradeoff.json`). Capture MMR + margin analytics for tuning and expose trade-off presets to ops.

---

## Epic EPC-003 – Quality & Cold-Start Experience
Improve relevance for underserved cohorts while maintaining diversity.

- [ ] **REC-301 – New-user onboarding profiles**  
  Build trait-based or curated starter profiles so `new_users` meet baseline quality. Validate with time-split evaluation showing ≥10% lift per metric (`analysis/quality_metrics.json` segment data).

- [ ] **REC-302 – Personalization decay tuning**  
  Review profile boost / half-life defaults to avoid over-weighting sparse history. Provide tunable config and tests proving stability across all segments.

- [ ] **REC-303 – Diversity knob documentation & presets**  
  Since MMR overrides work (scenario S4), publish recommended `mmr_lambda` presets per surface and expose through admin tooling with validation.

---

## Epic EPC-004 – Observability & Regression Guardrails
Codify verification so policy and merchandising features stay healthy.

- [ ] **REC-401 – Automated scenario suite**  
  Turn S1–S10 scripts (`analysis/scripts/run_scenarios.py`) into CI checks. Fail pipelines when any scenario regresses, storing evidence as build artifacts.

- [ ] **REC-402 – Production telemetry for rules/overrides**  
  Log applied rule IDs, boost deltas, and pin placements in decision traces. Build dashboards to monitor rule hit rates and override efficacy.

- [ ] **REC-403 – Alerting on zero-effect overrides**  
  Detect when boosts/pins are created but produce no ranking change within defined windows, notifying ops teams to investigate.

- [ ] **REC-404 – Documentation & runbooks refresh**  
  Update README/config docs once fixes land, ensuring merchants understand override precedence, policy guarantees, and troubleshooting steps.

---

## Tracking & Review
- Owners: Recommendation platform squad (lead dev prime).  
- Milestones: Policy enforcement hotfix (Week 1), merchandising controls (Week 2), cohort quality uplift (Week 4), guardrails + docs (Week 5).  
- Acceptance: All tickets closed with passing automated scenarios, updated documentation, and rerun evaluation achieving PASS criteria.
