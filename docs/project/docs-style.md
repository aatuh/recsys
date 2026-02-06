---
tags:
  - project
  - docs
---

# Documentation style guide

This repository follows the following structure:

- **Tutorials**: learning by doing
- **How-to**: goal-oriented steps
- **Explanation**: understanding and design rationale
- **Reference**: precise specification (APIs, schemas, config)

## Scope

These rules apply to **suite-level docs** under `docs/` (the MkDocs site).

Module docs may have their own style guides and can deviate when that improves clarity:

- `docs/recsys-eval/docs/` (see `docs/recsys-eval/docs/style.md`)
- `docs/recsys-pipelines/docs/` (see `docs/recsys-pipelines/docs/contributing/style.md`)
- `docs/recsys-algo/`

## Writing rules

- Prefer short sections and explicit headings.
- Use consistent terminology (see [`glossary.md`](glossary.md)).
- Use consistent product naming and capitalization (see "Naming" below).
- Always include:
  - prerequisites
  - expected outcomes
  - examples and sample outputs (when applicable)
- Prefer relative links inside `/docs`.
- Avoid linking to repository paths that do not exist in the rendered MkDocs site.

## Voice and tone

- Prefer **clear, concrete claims** over hype ("what is included", "what is not included").
- Write in **second person** for tutorials/how-tos ("you will", "run", "verify").
- Prefer **active voice** ("The service writes an exposure log", not "An exposure log is written").
- Use **imperative verbs** in steps ("Run", "Verify", "Copy").
- Define unfamiliar terms once, then link to the canonical definition (glossary or canonical page).

## Tags policy

This docs site uses Material for MkDocs **tags** for role- and topic-based browsing.

Tags are:

- **Required** for suite-level narrative pages under `docs/` (tutorials, how-to guides, explanations, hub pages,
  business/procurement docs).
- **Optional** for deep reference pages and module docs, especially when another style guide already applies
  (see "Scope" above).

Add tags as YAML front matter at the top of a page:

```md
---
tags:
  - how-to
  - integration
  - developer
---

# How-to: Integrate RecSys
```

Guidelines:

- Use **2–5 tags per page**.
- Prefer **consistent, shared tags** over one-off variants.
- Avoid `title:` in front matter; keep the page title as the `#` heading
  (our markdownlint config treats `title:` as an extra H1).

Recommended tags:

- Doc type: `overview`, `quickstart`, `tutorial`, `how-to`, `explanation`, `reference`, `runbook`
- Role: `business`, `developer`, `ops`, `ml`
- Components: `recsys-service`, `recsys-algo`, `recsys-pipelines`, `recsys-eval`
- Topics (examples): `architecture`, `integration`, `deployment`, `api`, `config`, `data-contracts`, `database`,
  `artifacts`, `evaluation`, `security`

## Suite-level page templates

Use these templates for **suite-level docs** under `docs/` (as opposed to module docs that have their own style guides).

### Exemptions (when "Who/What" is awkward)

Some pages are primarily **legal text** or **forms**. For these, it is acceptable to omit:

- `## Who this is for`
- `## What you will get`

Examples:

- License texts and pricing definitions (for example: evaluation license, pricing definitions)
- Order forms and order form templates

Rule of thumb: if the page is meant to be read verbatim by legal/procurement, start with a 1–2 sentence context
paragraph and link to the canonical buyer guide, then go straight into the content.

### Tutorials

- Goal (what you will build/run)
- Who this is for
- What you will get
- Prereqs (tools + access)
- Steps (numbered; each step has action + command + expected outcome)
- Verify (expected output shape)
- Troubleshooting (common failures)
- Read next (3 links)

### How-to guides

- Who this is for (optional but recommended)
- Goal / outcomes (what you will achieve)
- Prereqs
- Steps (each step has action + command + expected outcome)
- Verify
- Pitfalls / gotchas
- Read next (3 links)

### Explanations

- Who this is for
- What you will get
- Concepts and data flow (diagram if helpful)
- Failure modes and recovery notes
- Read next (3 links)

### Reference

- Who this is for
- What you will get
- Definitions / contracts / options (be precise)
- Examples (when it prevents ambiguity)
- Read next (optional)

### Runbooks

- Symptoms
- Quick triage (commands + what to look for)
- Likely causes
- Safe remediations
- Verification
- Escalation criteria

## Copy/paste templates

Use these skeletons for new suite-level pages. Keep headings consistent so readers can scan quickly.

### Tutorial skeleton

```md
# Title

## Who this is for

## What you will get

## Prereqs

## Steps

## Verify

## Troubleshooting

## Read next
```

### How-to skeleton

```md
# How-to: Title

## Who this is for

## Goal

## Prereqs

## Steps

## Verify

## Pitfalls

## Read next
```

### Explanation skeleton

```md
# Title

## Who this is for

## What you will get

## Concepts and data flow

## Failure modes and recovery

## Read next
```

### Reference skeleton

```md
# Title

## Who this is for

## What you will get

## Reference

## Examples

## Read next
```

## Markdown conventions

- Use fenced code blocks with language hints.
- Use admonitions for warnings and important notes.
- Keep line length readable (wrap long paragraphs).
- If you need a collapsible/accordion section that contains fenced code, prefer `<details markdown="1">`.
  Some markdown linters treat fences inside collapsible admonitions (`???`) as indented code blocks.

## Enforcement (what CI checks)

These are the checks we run on docs changes:

- `make mdlint`: markdownlint rules (including "one H1 per page")
- `make docs-check`: internal link check, spell check, strict MkDocs build, and a "reference stub" gate
  (`docs/reference/**` leaf pages must include `## Who this is for`, `## What you will get`, and at least one fenced
  code block).

## Glossary linking

- Add new suite terms to [`glossary.md`](glossary.md) (so we don’t define the same thing in five places).
- On first mention in a page, link glossary terms to the relevant entry (for example: `glossary.md#manifest`).
- Avoid over-linking: link once per page/section unless repetition prevents scanning.

## Canonical pages (avoid duplication)

When a concept is used across multiple pages (pricing, licensing, security posture, core definitions), use a single
canonical page and link to it elsewhere.

Rules:

- Put the full definition/decision tree on the canonical page.
- Other pages may include a short summary, then link to the canonical page.
- Avoid copy/pasting plan details, license rules, or “security posture” claims across multiple pages.

Example:

- Pricing plan details live on `docs/pricing/index.md`. Other pages should link there instead of repeating the plan
  table.

## Naming

Use these names consistently across pages, nav labels, and headings:

- Product: **RecSys suite** (or **RecSys** when the context is unambiguous)
- Service/API module: `recsys-service`
- Ranking core module: `recsys-algo`
- Pipelines module: `recsys-pipelines`
- Evaluation module: `recsys-eval`

Avoid mixed capitalization like “Recsys” in prose.

## API docs quality checklist

Use this as a PR checklist when changing the HTTP API or its docs.

- OpenAPI stays canonical: update `docs/reference/api/openapi.yaml` and re-run `cd api && make codegen`.
- Every changed endpoint has examples:
  - at least one success example
  - the common errors for that endpoint (e.g. 400/401/403/404/409/429)
- Errors are consistent: use Problem Details (RFC 7807) and document new problem types in
  [`reference/api/errors.md`](../reference/api/errors.md).
- Auth expectations are explicit: what headers/claims are required (dev headers vs JWT).
- The “happy path” is copy/paste runnable: examples in [`reference/api/examples.md`](../reference/api/examples.md)
  match the schema and response shapes in OpenAPI.
