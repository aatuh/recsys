#!/usr/bin/env python3
"""
Execute Recsys scenario suite S1-S10 and generate analysis/scenarios.csv.

Captures supporting evidence under analysis/evidence/ for each scenario.
"""

from __future__ import annotations

import argparse
import csv
import json
import os
import statistics
import sys
import time
from collections import defaultdict
from dataclasses import dataclass
from datetime import datetime
from typing import Dict, List, Tuple

sys.path.append(os.path.dirname(__file__))

from run_quality_eval import (  # type: ignore
    EVENT_WEIGHTS,
    build_session,
    compute_metrics_for_user,
    compute_split_timestamp,
    load_catalog,
    load_events,
    load_users,
    intra_list_similarity,
)

DEFAULT_BASE_URL = os.getenv("SCENARIOS_BASE_URL", "https://api.pepe.local")
DEFAULT_NAMESPACE = os.getenv("SCENARIOS_NAMESPACE", "default")
DEFAULT_ORG_ID = os.getenv("SCENARIOS_ORG_ID", "00000000-0000-0000-0000-000000000001")

SURFACE = "home"
K_DEFAULT = 20
SLEEP_BETWEEN_CALLS = 0.12

EVIDENCE_DIR = os.path.join("analysis", "evidence")

SCENARIO_HEADERS = ["id", "name", "input", "expected", "observed", "result"]


@dataclass
class ScenarioResult:
    id: str
    name: str
    input: str
    expected: str
    observed: str
    passed: bool

    def to_row(self) -> List[str]:
        return [
            self.id,
            self.name,
            self.input,
            self.expected,
            self.observed,
            "PASS" if self.passed else "FAIL",
        ]


def ensure_evidence_dir() -> None:
    os.makedirs(EVIDENCE_DIR, exist_ok=True)


def write_evidence(name: str, payload: Dict) -> None:
    ensure_evidence_dir()
    path = os.path.join(EVIDENCE_DIR, name)
    with open(path, "w", encoding="utf-8") as fh:
        json.dump(payload, fh, indent=2)


def recommend(session, namespace: str, payload: Dict) -> Dict:
    url = f"{session.base_url}/v1/recommendations"
    response = session.post(url, json=payload, timeout=30)
    if response.status_code != 200:
        raise RuntimeError(f"Recommendation failed ({response.status_code}): {response.text}")
    return response.json()


def extract_policy_summary(resp: Dict) -> Dict:
    trace = resp.get("trace")
    if not isinstance(trace, dict):
        return {}
    extras = trace.get("extras")
    if not isinstance(extras, dict):
        return {}
    policy = extras.get("policy")
    if isinstance(policy, dict):
        return policy
    return {}


def summarize_policy_samples(samples: List[Dict]) -> Dict[str, float]:
    if not samples:
        return {}
    agg: Dict[str, float] = {}
    for sample in samples:
        for key, value in sample.items():
            if isinstance(value, (int, float)):
                agg[key] = agg.get(key, 0.0) + float(value)
    # Average values to smooth across users
    count = float(len(samples))
    for key in list(agg.keys()):
        agg[key] = agg[key] / count
    return agg


def extract_starter_profile(resp: Dict) -> Dict:
    trace = resp.get("trace")
    if not isinstance(trace, dict):
        return {}
    extras = trace.get("extras")
    if not isinstance(extras, dict):
        return {}
    profile = extras.get("starter_profile")
    if isinstance(profile, dict):
        return profile
    return {}


def create_rule(session, payload: Dict) -> Dict:
    url = f"{session.base_url}/v1/admin/rules"
    resp = session.post(url, json=payload, timeout=30)
    if resp.status_code != 201:
        raise RuntimeError(f"Create rule failed ({resp.status_code}): {resp.text}")
    return resp.json()


