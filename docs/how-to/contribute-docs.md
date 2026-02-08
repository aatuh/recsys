---
diataxis: how-to
tags:
  - docs
  - contributing
---
# How-to: contribute to the docs

Use this guide when you add or change documentation pages.

## Step 1 — Pick exactly one Diátaxis type

Every page must be exactly one of:

- Tutorial
- How-to
- Reference
- Explanation

If a page mixes intents, split it (see quality gates).

Reference:

- Documentation quality gates: [Documentation quality gates](../project/docs-quality-gates.md)
- Docs templates: [Docs templates](../project/docs-templates.md)

## Step 2 — Add front matter and one H1

At the top of every Markdown page:

```yaml
---
diataxis: <tutorial|how-to|reference|explanation>
tags:
  - <tag>
---
# <One H1 title>
```

Rules:

- Exactly one `#` heading per page
- Link text must be intent-first (no raw file paths)

Reference:

- Linking and naming: [Linking and naming style reference](../project/linking-style.md)

## Step 3 — Wire the page into navigation

Add the page to `mkdocs.yml` under the correct Diátaxis group.

If you move or rename pages, also update redirects:

- `plugins.redirects.redirect_maps` in `mkdocs.yml`

## Step 4 — Run lint and fix issues

```bash
python3 scripts/docs_lint.py
```

Fix any violations (H1, Diátaxis missing, link labels, read-next hygiene) before merging.

## Step 5 — Validate persona journeys (optional but recommended)

If you touched persona hubs or major navigation, re-run the journey tests:

- [Persona journey tests and scoring rubric](../project/persona-journey-tests.md)

## Read next

- Run docs locally: [How-to: run the docs locally](run-docs-locally.md)
- Documentation quality gates: [Documentation quality gates](../project/docs-quality-gates.md)
- Linking and naming: [Linking and naming style reference](../project/linking-style.md)
- Persona rubric: [Persona journey tests and scoring rubric](../project/persona-journey-tests.md)
