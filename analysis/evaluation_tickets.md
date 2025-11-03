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

- [x] **REC-204 – New-item exposure controls**  
  Boost rules now inject off-candidate items so manual overrides surface fresh inventory even without baseline popularity. Added per-surface Prometheus counters and structured logs tracking boost/pin exposure rates to monitor exploration impact.

- [x] **REC-205 – Multi-objective trade-off curve**  
  Boost actions are now proportional to existing scores (with additive fallback) so adjusting `boost_value` yields predictable margin lifts and logged exposure counters. Scenario S9 captures the trade-off curve and policy telemetry in evidence.

---

## Epic EPC-003 – Quality & Cold-Start Experience
Improve relevance for underserved cohorts while maintaining diversity.

- [x] **REC-301 – New-user onboarding profiles**  
  Added trait-driven starter profiles for `new_users`, injecting curated tag weights when no history exists. Fallback personalization now logs in trace extras and powers scenario S7 to capture the applied profile.

- [x] **REC-302 – Personalization decay tuning**  
  Introduced configurable `PROFILE_MIN_EVENTS_FOR_BOOST` and `PROFILE_COLD_START_MULTIPLIER`, attenuating personalization when history is sparse. Added engine unit tests to guarantee the new scaling logic.

- [x] **REC-303 – Diversity knob documentation & presets**  
  Added `MMR_PRESETS` env parsing, exposed presets via `/v1/admin/recommendation/presets`, and documented the workflow so admin tooling can surface validated `mmr_lambda` options per surface.

---

## Epic EPC-004 – Observability & Regression Guardrails
Codify verification so policy and merchandising features stay healthy.

- [x] **REC-401 – Automated scenario suite**  
  Added a dedicated GitHub Actions workflow (`.github/workflows/scenario-suite.yml`) that spins up the stack, seeds baseline data, runs the S1–S10 harness via `make scenario-suite`, fails on regressions, and uploads the updated evidence.

- [x] **REC-402 – Production telemetry for rules/overrides**  
  Added structured `policy_rule_actions` / `policy_rule_exposure` logs, Prometheus counters (observed via `policyMetrics.Observe`), and documented the metrics so ops can monitor rule hit rates and surface override issues from production traffic.

- [x] **REC-403 – Alerting on zero-effect overrides**  
  Added `policy_rule_zero_effect` warnings plus the `policy_rule_zero_effect_total` Prometheus counter so ops can alert on boosts/pins that never impact the ranking.

- [x] **REC-404 – Documentation & runbooks refresh**  
  Update README/config docs once fixes land, ensuring merchants understand override precedence, policy guarantees, and troubleshooting steps.

---

## Tracking & Review
- Owners: Recommendation platform squad (lead dev prime).  
- Milestones: Policy enforcement hotfix (Week 1), merchandising controls (Week 2), cohort quality uplift (Week 4), guardrails + docs (Week 5).  
- Acceptance: All tickets closed with passing automated scenarios, updated documentation, and rerun evaluation achieving PASS criteria.
