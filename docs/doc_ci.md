# Documentation CI Checks

Use these lightweight scripts to keep Markdown healthy and ensure embedded code examples stay syntactically valid. Run them locally before opening a PR, and wire them into CI (GitHub Actions, etc.) so regressions are caught early.

## 1. Markdown link checker

Validates that local links (e.g., `[overview](overview.md)`) point to existing files.

```bash
python scripts/check_docs_links.py
```

- Ignores external URLs (`http`, `https`, `mailto`).
- Fails with a list of broken links when any target is missing.
- Use this as a pre-commit hook or CI step to avoid stale references.

## 2. Client example syntax tests

Compiles the Python and JavaScript snippets in [`docs/client_examples.md`](client_examples.md) to ensure they stay valid as the docs evolve.

```bash
python scripts/test_client_examples.py
```

- Requires Python 3.10+; JavaScript checks run only if the `node` binary is available.
- Does not call real APIs; it validates syntax only, so it’s safe in CI.

## 3. Suggested CI integration

- Add both commands to your documentation workflow (e.g., `.github/workflows/docs.yml`):

```yaml
- name: Check Markdown links
  run: python scripts/check_docs_links.py

- name: Compile client examples
  run: python scripts/test_client_examples.py
```

- Combine with spell-checkers or linters if desired.
- For repos with many docs, consider running the link checker only on changed files to keep CI fast.
- As part of manual review, run a quick search (`rg "Recsys|RecSYS" README.md docs`) to ensure product naming stays consistent as **RecSys** in prose. Keep uppercase variants (such as `RECSYS_GIT_COMMIT`) only for environment variables and identifiers.

## 4. Acronym review (manual)

Until we add a dedicated checker, keep acronym usage consistent by:

- Expanding key acronyms on first use in each doc (for example, “Maximal Marginal Relevance (MMR, a diversity-aware re-ranking method)”).
- Skimming for upper-case metric names such as `NDCG`, `MRR`, `MMR`, `CTR`, and `KPI` and confirming they are expanded nearby.
- Using the canonical phrases listed in [`docs/doc_style.md`](doc_style.md) under the acronym section.