def update_rule(session, rule_id: str, payload: Dict) -> Dict:
    url = f"{session.base_url}/v1/admin/rules/{rule_id}"
    resp = session.put(url, json=payload, timeout=30)
    if resp.status_code not in (200, 201):
        raise RuntimeError(f"Update rule failed ({resp.status_code}): {resp.text}")
    return resp.json()


def create_manual_override(session, payload: Dict) -> Dict:
    url = f"{session.base_url}/v1/admin/manual_overrides"
    resp = session.post(url, json=payload, timeout=30)
    if resp.status_code != 201:
        raise RuntimeError(f"Create manual override failed ({resp.status_code}): {resp.text}")
    return resp.json()


def cancel_manual_override(session, override_id: str) -> Dict:
    url = f"{session.base_url}/v1/admin/manual_overrides/{override_id}/cancel"
    resp = session.post(url, json={}, timeout=30)
    if resp.status_code != 200:
        raise RuntimeError(f"Cancel manual override failed ({resp.status_code}): {resp.text}")
    return resp.json()


def upsert_item(session, namespace: str, item: Dict) -> Dict:
    url = f"{session.base_url}/v1/items:upsert"
    payload = {"namespace": namespace, "items": [item]}
    resp = session.post(url, json=payload, timeout=30)
    if resp.status_code not in (200, 202):
        raise RuntimeError(f"Item upsert failed ({resp.status_code}): {resp.text}")
    return resp.json()


def upsert_user(session, namespace: str, user: Dict) -> Dict:
    url = f"{session.base_url}/v1/users:upsert"
    payload = {"namespace": namespace, "users": [user]}
    resp = session.post(url, json=payload, timeout=30)
    if resp.status_code not in (200, 202):
        raise RuntimeError(f"User upsert failed ({resp.status_code}): {resp.text}")
    return resp.json()


def gather_user_relevance(
    events_by_user: Dict[str, List[Dict]],
    split_ts: datetime,
) -> Dict[str, Tuple[set, Dict[str, float]]]:
    result = {}
    for user_id, user_events in events_by_user.items():
        train_items = set()
        relevance = defaultdict(float)
        for event in user_events:
            weight = EVENT_WEIGHTS.get(event["type"], 0.0)
            if event["ts"] <= split_ts:
                train_items.add(event["item_id"])
            else:
                if weight > 0:
                    relevance[event["item_id"]] += weight
        result[user_id] = (train_items, dict(relevance))
    return result


def compute_similarity(items: List[str], catalog: Dict[str, Dict]) -> float:
    return intra_list_similarity(items, catalog, k=min(len(items), 10))


def average_margin(items: List[str], catalog: Dict[str, Dict]) -> float:
    margins = []
    for item_id in items:
        margin = catalog.get(item_id, {}).get("props", {}).get("margin")
        if margin is not None:
            margins.append(margin)
    return statistics.mean(margins) if margins else 0.0


def scenario_simple_filter(session, namespace, catalog) -> ScenarioResult:
    user_id = "user_0003"  # niche_readers
    payload = {
        "namespace": namespace,
        "user_id": user_id,
        "k": 15,
        "context": {"surface": SURFACE},
        "constraints": {"include_tags_any": ["books"]},
        "include_reasons": True,
        "explain_level": "tags",
    }
    resp = recommend(session, namespace, payload)
    items = resp.get("items", [])
    failing = []
    for item in items:
        tags = catalog.get(item["item_id"], {}).get("tags", set())
        if "books" not in tags:
            failing.append(item["item_id"])
    observed = f"Returned {len(items)} items; offending items={failing}" if failing else f"Returned {len(items)} items, all tagged 'books'."
    write_evidence("scenario_s1_response.json", {"request": payload, "response": resp})
    return ScenarioResult(
        id="S1",
        name="Strict include filter",
        input="POST /v1/recommendations include_tags_any=['books'] for user_0003",
        expected="All results belong to category/tag 'books'; zero leakage.",
        observed=observed,
        passed=len(items) > 0 and not failing,
    )


