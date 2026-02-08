---
diataxis: reference
tags:
  - recsys-algo
  - scoring
  - reference
  - recsys-engineering
---
# Scoring model specification (recsys-algo)
This page is the canonical reference for Scoring model specification (recsys-algo).


## Who this is for

- Recommendation engineers reviewing scoring behavior and determinism guarantees
- Developers/operators interpreting `explain` output and evaluating knob changes

## What you will get

- The implemented normalization and blending formulas (no hand-wavy “ML model” language)
- Missing-signal behavior (what happens when stores/artifacts are absent)
- Tie-breaking rules (what happens when scores match)

## Reference

### Notation

Per candidate item `i`:

- `pop_raw(i)` is the baseline score from the primary candidate source (typically popularity; in `cooc`/`implicit`
  modes it is the score returned by those sources; if a candidate has no baseline score, this is `0`).
- `cooc_raw(i)` is the co-visitation score for `i` (may be missing/`0`).
- Similarity sub-signal raw scores (may be missing/`0`):
  - `emb_raw(i)` (embedding similarity, expected in `[0, 1]`)
  - `collab_raw(i)` (collaborative similarity, positive)
  - `content_raw(i)` (content/tag similarity, positive)
  - `session_raw(i)` (session sequence score, positive)
- Blend weights (non-negative):
  - `alpha` (popularity/baseline)
  - `beta` (co-visitation)
  - `gamma` (similarity bucket)

### Weight resolution

Weights are resolved in this order:

1. Start from configured values (`alpha`, `beta`, `gamma`) and clamp each to `>= 0`.
2. If the request does not specify weights and the selected algorithm mode is `popularity`, `cooc`, or `implicit`,
   force weights to popularity-only: `alpha = 1`, `beta = 0`, `gamma = 0`.
3. If the request specifies weights, clamp the provided values to `>= 0` and use them.
4. Safety fallback: if `alpha = beta = gamma = 0`, force `alpha = 1`.

Note: weights are not auto-normalized. If your weights sum to more than `1`, blended scores can exceed `1` (before
personalization).

### Normalization functions

`recsys-algo` normalizes scores into `[0, 1]`-ish ranges to make blending predictable.

For positive raw scores (popularity, co-vis, collaborative/content/session):

```text
norm_pos(s) = 0                  if s <= 0
norm_pos(s) = s / (s + 1)        if s > 0
```

For embedding similarity scores:

```text
norm_emb(s) = 0                  if s <= 0
norm_emb(s) = 1                  if s >= 1
norm_emb(s) = s                  otherwise
```

### Similarity bucket

Compute per-signal normalized values:

- `emb_norm(i) = norm_emb(emb_raw(i))`
- `collab_norm(i) = norm_pos(collab_raw(i))`
- `content_norm(i) = norm_pos(content_raw(i))`
- `session_norm(i) = norm_pos(session_raw(i))`

Then:

```text
sim_norm(i) = max(emb_norm(i), collab_norm(i), content_norm(i), session_norm(i))
```

If `sim_norm(i) == 0`, similarity contributes nothing to the blended score.

If multiple sub-signals share the same maximum normalized value (ties within epsilon `1e-9`), the item’s explanation
tracks all tied sources.

### Blended score (pre-personalization)

```text
pop_norm(i)  = norm_pos(pop_raw(i))
cooc_norm(i) = norm_pos(cooc_raw(i))

score_blend(i) = alpha*pop_norm(i) + beta*cooc_norm(i) + gamma*sim_norm(i)
```

Missing/disabled signals behave like `raw = 0` which yields `norm = 0` and therefore `0` contribution.

### Personalization boost (optional)

If `ProfileBoost > 0` and a user tag-profile exists, `recsys-algo` applies a multiplicative boost per candidate:

1. Build a user profile `profile[tag]` where weights sum to `1`.
1. For each candidate, compute tag overlap:

```text
overlap(i) = sum(profile[tag] for tag in tags(i) if tag is present in profile)
```

1. If `overlap(i) > 0`, compute the boost multiplier:

```text
multiplier_raw(i) = 1 + ProfileBoost*overlap(i)
```

1. If cold-start attenuation is enabled (`ProfileMinEventsForBoost > 0`) and the recent event count is lower than the
minimum, scale the boost down:

```text
multiplier(i) = 1 + (multiplier_raw(i) - 1)*ProfileColdStartMultiplier
```

1. Apply:

```text
score_final(i) = score_blend(i) * multiplier(i)
```

If `overlap(i) == 0` (or profiles are unavailable), no multiplier is applied.

### Output ordering (tie-breaking)

For non-pinned items, the response is ordered deterministically by:

1. `score_final` descending
2. `item_id` ascending (lexicographic) when scores are equal

Pinned items (from rules) are placed before scored items.

## Examples

Given:

- `alpha=1.0`, `beta=0.5`, `gamma=0.2`
- `pop_raw=3.0`, `cooc_raw=1.0`
- `emb_raw=0.8`, `collab_raw=2.0`, `content_raw=0`, `session_raw=0`

Compute:

```text
pop_norm  = 3/(3+1) = 0.75
cooc_norm = 1/(1+1) = 0.50
emb_norm  = 0.80
collab_norm = 2/(2+1) = 0.666...
sim_norm = max(0.80, 0.666..., 0, 0) = 0.80

score_blend = 1.0*0.75 + 0.5*0.50 + 0.2*0.80 = 1.16
```

If personalization is enabled and `overlap=0.2` with `ProfileBoost=0.5`:

```text
multiplier_raw = 1 + 0.5*0.2 = 1.1
score_final = 1.16 * 1.1 = 1.276
```

## Read next

- Ranking & constraints reference: [Ranking & constraints reference](ranking-reference.md)
- recsys-service knobs that affect scoring: [recsys-service configuration](../reference/config/recsys-service.md)
- Serving API fields (algorithm/weights/explain): [API Reference](../reference/api/api-reference.md)
