---
diataxis: how-to
tags:
  - docs
  - contributing
---
# How-to: run the docs locally

Use this guide to build and preview the documentation site on your machine.

## Prerequisites

- Python 3
- `pip`

## Install doc dependencies

From the repo root:

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements-docs.txt
```

## Preview locally (live reload)

```bash
mkdocs serve
```

Then open the URL printed by MkDocs (usually `http://127.0.0.1:8000`).

## Build the static site

```bash
mkdocs build --strict
```

## Run documentation linting

From the repo root:

```bash
python3 scripts/docs_lint.py
```

If the linter reports issues, fix them before merging changes.

## Read next

- Contribute docs changes: [How-to: contribute to the docs](contribute-docs.md)
- Documentation quality gates: [Documentation quality gates](../project/docs-quality-gates.md)
- Docs templates: [Docs templates](../project/docs-templates.md)
- Docs style guide: [Docs style guide](../project/docs-style.md)
