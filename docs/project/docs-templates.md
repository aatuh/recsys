---
diataxis: reference
tags:
  - project
  - docs
  - templates
---
# Docs templates

Use these templates for new pages to keep the site consistent and scannable.

## Tutorial template

```markdown
---
tags:
  - tutorial
  - <topic>
---

# Tutorial: <title>

## Who this is for

- ...

## What you will get

- ...

## Prereqs

- ...

## Steps

## Verify (Definition of Done)

- [ ] ...

## Troubleshooting

- ...

## Read next

- ...
```

## How-to template

```markdown
---
tags:
  - how-to
  - <topic>
---

# <task name>

## Who this is for

- ...

## What you will get

- ...

## Steps

## Verify

- [ ] ...

## Read next

- ...
```

## Reference template

```markdown
---
tags:
  - reference
  - <topic>
---

# <thing name>

## Purpose

## Schema / contract

## Examples

## Compatibility notes

## Read next
```

## Explanation template

```markdown
---
tags:
  - explanation
  - <topic>
---

# <concept name>

## Problem

## Why this exists

## How it works (mental model)

## Tradeoffs

## Read next
```

## Read next

- Docs style guide: [Documentation style guide](docs-style.md)
- Docs per release policy: [Docs per release policy](docs-per-release.md)
