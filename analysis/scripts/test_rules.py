#!/usr/bin/env python3
"""
Exercise merchandising rule actions (BLOCK, BOOST, PIN) and capture evidence.

The script:
1. Calls /v1/recommendations to capture a baseline payload.
2. Creates temporary rules targeting the top results.
3. Re-runs recommendations to ensure rule counters fire.
4. Writes before/after payloads plus rule metadata to analysis/evidence/.
5. Disables the temporary rules (unless --keep-rules is set).
"""

from __future__ import annotations

import argparse
import json
import os
import sys
import time
from datetime import datetime, timezone
from typing import Dict, List, Tuple

import requests


DEFAULT_BASE_URL = os.getenv("RECSYS_BASE_URL", "https://api.pepe.local")
DEFAULT_NAMESPACE = os.getenv("RECSYS_NAMESPACE", "default")
DEFAULT_ORG_ID = os.getenv("RECSYS_ORG_ID", "00000000-0000-0000-0000-000000000001")
DEFAULT_SURFACE = "home"
DEFAULT_USER_ID = "user_0001"
DEFAULT_EVIDENCE = os.path.join("analysis", "evidence", "rules_test_report.json")


class APIClient:
    def __init__(self, base_url: str, org_id: str):
        self.base_url = base_url.rstrip("/")
        self.session = requests.Session()
        self.session.verify = False
        self.session.headers.update(
            {
                "Content-Type": "application/json",
                "X-Org-ID": org_id,
            }
        )

    def post_json(self, path: str, payload: Dict, expected: Tuple[int, ...] = (200,)) -> Dict:
        resp = self.session.post(f"{self.base_url}{path}", json=payload, timeout=30)
        if resp.status_code not in expected:
            raise RuntimeError(f"POST {path} failed ({resp.status_code}): {resp.text}")
        if resp.content:
            return resp.json()
        return {}

    def put_json(self, path: str, payload: Dict, expected: Tuple[int, ...] = (200,)) -> Dict:
        resp = self.session.put(f"{self.base_url}{path}", json=payload, timeout=30)
        if resp.status_code not in expected:
            raise RuntimeError(f"PUT {path} failed ({resp.status_code}): {resp.text}")
        if resp.content:
            return resp.json()
        return {}


def recommend(client: APIClient, namespace: str, user_id: str, surface: str, k: int) -> Dict:
    payload = {
        "namespace": namespace,
        "user_id": user_id,
        "k": k,
        "include_reasons": True,
        "explain_level": "full",
    }
    if surface:
        payload["context"] = {"surface": surface}
    return client.post_json("/v1/recommendations", payload)


def ensure_dir(path: str) -> None:
    directory = os.path.dirname(path)
    if directory:
        os.makedirs(directory, exist_ok=True)


def extract_policy(resp: Dict) -> Dict:
    trace = resp.get("trace")
    if not isinstance(trace, dict):
        return {}
    extras = trace.get("extras")
    if not isinstance(extras, dict):
        return {}
    policy = extras.get("policy")
    return policy if isinstance(policy, dict) else {}


