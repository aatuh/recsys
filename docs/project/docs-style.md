# Documentation style guide

This repository follows the **Diataxis** structure:

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

## Suite-level page templates

Use these templates for **suite-level docs** under `docs/` (as opposed to module docs that have their own style guides).

### Tutorials

- Goal (what you will build/run)
- Who this is for
- What you will get
- Prereqs (tools + access)
- Steps (numbered, copy/paste)
- Verify (expected output shape)
- Troubleshooting (common failures)
- Read next (links)

### How-to guides

- Who this is for (optional but recommended)
- Goal / outcomes (what you will achieve)
- Prereqs
- Steps
- Verify
- Pitfalls / gotchas
- Read next

### Explanations

- Who this is for
- What you will get
- Concepts and data flow (diagram if helpful)
- Failure modes and recovery notes
- Read next

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

## Glossary linking

- Add new suite terms to [`glossary.md`](glossary.md) (so we don’t define the same thing in five places).
- On first mention in a page, link glossary terms to the relevant entry (for example: `glossary.md#manifest`).
- Avoid over-linking: link once per page/section unless repetition prevents scanning.

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
