#!/usr/bin/env python3
"""Lightweight internal link checker for /docs.

Rules:
- Checks relative links in markdown under docs/ (excluding http(s), mailto, # anchors).
- For links ending with '/', expects <path>/index.md.
- For links to .md files, expects they exist.
- Ignores links that point outside docs/ (e.g. ../) to avoid false positives.

This is intentionally simple: it catches most drift without requiring external tools.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

DOCS = Path(__file__).resolve().parent.parent / "docs"

LINK_RE = re.compile(r"\]\(([^)]+)\)")

IGNORE_PREFIXES = (
    "http://",
    "https://",
    "mailto:",
    "#",
)


def norm_target(md_path: Path, href: str) -> Path | None:
    # Strip anchors and query strings
    href = href.split("#", 1)[0].split("?", 1)[0].strip()
    if not href or href.startswith(IGNORE_PREFIXES):
        return None
    # Ignore templated/placeholder links
    if "{" in href or "}" in href:
        return None
    # Ignore links to repo-root or higher
    if href.startswith("../"):
        return None

    base = md_path.parent
    target = (base / href).resolve()

    try:
        target.relative_to(DOCS.resolve())
    except Exception:
        return None

    # Directory-style links => index.md
    if href.endswith("/"):
        return target / "index.md"

    # If it points to a directory without trailing slash, allow index.md
    if target.is_dir():
        return target / "index.md"

    # Otherwise treat as file path
    return target


def main() -> int:
    missing: list[tuple[Path, str]] = []
    for md in DOCS.rglob("*.md"):
        txt = md.read_text(encoding="utf-8", errors="ignore")
        for m in LINK_RE.finditer(txt):
            href = m.group(1).strip()
            target = norm_target(md, href)
            if target is None:
                continue
            if not target.exists():
                missing.append((md.relative_to(DOCS), href))

    if missing:
        print("Broken internal links detected:\n")
        for src, href in missing:
            print(f"- {src}: {href}")
        return 1

    print("Docs link check OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