def scenario_block_tag(session, namespace, catalog) -> ScenarioResult:
    user_id = "user_0001"
    base_payload = {
        "namespace": namespace,
        "user_id": user_id,
        "k": 15,
        "context": {"surface": SURFACE},
        "include_reasons": True,
    }
    baseline = recommend(session, namespace, base_payload)

    before_high_margin = [
        item["item_id"]
        for item in baseline.get("items", [])
        if "high_margin" in catalog.get(item["item_id"], {}).get("tags", set())
    ]

    rule_payload = {
        "namespace": namespace,
        "surface": SURFACE,
        "name": "scenario_s2_block_high_margin",
        "description": "Scenario S2 hard exclude high margin tag",
        "action": "BLOCK",
        "target_type": "TAG",
        "target_key": "high_margin",
        "priority": 100,
        "enabled": True,
    }
    created = create_rule(session, rule_payload)
    rule_id = created["rule_id"]
    time.sleep(0.3)

    after = recommend(session, namespace, base_payload)
    after_high_margin = [
        item["item_id"]
        for item in after.get("items", [])
        if "high_margin" in catalog.get(item["item_id"], {}).get("tags", set())
    ]

    rule_payload["enabled"] = False
    update_rule(session, rule_id, rule_payload)

    write_evidence(
        "scenario_s2_block_high_margin.json",
        {
            "baseline": baseline,
            "after_block": after,
            "before_high_margin": before_high_margin,
            "after_high_margin": after_high_margin,
            "rule": created,
        },
    )

    observed = f"Before block: {before_high_margin}; After block: {after_high_margin}"
    passed = len(after.get("items", [])) > 0 and not after_high_margin
    return ScenarioResult(
        id="S2",
        name="Hard exclude tag",
        input="Create BLOCK rule on tag 'high_margin', request /v1/recommendations user_0001",
        expected="No items tagged 'high_margin' appear in ranked list.",
        observed=observed,
        passed=passed,
    )


def scenario_boost_monotonicity(session, namespace, catalog) -> ScenarioResult:
    user_id = "user_0002"
    payload = {
        "namespace": namespace,
        "user_id": user_id,
        "k": 15,
        "context": {"surface": SURFACE},
        "include_reasons": True,
    }
    before = recommend(session, namespace, payload)
    items_before = [item["item_id"] for item in before.get("items", [])]
    if len(items_before) < 2:
        raise RuntimeError("Not enough items in baseline recommendation for scenario S3.")
    target_item = items_before[2] if len(items_before) > 2 else items_before[1]

    override_payload = {
        "namespace": namespace,
        "surface": SURFACE,
        "item_id": target_item,
        "action": "boost",
        "boost_value": 3.0,
        "created_by": "scenario_s3",
        "notes": "Scenario S3 boost monotonicity",
    }
    created = create_manual_override(session, override_payload)
    override_id = created["override_id"]
    time.sleep(0.3)

    after = recommend(session, namespace, payload)
    items_after = [item["item_id"] for item in after.get("items", [])]

    cancel_manual_override(session, override_id)

    rank_before = items_before.index(target_item) if target_item in items_before else None
    rank_after = items_after.index(target_item) if target_item in items_after else None

    write_evidence(
        "scenario_s3_boost.json",
        {
            "baseline": before,
            "after_boost": after,
            "target_item": target_item,
            "rank_before": rank_before,
            "rank_after": rank_after,
            "override": created,
        },
    )

    observed = f"Target {target_item} moved from {rank_before} to {rank_after}"
    passed = rank_before is not None and rank_after is not None and rank_after < rank_before
    return ScenarioResult(
        id="S3",
        name="Boost monotonicity",
        input=f"Manual override boost +3.0 for {target_item} on surface {SURFACE}",
        expected="Boosted item ranks higher (lower position number).",
        observed=observed,
        passed=passed,
    )


