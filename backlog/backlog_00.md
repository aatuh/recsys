# TODO_backlog_00.md

Short answer: build “Why this rec?” deterministically from your scoring
pipeline and ship it in structured form. Use an LLM only as an optional
front-end paraphraser (on-prem or disabled by default). You already have all
the signals to do this cleanly and fast.

Here’s a concrete plan that fits our codebase and buyer needs.

What to explain (from our pipeline)

We already compute exactly the things business users care about:

- Blended scoring components: normalized popularity, co-visitation, and
embedding similarity with per-request weights α/β/γ. We compute and normalize
them before blending, so we can expose each component’s share.
- Light personalization: tag-overlap boost and whether an item was boosted
for this user. Expose the overlap and multiplier.
- Diversity & caps: MMR selection and brand/category caps are applied during
re-rank; we can show if diversity influenced the pick and whether caps
constrained it.
- Anchors & similar: for similar-items and blended flows, we already track
which signals fired (embedding/co-vis) and can include the relevant anchor
item(s).

We already return reasons []string and gate it with
include_reasons, so the API shape is in place; we just need to enrich
it.

API design (backward compatible)

Add a structured, machine-readable block while keeping the current reasons
array:

{
  "item_id": "slot_42",
  "score": 0.813,
  "reasons": ["recent_popularity", "co_visitation", "personalization", "diversity"],
  "explain": {
    "blend": {
      "alpha": 0.7, "beta": 0.2, "gamma": 0.1,
      "pop_norm": 0.62, "cooc_norm": 0.28, "emb_norm": 0.10,
      "contrib": {"pop": 0.434, "cooc": 0.056, "emb": 0.010}
    },
    "personalization": { "overlap": 0.44, "boost_multiplier": 1.11 },
    "mmr": { "lambda": 0.75, "max_sim": 0.33, "penalty": 0.083 },
    "caps": { "brand": {"applied": true}, "category": {"applied": false} },
    "anchors": ["item:slot_17", "item:slot_08"]
  }
}

How to request it:

- Keep include_reasons (today’s behavior).
- Add explain_level with values: "tags" (default), "numeric", "full".
  - tags: current string array only.
  - numeric: add explain.blend/personalization/mmr/caps as above.
  - full: add raw values useful for admins (e.g., raw popularity sums,
  window sizes, event-type weights, and bandit metadata if you called the
  bandit variant). Your swagger already distinguishes the bandit response and
  has a bandit_explain object; mirror that spirit.

Backend implementation steps (Go)

    Capture normalized components during blending

        In applyBlendedScoringWithWeights, you already compute popNorm,
        coocNorm, embNorm. Persist these into maps keyed by item ID on
        CandidateData (e.g., PopNorm[item], CoocNorm[item],
        EmbNorm[item]). Then compute per-component contribution as
        alpha*popNorm, etc., for the explain payload. 

Record personalization overlap

    In applyPersonalizationBoost, compute and store the per-item
    overlap and boost_multiplier when you set Boosted[item]=true.
    You already mark boosted items; add, e.g., ProfileOverlap[item]=overlap. 

Return MMR & caps details

    Extend the MMR pass to capture for each selected item the maxSim
    observed at its selection step and expose the current lambda. You can
    return a small MMRInfo map {itemID: {maxSim, lambda}}. Caps are already
    checked in canSelectWithCaps; mark which cap blocked candidates or was
    active at pick time.

Anchors

    You pass anchors []string into embedding/co-vis gathering. Keep a
    deduped slice and ship it in explain.anchors. 

Build the structured explain block

    Extend buildResponse to assemble explain when
    explain_level != "tags". Keep reasons []string for compatibility (you
    already construct reason tags there). 

    Swagger & tests

        Add the new fields to swagger.yaml and add tests similar to your existing
        include_reasons tests to assert that numerical pieces appear when
        requested. You already test reasons and exclude-logic; mirror that style.

UI plan (Demo & admin)

    Your ExplainModal.tsx already parses reason strings and shows a glossary.
    Switch it to prefer item.explain.* when present, and fall back to parsing
    strings if not. Render small horizontal bars for the 3 blend contributions,
    show overlap/boost, and list anchors as badges.

    For non-technical viewers, keep concise labels:

        “Trending now” for popularity, “Often played with X” for co-vis, “Similar
        to your taste” for embeddings/personalization. Your glossary copy can live
        right in the modal as it does today. 

Should we use an external LLM?

Default: No. Reasons:

    You already have deterministic attributions that map 1:1 to your scoring
    math (alpha/beta/gamma, overlap, MMR). That’s faster, cheaper, auditable, and
    stable for regulated buyers. 

    You already return “reason tags” and have tests and swagger types for them,
    so enriching deterministically is low-risk and consistent.

When an LLM makes sense (optional add-on):

    If a client wants natural-language narratives (“We suggested MegaJackpot
    because it’s trending and similar to your recent picks”), consider an on-prem
    or self-hosted model (for data control) that takes the structured explain as
    input and emits a short sentence. Keep it stateless and PII-free by
    passing only tags, reason names, and item display names.

    Do not use the LLM for the actual attribution. Use it only to paraphrase
    the structured facts you compute deterministically.

Guardrails (important for gambling)

    Latency: Gate full explains behind a flag and only compute for the top-K
    items shown. The incremental cost is tiny since you already have the numbers.

    Privacy: Never include raw user events; only derived values (overlap,
    anchors by ID). You already operate on opaque IDs and decayed aggregates. 

    Security: Keep “full” details for admin/staff surfaces only.

    Determinism: All numbers should recompute to the same values using
    request inputs; tests should assert this on a seed dataset (you already have
    examples and golden-style tests).

Minimal implementation checklist

    Add fields on CandidateData: PopNorm, CoocNorm, EmbNorm,
    ProfileOverlap, MMRInfo, CapsApplied. Populate in existing functions.

    Extend ScoredItem (or add ScoredItemExplain) to include explain.
    Keep reasons []string as is. 

Wire explain_level request param and serialize the new block in
buildResponse.

Update Swagger and web UI ExplainModal to render bars, badges, glossary.

    Tests:

        Assert reasons for “tags” (existing).

        Assert numeric parts present and consistent for “numeric”/“full”
        (e.g., contrib.pop + contrib.cooc + contrib.emb ≈ score before MMR).
        You already have end-to-end tests checking reasons/caps/excludes; extend
        those.

This route gives you explainability that’s instant, honest, and audit-ready
— perfect for non-technical buyers and regulated iGaming use-cases. If a buyer
wants friendlier prose, layer an optional, local LLM paraphrase over your
own structured truth.
