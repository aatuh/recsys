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
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Tuple

import requests
import urllib3
import sys

from env_utils import env_metadata

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


DEFAULT_BASE_URL = "https://api.pepe.local"
DEFAULT_NAMESPACE = "default"
DEFAULT_ORG_ID = "00000000-0000-0000-0000-000000000001"
DEFAULT_REQUEST_TIMEOUT = 30.0

EVENT_WEIGHTS = {
    0: 0.1,  # view
    1: 0.3,  # click
    2: 0.6,  # add-to-cart
    3: 1.0,  # purchase
    4: 0.2,  # custom
}

MAX_K = 20
INTRA_K = 10
DEFAULT_MIN_USERS = 100
DEFAULT_WARM_MIN_EVENTS = 50
DEFAULT_RERANK_QUERIES = 50
DEFAULT_RERANK_CANDIDATES = 200
DEFAULT_SIMILAR_HEAD = 50
DEFAULT_SIMILAR_TAIL = 50
EVIDENCE_DIR = Path("analysis/evidence")


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


@dataclass
class WarmCandidate:
    user_id: str
    segment: str
    train_items: set[str]
    train_event_count: int
    test_relevance: Dict[str, float]


@dataclass
class EvaluationContext:
    session: requests.Session
    catalog: Dict[str, Dict]
    users: Dict[str, Dict]
    popularity: Counter
    long_tail_threshold: float
    global_pop_ranking: List[str]
    warm_candidates: List[WarmCandidate]
    selected_candidates: List[WarmCandidate]
    split_timestamp: datetime


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Run Recsys quality evaluation.")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL)
    parser.add_argument("--namespace", default=DEFAULT_NAMESPACE)
    parser.add_argument("--org-id", default=DEFAULT_ORG_ID)
    parser.add_argument(
        "--env-file",
        default="api/.env",
        help="Path to the env file whose hash should be recorded for provenance (default: %(default)s).",
    )
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
    parser.add_argument(
        "--min-catalog-coverage",
        type=float,
        default=0.0,
        help="Minimum system catalog coverage required (default: disabled).",
    )
    parser.add_argument(
        "--min-long-tail-share",
        type=float,
        default=0.0,
        help="Minimum long-tail share required (default: disabled).",
    )
    parser.add_argument(
        "--min-users",
        type=int,
        default=DEFAULT_MIN_USERS,
        help="Minimum warm users to evaluate per namespace (default: %(default)s).",
    )
    parser.add_argument(
        "--warm-min-events",
        type=int,
        default=DEFAULT_WARM_MIN_EVENTS,
        help="Minimum historical interactions (pre-split) required to treat a user as warm (default: %(default)s).",
    )
    parser.add_argument(
        "--rerank-queries",
        type=int,
        default=DEFAULT_RERANK_QUERIES,
        help="Number of rerank queries to evaluate (default: %(default)s).",
    )
    parser.add_argument(
        "--rerank-candidates",
        type=int,
        default=DEFAULT_RERANK_CANDIDATES,
        help="Number of candidates to send to /v1/rerank for each query (default: %(default)s).",
    )
    parser.add_argument(
        "--similar-head",
        type=int,
        default=DEFAULT_SIMILAR_HEAD,
        help="Number of head items to sample for /similar evaluation (default: %(default)s).",
    )
    parser.add_argument(
        "--similar-tail",
        type=int,
        default=DEFAULT_SIMILAR_TAIL,
        help="Number of long-tail items to sample for /similar evaluation (default: %(default)s).",
    )
    parser.add_argument(
        "--results-dir",
        default="analysis/results",
        help="Directory for storing suite outputs (default: %(default)s).",
    )
    parser.add_argument(
        "--recommendation-dump",
        default="analysis/results/recommendation_dump.json",
        help="Path to write aggregated recommendation dumps for exposure analysis (set to blank to skip).",
    )
    parser.add_argument(
        "--dump-sample-limit",
        type=int,
        default=10,
        help="How many recommendation payloads to retain for the dump/exposure dashboard (0 = unlimited).",
    )
    parser.add_argument(
        "--request-timeout",
        type=float,
        default=DEFAULT_REQUEST_TIMEOUT,
        help="HTTP timeout (seconds) for API calls (default: %(default)s).",
    )
    return parser.parse_args()


