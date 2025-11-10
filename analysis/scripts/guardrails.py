#!/usr/bin/env python3
"""
Guardrail configuration loader.

Reads guardrails.yml (YAML/JSON) so automation can apply per-customer thresholds.
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import Dict, Optional

import yaml


def load_guardrails(path: str) -> Dict:
    file_path = Path(path)
    if not file_path.exists():
        raise FileNotFoundError(f"Guardrails file '{path}' not found.")
    text = file_path.read_text(encoding="utf-8")
    if file_path.suffix.lower() in {".yml", ".yaml"}:
        data = yaml.safe_load(text)
    else:
        data = json.loads(text)
    if not isinstance(data, dict):
        raise ValueError("Guardrails file must describe an object.")
    return data


def resolve_guardrails(config: Dict, customer: str, namespace: Optional[str]) -> Dict:
    defaults = config.get("defaults", {})
    customers = config.get("customers", {})
    entry = customers.get(customer, {})
    if entry and namespace and "namespace" in entry and entry["namespace"] != namespace:
        # Allow multiple entries by namespace under same customer
        # Support structure: customers: { customer: { namespaces: { ns: {...} } } }
        namespaces = entry.get("namespaces")
        if isinstance(namespaces, dict) and namespace in namespaces:
            entry = namespaces[namespace]
    resolved = {"quality": {}, "scenarios": {}}
    for key in ("quality", "scenarios"):
        resolved[key].update(defaults.get(key, {}))
        resolved[key].update(entry.get(key, {}))
    return resolved