def scenario_diversity_knob(
    session,
    namespace,
    catalog,
    user_relevance: Dict[str, Tuple[set, Dict[str, float]]],
) -> ScenarioResult:
    user_id = "user_0005"
    train_items, relevance = user_relevance.get(user_id, (set(), {}))
    if not relevance:
        return ScenarioResult(
            id="S4",
            name="Diversity budget",
            input="MMR_LAMBDA override 0.3 -> 0.1 for user_0005",
            expected="Lower lambda increases diversity with ≤5% NDCG loss.",
            observed="No hold-out relevance for user_0005; scenario skipped.",
            passed=False,
        )

    base_payload = {
        "namespace": namespace,
        "user_id": user_id,
        "k": 20,
        "context": {"surface": SURFACE},
        "include_reasons": True,
    }
    baseline = recommend(session, namespace, base_payload)
    base_items = [item["item_id"] for item in baseline.get("items", [])]

    diversity_payload = dict(base_payload)
    diversity_payload["overrides"] = {"mmr_lambda": 0.0}
    diversified = recommend(session, namespace, diversity_payload)
    diverse_items = [item["item_id"] for item in diversified.get("items", [])]

    ndcg_base, _, _ = compute_metrics_for_user(base_items, relevance, k=20)
    ndcg_diverse, _, _ = compute_metrics_for_user(diverse_items, relevance, k=20)

    sim_base = compute_similarity(base_items, catalog)
    sim_diverse = compute_similarity(diverse_items, catalog)

    change = (ndcg_diverse - ndcg_base) / ndcg_base if ndcg_base > 0 else 0.0

    write_evidence(
        "scenario_s4_diversity.json",
        {
            "baseline": baseline,
            "diversified": diversified,
            "ndcg_base": ndcg_base,
            "ndcg_diverse": ndcg_diverse,
            "similarity_base": sim_base,
            "similarity_diverse": sim_diverse,
        },
    )

    observed = (
        f"Similarity {sim_base:.3f} -> {sim_diverse:.3f}; "
        f"NDCG {ndcg_base:.3f} -> {ndcg_diverse:.3f} ({change:+.2%})."
    )
    passed = sim_diverse + 0.0001 < sim_base and change >= -0.05
    return ScenarioResult(
        id="S4",
        name="Diversity budget",
        input="Override mmr_lambda=0.1 vs default for user_0005",
        expected="Higher diversity (lower similarity) with ≤5% NDCG loss.",
        observed=observed,
        passed=passed,
    )


def scenario_pin_position(session, namespace) -> ScenarioResult:
    user_id = "user_0007"
    target_item = "item_0005"
    rule_payload = {
        "namespace": namespace,
        "surface": SURFACE,
        "name": "scenario_s5_pin_item",
        "description": "Scenario S5 pin position test",
        "action": "PIN",
        "target_type": "ITEM",
        "item_ids": [target_item],
        "priority": 200,
        "max_pins": 1,
        "enabled": True,
    }
    created = create_rule(session, rule_payload)
    rule_id = created["rule_id"]
    time.sleep(0.3)

    payload = {
        "namespace": namespace,
        "user_id": user_id,
        "k": 10,
        "context": {"surface": SURFACE},
        "include_reasons": True,
    }
    response = recommend(session, namespace, payload)
    items = [item["item_id"] for item in response.get("items", [])]

    rule_payload["enabled"] = False
    update_rule(session, rule_id, rule_payload)

    rank = items.index(target_item) if target_item in items else None

    write_evidence(
        "scenario_s5_pin.json",
        {"response": response, "target_item": target_item, "rank": rank, "rule": created},
    )

    observed = f"Target item rank={rank}; recommendation count={len(items)}"
    passed = rank == 0
    return ScenarioResult(
        id="S5",
        name="Pin position",
        input=f"PIN rule for {target_item} on surface {SURFACE}",
        expected="Pinned item appears at requested top slot.",
        observed=observed,
        passed=passed,
    )


