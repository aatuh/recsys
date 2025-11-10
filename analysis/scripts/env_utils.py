#!/usr/bin/env python3
"""
Helpers for recording environment provenance (env file hash, metadata).
"""

from __future__ import annotations

import hashlib
from pathlib import Path
from typing import Dict


def compute_env_hash(path: str) -> str:
    file_path = Path(path)
    if not file_path.exists():
        return ""
    try:
        data = file_path.read_bytes()
    except OSError:
        return ""
    return hashlib.sha256(data).hexdigest()


def env_metadata(path: str) -> Dict[str, str]:
    return {
        "env_file": path,
        "env_hash": compute_env_hash(path),
    }
