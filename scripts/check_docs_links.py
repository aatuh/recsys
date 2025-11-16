#!/usr/bin/env python3
"""
Verify that local Markdown links point to existing files.

Usage:
  python scripts/check_docs_links.py
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
MD_LINK = re.compile(r"\[([^\]]+)\]\(([^)]+)\)")
IGNORE_PREFIXES = ("http://", "https://", "mailto:", "#")


def main() -> int:
    failures: list[str] = []
    for md_path in ROOT.rglob("*.md"):
        # skip vendored or generated areas if needed
        if any(part.startswith(".git") or part in {"node_modules", ".trash"} for part in md_path.parts):
            continue
        text = md_path.read_text(encoding="utf-8")
        for match in MD_LINK.finditer(text):
            target = match.group(2).strip()
            if not target or target.startswith(IGNORE_PREFIXES):
                continue
            # Ignore inline anchors like `[foo](#bar)`
            if target.startswith("#"):
                continue
            # Strip anchors or query strings (`docs/foo.md#section`)
            path_part = target.split("#", 1)[0].split("?", 1)[0]
            # Skip absolute paths (/, C:, etc.)
            if Path(path_part).is_absolute():
                continue
            candidate = (md_path.parent / path_part).resolve()
            if not candidate.exists():
                failures.append(f"{md_path.relative_to(ROOT)} -> {target}")
    if failures:
        print("Broken Markdown links found:")
        for item in failures:
            print(f"  - {item}")
        return 1
    print("All Markdown links point to existing files.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