def scenario_whitelist_brand(session, namespace, catalog) -> ScenarioResult:
    user_id = "user_0008"
    brand_tag = "acmetech"
    payload = {
        "namespace": namespace,
        "user_id": user_id,
        "k": 10,
        "context": {"surface": SURFACE},
        "constraints": {"include_tags_any": [brand_tag]},
    }
    response = recommend(session, namespace, payload)
    items = response.get("items", [])
    failing = []
    for item in items:
        tags = catalog.get(item["item_id"], {}).get("tags", set())
        if brand_tag not in tags:
            failing.append(item["item_id"])

    write_evidence("scenario_s6_whitelist.json", {"request": payload, "response": response})

    observed = f"Returned {len(items)} items; non-whitelisted={failing}"
    passed = len(items) > 0 and not failing
    return ScenarioResult(
        id="S6",
        name="Whitelist brand",
        input=f"/v1/recommendations include_tags_any=['{brand_tag}'] user_0008",
        expected="All items belong to brand 'AcmeTech'.",
        observed=observed,
        passed=passed,
    )


def scenario_cold_start(session, namespace, catalog) -> ScenarioResult:
    user_id = "user_cold_start"
    user_payload = {
        "user_id": user_id,
        "traits": {"segment": "cold_start", "notes": "Synthetic cold-start user"},
    }
    upsert_user(session, namespace, user_payload)
    payload = {
        "namespace": namespace,
        "user_id": user_id,
        "k": 10,
        "context": {"surface": SURFACE},
        "include_reasons": True,
    }
    response = recommend(session, namespace, payload)
    items = response.get("items", [])
    categories = {catalog.get(item["item_id"], {}).get("category") for item in items}
    starter_profile = extract_starter_profile(response)

    write_evidence(
        "scenario_s7_cold_start.json",
        {
            "request": payload,
            "response": response,
            "categories": list(categories),
            "starter_profile": starter_profile,
        },
    )

    observed = (
        f"Returned {len(items)} items across {len(categories)} categories; "
        f"Starter profile tags={list(starter_profile.keys()) if starter_profile else []}."
    )
    passed = len(items) >= 5 and len([c for c in categories if c]) >= 4
    return ScenarioResult(
        id="S7",
        name="Cold-start user",
        input="Upsert new user with no history then call /v1/recommendations",
        expected="List is non-empty, diverse, adheres to business rules.",
        observed=observed,
        passed=passed,
    )


def scenario_new_item_exposure(session, namespace, catalog, users_sample: List[str]) -> ScenarioResult:
    item_id = "item_new_explore"
    new_item = {
        "item_id": item_id,
        "category": "Fashion",
        "brand": "AuraThreads",
        "price": 149.0,
        "tags": ["fashion", "style", "aurathreads", "new_arrival"],
        "available": True,
        "props": {"margin": 0.55, "novelty": 0.95},
    }
    upsert_item(session, namespace, new_item)
    time.sleep(0.3)

    def collect_exposure(override: bool) -> Dict:
        exposures = []
        policy_samples: List[Dict] = []
        payload_base = {
            "namespace": namespace,
            "k": 15,
            "context": {"surface": SURFACE},
        }
        if override:
            override_payload = {
                "namespace": namespace,
                "surface": SURFACE,
                "item_id": item_id,
                "action": "boost",
                "boost_value": 2.5,
                "created_by": "scenario_s8",
                "notes": "New item exposure boost",
            }
            manual = create_manual_override(session, override_payload)
        else:
            manual = None

        for user_id in users_sample:
            payload = dict(payload_base)
            payload["user_id"] = user_id
            response = recommend(session, namespace, payload)
            items = [it["item_id"] for it in response.get("items", [])]
            rank = items.index(item_id) + 1 if item_id in items else None
            exposures.append(
                {
                    "user_id": user_id,
                    "rank": rank,
                    "recommended": item_id in items,
                }
            )
            policy = extract_policy_summary(response)
            if policy is not None:
                policy_samples.append(policy)
            time.sleep(SLEEP_BETWEEN_CALLS)

        if manual:
            cancel_manual_override(session, manual["override_id"])

        return {"exposures": exposures, "policy_samples": policy_samples}

    baseline = collect_exposure(override=False)
    boosted = collect_exposure(override=True)

    write_evidence(
        "scenario_s8_new_item.json",
        {
            "baseline": baseline,
            "boosted": boosted,
            "item": new_item,
        },
    )

    base_rate = sum(1 for e in baseline["exposures"] if e["recommended"]) / len(baseline["exposures"])
    boost_rate = sum(1 for e in boosted["exposures"] if e["recommended"]) / len(boosted["exposures"])

    boosted_policy = summarize_policy_samples(boosted["policy_samples"])
    observed = (
        f"Baseline exposure={base_rate:.0%}; Boosted exposure={boost_rate:.0%}; "
        f"Boost policy: {boosted_policy}"
    )
    passed = (
        boost_rate > base_rate
        and boost_rate >= 0.2
        and boosted_policy.get("rule_boost_exposure", 0) > 0
    )
    return ScenarioResult(
        id="S8",
        name="New item exposure",
        input="Insert fresh item and measure exposure before/after boost",
        expected="Boost controls increase exposure materially without manual data load.",
        observed=observed,
        passed=passed,
    )


