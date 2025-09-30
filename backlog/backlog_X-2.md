# X-2 — ExplainLLM (RCA Assist) v1

> Status: **Proposed — small, shippable v1**
> Goal: Give ops/product a one-click **root-cause analysis (RCA) summary** for “why X isn’t working,” using an LLM constrained by our own facts (audit, rules, metrics, dry‑run). Output is human‑readable but consistently structured.

---

## TL;DR

* Provide an **Explain** panel backed by an LLM that reads a compact **Facts Pack** about a target (banner/item/surface/segment) and time window.
* Response is **free‑form** for humans, but must include sections: **Summary, Status, Key findings, Likely causes**, and **Suggested fix** (optional, only if warranted).
* Tight **fact bounds**: we *only* send minimized, relevant facts; no PII. Clear failure modes and fallbacks to a facts‑only panel.

---

## Why now?

* Teams currently hop across rules, audit decisions, caps, segments, and dashboards to answer simple “why not?” questions.
* We already capture rich telemetry (DecisionTrace, rules, metrics). This feature **interprets** it for humans, with zero config per investigation.

---

## Model recommendation (API)

* **Primary (default):** **o4‑mini** — fast, cost‑efficient, strong at synthesis and following instructions.
* **Escalation (big/ambiguous cases):** **o3** — higher‑end reasoning for long/complex fact sets.
* Selection heuristic (automatic): if Facts Pack > threshold (e.g., >2k tokens) or many conflicting signals → use **o3**, else **o4‑mini**.

> Note: keep provider keys in our existing secrets store; add `LLM_PROVIDER=openai`, `LLM_MODEL_PRIMARY=o4-mini`, `LLM_MODEL_ESCALATE=o3`.

---

## Response format (human‑readable with sections)

The LLM must produce **markdown** with the following sections in order. If a section is empty, include it with a one‑line explanation (e.g., “None found”).

1. **Summary** — 2–4 sentences, plain language.
2. **Status** — one of: `working_as_configured | degraded | not_working | unknown`.
3. **Key findings** — 3–6 bullets, each must cite `evidence_id`(s) from facts.
4. **Likely causes** — 2–5 bullets: `cause · confidence=[0..1] · evidence_id=[...]`.
5. **Suggested fix (optional)** — 1–3 bullets with concrete next steps + deep links.

---

## Prompt design (v1)

**System prompt (fixed):**

> You are a reliability analyst for our Recommendation System. Use **only** the provided FACTS. Do **not** invent data. When a plausible cause lacks evidence, mark it **uncertain**. Be concise, action‑oriented, and cite **evidence\_id**s for claims. Output **markdown** with sections: Summary, Status, Key findings, Likely causes, Suggested fix (only if warranted). Never include private reasoning.

**User prompt (templated):**

```
Investigate why {target_type} "{target_id}" appears to be “not working” on surface "{surface}" in namespace "{namespace}" during {from}..{to} for segment "{segment_id}".

FACTS (compact JSON):
{facts_json}

If facts are insufficient, say so explicitly and suggest the top 3 diagnostics.
```

**Guardrails in prompt:**

* Hard cap: avoid speculation; prefer “unknown” with diagnostics.
* No PII; never output raw IDs if marked `private_id`. Use our `public_ref` when present.
* Prefer actionable checks over generic advice; include direct links if given.

---

## Facts Pack (what we send)

**Objective:** minimal, relevant, numeric.

**Sources**

* **Audit/DecisionTrace:** generators used, caps/MMR hits, personalization, anchors, matched **rules** (block/pin/boost), rule effects per item, evidence IDs.
* **Rules snapshot (active now & during window):** affecting target via `item_id`, `tag`, `brand`, `category`; include TTL & priority.
* **Serving metrics:** impressions, clicks, CTR, err rates, eligibility drop‑offs per stage, sample sizes (with window).
* **Eligibility & context:** segment size, A/B group, frequency caps, purchased‑exclude, profile params deltas.
* **Dry‑run simulation (optional):** would the target show *now*? (yes/no + reason).

**Compaction**

