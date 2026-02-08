---
diataxis: reference
tags:
  - docs
  - personas
  - scoring
  - quality
---
# Persona journey tests and scoring rubric

This page defines **stable 10-minute journey tests** and a **scoring rubric** for RecSys docs.

Use it to:

- keep documentation improvements objective (scores should not fluctuate based on reviewer mood)
- validate that each persona can reach first value quickly
- keep quality gates (Diátaxis, linking, scannability) aligned with outcomes

## How to use this rubric

1. Pick a persona below.
2. Run the 10-minute task script **without jumping outside the docs site**.
3. Score each dimension using the thresholds in this page.
4. File issues only when you can point to a concrete blocker (page/section + why it blocks the task).

!!! tip "Stability rule"
    If you change the rubric, update this page and note the change in the repo's "What's new".

## Dimensions and thresholds

Each dimension is scored 0–10. Use the thresholds below.

### Time-to-first-success (TTFS)

Score based on whether the persona completes their **primary task** within 10 minutes.

- **9–10:** primary task completed; 0–1 backtracks; key steps are in one tutorial/how-to; no dead ends.
- **7–8:** primary task completed but required 1–2 backtracks or 3+ page hops.
- **5–6:** partial success; missing detail forces guesswork or code-reading.
- **0–4:** cannot complete primary task from docs.

### Cohesiveness / findability

Score based on whether the next needed page is discoverable from a persona hub in ≤2 clicks.

- **9–10:** persona hub exists; 0 dead ends; the next needed page is always ≤2 clicks.
- **7–8:** 1–2 moments of searching; occasional 3-click chains.
- **5–6:** frequent searching; unclear nav placement; hub missing or incomplete.
- **0–4:** repeated backtracking; pages feel scattered or hidden.

### Clarity

Score based on whether the persona can restate "what this is" and "what to do next" using only the current page.

- **9–10:** page starts with a useful 1–3 sentence intro; headings match intent; next step is obvious.
- **7–8:** mostly clear but key assumptions are implied, not stated.
- **5–6:** dense paragraphs; mixed intents; unclear next step.
- **0–4:** ambiguous, or requires external context.

### Trust / credibility

Score based on whether evidence, limitations, and security/licensing posture are discoverable from **one canonical entry point**.

- **9–10:** has one "trust center" path; evidence + limitations + security + licensing are linked and concrete.
- **7–8:** most trust items exist but are scattered or not clearly canonical.
- **5–6:** trust items exist but are hard to find; no obvious starting point.
- **0–4:** missing critical trust artifacts.

### Conversion readiness

Score based on whether a buyer can reach "pricing → licensing choice → evaluation/order next step" in ≤3 clicks.

- **9–10:** path is explicit and short; order form and contact path are obvious.
- **7–8:** path exists but requires searching or interpretation.
- **5–6:** pricing/licensing/eval pages exist but are not connected as a flow.
- **0–4:** no clear buying/evaluation flow.


## 10-minute task scripts

Each persona script includes 3–5 tasks. The **first task** is the primary task for TTFS.

### Developer

1. **Run a minimal local quickstart** and see a non-empty recommendation response.
2. Find the minimum instrumentation requirements (request IDs, exposures, outcomes).
3. Find the integration checklist and the most common integration troubleshooting page.
4. Find the operational rollback story (what is safe to change and how to revert).

Start points:

- Persona hub: [Developer](../personas/lead-developer.md)
- Primary tutorial: [Quickstart (minimal)](../tutorials/quickstart-minimal.md)

### Business representative

1. **Understand what RecSys is and whether a pilot is credible** (outcomes, time box, evidence).
2. Find “what outputs look like” (evidence examples).
3. Reach “pricing → licensing choice → how to procure” in ≤3 clicks.
4. Find the security posture and known limitations.

Start points:

- Persona hub: [Business representative](../personas/business-representative.md)
- Business entry point: [For businesses](../for-businesses/index.md)

### Technical writer

1. **Classify a page as exactly one Diátaxis type and fix violations** (using repo gates).
2. Add a new page and wire it into `mkdocs.yml` nav.
3. Run the docs linter and fix link-label and H1 issues.
4. Find the canonical map to avoid duplicating glossary/style.

Start points:

- Persona hub: [Technical writer](../personas/technical-writer.md)
- Quality gates: [Documentation quality gates](docs-quality-gates.md)

### Recommendation systems expert

1. **Understand ranking behavior and what is deterministic** (scoring, constraints, tie-break).
2. Find how to evaluate a change (offline gate) and interpret results.
3. Find how to add/adjust a signal end-to-end.
4. Find how to validate exposure/outcome logging and joinability.

Start points:

- Persona hub: [Recommendation systems expert](../personas/recommendation-systems-expert.md)
- Engineering hub: [RecSys engineering hub](../recsys-engineering/index.md)

## Read next

- Docs style guide: [Docs style guide](docs-style.md)
- Documentation quality gates: [Documentation quality gates](docs-quality-gates.md)
- Canonical content map: [Canonical content map](canonical-map.md)