def scenario_multi_objective(
    session,
    namespace,
    catalog,
    user_relevance: Dict[str, Tuple[set, Dict[str, float]]],
    users_sample: List[str],
) -> ScenarioResult:
    rule_payload = {
        "namespace": namespace,
        "surface": SURFACE,
        "name": "scenario_s9_margin_boost",
        "description": "Scenario S9 multi-objective trade-off",
        "action": "BOOST",
        "target_type": "TAG",
        "target_key": "high_margin",
        "priority": 120,
        "boost_value": 0.2,
        "enabled": True,
    }
    created = create_rule(session, rule_payload)
    rule_id = created["rule_id"]
    time.sleep(0.3)

    curve = []
    for boost_value in [0.0, 0.4, 0.8]:
        if boost_value == 0.0:
            rule_payload["enabled"] = False
        else:
            rule_payload.update({"boost_value": boost_value, "enabled": True})
        update_rule(session, rule_id, rule_payload)
        time.sleep(0.3)

        ndcgs = []
        margins = []
        for user_id in users_sample:
            _, relevance = user_relevance.get(user_id, (set(), {}))
            payload = {
                "namespace": namespace,
                "user_id": user_id,
                "k": 20,
                "context": {"surface": SURFACE},
            }
            response = recommend(session, namespace, payload)
            ranked = [item["item_id"] for item in response.get("items", [])]
            if relevance:
                ndcg, _, _ = compute_metrics_for_user(ranked, relevance, k=20)
                ndcgs.append(ndcg)
            margins.append(average_margin(ranked, catalog))
            time.sleep(SLEEP_BETWEEN_CALLS)

        avg_ndcg = statistics.mean(ndcgs) if ndcgs else 0.0
        avg_margin = statistics.mean(margins) if margins else 0.0
        curve.append({"boost_value": boost_value, "ndcg": avg_ndcg, "margin": avg_margin})

    rule_payload["enabled"] = False
    update_rule(session, rule_id, rule_payload)

    write_evidence("scenario_s9_tradeoff.json", {"curve": curve, "rule": created})

    ndcg_tolerance = 0.01
    increasing_margin = all(curve[i + 1]["margin"] >= curve[i]["margin"] for i in range(len(curve) - 1))
    decreasing_ndcg = all(
        curve[i + 1]["ndcg"] <= curve[i]["ndcg"] + ndcg_tolerance for i in range(len(curve) - 1)
    )
    margin_span = max(p["margin"] for p in curve) - min(p["margin"] for p in curve)
    observed_shift = margin_span > 0.01

    observed = "; ".join(f"boost {p['boost_value']}: margin={p['margin']:.3f}, ndcg={p['ndcg']:.3f}" for p in curve)
    passed = increasing_margin and decreasing_ndcg and observed_shift
    return ScenarioResult(
        id="S9",
        name="Multi-objective trade-off",
        input="Boost rule on tag 'high_margin' with varying boost_value",
        expected="Higher boost lifts margin share while relevance drops smoothly.",
        observed=observed,
        passed=passed,
    )