# -----------------------------------------------------------------------------
# API helpers
# -----------------------------------------------------------------------------


def build_session(base_url: str, org_id: str, request_timeout: float = DEFAULT_REQUEST_TIMEOUT) -> requests.Session:
    session = requests.Session()
    session.verify = False
    session.headers.update(
        {
            "Content-Type": "application/json",
            "X-Org-ID": org_id,
        }
    )
    session.request_timeout = request_timeout
    session.base_url = base_url.rstrip("/")
    return session


def fetch_paginated(session: requests.Session, path: str, namespace: str, limit: int = 500) -> List[Dict]:
    records: List[Dict] = []
    offset = 0
    while True:
        params = {"namespace": namespace, "limit": limit, "offset": offset}
        response = session.get(f"{session.base_url}{path}", params=params, timeout=session.request_timeout)
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
# Helper utilities
# -----------------------------------------------------------------------------


def ensure_dir(path: Path) -> None:
    path.mkdir(parents=True, exist_ok=True)


def write_json(path: Path, payload: Any) -> None:
    ensure_dir(path.parent)
    with path.open("w", encoding="utf-8") as fh:
        json.dump(payload, fh, indent=2)


def build_warm_candidates(
    events_by_user: Dict[str, List[Dict]],
    users: Dict[str, Dict],
    split_ts: datetime,
    warm_min_events: int,
) -> List[WarmCandidate]:
    candidates: List[WarmCandidate] = []
    for user_id, user_events in events_by_user.items():
        train_items: set[str] = set()
        train_event_count = 0
        test_relevance: Dict[str, float] = defaultdict(float)
        for event in user_events:
            weight = EVENT_WEIGHTS.get(event["type"], 0.0)
            if event["ts"] <= split_ts:
                train_items.add(event["item_id"])
                train_event_count += 1
            elif weight > 0:
                test_relevance[event["item_id"]] += weight
        if not test_relevance:
            continue
        if train_event_count < warm_min_events:
            continue
        candidates.append(
            WarmCandidate(
                user_id=user_id,
                segment=users.get(user_id, {}).get("segment", "unknown"),
                train_items=train_items,
                train_event_count=train_event_count,
                test_relevance=dict(test_relevance),
            )
        )
    candidates.sort(key=lambda c: (-c.train_event_count, c.user_id))
    return candidates


def select_warm_candidates(
    candidates: List[WarmCandidate], min_users: int, limit_users: int
) -> List[WarmCandidate]:
    if len(candidates) < min_users:
        raise RuntimeError(
            f"Only {len(candidates)} warm users available; need at least {min_users}. "
            "Seed more data or adjust --warm-min-events."
        )
    if limit_users > 0:
        if limit_users < min_users:
            raise ValueError(
                f"--limit-users ({limit_users}) is lower than --min-users ({min_users}); "
                "increase the limit or lower the minimum."
            )
        return candidates[:limit_users]
    return candidates


def score_warm_users(
    session: requests.Session,
    namespace: str,
    candidates: List[WarmCandidate],
    catalog: Dict[str, Dict],
    popularity: Counter,
    long_tail_threshold: float,
    global_pop_ranking: List[str],
    sleep_ms: int,
    sample_limit: int,
) -> Tuple[List[UserMetrics], List[Dict]]:
    user_metrics: List[UserMetrics] = []
    sample_recommendations: List[Dict] = []

    for candidate in candidates:
        baseline_candidates = [
            item for item in global_pop_ranking if item not in candidate.train_items
        ][:MAX_K]
        payload = {
            "namespace": namespace,
            "user_id": candidate.user_id,
            "k": MAX_K,
            "include_reasons": True,
            "explain_level": "numeric",
        }
        response = session.post(f"{session.base_url}/v1/recommendations", json=payload, timeout=session.request_timeout)
        if response.status_code != 200:
            raise RuntimeError(
                f"Recommendations failed for {candidate.user_id}: "
                f"{response.status_code} {response.text}"
            )
        rec_payload = response.json()
        rec_items = [item["item_id"] for item in rec_payload.get("items", [])]

        if sample_limit <= 0 or len(sample_recommendations) < sample_limit:
            sample_recommendations.append(
                {
                    "user_id": candidate.user_id,
                    "segment": candidate.segment,
                    "train_items": sorted(candidate.train_items)[:10],
                    "test_relevance": candidate.test_relevance,
                    "recommendations": rec_payload,
                    "baseline": baseline_candidates,
                }
            )

        baseline_ndcg, baseline_recall, baseline_mrr = compute_metrics_for_user(
            baseline_candidates, candidate.test_relevance, k=MAX_K
        )
        system_ndcg, system_recall, system_mrr = compute_metrics_for_user(
            rec_items, candidate.test_relevance, k=MAX_K
        )

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
                user_id=candidate.user_id,
                segment=candidate.segment,
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

    return user_metrics, sample_recommendations


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


