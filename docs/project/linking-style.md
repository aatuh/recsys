---
diataxis: reference
tags:
  - docs
  - style
  - links
---
# Linking and naming style reference

This page is the canonical reference for **link text, cross-links, and page naming**.

## Link text rules (intent-first)

- Link labels must describe **user intent**, not repo paths or filenames.
- Assume the link will be read out of context (screen readers, search results).

✅ Good

- "Run the minimal quickstart"
- "See pricing and plan scope"
- "Validate joinability (request_id → outcomes)"

❌ Avoid

- "tutorials/quickstart.md"
- "mkdocs.yml"
- "LICENSE" (as a bare label)

## Location line (optional)

If a filesystem path is helpful, add a small Location line near the link:

- Location: `docs/<path>`

## Canonical vs non-canonical pages

- Prefer **one canonical page** for each repeated concept (glossary, style, quality gates, licensing decision tree).
- Component docs must link to canonical pages instead of copying them.

See the canonical map:

- [Canonical content map](canonical-map.md)

## Cross-linking rules

- Every persona hub must link to its primary tutorial/how-to within **one click**.
- Tutorials and how-to guides must end with exactly one **Read next** section (3–5 links).
- Explanations should link to the most relevant how-to or tutorial when a user might want to act.

## Page naming conventions

- Tutorials: `Tutorial: <goal>`
- How-to: `How-to: <task>`
- Reference: `<thing> reference` (or just `<thing>` if unambiguous)
- Explanation: `<concept>`

## Read next

- Documentation quality gates: [Documentation quality gates](docs-quality-gates.md)
- Docs templates: [Docs templates](docs-templates.md)
- Docs style guide: [Docs style guide](docs-style.md)