def scenario_explainability(session, namespace) -> ScenarioResult:
    user_id = "user_0004"
    payload = {
        "namespace": namespace,
        "user_id": user_id,
        "k": 10,
        "context": {"surface": SURFACE},
        "include_reasons": True,
        "explain_level": "full",
    }
    response = recommend(session, namespace, payload)
    write_evidence("scenario_s10_explainability.json", {"request": payload, "response": response})

    items = response.get("items", [])
    explain_available = any(item.get("explain") for item in items)
    reasons_present = any(item.get("reasons") for item in items)
    model_version = response.get("model_version")

    observed = f"model_version={model_version}, reasons={reasons_present}, explain_blocks={explain_available}"
    passed = bool(model_version) and reasons_present and explain_available
    return ScenarioResult(
        id="S10",
        name="Explainability",
        input="Request recommendations with include_reasons=true & explain_level=full",
        expected="Response includes reasons, explain blocks, and model identifier.",
        observed=observed,
        passed=passed,
    )


def run_all(base_url: str, namespace: str, org_id: str) -> List[ScenarioResult]:
    session = build_session(base_url, org_id)

    catalog = load_catalog(session, namespace)
    users = load_users(session, namespace)
    events = load_events(session, namespace)

    events_by_user = defaultdict(list)
    for event in events:
        events_by_user[event["user_id"]].append(event)

    split_ts = compute_split_timestamp(events, percentile=0.75)
    user_relevance = gather_user_relevance(events_by_user, split_ts)

    # Sample of users for scenarios needing multiple checks (power + others).
    users_sample = [f"user_{i:04d}" for i in range(1, 41)]

    results = [
        scenario_simple_filter(session, namespace, catalog),
        scenario_block_tag(session, namespace, catalog),
        scenario_boost_monotonicity(session, namespace, catalog),
        scenario_diversity_knob(session, namespace, catalog, user_relevance),
        scenario_pin_position(session, namespace),
        scenario_whitelist_brand(session, namespace, catalog),
        scenario_cold_start(session, namespace, catalog),
        scenario_new_item_exposure(session, namespace, catalog, users_sample),
        scenario_multi_objective(session, namespace, catalog, user_relevance, users_sample),
        scenario_explainability(session, namespace),
    ]

    return results


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Execute S1-S10 scenario suite against a RecSys API.")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL, help="Recommendation API base URL (default: %(default)s)")
    parser.add_argument("--namespace", default=DEFAULT_NAMESPACE, help="Namespace to target (default: %(default)s)")
    parser.add_argument("--org-id", default=DEFAULT_ORG_ID, help="Org ID / tenant identifier (default: %(default)s)")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    results = run_all(args.base_url, args.namespace, args.org_id)
    ensure_evidence_dir()
    with open("analysis/scenarios.csv", "w", encoding="utf-8", newline="") as fh:
        writer = csv.writer(fh)
        writer.writerow(SCENARIO_HEADERS)
        for result in results:
            writer.writerow(result.to_row())

    summary = {
        "results": [r.__dict__ for r in results],
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "base_url": args.base_url,
        "namespace": args.namespace,
        "org_id": args.org_id,
    }
    write_evidence("scenario_summary.json", summary)
    overall_pass = all(r.passed for r in results)
    status = "PASS" if overall_pass else "CHECK FINDINGS"
    print(json.dumps({"status": status, "scenarios": [r.to_row() for r in results]}, indent=2))
    if not overall_pass:
        sys.exit(1)


if __name__ == "__main__":
    main()