def build_rerank_candidate_items(
    candidate: WarmCandidate,
    global_pop_ranking: List[str],
    popularity: Counter,
    pool_size: int,
) -> List[Dict]:
    items: List[Dict] = []
    used: set[str] = set()
    for item in global_pop_ranking:
        if item in used or item in candidate.train_items:
            continue
        items.append({"item_id": item, "score": float(popularity.get(item, 0.0))})
        used.add(item)
        if len(items) >= pool_size:
            break
    for item, _ in sorted(candidate.test_relevance.items(), key=lambda kv: kv[1], reverse=True):
        if len(items) >= pool_size:
            break
        if item in used:
            continue
        items.append({"item_id": item, "score": 0.0})
        used.add(item)
    return items


def summarize_rerank_metrics(rows: List[Dict]) -> Dict:
    def avg(key: str) -> float:
        return statistics.mean(row[key] for row in rows) if rows else 0.0

    baseline = {
        "ndcg@10": avg("baseline_ndcg10"),
        "recall@20": avg("baseline_recall20"),
        "mrr@10": avg("baseline_mrr10"),
    }
    system = {
        "ndcg@10": avg("system_ndcg10"),
        "recall@20": avg("system_recall20"),
        "mrr@10": avg("system_mrr10"),
    }

    def lift(metric: str) -> float:
        base = baseline[metric]
        sys_val = system[metric]
        if base == 0:
            return float("inf") if sys_val > 0 else 0.0
        return (sys_val - base) / base

    return {
        "queries": len(rows),
        "baseline": baseline,
        "system": system,
        "lift": {
            "ndcg@10": lift("ndcg@10"),
            "recall@20": lift("recall@20"),
            "mrr@10": lift("mrr@10"),
        },
    }


def evaluate_rerank_suite(
    session: requests.Session,
    namespace: str,
    candidates: List[WarmCandidate],
    global_pop_ranking: List[str],
    popularity: Counter,
    sleep_ms: int,
    query_limit: int,
    pool_size: int,
) -> Tuple[Optional[Dict], List[Dict]]:
    rows: List[Dict] = []
    samples: List[Dict] = []
    for candidate in candidates:
        if len(rows) >= query_limit:
            break
        items = build_rerank_candidate_items(candidate, global_pop_ranking, popularity, pool_size)
        if len(items) < MAX_K:
            continue
        baseline_order = [
            entry["item_id"]
            for entry in sorted(items, key=lambda e: (-e.get("score", 0.0), e["item_id"]))
        ]
        payload = {
            "namespace": namespace,
            "user_id": candidate.user_id,
            "k": MAX_K,
            "items": items,
            "include_reasons": False,
        }
        response = session.post(f"{session.base_url}/v1/rerank", json=payload, timeout=session.request_timeout)
        if response.status_code != 200:
            raise RuntimeError(
                f"Rerank failed for {candidate.user_id}: "
                f"{response.status_code} {response.text}"
            )
        rerank_payload = response.json()
        rerank_items = [item["item_id"] for item in rerank_payload.get("items", [])]
        baseline_ndcg, baseline_recall, baseline_mrr = compute_metrics_for_user(
            baseline_order, candidate.test_relevance, k=MAX_K
        )
        system_ndcg, system_recall, system_mrr = compute_metrics_for_user(
            rerank_items, candidate.test_relevance, k=MAX_K
        )
        rows.append(
            {
                "user_id": candidate.user_id,
                "baseline_ndcg10": baseline_ndcg,
                "baseline_recall20": baseline_recall,
                "baseline_mrr10": baseline_mrr,
                "system_ndcg10": system_ndcg,
                "system_recall20": system_recall,
                "system_mrr10": system_mrr,
            }
        )
        if len(samples) < 10:
            samples.append(
                {
                    "user_id": candidate.user_id,
                    "baseline_order": baseline_order[:MAX_K],
                    "rerank": rerank_payload,
                    "test_relevance": candidate.test_relevance,
                }
            )
        time.sleep(sleep_ms / 1000.0)

    if not rows:
        return None, samples
    return summarize_rerank_metrics(rows), samples


