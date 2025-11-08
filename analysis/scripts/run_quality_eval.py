#!/usr/bin/env python3
"""
Run offline quality evaluation against live Recsys API.

Outputs metrics into analysis/quality_metrics.json and captures sample
recommendation payloads for evidence review.
"""

from __future__ import annotations

import argparse
import json
import math
import statistics
import time
from collections import Counter, defaultdict
from dataclasses import dataclass
from datetime import datetime, timezone
from itertools import combinations
from typing import Dict, Iterable, List, Optional, Tuple

import requests
import urllib3
import sys

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


DEFAULT_BASE_URL = "https://api.pepe.local"
DEFAULT_NAMESPACE = "default"
DEFAULT_ORG_ID = "00000000-0000-0000-0000-000000000001"

EVENT_WEIGHTS = {
    0: 0.1,  # view
    1: 0.3,  # click
    2: 0.6,  # add-to-cart
    3: 1.0,  # purchase
    4: 0.2,  # custom
}

MAX_K = 20
INTRA_K = 10


@dataclass
class UserMetrics:
    user_id: str
    segment: str
    baseline_ndcg10: float
    baseline_recall20: float
    baseline_mrr10: float
    system_ndcg10: float
    system_recall20: float
    system_mrr10: float
    baseline_long_tail_share: float
    system_long_tail_share: float
    baseline_coverage_items: List[str]
    system_coverage_items: List[str]
    baseline_intra_sim: float
    system_intra_sim: float
    baseline_novelty: float
    system_novelty: float


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Run Recsys quality evaluation.")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL)
    parser.add_argument("--namespace", default=DEFAULT_NAMESPACE)
    parser.add_argument("--org-id", default=DEFAULT_ORG_ID)
    parser.add_argument("--limit-users", type=int, default=0, help="Optional cap for evaluation users (0 = no cap).")
    parser.add_argument("--sleep-ms", type=int, default=120, help="Sleep between recommendation calls to avoid throttling.")
    parser.add_argument(
        "--min-segment-lift-ndcg",
        type=float,
        default=0.1,
        help="Minimum NDCG lift required for each segment (default: %(default)s = +10%%).",
    )
    parser.add_argument(
        "--min-segment-lift-mrr",
        type=float,
        default=0.1,
        help="Minimum MRR lift required for each segment (default: %(default)s = +10%%).",
    )
    return parser.parse_args()


# -----------------------------------------------------------------------------
# API helpers
# -----------------------------------------------------------------------------


def build_session(base_url: str, org_id: str) -> requests.Session:
    session = requests.Session()
    session.verify = False
    session.headers.update(
        {
            "Content-Type": "application/json",
            "X-Org-ID": org_id,
        }
    )
    session.base_url = base_url.rstrip("/")
    return session


def fetch_paginated(session: requests.Session, path: str, namespace: str, limit: int = 500) -> List[Dict]:
    records: List[Dict] = []
    offset = 0
    while True:
        params = {"namespace": namespace, "limit": limit, "offset": offset}
        response = session.get(f"{session.base_url}{path}", params=params, timeout=30)
        if response.status_code != 200:
            raise RuntimeError(f"GET {path} failed: {response.status_code} {response.text}")
        payload = response.json()
        items = payload.get("items", [])
        records.extend(items)
        if not payload.get("has_more"):
            break
        next_offset = payload.get("next_offset")
        if next_offset is None:
            offset += len(items)
        else:
            offset = next_offset
    return records


# -----------------------------------------------------------------------------
# Data preparation
# -----------------------------------------------------------------------------


def parse_iso(ts: str) -> datetime:
    return datetime.fromisoformat(ts.replace("Z", "+00:00")).astimezone(timezone.utc)