* **Top‑K** per stage (e.g., K=10) with reason codes & counts.
* **Roll‑ups**: totals, rates, min/max, last change timestamps.
* **Redaction**: drop user IDs, IPs, device IDs.
* **Token budget**: aim ≤ 2k tokens; chunk and summarize if over (bucket small events; collapse long lists into `+N more`).

Schema sketch:

```json
{
  "target": {"type":"banner|item|surface|segment","id":"…"},
  "window": {"from":"…","to":"…"},
  "context": {"namespace":"…","surface":"…","segment_id":"…"},
  "metrics": {"impressions":..., "clicks":..., "ctr":..., "errors":...},
  "rules_active": [{"rule_id":"…","action":"BLOCK|PIN|BOOST","target":"ITEM|TAG|BRAND|CATEGORY","priority":...,"ttl":"..."}],
  "audit": [{"evidence_id":"…","stage":"mmr|caps|rules|profile|generator","message":"…","count":...}],
  "sim": {"eligible_now":true, "why":"…"},
  "links": {"rules":"/admin/rules?...","metrics":"/dash?..."}
}
```

---

## API & UI (v1)

**Endpoint**

```
POST /explain/llm
{
  "target_type": "banner|item|surface|segment",
  "target_id": "…",
  "namespace": "…",
  "surface": "…",
  "segment_id": "…",
  "from": "RFC3339",
  "to":   "RFC3339",
  "question": "optional free text"
}
```

**Response (markdown)**: the LLM’s formatted explanation.

**UI**

* A right‑side drawer/panel: selector (target + window) → **Explain** button → shows markdown output with copy/export.
* Toggle: **Facts shown to LLM** (collapsible) for transparency.

---

## Implementation plan

1. **FactsCollector**: compose & compact facts from audit, rules, metrics, dry‑run; redact PII.
2. **ModelSelector**: choose **o4‑mini** or **o3** by budget/complexity.
3. **PromptBuilder**: render system+user prompts and inject compact facts JSON.
4. **LLMClient**: call OpenAI Responses API; timeouts, retries, token caps; log request/response with redaction.
5. **Presenter**: render markdown in UI; provide “show facts” toggle and deep links.
6. **Caching**: cache `(target, window, hash(facts))` for 15–60 min; include `X-Explain-Cache: hit/miss`.

---

## Config (env)

* `LLM_EXPLAIN_ENABLED=false`
* `LLM_PROVIDER=openai`
* `LLM_MODEL_PRIMARY=o4-mini`
* `LLM_MODEL_ESCALATE=o3`
* `LLM_TIMEOUT=6s`
* `LLM_MAX_TOKENS=1200`

---

## Acceptance criteria

* Returns markdown with all required sections; **Status** ∈ allowed values.
* Each **Key finding** and **Likely cause** cites at least one `evidence_id`.
* If facts are insufficient, output explicitly states this and lists top 3 diagnostics.
* No PII leaves the service; facts are redacted and size‑bounded.
* Average latency within budget (p95 < 2.0s on o4‑mini); graceful degradation to **facts‑only** when LLM fails.
* Cached repeats for same (target, window) serve instantly.

---

## Testing

* **Unit**: FactsCollector compaction (Top‑K, roll‑ups, redaction).
* **Contract**: PromptBuilder renders required sections.
* **Golden**: Fixed Facts Pack → snapshot of LLM output via stub (no live API).
* **Load**: worst‑case Facts Pack within token budget; ensure model selector escalates.
* **Security**: verify PII redaction and secrets not logged.

---

## Risks & mitigations

* **Hallucination** → strict instructions, evidence‑id citations, facts toggle for humans, easy fallback panel.
* **Token bloat** → hard size caps, compaction, escalation only when necessary.
* **Cost drift** → cache results; expose per‑call token usage in logs/metrics.

---

## Future (v1.1+)

* **Tool calling** for controlled follow‑ups (e.g., `get_latest_metrics`, `run_simulation`).
* **Structured outputs** option for programmatic consumption beside markdown.
* **RBAC** + **audit** trail of explanations shown.
* Multi‑entity diffs: “Why did CTR drop vs last week?”
