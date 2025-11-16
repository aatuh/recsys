# Documentation CI Checks

Use these lightweight scripts to keep Markdown healthy and ensure embedded code examples stay syntactically valid. Run them locally before opening a PR, and wire them into CI (GitHub Actions, etc.) so regressions are caught early.

## 1. Markdown link checker

Validates that local links (e.g., `[overview](docs/overview.md)`) point to existing files.

```bash
python scripts/check_docs_links.py
```

- Ignores external URLs (`http`, `https`, `mailto`).
- Fails with a list of broken links when any target is missing.
- Use this as a pre-commit hook or CI step to avoid stale references.

## 2. Client example syntax tests

Compiles the Python and JavaScript snippets in `docs/client_examples.md` to ensure they stay valid as the docs evolve.

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