def load_catalog(session: requests.Session, namespace: str) -> Dict[str, Dict]:
    items = fetch_paginated(session, "/v1/items", namespace, limit=500)
    parsed: Dict[str, Dict] = {}
    for item in items:
        props_raw = item.get("props")
        if isinstance(props_raw, str) and props_raw:
            try:
                props = json.loads(props_raw)
            except json.JSONDecodeError:
                props = {}
        elif isinstance(props_raw, dict):
            props = props_raw
        else:
            props = {}
        tags = item.get("tags") or []
        parsed[item["item_id"]] = {
            "category": item.get("category"),
            "brand": item.get("brand"),
            "tags": set(tags),
            "price": item.get("price"),
            "props": props,
            "available": item.get("available", True),
        }
    return parsed


def load_users(session: requests.Session, namespace: str) -> Dict[str, Dict]:
    users = fetch_paginated(session, "/v1/users", namespace, limit=200)
    parsed: Dict[str, Dict] = {}
    for user in users:
        traits_raw = user.get("traits")
        if isinstance(traits_raw, str) and traits_raw:
            try:
                traits = json.loads(traits_raw)
            except json.JSONDecodeError:
                traits = {}
        elif isinstance(traits_raw, dict):
            traits = traits_raw
        else:
            traits = {}
        parsed[user["user_id"]] = traits
    return parsed


def load_events(session: requests.Session, namespace: str) -> List[Dict]:
    events = fetch_paginated(session, "/v1/events", namespace, limit=1000)
    parsed: List[Dict] = []
    for event in events:
        meta_raw = event.get("meta")
        if isinstance(meta_raw, str) and meta_raw:
            try:
                meta = json.loads(meta_raw)
            except json.JSONDecodeError:
                meta = {}
        elif isinstance(meta_raw, dict):
            meta = meta_raw
        else:
            meta = {}
        parsed.append(
            {
                "user_id": event["user_id"],
                "item_id": event["item_id"],
                "type": event.get("type", 0),
                "ts": parse_iso(event["ts"]),
                "value": event.get("value", 1),
                "meta": meta,
            }
        )
    parsed.sort(key=lambda e: e["ts"])
    return parsed


# -----------------------------------------------------------------------------
# Metrics
# -----------------------------------------------------------------------------


def compute_split_timestamp(events: List[Dict], percentile: float = 0.75) -> datetime:
    if not events:
        raise ValueError("No events found for split calculation.")
    idx = int(len(events) * percentile)
    idx = min(idx, len(events) - 1)
    return events[idx]["ts"]


def aggregate_popularity(events: List[Dict]) -> Counter:
    popularity = Counter()
    for event in events:
        weight = EVENT_WEIGHTS.get(event["type"], 0.1)
        popularity[event["item_id"]] += weight
    return popularity


def derive_long_tail_threshold(popularity: Counter, tail_fraction: float = 0.4) -> float:
    if not popularity:
        return 0.0
    ranked = sorted(popularity.values())
    idx = int(len(ranked) * (1 - tail_fraction))
    idx = min(idx, len(ranked) - 1)
    return ranked[idx]


def dcg(scores: List[float]) -> float:
    return sum((2 ** rel - 1) / math.log2(idx + 2) for idx, rel in enumerate(scores))


def compute_metrics_for_user(rankings: List[str], relevance: Dict[str, float], k: int) -> Tuple[float, float, float]:
    gains = []
    hits = 0
    mrr = 0.0
    total_relevant = len(relevance)
    for idx, item_id in enumerate(rankings[:k]):
        rel = relevance.get(item_id, 0.0)
        gains.append(rel)
        if rel > 0:
            hits += 1
            if mrr == 0.0:
                mrr = 1.0 / (idx + 1)
    dcg_k = dcg(gains)
    ideal_gains = sorted(relevance.values(), reverse=True)[:k]
    idcg_k = dcg(ideal_gains)
    ndcg = dcg_k / idcg_k if idcg_k > 0 else 0.0
    recall = hits / total_relevant if total_relevant > 0 else 0.0
    return ndcg, recall, mrr


