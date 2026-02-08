---
diataxis: reference
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

This is the **canonical** style guide for the entire RecSys documentation set rendered by MkDocs:

- suite-level docs under `docs/`
- module docs included in the same site (for example `docs/recsys-eval/docs/` and `docs/recsys-pipelines/docs/`)

If a module needs a module-specific convention, document it **briefly** on the module page and link back here. Avoid duplicating rules.

## Writing rules

- Prefer short sections and explicit headings.
- Use consistent terminology (see [Glossary](glossary.md)).
- Use consistent product naming and capitalization (see "Naming" below).
- Always include:
  - prerequisites
  - expected outcomes
  - examples and sample outputs (when applicable)
- Prefer relative links inside `/docs`.
- Avoid linking to repository paths that do not exist in the rendered MkDocs site.


## Page skeleton

For narrative pages (tutorials, how-to guides, explanations, and persona hubs):

- Start with **Who this is for** and **What you will get**.
- Keep the intro short (2–5 sentences).
- Use headings that read like decisions or actions.
- Prefer lists over long paragraphs.
- Add callouts when a mistake is costly (privacy, production, irreversible changes).

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
- **Optional** for deep reference pages and module docs.

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

For a ready-to-copy set of templates, see: [Docs templates](docs-templates.md)

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

- Add new suite terms to [Glossary](glossary.md) (so we don’t define the same thing in five places).
- On first mention in a page, link glossary terms to the relevant entry (for example: `glossary.md#manifest`).
- Avoid over-linking: link once per page/section unless repetition prevents scanning.

## Canonical pages (avoid duplication)

When a concept is used across multiple pages (pricing, licensing, security posture, core definitions), use a single
canonical page and link to it elsewhere.

See: [Canonical content map](canonical-map.md).

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
  [Error handling & troubleshooting API calls](../reference/api/errors.md).
- Auth expectations are explicit: what headers/claims are required (dev headers vs JWT).
- The “happy path” is copy/paste runnable: examples in [API examples (HTTP files)](../reference/api/examples.md)
  match the schema and response shapes in OpenAPI.

## Tag taxonomy

Tags are for discovery and filtering. Prefer **few and consistent** tags.

Recommended stable tags:

- **Diátaxis**: `tutorial`, `how-to`, `reference`, `explanation`
- **Personas**: `developer`, `business`, `ops`, `recsys-engineering`, `technical-writer`
- **Modules**: `recsys-algo`, `recsys-eval`, `recsys-pipelines`, `recsys-service`
- **Topics**: `integration`, `security`, `privacy`, `pricing`, `runbook`, `checklist`, `api`, `database`

Rules:

- Keep tags lowercase and hyphenated.
- Avoid synonyms (`ops` vs `operations`): pick one and stick to it.
- Prefer tagging a page rather than duplicating content.

## Docs QA checklist

Before merging documentation changes, ensure:

- [ ] One clear H1 per page
- [ ] No truncated link labels (avoid `...` in visible text)
- [ ] Links are relative (within docs) and not broken
- [ ] Code blocks have language fences where applicable
- [ ] Every tutorial has:
  - [ ] prerequisites
  - [ ] steps
  - [ ] definition of done
  - [ ] troubleshooting or failure matrix
- [ ] Every page has tags and a “Read next” section (unless explicitly exempted)

## No truncation and link hygiene

- Do not use truncated labels like `chatgpt.com/g/...` in visible link text.
- Prefer descriptive labels (e.g., “Recsys Copilot”) and keep full URLs only in the link target.

## Docs contribution workflow

When contributing docs:

1. Update the relevant Markdown page(s).
2. Run a local build (or CI) to catch broken links.
3. Ensure the nav still makes sense (no page-hunting).
4. Add or update “Read next” links to keep journeys cohesive.

See: [Contributing](contributing.md)