def build_rule_payload(
    namespace: str,
    surface: str,
    action: str,
    item_id: str,
    *,
    boost: float | None = None,
    max_pins: int | None = None,
    enabled: bool = True,
    prefix: str,
) -> Dict:
    payload: Dict = {
        "namespace": namespace,
        "surface": surface,
        "name": f"{prefix}-{action.lower()}-{item_id}",
        "action": action,
        "target_type": "ITEM",
        "item_ids": [item_id],
        "enabled": enabled,
    }
    if boost is not None:
        payload["boost_value"] = boost
    if max_pins is not None:
        payload["max_pins"] = max_pins
    return payload


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Smoke-test merchandising rules.")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL)
    parser.add_argument("--org-id", default=DEFAULT_ORG_ID)
    parser.add_argument("--namespace", default=DEFAULT_NAMESPACE)
    parser.add_argument("--surface", default=DEFAULT_SURFACE)
    parser.add_argument("--user-id", default=DEFAULT_USER_ID)
    parser.add_argument("--k", type=int, default=10)
    parser.add_argument("--evidence-path", default=DEFAULT_EVIDENCE)
    parser.add_argument("--sleep-seconds", type=float, default=0.5, help="Wait after creating rules.")
    parser.add_argument(
        "--keep-rules",
        action="store_true",
        help="Skip disabling the temporary rules after verification.",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    client = APIClient(args.base_url, args.org_id)

    baseline = recommend(client, args.namespace, args.user_id, args.surface, args.k)
    base_items = baseline.get("items", [])
    if len(base_items) < 3:
        raise RuntimeError("Need at least 3 baseline items to exercise block/boost/pin.")

    timestamp = datetime.now(timezone.utc).strftime("%Y%m%d%H%M%S")
    rule_prefix = f"auto-rule-test-{timestamp}"
    targets = base_items[:3]
    rules_to_disable: List[Tuple[str, Dict]] = []

    block_payload = build_rule_payload(
        args.namespace, args.surface, "BLOCK", targets[0]["item_id"], prefix=rule_prefix
    )
    boost_payload = build_rule_payload(
        args.namespace,
        args.surface,
        "BOOST",
        targets[1]["item_id"],
        boost=0.15,
        prefix=rule_prefix,
    )
    pin_payload = build_rule_payload(
        args.namespace,
        args.surface,
        "PIN",
        targets[2]["item_id"],
        max_pins=1,
        prefix=rule_prefix,
    )

    created_rules: List[Dict] = []
    disable_specs: List[Tuple[str, Dict]] = []
    failure_message: str | None = None
    try:
        for payload in (block_payload, boost_payload, pin_payload):
            resp = client.post_json("/v1/admin/rules", payload, expected=(201,))
            rule_id = resp.get("rule_id")
            if not isinstance(rule_id, str):
                raise RuntimeError(f"Rule creation response missing rule_id: {resp}")
            created_rules.append({"rule_id": rule_id, "response": resp, "payload": payload})
            disable_specs.append((rule_id, payload))

        time.sleep(args.sleep_seconds)
        mutated = recommend(client, args.namespace, args.user_id, args.surface, args.k)
        policy = extract_policy(mutated)
        validation = {
            "block_hits": policy.get("rule_block_count", 0),
            "boost_hits": policy.get("rule_boost_count", 0),
            "pin_hits": policy.get("rule_pin_count", 0),
            "pin_exposure": policy.get("rule_pin_exposure", 0),
            "boost_exposure": policy.get("rule_boost_exposure", 0),
        }
        block_ok = validation["block_hits"] >= 1
        boost_ok = validation["boost_hits"] >= 1 or validation["boost_exposure"] >= 1
        pin_ok = validation["pin_hits"] >= 1 or validation["pin_exposure"] >= 1
        validation["status"] = "PASSED" if (block_ok and boost_ok and pin_ok) else "FAILED"

        evidence = {
            "metadata": {
                "base_url": args.base_url,
                "namespace": args.namespace,
                "surface": args.surface,
                "user_id": args.user_id,
                "k": args.k,
                "timestamp": timestamp,
            },
            "baseline": baseline,
            "with_rules": mutated,
            "policy_summary": policy,
            "validation": validation,
            "rules": created_rules,
        }

        ensure_dir(args.evidence_path)
        with open(args.evidence_path, "w", encoding="utf-8") as fh:
            json.dump(evidence, fh, indent=2)

        if validation["status"] != "PASSED":
            missing = []
            if not block_ok:
                missing.append("block")
            if not boost_ok:
                missing.append("boost")
            if not pin_ok:
                missing.append("pin")
            failure_message = (
                "Rules validation failed; missing effects for " + ", ".join(missing)
            )
        else:
            print(
                f"Rules test complete. Block hits={validation['block_hits']}, "
                f"Boost hits={validation['boost_hits']}, Pin hits={validation['pin_hits']}."
            )
            print(f"Evidence written to {args.evidence_path}")
    finally:
        if not args.keep_rules and disable_specs:
            for rule_id, payload in disable_specs:
                disable_payload = dict(payload)
                disable_payload["enabled"] = False
                try:
                    client.put_json(f"/v1/admin/rules/{rule_id}", disable_payload, expected=(200,))
                except Exception as exc:  # pragma: no cover - best effort cleanup
                    print(f"Failed to disable rule {rule_id}: {exc}", file=sys.stderr)

    if failure_message:
        raise RuntimeError(failure_message)


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        sys.exit(1)