def intra_list_similarity(rankings: List[str], catalog: Dict[str, Dict], k: int = INTRA_K) -> float:
    subset = rankings[:k]
    if len(subset) < 2:
        return 0.0
    sims = []
    for a, b in combinations(subset, 2):
        tags_a = catalog.get(a, {}).get("tags", set())
        tags_b = catalog.get(b, {}).get("tags", set())
        if not tags_a and not tags_b:
            sim = 0.0
        else:
            union = tags_a | tags_b
            inter = tags_a & tags_b
            sim = len(inter) / len(union) if union else 0.0
        sims.append(sim)
    return statistics.mean(sims) if sims else 0.0


def average_novelty(rankings: List[str], catalog: Dict[str, Dict], k: int = MAX_K) -> float:
    scores = []
    for item_id in rankings[:k]:
        novelty = catalog.get(item_id, {}).get("props", {}).get("novelty")
        if novelty is not None:
            scores.append(novelty)
    return statistics.mean(scores) if scores else 0.0


# -----------------------------------------------------------------------------
# Evaluation workflow
# -----------------------------------------------------------------------------


def evaluate(
    base_url: str,
    namespace: str,
    org_id: str,
    limit_users: int,
    sleep_ms: int,
) -> Dict:
    session = build_session(base_url, org_id)

    catalog = load_catalog(session, namespace)
    users = load_users(session, namespace)
    events = load_events(session, namespace)

    split_ts = compute_split_timestamp(events, percentile=0.75)
    popularity = aggregate_popularity([e for e in events if e["ts"] <= split_ts])
    long_tail_threshold = derive_long_tail_threshold(popularity, tail_fraction=0.4)

    events_by_user = defaultdict(list)
    for event in events:
        events_by_user[event["user_id"]].append(event)

    evaluation_users: List[str] = []
    user_metrics: List[UserMetrics] = []
    sample_recommendations: List[Dict] = []

    global_pop_ranking = [item for item, _ in popularity.most_common()]

    for user_id, user_events in events_by_user.items():
        train_items = set()
        test_relevance = defaultdict(float)
        for event in user_events:
            weight = EVENT_WEIGHTS.get(event["type"], 0.0)
            if event["ts"] <= split_ts:
                train_items.add(event["item_id"])
            else:
                if weight > 0:
                    test_relevance[event["item_id"]] += weight

        if not test_relevance:
            continue

        evaluation_users.append(user_id)
        if limit_users and len(evaluation_users) > limit_users:
            break

        baseline_candidates = [item for item in global_pop_ranking if item not in train_items][:MAX_K]

        payload = {
            "namespace": namespace,
            "user_id": user_id,
            "k": MAX_K,
            "include_reasons": True,
            "explain_level": "numeric",
        }
        response = session.post(f"{session.base_url}/v1/recommendations", json=payload, timeout=30)
        if response.status_code != 200:
            raise RuntimeError(f"Recommendations failed for {user_id}: {response.status_code} {response.text}")
        rec_payload = response.json()
        rec_items = [item["item_id"] for item in rec_payload.get("items", [])]

        if len(sample_recommendations) < 10:
            sample_recommendations.append(
                {
                    "user_id": user_id,
                    "segment": users.get(user_id, {}).get("segment"),
                    "train_items": sorted(train_items)[:10],
                    "test_relevance": test_relevance,
                    "recommendations": rec_payload,
                    "baseline": baseline_candidates,
                }
            )

        segment = users.get(user_id, {}).get("segment", "unknown")

        baseline_ndcg, baseline_recall, baseline_mrr = compute_metrics_for_user(
            baseline_candidates, test_relevance, k=MAX_K
        )
        system_ndcg, system_recall, system_mrr = compute_metrics_for_user(rec_items, test_relevance, k=MAX_K)

        def long_tail_share(rankings: List[str]) -> float:
            if not rankings:
                return 0.0
            tail_hits = 0
            for item_id in rankings[:MAX_K]:
                if popularity.get(item_id, 0.0) <= long_tail_threshold:
                    tail_hits += 1
            return tail_hits / min(len(rankings), MAX_K)

        baseline_lts = long_tail_share(baseline_candidates)
        system_lts = long_tail_share(rec_items)

        baseline_sim = intra_list_similarity(baseline_candidates, catalog, INTRA_K)
        system_sim = intra_list_similarity(rec_items, catalog, INTRA_K)

        baseline_nov = average_novelty(baseline_candidates, catalog, MAX_K)
        system_nov = average_novelty(rec_items, catalog, MAX_K)

        user_metrics.append(
            UserMetrics(
                user_id=user_id,
                segment=segment,
                baseline_ndcg10=baseline_ndcg,
                baseline_recall20=baseline_recall,
                baseline_mrr10=baseline_mrr,
                system_ndcg10=system_ndcg,
                system_recall20=system_recall,
                system_mrr10=system_mrr,
                baseline_long_tail_share=baseline_lts,
                system_long_tail_share=system_lts,
                baseline_coverage_items=baseline_candidates[:MAX_K],
                system_coverage_items=rec_items[:MAX_K],
                baseline_intra_sim=baseline_sim,
                system_intra_sim=system_sim,
                baseline_novelty=baseline_nov,
                system_novelty=system_nov,
            )
        )

        time.sleep(sleep_ms / 1000.0)

    if not user_metrics:
        raise RuntimeError("No evaluable users with hold-out interactions were found.")

    results = summarize_metrics(user_metrics, catalog, popularity, MAX_K, long_tail_threshold)
    results["meta"] = {
        "base_url": base_url,
        "namespace": namespace,
        "org_id": org_id,
        "evaluated_users": len(user_metrics),
        "split_timestamp": split_ts.isoformat(),
        "long_tail_threshold": long_tail_threshold,
    }

    with open("analysis/quality_metrics.json", "w", encoding="utf-8") as fh:
        json.dump(results, fh, indent=2)

    with open("analysis/evidence/recommendation_samples_after_seed.json", "w", encoding="utf-8") as fh:
        json.dump(sample_recommendations, fh, indent=2)

    return results


