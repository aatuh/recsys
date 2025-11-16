# Algorithm & Environment Configuration Reference

This is the **canonical** list of environment variables (and matching runtime overrides) that influence RecSys ranking. Other docs intentionally link here instead of duplicating tables—update this file first whenever a knob changes.

> ⚠️ **Advanced topic**
>
> Read this after you understand the basic configuration concepts and need to tune or debug behavior.
>
> **Where this fits:** Ranking & personalization.
>
> **Who should read this?** Integration engineers and developers tuning ranking behavior. Business stakeholders can skim the interaction notes to understand how overrides relate to guardrails.

## Contents

- [Algorithm \& Environment Configuration Reference](#algorithm--environment-configuration-reference)
  - [Contents](#contents)
  - [Core retriever weights](#core-retriever-weights)
  - [Personalization \& starter profiles](#personalization--starter-profiles)
  - [Diversity, MMR, and coverage](#diversity-mmr-and-coverage)
  - [Rules \& overrides](#rules--overrides)
  - [Bandit experimentation](#bandit-experimentation)
  - [Runtime overrides vs. env profiles](#runtime-overrides-vs-env-profiles)
    - [Interaction considerations](#interaction-considerations)
    - [Reference](#reference)
    - [Service metadata](#service-metadata)
  - [Related docs](#related-docs)

---

## Core retriever weights

**`BLEND_ALPHA`, `BLEND_BETA`, `BLEND_GAMMA`** — Blend weights for popularity, co-visitation, embeddings. Higher weight raises that retriever’s influence (keep weights summing to ~1). Related knobs: `NEW_USER_BLEND_*`, `BLEND_WEIGHTS_OVERRIDES`. Override via request: `overrides.blend_alpha/beta/gamma`.

**`BLEND_SEGMENT_OVERRIDES`** — Per-segment blend weights (`segment=alpha|beta|gamma`). This was the legacy approach; prefer env profile manager (`/v1/admin/recommendation/config`). Related knobs: starter profiles, `PROFILE_*`. No request override; edit segment profiles instead.

**`POPULARITY_FANOUT`** — Number of popularity candidates injected before ranking. Larger fanout increases coverage but costs compute. Tied to `MMR_LAMBDA` and coverage guardrails. Override via `overrides.popularity_fanout`.

**`COVIS_WINDOW_DAYS`, `POPULARITY_HALFLIFE_DAYS`** — Windows/decay for co-visitation and popularity. Shorter windows react faster but drop historical context. Override via `overrides.covis_window_days` and `overrides.popularity_halflife_days`.

**Tips:** Increase `BLEND_GAMMA` if embeddings are under-weighted (and lower `ALPHA` to compensate). For surge events, shorten `COVIS_WINDOW_DAYS`. When raising `BLEND_ALPHA`, consider increasing `POPULARITY_FANOUT` or lowering `MMR_LAMBDA` to avoid stale lists.

## Personalization & starter profiles

**`PROFILE_BOOST`** — Multiplicative boost applied to tag overlap. Raise (0.6–0.8) when you trust historical signals; lower (0.3–0.5) for sparse cohorts. Related knobs: `PROFILE_COLD_START_MULTIPLIER`, `PROFILE_MIN_EVENTS_FOR_BOOST`. Override via `overrides.profile_boost`.

**`PROFILE_MIN_EVENTS_FOR_BOOST`, `PROFILE_COLD_START_MULTIPLIER`** — Minimum event count before full boost plus the attenuation factor for sparse history. Lower the threshold (2–4) to kick in personalization sooner; keep the multiplier near 1.0 when you want starters to stay prominent. Related knobs: starter decay, `PROFILE_STARTER_BLEND_WEIGHT`. `PROFILE_MIN_EVENTS_FOR_BOOST` lives in admin configs; multiplier is static.

**`PROFILE_STARTER_BLEND_WEIGHT`, `PROFILE_STARTER_DECAY_EVENTS`, `PROFILE_STARTER_PRESETS`** — Control curated starter profiles for new users. Higher weight/decay values keep curated anchors longer; adjust per segment to guide cold-start experiences. Tie into fixtures noted in `docs/simulations_and_guardrails.md` and the tuning harness `--starter-blend-weights`. Override via `overrides.profile_starter_blend_weight`.

**`NEW_USER_BLEND_ALPHA/BETA/GAMMA`, `NEW_USER_POP_FANOUT`, `NEW_USER_MMR_LAMBDA`** — Optional overrides for `surface=new_user`. Use them when onboarding flows need more exploration or different trade-offs. Coverage and starter-profile guardrails keep them honest. All are overrideable via `overrides.*`.

## Diversity, MMR, and coverage

**`MMR_LAMBDA`** — Diversification trade-off (0 = diversity, 1 = relevance). Lower values increase novelty; monitor NDCG guardrails. Related knobs: `MMR_PRESETS`, blend weights. Override via `overrides.mmr_lambda`.

**`MMR_PRESETS`** — Surface-specific `mmr_lambda` values (e.g., `home=0.25`). Keeps surfaces aligned with business goals; edit via env files or admin API (`/v1/admin/recommendation/presets`). No per-request override.

**`BRAND_CAP`, `CATEGORY_CAP`** — Maximum items per brand/category in the final list. Prevents a single supplier from dominating. Coordinate with rules + guardrails. Override via `overrides.brand_cap` / `overrides.category_cap`.

**`RULE_EXCLUDE_EVENTS`, `PURCHASED_WINDOW_DAYS`, `EXCLUDE_EVENT_TYPES`** — Purchase suppression controls. When enabled, set `PURCHASED_WINDOW_DAYS > 0`. Audit traces show filtered items. Override via `overrides.rule_exclude_events` and `overrides.purchased_window_days`.

**`COVERAGE_CACHE_TTL`, `COVERAGE_LONG_TAIL_HINT_THRESHOLD`** — Cache behavior and hint thresholds for long-tail classification. Shorter TTL refreshes metrics faster; threshold defines what counts as “long tail.” Guardrails target ≥0.60 catalog coverage and ≥0.20 long-tail share. No request override.

## Rules & overrides

**`RULES_ENABLE`** — Global kill-switch for the rules engine. Setting it to `false` disables boost/pin/block evaluation entirely; guardrails expect it to stay on. No request override.

**`RULES_MAX_PIN_SLOTS`, `RULES_CACHE_REFRESH`, `RULES_AUDIT_SAMPLE`** — Pin capacity, cache TTL, and audit sampling knobs. Increase pin slots when hero modules need more space, shorten cache TTL for high-churn campaigns, raise sampling to gather more evidence. Not overrideable.

**`AUDIT_DECISIONS_*`** — Controls decision trace capture (queue/batch sizes, sample rates, namespace filters). Tune when deep guardrail debugging is required; `AUDIT_DECISIONS_SAMPLE_OVERRIDES` targets traffic heavy in manual overrides. No per-request override.

Manual overrides rely on these same knobs because they compile to rules. Always dry-run complex changes via `/v1/admin/rules/dry-run`.

## Bandit experimentation

**`BANDIT_ALGO`** — Global policy algorithm (`thompson`, `ucb1`, etc.). Determines exploration strategy. Overrideable via `overrides.bandit_algo`.

**`BANDIT_EXPERIMENT_ENABLED`, `BANDIT_EXPERIMENT_HOLDOUT_PERCENT`, `BANDIT_EXPERIMENT_SURFACES`, `BANDIT_EXPERIMENT_LABEL`** — Enable and shape the default holdout experiment. Holdout traffic receives control content; metrics flow into Prometheus dashboards. Adjust via env profiles/admin configs (no per-request override). Update `/v1/bandit/policies` when changing defaults so tooling stays in sync.

Remember to update `/v1/bandit/policies` when changing env defaults so experimentation tooling stays in sync.

## Runtime overrides vs. env profiles

- **Per-request overrides (rapid experimentation):** Use the `overrides` object in `/v1/recommendations` for short-lived experiments or customer-specific tuning. Every override is captured in audit traces and scenario evidence.
- **Env profiles (namespace defaults):** Use env files (`api/env/*.env`) plus `config/profiles.yml` when setting the baseline for a namespace/environment. Apply via `analysis/scripts/configure_env.py --namespace <ns>`; the script records the before/after diff in `analysis/env_history/`. When you need hot-swappable profiles without restarting containers, capture/apply them via `analysis/scripts/env_profile_manager.py fetch/apply` so `/v1/admin/recommendation/config` is updated directly.
- **Admin APIs:** For structural changes (segments, event types, starter profiles), prefer the dedicated endpoints so changes are versioned in the DB and surfaced via `/v1/segments`, `/v1/event-types`, etc.
- **Guardrails:** Regardless of approach, run `analysis/scripts/run_simulation.py` with guardrails enabled before rollout. This ensures coverage and segment lifts stay above thresholds.

### Interaction considerations

- Raising `PROFILE_BOOST` without adjusting `PROFILE_COLD_START_MULTIPLIER` can cause cold-start users to see overly personalized lists; adjust both together.
- Increasing `BLEND_ALPHA` (popularity weight) may require a larger `POPULARITY_FANOUT` or lower `MMR_LAMBDA` to keep diversity intact.
- Turning on `RULE_EXCLUDE_EVENTS` without setting `PURCHASED_WINDOW_DAYS` results in no suppression—always set both.
- Changing starter profiles (env or admin API) should be followed by a simulation run; starter-profile and boost-exposure guardrails catch regressions in cold-start behavior.

### Reference

- Overrides supported in `/v1/recommendations`: see “Per-request algorithm overrides” in README.
- Env profile parity: run `python analysis/scripts/check_env_profiles.py --strict` after editing any `api/env/*.env`.
- Profiles ↔ namespaces: maintain `config/profiles.yml`; `run_simulation.py` uses it to pick the correct profile per namespace.

### Service metadata

**`RECSYS_GIT_COMMIT`** — Populates `/version.git_commit`. Set during deploys (`git rev-parse HEAD`); falls back to the `git` CLI or `"unknown"` if unset.

**`RECSYS_BUILD_TIME`** — Populates `/version.build_time` (UTC ISO-8601). Optional; defaults to process start time if not provided.

Supplying both makes `/version` output—and therefore determinism/quality evidence—traceable to a single build without digging through CI logs.

---

## Related docs

- `docs/configuration.md` – conceptual explanation of how these knobs fit into the ranking pipeline.
- `docs/api_reference.md` – request/response schemas and where overrides live in API payloads.
- `docs/rules_runbook.md` – how overrides interact with rules and incident workflows.
- `docs/simulations_and_guardrails.md` – how to run simulations and guardrail checks for profile changes.
- `docs/analysis_scripts_reference.md` – which script manages each profile or env change.
