---
diataxis: reference
tags:
  - docs
  - reference
---
# Documentation quality gates

These gates keep the docs stable across iterations. If a gate is violated, fix the docs before adding new content.

## G1) Diátaxis purity

Every page must be classified as exactly one of:

- Tutorial
- How-to
- Reference
- Explanation

**Rule:** a page must do one job. If it tries to teach *and* be a lookup *and* justify design, split it.

**Implementation in this repo:** every Markdown page has front matter `diataxis: tutorial|how-to|reference|explanation`.

## G2) One H1 rule

Each page has exactly one top-level `#` heading.

- If you need multiple top-level topics, split pages.
- If you want a section, demote headings to `##` or lower.

## G3) Scannability baseline

A reader must be able to skim any page and understand:

- what this page is for
- what to do next (if applicable)

**Minimum page structure:**

1. 1–3 sentence intro
2. Clear section headings (reader intent)
3. Lists over dense paragraphs
4. Callouts for hazards / decisions (`!!! warning`, `!!! info`, `!!! tip`)

## G4) Intent-first links

Link text must describe intent, not repo paths.

- Good: "Run the minimal quickstart"
- Bad: "tutorials/quickstart.md"

If a file path is useful, include it only as:

- `Location: docs/<path>`

## G5) “Read next” hygiene

Tutorials and How-to guides must end with a single **Read next** section:

- 3–5 links
- directly relevant
- no more content after it

## Repo conventions

### Page title format

- Tutorials: `Tutorial: <goal>`
- How-to: `How-to: <task>`
- Reference: `<thing> reference` or `<thing>` (if unambiguous)
- Explanation: `<concept>`

### Canonical content

- One canonical glossary: [Glossary](glossary.md)
- One canonical docs style guide: [Docs style guide](docs-style.md)

Component docs may link to canonical content, but must not fork it.

### Automated checks

If you run linting in CI, use:

- `python3 scripts/docs_lint.py`


## Read next

- Persona journey tests and scoring rubric: [Persona journey tests and scoring rubric](persona-journey-tests.md)
- Linking and naming: [Linking and naming style reference](linking-style.md)
- Docs templates: [Docs templates](docs-templates.md)