def summarize_metrics(
    user_metrics: List[UserMetrics],
    catalog: Dict[str, Dict],
    popularity: Counter,
    k: int,
    long_tail_threshold: float,
) -> Dict:
    def avg(values: Iterable[float]) -> float:
        values = list(values)
        return float(sum(values) / len(values)) if values else 0.0

    def lift(system: float, baseline: float) -> float:
        if baseline == 0:
            return float("inf") if system > 0 else 0.0
        return (system - baseline) / baseline

    overall = {
        "baseline": {
            "ndcg@10": avg(m.baseline_ndcg10 for m in user_metrics),
            "recall@20": avg(m.baseline_recall20 for m in user_metrics),
            "mrr@10": avg(m.baseline_mrr10 for m in user_metrics),
            "long_tail_share@20": avg(m.baseline_long_tail_share for m in user_metrics),
            "intra_list_similarity@10": avg(m.baseline_intra_sim for m in user_metrics),
            "novelty@20": avg(m.baseline_novelty for m in user_metrics),
        },
        "system": {
            "ndcg@10": avg(m.system_ndcg10 for m in user_metrics),
            "recall@20": avg(m.system_recall20 for m in user_metrics),
            "mrr@10": avg(m.system_mrr10 for m in user_metrics),
            "long_tail_share@20": avg(m.system_long_tail_share for m in user_metrics),
            "intra_list_similarity@10": avg(m.system_intra_sim for m in user_metrics),
            "novelty@20": avg(m.system_novelty for m in user_metrics),
        },
    }
    overall["lift"] = {
        "ndcg@10": lift(overall["system"]["ndcg@10"], overall["baseline"]["ndcg@10"]),
        "recall@20": lift(overall["system"]["recall@20"], overall["baseline"]["recall@20"]),
        "mrr@10": lift(overall["system"]["mrr@10"], overall["baseline"]["mrr@10"]),
    }

    segments: Dict[str, List[UserMetrics]] = defaultdict(list)
    for m in user_metrics:
        segments[m.segment].append(m)

    segment_metrics = {}
    for segment, metrics in segments.items():
        segment_metrics[segment] = {
            "users": len(metrics),
            "baseline": {
                "ndcg@10": avg(m.baseline_ndcg10 for m in metrics),
                "recall@20": avg(m.baseline_recall20 for m in metrics),
                "mrr@10": avg(m.baseline_mrr10 for m in metrics),
                "long_tail_share@20": avg(m.baseline_long_tail_share for m in metrics),
                "intra_list_similarity@10": avg(m.baseline_intra_sim for m in metrics),
            },
            "system": {
                "ndcg@10": avg(m.system_ndcg10 for m in metrics),
                "recall@20": avg(m.system_recall20 for m in metrics),
                "mrr@10": avg(m.system_mrr10 for m in metrics),
                "long_tail_share@20": avg(m.system_long_tail_share for m in metrics),
                "intra_list_similarity@10": avg(m.system_intra_sim for m in metrics),
            },
        }
        segment_metrics[segment]["lift"] = {
            "ndcg@10": lift(
                segment_metrics[segment]["system"]["ndcg@10"],
                segment_metrics[segment]["baseline"]["ndcg@10"],
            ),
            "recall@20": lift(
                segment_metrics[segment]["system"]["recall@20"],
                segment_metrics[segment]["baseline"]["recall@20"],
            ),
            "mrr@10": lift(
                segment_metrics[segment]["system"]["mrr@10"],
                segment_metrics[segment]["baseline"]["mrr@10"],
            ),
        }

    coverage_system = Counter()
    coverage_baseline = Counter()
    for m in user_metrics:
        coverage_system.update(m.system_coverage_items)
        coverage_baseline.update(m.baseline_coverage_items)

    total_catalog = len(catalog)
    coverage = {
        "system_unique_items": len(coverage_system),
        "baseline_unique_items": len(coverage_baseline),
        "catalog_size": total_catalog,
        "system_catalog_coverage": len(coverage_system) / total_catalog if total_catalog else 0.0,
        "baseline_catalog_coverage": len(coverage_baseline) / total_catalog if total_catalog else 0.0,
        "system_long_tail_unique": sum(
            1 for item in coverage_system if popularity.get(item, 0.0) <= long_tail_threshold
        ),
    }

    return {
        "overall": overall,
        "segments": segment_metrics,
        "coverage": coverage,
        "user_count": len(user_metrics),
    }


def main() -> None:
    args = parse_args()
    results = evaluate(
        base_url=args.base_url,
        namespace=args.namespace,
        org_id=args.org_id,
        limit_users=args.limit_users,
        sleep_ms=args.sleep_ms,
    )
    print(json.dumps(results, indent=2))
    failing_segments = []
    for segment, metrics in results.get("segments", {}).items():
        ndcg_lift = metrics.get("lift", {}).get("ndcg@10")
        mrr_lift = metrics.get("lift", {}).get("mrr@10")
        if ndcg_lift is None or mrr_lift is None:
            continue
        if ndcg_lift < args.min_segment_lift_ndcg or mrr_lift < args.min_segment_lift_mrr:
            failing_segments.append(
                {
                    "segment": segment,
                    "ndcg_lift": ndcg_lift,
                    "mrr_lift": mrr_lift,
                    "threshold_ndcg": args.min_segment_lift_ndcg,
                    "threshold_mrr": args.min_segment_lift_mrr,
                }
            )
    if failing_segments:
        sys.stderr.write(
            "Segment guardrail failure detected: "
            + json.dumps(failing_segments, indent=2)
            + "\n"
        )
        sys.exit(1)


if __name__ == "__main__":
    main()
