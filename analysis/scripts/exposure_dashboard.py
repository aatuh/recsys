#!/usr/bin/env python3
"""Compute brand/category exposure ratios from recommendation dump evidence."""

from __future__ import annotations

import argparse
import json
import math
import os
import sys
from typing import Dict, Any


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Validate exposure diversity per namespace.")
    parser.add_argument(
        "--input",
        default="analysis/results/recommendation_dump.json",
        help="Path to recommendation dump JSON (default: %(default)s)",
    )
    parser.add_argument(
        "--output",
        default="analysis/results/exposure_dashboard.json",
        help="Where to write the computed dashboard (default: %(default)s)",
    )
    parser.add_argument(
        "--threshold",
        type=float,
        default=1.4,
        help="Maximum allowed max/mean exposure ratio (default: %(default)s)",
    )
    return parser.parse_args()


def load_dump(path: str) -> Dict[str, Any]:
    if not os.path.exists(path):
        raise FileNotFoundError(f"Input dump {path} not found")
    with open(path, "r", encoding="utf-8") as fh:
        return json.load(fh)


def compute_ratio(counts: Dict[str, int]) -> float:
    if not counts:
        return 0.0
    values = [v for v in counts.values() if v is not None]
    if not values:
        return 0.0
    mean = sum(values) / len(values)
    max_val = max(values)
    if mean <= 0:
        return math.inf
    return max_val / mean


def main() -> None:
    args = parse_args()
    data = load_dump(args.input)
    summary = data.get("summary") or {}
    dashboard = {
        "threshold": args.threshold,
        "namespaces": {},
    }
    failures = []
    for namespace, stats in summary.items():
        ns_entry: Dict[str, Any] = {}
        brand_counts = stats.get("brand_counts") or {}
        ratio = compute_ratio(brand_counts)
        ns_entry["brand_ratio"] = ratio
        ns_entry["max_brand_exposure"] = stats.get("max_brand_exposure")
        ns_entry["mean_brand_exposure"] = stats.get("mean_brand_exposure")
        ns_entry["brand_counts"] = brand_counts
        ns_entry["user_count"] = stats.get("user_count")
        ns_entry["exposure_ratio"] = stats.get("exposure_ratio", ratio)
        ns_entry["status"] = "pass" if ratio <= args.threshold else "fail"
        dashboard["namespaces"][namespace] = ns_entry
        if ratio > args.threshold:
            failures.append((namespace, ratio))

    os.makedirs(os.path.dirname(args.output), exist_ok=True)
    with open(args.output, "w", encoding="utf-8") as fh:
        json.dump(dashboard, fh, indent=2)

    if failures:
        details = ", ".join(f"{ns}={ratio:.2f}" for ns, ratio in failures)
        raise SystemExit(f"Exposure ratio exceeded threshold {args.threshold:.2f}: {details}")


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:  # pragma: no cover - CLI surface
        print(f"ERROR: {exc}", file=sys.stderr)
        raise