def select_seed_items(
    catalog: Dict[str, Dict],
    popularity: Counter,
    head_count: int,
    tail_count: int,
) -> List[Tuple[str, str]]:
    scored_items = sorted(
        catalog.keys(),
        key=lambda item: (popularity.get(item, 0.0), item),
        reverse=True,
    )
    head = []
    seen = set()
    for item in scored_items:
        head.append(item)
        seen.add(item)
        if len(head) >= head_count:
            break
    tail = []
    for item in reversed(scored_items):
        if item in seen:
            continue
        tail.append(item)
        if len(tail) >= tail_count:
            break
    seeds = [(item, "head") for item in head]
    seeds.extend((item, "tail") for item in tail)
    return seeds


def evaluate_similar_suite(
    session: requests.Session,
    namespace: str,
    catalog: Dict[str, Dict],
    popularity: Counter,
    seeds: List[Tuple[str, str]],
    long_tail_threshold: float,
) -> Tuple[Optional[Dict], List[Dict]]:
    seed_results: List[Dict] = []
    for seed_id, seed_type in seeds:
        response = session.get(
            f"{session.base_url}/v1/items/{seed_id}/similar",
            params={"namespace": namespace, "k": MAX_K},
            timeout=session.request_timeout,
        )
        if response.status_code != 200:
            continue
        items = response.json()
        if not isinstance(items, list):
            continue
        seed_tags = catalog.get(seed_id, {}).get("tags", set())
        seed_category = catalog.get(seed_id, {}).get("category")
        seed_brand = catalog.get(seed_id, {}).get("brand")
        j_scores = []
        same_category = 0
        same_brand = 0
        long_tail_hits = 0
        for rec in items:
            rec_id = rec.get("item_id")
            if not rec_id:
                continue
            rec_tags = catalog.get(rec_id, {}).get("tags", set())
            union = seed_tags | rec_tags
            if union:
                j_scores.append(len(seed_tags & rec_tags) / len(union))
            else:
                j_scores.append(0.0)
            if seed_category and catalog.get(rec_id, {}).get("category") == seed_category:
                same_category += 1
            if seed_brand and catalog.get(rec_id, {}).get("brand") == seed_brand:
                same_brand += 1
            if popularity.get(rec_id, 0.0) <= long_tail_threshold:
                long_tail_hits += 1
        total_items = max(len(items), 1)
        seed_results.append(
            {
                "seed": seed_id,
                "type": seed_type,
                "returned": len(items),
                "avg_jaccard": statistics.mean(j_scores) if j_scores else 0.0,
                "same_category_ratio": same_category / total_items,
                "same_brand_ratio": same_brand / total_items,
                "long_tail_ratio": long_tail_hits / total_items,
                "items": items[:MAX_K],
            }
        )

    if not seed_results:
        return None, seed_results

    overall = {
        "avg_jaccard": statistics.mean(seed["avg_jaccard"] for seed in seed_results),
        "same_category_ratio": statistics.mean(seed["same_category_ratio"] for seed in seed_results),
        "same_brand_ratio": statistics.mean(seed["same_brand_ratio"] for seed in seed_results),
        "long_tail_ratio": statistics.mean(seed["long_tail_ratio"] for seed in seed_results),
        "evaluated_seeds": len(seed_results),
    }
    return overall, seed_results


# -----------------------------------------------------------------------------
# Evaluation workflow
# -----------------------------------------------------------------------------


