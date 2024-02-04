#!/usr/bin/env python3
"""External link checker for user-facing docs.

Goal:
Catch dead external links (404/410) without making CI flaky.

Scope:
- Scans markdown under docs/ plus a small set of root docs files.
- Skips local/dev URLs (localhost, docker hostnames).
- Uses HEAD where possible; falls back to a tiny GET (Range: bytes=0-0).
"""

from __future__ import annotations

import re
import sys
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path
from urllib.error import HTTPError, URLError
from urllib.parse import urlparse
from urllib.request import Request, urlopen

ROOT = Path(__file__).resolve().parent.parent
DOCS_DIR = ROOT / "docs"

EXTRA_DOC_FILES = [
    ROOT / "README.md",
    ROOT / "mkdocs.yml",
]

# Extract URLs from plain text and markdown. We intentionally keep this simple and
# run the result through urlparse-based validation afterwards.
URL_RE = re.compile(r"https?://[^\s<>()]+")

SKIP_HOSTS = {
    "localhost",
    "127.0.0.1",
    "0.0.0.0",
    "minio",  # docker compose hostname used in tutorials
    "recsys-svc",  # docker compose hostname used in tests/tutorials
    "example.com",  # placeholder host used in docs examples
}

USER_AGENT = "recsys-docs-linkcheck/1.0"
TIMEOUT_SECS = 15
MAX_RETRIES = 2
MAX_WORKERS = 8


def canonical_url(raw: str) -> str | None:
    url = raw.strip().lstrip("<").rstrip(">")
    url = url.rstrip("`'\".,;:!?)[]}")
    parsed = urlparse(url)
    if parsed.scheme not in {"http", "https"}:
        return None
    if not parsed.netloc:
        return None
    return url


def should_skip(url: str) -> bool:
    host = (urlparse(url).hostname or "").lower()
    return host in SKIP_HOSTS


def fetch_status(url: str) -> int | None:
    headers = {"User-Agent": USER_AGENT}
    for method in ("HEAD", "GET"):
        request_headers = dict(headers)
        if method == "GET":
            request_headers["Range"] = "bytes=0-0"
        req = Request(url, headers=request_headers, method=method)
        try:
            with urlopen(req, timeout=TIMEOUT_SECS) as resp:
                return resp.getcode()
        except HTTPError as exc:
            # Some servers reject HEAD. If so, try a tiny GET.
            if method == "HEAD" and exc.code in {403, 405}:
                continue
            return exc.code
        except URLError:
            return None
        except Exception:
            return None
    return None


def check_url(url: str) -> int | None:
    for attempt in range(MAX_RETRIES + 1):
        status = fetch_status(url)
        if status is not None:
            return status
        if attempt < MAX_RETRIES:
            time.sleep(1.0 * (attempt + 1))
    return None


def iter_doc_sources() -> list[Path]:
    sources = list(DOCS_DIR.rglob("*.md"))
    for extra in EXTRA_DOC_FILES:
        if extra.exists():
            sources.append(extra)
    return sources


def main() -> int:
    url_sources: dict[str, list[str]] = {}
    for path in iter_doc_sources():
        try:
            lines = path.read_text(encoding="utf-8", errors="ignore").splitlines()
        except Exception:
            continue
        for lineno, line in enumerate(lines, start=1):
            for m in URL_RE.finditer(line):
                url = canonical_url(m.group(0))
                if url is None or should_skip(url):
                    continue
                src = path.relative_to(ROOT)
                url_sources.setdefault(url, []).append(f"{src}:{lineno}")

    if not url_sources:
        print("External docs link check OK (no URLs found)")
        return 0

    broken: list[tuple[str, int | None]] = []
    warnings: list[tuple[str, int]] = []

    with ThreadPoolExecutor(max_workers=MAX_WORKERS) as pool:
        futures = {pool.submit(check_url, url): url for url in url_sources}
        for fut in as_completed(futures):
            url = futures[fut]
            status = fut.result()

            if status is None:
                broken.append((url, None))
                continue

            if 200 <= status < 400:
                continue
            if status in {401, 403, 429}:
                # Reachable but protected / rate-limited. Do not fail CI on this.
                continue
            if status in {404, 410}:
                broken.append((url, status))
                continue
            if 500 <= status < 600:
                warnings.append((url, status))
                continue

            broken.append((url, status))

    if warnings:
        print("External docs link check warnings (transient errors):\n")
        for url, status in sorted(warnings, key=lambda x: x[0]):
            print(f"- {status} {url}")
        print("")

    if broken:
        print("Broken external links detected:\n")
        for url, status in sorted(broken, key=lambda x: x[0]):
            status_str = "error" if status is None else str(status)
            print(f"- {status_str} {url}")
            for src in sorted(url_sources.get(url, [])):
                print(f"  - {src}")
        return 1

    print(f"External docs link check OK ({len(url_sources)} URL(s) checked)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
