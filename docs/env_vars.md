# Algorithm & Environment Configuration Reference

This guide explains every environment variable (and matching runtime override)
that influences Recsys ranking. Use it alongside:
- `docs/api_endpoints.md` (where overrides live in API requests)
- `docs/rules-runbook.md` (how overrides affect guardrails)
- `docs/bespoke_simulations.md` (how to run simulations with different profiles)

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

---

## Core retriever weights

| Env var                                         | Description                                              | Impact                                                                                          | Related knobs                                                                                   | Override?                                                                  |
|-------------------------------------------------|----------------------------------------------------------|-------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------|
| `BLEND_ALPHA`, `BLEND_BETA`, `BLEND_GAMMA`      | Blend weights for popularity, co-visitation, embeddings. | Higher weight raises that retriever’s influence. Ensure weights sum to ~1 for interpretability. | `NEW_USER_BLEND_*` for cold-start surfaces, `BLEND_WEIGHTS_OVERRIDES` for per-namespace tweaks. | Yes (`overrides.blend_alpha/beta/gamma`).                                  |
| `BLEND_SEGMENT_OVERRIDES`                      | Per-segment blend weights (`segment=alpha\|beta\|gamma`). | Lets you tailor mix for named cohorts (e.g., `trend_seekers`, `weekend_adventurers`).           | Starter profiles, `PROFILE_*` knobs, namespace overrides.                                        | Not via request; configure via env/segment profiles.            |
| `POPULARITY_FANOUT`                             | Number of popularity candidates inserted before ranking. | Larger fan-out increases catalog coverage but adds compute.                                     | `MMR_LAMBDA`, coverage guardrails.                                                              | Yes (`overrides.popularity_fanout`).                                       |
| `COVIS_WINDOW_DAYS`, `POPULARITY_HALFLIFE_DAYS` | Time windows for co-visitation and popularity decay.     | Short windows react faster but reduce historical signal.                                        | Event ingestion cadence.                                                                        | Yes (`overrides.covis_window_days`, `overrides.popularity_halflife_days`). |

**Tips**

- Increase `BLEND_GAMMA` if embeddings retriever is weak; balance by lowering `ALPHA` to avoid popularity dominance.
- For surge events (holiday promotions), shorten `COVIS_WINDOW_DAYS` to focus on fresh trends.
- When raising `BLEND_ALPHA`, also consider bumping `POPULARITY_FANOUT` or lowering `MMR_LAMBDA` to avoid stale lists.

## Personalization & starter profiles

| Env var                                                                                   | Description                                                               | Impact                                                                      | Related knobs                                                    | Override?                                       |
|-------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|-----------------------------------------------------------------------------|------------------------------------------------------------------|-------------------------------------------------|
| `PROFILE_BOOST`                                                                           | Multiplicative boost applied to tag overlap.                              | Raising it can overfit to existing preferences; keep between 0.3–0.8.       | `PROFILE_COLD_START_MULTIPLIER`, `PROFILE_MIN_EVENTS_FOR_BOOST`. | Yes (`overrides.profile_boost`).                |
| `PROFILE_MIN_EVENTS_FOR_BOOST`, `PROFILE_COLD_START_MULTIPLIER`                           | Minimum events required before full boost; attenuation factor beforehand. | Prevents over-personalized lists when data is sparse.                       | Starter decay, `PROFILE_STARTER_BLEND_WEIGHT`.                   | Yes.                                            |
| `PROFILE_STARTER_BLEND_WEIGHT`, `PROFILE_STARTER_DECAY_EVENTS`, `PROFILE_STARTER_PRESETS` | Control curated starter profiles for new users.                           | Weight determines how much curated tags influence ranking before decay.     | `docs/bespoke_simulations.md` fixtures.                          | Yes (`overrides.profile_starter_blend_weight`). |
| `NEW_USER_BLEND_ALPHA/BETA/GAMMA`, `NEW_USER_POP_FANOUT`, `NEW_USER_MMR_LAMBDA`           | Optional overrides for cold-start surfaces.                               | Use when `surface=new_user` needs more exploration or different trade-offs. | Guardrails (coverage, S7).                                       | Yes (`overrides.*`).                            |

## Diversity, MMR, and coverage