def evaluate_warm_users(
    base_url: str,
    namespace: str,
    org_id: str,
    limit_users: int,
    sleep_ms: int,
    min_users: int,
    warm_min_events: int,
    env_meta: Optional[Dict[str, str]] = None,
    sample_limit: int = 10,
    request_timeout: float = DEFAULT_REQUEST_TIMEOUT,
) -> Tuple[EvaluationContext, Dict, List[Dict]]:
    session = build_session(base_url, org_id, request_timeout=request_timeout)

    catalog = load_catalog(session, namespace)
    users = load_users(session, namespace)
    events = load_events(session, namespace)

    split_ts = compute_split_timestamp(events, percentile=0.75)
    popularity = aggregate_popularity([e for e in events if e["ts"] <= split_ts])
    long_tail_threshold = derive_long_tail_threshold(popularity, tail_fraction=0.4)

    events_by_user = defaultdict(list)
    for event in events:
        events_by_user[event["user_id"]].append(event)

    global_pop_ranking = [item for item, _ in popularity.most_common()]

    warm_candidates = build_warm_candidates(events_by_user, users, split_ts, warm_min_events)
    selected_candidates = select_warm_candidates(warm_candidates, min_users, limit_users)
    user_metrics, sample_recommendations = score_warm_users(
        session=session,
        namespace=namespace,
        candidates=selected_candidates,
        catalog=catalog,
        popularity=popularity,
        long_tail_threshold=long_tail_threshold,
        global_pop_ranking=global_pop_ranking,
        sleep_ms=sleep_ms,
        sample_limit=sample_limit,
    )

    if not user_metrics:
        raise RuntimeError("No evaluable warm users with hold-out interactions were found.")

    results = summarize_metrics(user_metrics, catalog, popularity, MAX_K, long_tail_threshold)
    results["meta"] = {
        "base_url": base_url,
        "namespace": namespace,
        "org_id": org_id,
        "evaluated_users": len(user_metrics),
        "split_timestamp": split_ts.isoformat(),
        "long_tail_threshold": long_tail_threshold,
        "min_users": min_users,
        "warm_min_events": warm_min_events,
    }
    if env_meta:
        results["meta"].update(env_meta)

    context = EvaluationContext(
        session=session,
        catalog=catalog,
        users=users,
        popularity=popularity,
        long_tail_threshold=long_tail_threshold,
        global_pop_ranking=global_pop_ranking,
        warm_candidates=warm_candidates,
        selected_candidates=selected_candidates,
        split_timestamp=split_ts,
    )

    return context, results, sample_recommendations


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


def compute_ratio(counter: Counter) -> float:
    if not counter:
        return 0.0
    values = list(counter.values())
    if not values:
        return 0.0
    mean_value = sum(values) / len(values)
    if mean_value <= 0:
        return 0.0
    return max(values) / mean_value


def build_recommendation_dump_payload(
    namespace: str,
    samples: List[Dict],
    catalog: Dict[str, Dict],
    base_url: str,
    org_id: str,
    env_meta: Optional[Dict[str, str]],
) -> Optional[Dict]:
    brand_counts: Counter = Counter()
    category_counts: Counter = Counter()
    sample_entries: List[Dict[str, Any]] = []

    for sample in samples:
        rec_payload = sample.get("recommendations") or {}
        rec_items = rec_payload.get("items") or []
        if not rec_items:
            continue
        entry_items: List[Dict[str, Any]] = []
        for item in rec_items:
            item_id = item.get("item_id")
            catalog_entry = catalog.get(item_id, {})
            brand = catalog_entry.get("brand") or catalog_entry.get("props", {}).get("brand")
            category = catalog_entry.get("category")
            if brand:
                brand_counts[brand] += 1
            if category:
                category_counts[category] += 1
            entry_items.append(
                {
                    "item_id": item_id,
                    "score": item.get("score"),
                    "brand": brand,
                    "category": category,
                }
            )
        sample_entries.append(
            {
                "namespace": namespace,
                "user_id": sample.get("user_id"),
                "segment": sample.get("segment"),
                "items": entry_items,
            }
        )

    if not sample_entries:
        return None

    brand_ratio = compute_ratio(brand_counts)
    summary = {
        namespace: {
            "user_count": len(sample_entries),
            "brand_counts": dict(brand_counts),
            "category_counts": dict(category_counts),
            "max_brand_exposure": max(brand_counts.values()) if brand_counts else 0,
            "mean_brand_exposure": (sum(brand_counts.values()) / len(brand_counts)) if brand_counts else 0.0,
            "exposure_ratio": brand_ratio,
        }
    }

    payload = {
        "meta": {
            "base_url": base_url,
            "namespace": namespace,
            "org_id": org_id,
            "generated_at": datetime.now(timezone.utc).isoformat(),
            "sample_count": len(sample_entries),
            "env": env_meta or {},
        },
        "summary": summary,
        "samples": sample_entries,
    }
    return payload


