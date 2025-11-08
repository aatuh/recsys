#!/usr/bin/env python3
"""
Probe determinism by replaying a fixed request multiple times and comparing rank variance.
"""

from __future__ import annotations

import argparse
import json
import os
import statistics
import sys
from typing import Dict, List

import requests


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Check recommendation determinism.")
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--org-id", required=True)
    parser.add_argument("--namespace", default="default")
    parser.add_argument("--baseline", required=True, help="Path to baseline determinism JSON")
    parser.add_argument("--calls", type=int, default=10, help="Number of repeated calls to perform")
    parser.add_argument("--max-rank-delta", type=float, default=0.01, help="Allowed fraction of rank changes")
    parser.add_argument("--output", default="analysis_v2/evidence/determinism_ci.json")
    return parser.parse_args()


def load_request(path: str) -> Dict:
    with open(path, "r", encoding="utf-8") as fh:
        payload = json.load(fh)
    req = payload.get("request")
    if not isinstance(req, dict):
        raise ValueError("baseline missing 'request'")
    return req


def call_recommendations(session: requests.Session, base_url: str, org_id: str, payload: Dict) -> List[str]:
    url = f"{base_url.rstrip('/')}/v1/recommendations"
    resp = session.post(url, json=payload, headers={"X-Org-ID": org_id}, timeout=30)
    resp.raise_for_status()
    data = resp.json()
    return [item["item_id"] for item in data.get("items", [])]


def compute_rank_deltas(base_runs: List[List[str]], window: int) -> Dict[int, float]:
    if not base_runs:
        return {}
    first = base_runs[0][:window]
    deltas: Dict[int, float] = {}
    comparisons = max(1, len(base_runs) - 1)
    for idx in range(1, len(base_runs)):
        items = base_runs[idx][:window]
        for rank, expected in enumerate(first):
            if rank >= len(items):
                continue
            if items[rank] != expected:
                deltas[rank] = deltas.get(rank, 0) + 1
    for rank in deltas:
        deltas[rank] /= comparisons
    return deltas


def main() -> None:
    args = parse_args()
    payload = load_request(args.baseline)
    session = requests.Session()

    runs: List[List[str]] = []
    for _ in range(args.calls):
        runs.append(call_recommendations(session, args.base_url, args.org_id, payload))

    deltas = compute_rank_deltas(runs, payload.get("k", 20))
    worst_rank = max(deltas, key=deltas.get, default=None)
    worst_delta = deltas.get(worst_rank, 0.0)

    result = {
        "calls": args.calls,
        "max_rank_delta": worst_delta,
        "worst_rank": worst_rank,
        "threshold": args.max_rank_delta,
    }
    os.makedirs(os.path.dirname(args.output), exist_ok=True)
    with open(args.output, "w", encoding="utf-8") as fh:
        json.dump(result, fh, indent=2)

    if worst_delta > args.max_rank_delta:
        print(f"Determinism failed: rank {worst_rank} delta {worst_delta:.3f} > {args.max_rank_delta:.3f}")
        sys.exit(1)
    print(f"Determinism passed: max delta {worst_delta:.3f}")


if __name__ == "__main__":
    main()
