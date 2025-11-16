# Documentation Style & Terminology Guide

Use this guide to keep RecSys docs consistent and easy to scan. Apply these conventions when writing new docs or updating existing ones.

## 1. Tone & voice

- Write in plain English, with “you” and “we” sparingly. Favor direct, imperative sentences (“Run `make dev`.”).
- Keep paragraphs short (2–4 sentences). Break long lists into subsections with headings.
- Avoid marketing fluff; describe what the system does, why it matters, and how to use it.

## 2. Headings & structure

- Every doc starts with an `# H1` title followed by a short introductory paragraph explaining audience and purpose. Use the `<Who should read this?>` pattern where appropriate.
- Use `##` headings for major sections and `###` for sub-sections. Avoid going deeper than `###` unless absolutely necessary.
- Include “Where to go next” or “Related docs” when the reader should continue elsewhere.

## 3. Links & cross-references

- Use relative links (e.g., `docs/overview.md`) so they render correctly in GitHub and other viewers.
- When referencing scripts or commands, specify the relative path (`analysis/scripts/run_scenarios.py`, `make scenario-suite`).
- Avoid bare URLs; wrap them in descriptive text (e.g., `[Install guide](https://docs.docker.com/get-docker/)`).

## 4. Terminology & acronyms

- Introduce acronyms on first use with a short description, e.g., “Maximal Marginal Relevance (MMR)”.
- Use “guardrails” instead of “constraints” when referring to the YAML/CI policy system.
- Refer to RecSys as a “recommendation control plane” when describing the platform at a high level.
- Namespaces are lower_snake_case in examples (`retail_demo`); org IDs are UUIDs.

Preferred terms:

- Use **org** and **namespace** instead of “tenant” when describing isolation.
- Use **env profile** for configuration bundles, and prefer that over vague “environment” when you mean a named profile applied to a namespace.

### Canonical acronym expansions

Use these phrases the first time each acronym appears in a doc:

- **MMR** – “Maximal Marginal Relevance (MMR, a diversity-aware re-ranking method)”
- **NDCG** – “Normalized Discounted Cumulative Gain (NDCG, a ranking quality score)”
- **MRR** – “Mean Reciprocal Rank (MRR, ‘how early do good items appear?’)”
- **CTR** – “click-through rate (CTR)”
- **KPI** – “key performance indicator (KPI)”

For key ranking terms, prefer:

- **Blend / blend weights** – “blend weights (how much each signal contributes to the final score)”

## 5. Formatting & samples

- Code blocks use triple backticks with a language hint: ```bash, ```python, ```javascript, etc.
- Use bullet lists instead of Markdown tables to improve readability, especially for “document – question – highlight” lists.
- Include the minimal context needed to run a command (env vars, base URLs) before showing the code block.
- When providing JSON examples, keep them under ~20 lines and highlight key fields inline (“`tags` power rules/personalization”).

## 6. Grammar & conventions

- Use American English spelling (behavior, favor, personalization).
- Use “guardrail” (singular) and “guardrails” (plural); avoid “guard rails”.
- “RecSys” capitalizes the R and S; write “API” in uppercase.
- Avoid placeholder ellipses (`...`). Replace them with real paths, URLs, or concrete instructions.

## 7. Review checklist

Before merging documentation changes:

- [ ] Does the doc start with audience/purpose context?
- [ ] Are acronyms expanded on first use?
- [ ] Are code samples tested or at least syntax-checked? (`docs/client_examples.md` + `scripts/test_client_examples.py`)
- [ ] Did you run `python scripts/check_docs_links.py` to catch broken links?
- [ ] Are there any lingering “TODO”, “...” placeholders, or inconsistent terms?

Following these guidelines keeps the docs approachable for business stakeholders and practical for engineers.
