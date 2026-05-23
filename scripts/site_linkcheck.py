#!/usr/bin/env python3
"""Validate internal links in a generated static site."""

from __future__ import annotations

import sys
from html.parser import HTMLParser
from pathlib import Path
from urllib.parse import unquote, urlparse


class LinkParser(HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self.links: list[str] = []

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        attr_names = ("href", "src") if tag in {"a", "link", "script", "img", "source"} else ()
        if not attr_names:
            return
        data = dict(attrs)
        for name in attr_names:
            value = data.get(name)
            if value:
                self.links.append(value)


SKIP_SCHEMES = {"http", "https", "mailto", "tel", "data"}


def target_for(root: Path, html_file: Path, raw: str) -> Path | None:
    parsed = urlparse(raw)
    if parsed.scheme in SKIP_SCHEMES or parsed.netloc:
        return None
    if raw.startswith("#"):
        return None

    path = unquote(parsed.path)
    if not path:
        return None

    if path.startswith("/"):
        target = root / path.lstrip("/")
    else:
        target = (html_file.parent / path).resolve()
        try:
            target.relative_to(root)
        except ValueError:
            return None

    if path.endswith("/"):
        return target / "index.html"
    if target.suffix:
        return target
    if target.is_dir():
        return target / "index.html"
    return target.with_name(target.name) / "index.html"


def main(argv: list[str]) -> int:
    root = Path(argv[1] if len(argv) > 1 else ".site").resolve()
    if not root.is_dir():
        print(f"site directory is missing: {root}", file=sys.stderr)
        return 1

    missing: list[tuple[Path, str, Path]] = []
    for html in sorted(root.rglob("*.html")):
        parser = LinkParser()
        parser.feed(html.read_text(encoding="utf-8", errors="ignore"))
        for link in parser.links:
            target = target_for(root, html, link)
            if target is not None and not target.exists():
                missing.append((html.relative_to(root), link, target.relative_to(root)))

    if missing:
        print("Broken generated-site links detected:\n")
        for src, link, target in missing:
            print(f"- {src}: {link} -> {target}")
        return 1

    print("Generated site link check OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
