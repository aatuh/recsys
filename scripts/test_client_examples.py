#!/usr/bin/env python3
"""
Ensure the code snippets in docs/client_examples.md compile.

Checks:
  - Python example compiles (no syntax errors).
  - JavaScript example passes `node --check` when Node.js is available.
"""

from __future__ import annotations

import shutil
import subprocess
import sys
import tempfile
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[1]
CLIENT_DOC = ROOT / "docs" / "client_examples.md"


def extract_code(language: str) -> str:
    pattern = re.compile(rf"```{language}\s+(.*?)```", re.DOTALL)
    text = CLIENT_DOC.read_text(encoding="utf-8")
    match = pattern.search(text)
    if not match:
        raise RuntimeError(f"No ```{language}``` block found in {CLIENT_DOC}")
    return match.group(1).strip() + "\n"


def check_python() -> None:
    code = extract_code("python")
    compile(code, str(CLIENT_DOC), "exec")


def check_node() -> None:
    node = shutil.which("node")
    if node is None:
        print("node executable not found; skipping JavaScript example check.")
        return
    code = extract_code("javascript")
    with tempfile.NamedTemporaryFile("w", suffix=".js", delete=False, encoding="utf-8") as tmp:
        tmp.write(code)
        tmp_path = tmp.name
    try:
        subprocess.run([node, "--check", tmp_path], check=True)
    finally:
        Path(tmp_path).unlink(missing_ok=True)


def main() -> int:
    check_python()
    check_node()
    print("Client examples compiled successfully.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
