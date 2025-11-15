#!/usr/bin/env python3
"""Profile candidate generation and pruning stages for coverage analysis."""

from __future__ import annotations

import argparse
import json
import os
import random
import statistics
import sys
from collections import Counter, defaultdict
from datetime import datetime
from typing import Dict, List

sys.path.append(os.path.dirname(__file__))

from run_quality_eval import build_session, load_catalog, load_events, load_users  # type: ignore

DEFAULT_BASE_URL = os.getenv("SCENARIOS_BASE_URL", "https://api.pepe.local")
DEFAULT_NAMESPACE = os.getenv("SCENARIOS_NAMESPACE", "default")
DEFAULT_ORG_ID = os.getenv("SCENARIOS_ORG_ID", "00000000-0000-0000-0000-000000000001")

EVIDENCE_PATH = "analysis/evidence/coverage_profile.json"
SAMPLE_LIMIT = 120
K = 20


def recommend(session, namespace: str, user_id: str) -> Dict:
    payload = {"namespace": namespace, "user_id": user_id, "k": K}
    url = f"{session.base_url}/v1/recommendations"
    resp = session.post(url, json=payload, timeout=30)
    if resp.status_code != 200:
        raise RuntimeError(f"Recommendation failed ({resp.status_code}): {resp.text}")
    return resp.json()


def extract_source_counts(resp: Dict) -> Dict[str, int]:
    trace = resp.get("trace")
    if not isinstance(trace, dict):
        return {}
    extras = trace.get("extras")
    if not isinstance(extras, dict):
        return {}
    sources = extras.get("candidate_sources")
    if isinstance(sources, dict):
        return {k: int(v.get("count", 0)) if isinstance(v, dict) else 0 for k, v in sources.items()}
    return {}


def extract_policy_counts(resp: Dict) -> Dict[str, int]:
    trace = resp.get("trace")
    if not isinstance(trace, dict):
        return {}
    extras = trace.get("extras")
    if not isinstance(extras, dict):
        return {}
    policy = extras.get("policy")
    if isinstance(policy, dict):
        return {
            "total_candidates": int(policy.get("total_candidates", 0)),
            "after_exclusions": int(policy.get("after_exclusions", 0)),
            "after_rules": int(policy.get("after_rules", 0)),
            "final_count": int(policy.get("final_count", 0)),
        }
    return {}


def profile_coverage(base_url: str, namespace: str, org_id: str) -> Dict:
    session = build_session(base_url, org_id)
    catalog = load_catalog(session, namespace)
    users = load_users(session, namespace)
    events = load_events(session, namespace)

    events_by_user: Dict[str, List[Dict]] = defaultdict(list)
    for event in events:
        events_by_user[event["user_id"]].append(event)

    user_ids = list(events_by_user.keys())
    random.shuffle(user_ids)
    sample_ids = user_ids[:SAMPLE_LIMIT]

    unique_items = set()
    source_counter = Counter()
    policy_counter = Counter()
    final_lengths: List[int] = []

    per_user: List[Dict] = []

    for user_id in sample_ids:
        resp = recommend(session, namespace, user_id)
        items = [item.get("item_id") for item in resp.get("items", []) if item.get("item_id")]
        final_lengths.append(len(items))
        unique_items.update(items)

        sources = extract_source_counts(resp)
        policy = extract_policy_counts(resp)
        source_counter.update(sources)
        policy_counter.update(policy)

        per_user.append(
            {
                "user_id": user_id,
                "candidate_sources": sources,
                "policy_counts": policy,
                "final_items": items,
            }
        )

    coverage_ratio = len(unique_items) / len(catalog) if catalog else 0.0
    summary = {
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "base_url": base_url,
        "namespace": namespace,
        "sampled_users": len(sample_ids),
        "unique_items": len(unique_items),
        "catalog_size": len(catalog),
        "coverage_ratio": coverage_ratio,
        "avg_final_list_len": statistics.mean(final_lengths) if final_lengths else 0,
        "avg_candidates": {k: v / len(sample_ids) for k, v in source_counter.items()},
        "avg_policy_counts": {k: v / len(sample_ids) for k, v in policy_counter.items()},
        "users": per_user,
    }

    os.makedirs(os.path.dirname(EVIDENCE_PATH), exist_ok=True)
    with open(EVIDENCE_PATH, "w", encoding="utf-8") as fh:
        json.dump(summary, fh, indent=2)

    return summary


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Profile coverage across candidate stages.")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL)
    parser.add_argument("--namespace", default=DEFAULT_NAMESPACE)
    parser.add_argument("--org-id", default=DEFAULT_ORG_ID)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    summary = profile_coverage(args.base_url, args.namespace, args.org_id)
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