def main() -> None:
    args = parse_args()
    env_meta = env_metadata(args.env_file)
    context, warm_results, warm_samples = evaluate_warm_users(
        base_url=args.base_url,
        namespace=args.namespace,
        org_id=args.org_id,
        limit_users=args.limit_users,
        sleep_ms=args.sleep_ms,
        min_users=args.min_users,
        warm_min_events=args.warm_min_events,
        env_meta=env_meta,
        sample_limit=args.dump_sample_limit,
        request_timeout=args.request_timeout,
    )
    quality_path = Path("analysis/quality_metrics.json")
    write_json(quality_path, warm_results)
    write_json(EVIDENCE_DIR / "recommendation_samples_after_seed.json", warm_samples)
    results_dir = Path(args.results_dir)
    write_json(results_dir / f"{args.namespace}_warm_quality.json", warm_results)
    if args.recommendation_dump:
        dump_payload = build_recommendation_dump_payload(
            namespace=args.namespace,
            samples=warm_samples,
            catalog=context.catalog,
            base_url=args.base_url,
            org_id=args.org_id,
            env_meta=env_meta,
        )
        if dump_payload:
            write_json(Path(args.recommendation_dump), dump_payload)
        else:
            print(f"No sample recommendations captured; skipping dump write to {args.recommendation_dump}")

    rerank_summary, rerank_samples = evaluate_rerank_suite(
        session=context.session,
        namespace=args.namespace,
        candidates=context.selected_candidates,
        global_pop_ranking=context.global_pop_ranking,
        popularity=context.popularity,
        sleep_ms=args.sleep_ms,
        query_limit=args.rerank_queries,
        pool_size=args.rerank_candidates,
    )
    if rerank_summary:
        rerank_summary["meta"] = {
            "base_url": args.base_url,
            "namespace": args.namespace,
            "org_id": args.org_id,
            "queries": rerank_summary["queries"],
            "candidate_pool": args.rerank_candidates,
        }
        write_json(results_dir / f"{args.namespace}_rerank.json", rerank_summary)
    if rerank_samples:
        write_json(EVIDENCE_DIR / "rerank_samples.json", rerank_samples)

    seeds = select_seed_items(
        context.catalog,
        context.popularity,
        args.similar_head,
        args.similar_tail,
    )
    similar_summary, similar_details = evaluate_similar_suite(
        session=context.session,
        namespace=args.namespace,
        catalog=context.catalog,
        popularity=context.popularity,
        seeds=seeds,
        long_tail_threshold=context.long_tail_threshold,
    )
    if similar_summary:
        similar_summary["meta"] = {
            "base_url": args.base_url,
            "namespace": args.namespace,
            "org_id": args.org_id,
            "head_seeds": args.similar_head,
            "tail_seeds": args.similar_tail,
        }
        write_json(results_dir / f"{args.namespace}_similar.json", similar_summary)
    if similar_details:
        write_json(EVIDENCE_DIR / "similar_samples.json", similar_details)

    print(json.dumps(warm_results, indent=2))
    failing_segments = []
    for segment, metrics in warm_results.get("segments", {}).items():
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

    coverage = warm_results.get("coverage", {})
    overall_system = warm_results.get("overall", {}).get("system", {})
    coverage_failures = []
    if args.min_catalog_coverage > 0:
        system_cov = coverage.get("system_catalog_coverage", 0.0)
        if system_cov < args.min_catalog_coverage:
            coverage_failures.append(
                {
                    "metric": "system_catalog_coverage",
                    "value": system_cov,
                    "threshold": args.min_catalog_coverage,
                }
            )
    if args.min_long_tail_share > 0:
        system_lts = overall_system.get("long_tail_share@20", 0.0)
        if system_lts < args.min_long_tail_share:
            coverage_failures.append(
                {
                    "metric": "system_long_tail_share",
                    "value": system_lts,
                    "threshold": args.min_long_tail_share,
                }
            )
    if coverage_failures:
        sys.stderr.write(
            "Coverage guardrail failure detected: "
            + json.dumps(coverage_failures, indent=2)
            + "\n"
        )
        sys.exit(1)


if __name__ == "__main__":
    main()
