#!/usr/bin/env python3
"""Extract code snippets from docs into a single ordered markdown file.

Why this exists:
- To support snippet testing/smoke checks so docs examples can be executed and
  validated in sequence, proving they still work.

What this script does:
- Resolves repo root as `scripts/..` so it works from any current directory.
- Reads all `docs/**/*.md` files.
- Orders files by `mkdocs.yml` nav order, then appends remaining docs files
  alphabetically.
- Extracts fenced code blocks (including indented fences), keeps their source
  file + start line, and writes them to `/.snippets.md`.
- Skips non-executable/presentation fences: empty-lang, `text`, `mermaid`,
  `md`, `markdown`.
"""

import re
from pathlib import Path

# Resolve repository root as one level up from this script's directory.
root = Path(__file__).resolve().parent.parent
docs = root / "docs"
mkdocs = root / "mkdocs.yml"
out = root / ".snippets.md"

all_files = sorted(p for p in docs.rglob("*.md"))
all_rel = [p.relative_to(root).as_posix() for p in all_files]

# Build file order from mkdocs nav list items, then append remaining docs files.
nav_item_re = re.compile(
    r"^\s*-\s+(?:(?:[^:#\n]+):\s+)?([A-Za-z0-9_./-]+\.md)\s*$")
nav_order = []
seen = set()
for line in mkdocs.read_text(encoding="utf-8").splitlines():
    m = nav_item_re.match(line)
    if not m:
        continue
    p = m.group(1)
    rel = f"docs/{p}" if not p.startswith("docs/") else p
    full = root / rel
    if full.exists() and rel not in seen:
        seen.add(rel)
        nav_order.append(rel)

remaining = [r for r in all_rel if r not in seen]
ordered_files = nav_order + remaining

# Include executable-ish code fences; skip presentation-only fences.
skip_langs = {"", "text", "mermaid", "md", "markdown"}
open_re = re.compile(r"^\s*```\s*([A-Za-z0-9_+-]*)\s*$")  # allows indentation
close_re = re.compile(r"^\s*```\s*$")

blocks = []
for rel in ordered_files:
    p = root / rel
    lines = p.read_text(encoding="utf-8").splitlines()
    i = 0
    while i < len(lines):
        m = open_re.match(lines[i])
        if not m:
            i += 1
            continue

        lang = (m.group(1) or "").lower()
        start_line = i + 1
        i += 1

        content = []
        while i < len(lines) and not close_re.match(lines[i]):
            content.append(lines[i])
            i += 1
        if i < len(lines) and close_re.match(lines[i]):
            i += 1

        if lang in skip_langs:
            continue

        blocks.append((rel, start_line, lang, "\n".join(content).rstrip()))

with out.open("w", encoding="utf-8") as f:
    f.write("# Snippets\n\n")
    f.write(
        "Assumed execution order: docs nav order from `mkdocs.yml` "
        "(tutorial/how-to/reference flow), then remaining `docs/**/*.md` "
        "alphabetically; in-file snippet order is preserved.\n\n"
    )
    for idx, (rel, line, lang, body) in enumerate(blocks, 1):
        f.write(f"## {idx}. `{rel}:{line}` ({lang})\n\n")
        f.write(f"```{lang}\n")
        if body:
            f.write(body + "\n")
        f.write("```\n\n")

print(f"wrote {out} with {len(blocks)} snippets")