| Env var                                                               | Description                                                     | Impact                                                                              | Related knobs                                                          | Override?                                                                 |
|-----------------------------------------------------------------------|-----------------------------------------------------------------|-------------------------------------------------------------------------------------|------------------------------------------------------------------------|---------------------------------------------------------------------------|
| `MMR_LAMBDA`                                                          | Diversification trade-off (0=diversity, 1=relevance).           | Lower values increase novelty; watch guardrails (NDCG).                             | `MMR_PRESETS`, `blend_*`.                                              | Yes (`overrides.mmr_lambda`).                                             |
| `MMR_PRESETS`                                                         | Surface-specific `mmr_lambda` values (e.g., `home=0.25`).       | Keeps surfaces aligned with business goals.                                         | `/v1/admin/recommendation/presets`.                                    | Not via overrides; use env or admin API.                                  |
| `BRAND_CAP`, `CATEGORY_CAP`                                           | Max items per brand/category in final list.                     | Prevents domination by a single brand or category.                                  | Rules, guardrails.                                                     | Yes (`overrides.brand_cap`, `overrides.category_cap`).                    |
| `RULE_EXCLUDE_EVENTS`, `PURCHASED_WINDOW_DAYS`, `EXCLUDE_EVENT_TYPES` | Controls purchase suppression.                                  | When enabled, `PURCHASED_WINDOW_DAYS` must be >0.                                   | Audit traces show filtered items.                                      | Yes (`overrides.rule_exclude_events`, `overrides.purchased_window_days`). |
| `COVERAGE_CACHE_TTL`, `COVERAGE_LONG_TAIL_HINT_THRESHOLD`             | Cache behavior and hint threshold for long-tail classification. | Shorter TTL refreshes metrics faster; threshold defines what counts as “long tail”. | Coverage guardrails (>=0.60 catalog coverage, >=0.20 long-tail share). | No.                                                                       |

## Rules & overrides

| Env var                                                            | Description                                                             | Impact                                         | Override?                                                          |
|--------------------------------------------------------------------|-------------------------------------------------------------------------|------------------------------------------------|--------------------------------------------------------------------|
| `RULES_ENABLE`                                                     | Global kill-switch for rules engine.                                    | `false` disables boost/pin/block evaluation.   | No.                                                                |
| `RULES_MAX_PIN_SLOTS`, `RULES_CACHE_REFRESH`, `RULES_AUDIT_SAMPLE` | Pin capacity, cache TTL, audit sampling.                                | Higher sample rate increases audit storage.    | No.                                                                |
| `AUDIT_DECISIONS_*`                                                | Controls decision trace capture (queue size, batch size, sample rates). | Tune when heavy guardrail debugging is needed. | No (but `AUDIT_DECISIONS_SAMPLE_OVERRIDES` can target namespaces). |

Manual overrides share the same knobs because they compile to rules. Always dry-run complex changes via `/v1/admin/rules/dry-run`.

## Bandit experimentation

| Env var                                                                                                                   | Description                                   | Impact                                                                       | Override?                                  |
|---------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------|------------------------------------------------------------------------------|--------------------------------------------|
| `BANDIT_ALGO`                                                                                                             | Global policy algorithm (`thompson`, `ucb1`). | Determines exploration strategy.                                             | Yes (`overrides.bandit_algo` in requests). |
| `BANDIT_EXPERIMENT_ENABLED`, `BANDIT_EXPERIMENT_HOLDOUT_PERCENT`, `BANDIT_EXPERIMENT_SURFACES`, `BANDIT_EXPERIMENT_LABEL` | Enable/shape the default holdout experiment.  | Holdout traffic receives “control” content; metrics exported via Prometheus. | No (set via env profiles).                 |

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
- Changing starter profiles (env or admin API) should be followed by a simulation run; guardrails S7/S8 catch regressions in cold-start behavior.

### Reference

- Overrides supported in `/v1/recommendations`: see “Per-request algorithm overrides” in README.
- Env profile parity: run `python analysis/scripts/check_env_profiles.py --strict` after editing any `api/env/*.env`.
- Profiles ↔ namespaces: maintain `config/profiles.yml`; `run_simulation.py` uses it to pick the correct profile per namespace.

### Service metadata

| Env var             | Description                                               | Notes                                                                                           |
|---------------------|-----------------------------------------------------------|-------------------------------------------------------------------------------------------------|
| `RECSYS_GIT_COMMIT` | Populates `/version.git_commit`.                           | Set during deploys (`git rev-parse HEAD`); falls back to the `git` CLI or `"unknown"` if unset. |
| `RECSYS_BUILD_TIME` | Populates `/version.build_time` (UTC ISO-8601).            | Optional; defaults to process start time if not supplied.                                       |

Supplying both makes `/version` output—and therefore determinism/quality evidence—traceable to a single build without digging through CI logs.
