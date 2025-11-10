#!/usr/bin/env python3
"""
Delete all items, users, and events within a namespace via the public API.

Intended to run ahead of bespoke seeding so fixtures never mingle with the
previous dataset. Records a small evidence artifact under analysis/evidence/.
"""

from __future__ import annotations

import argparse
import json
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, List

import requests
import urllib3

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

DEFAULT_BASE_URL = "https://api.pepe.local"
DEFAULT_NAMESPACE = "default"
DEFAULT_ORG_ID = "00000000-0000-0000-0000-000000000001"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Reset a namespace via delete endpoints.")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL, help="API base URL.")
    parser.add_argument("--namespace", default=DEFAULT_NAMESPACE, help="Namespace to wipe.")
    parser.add_argument("--org-id", default=DEFAULT_ORG_ID, help="Org ID header for multi-tenancy.")
    parser.add_argument(
        "--force",
        action="store_true",
        help="Skip the safety prompt (useful for CI / scripted runs).",
    )
    parser.add_argument(
        "--evidence-dir",
        default="analysis/evidence",
        help="Directory for reset evidence artifacts.",
    )
    return parser.parse_args()


def confirm(namespace: str, base_url: str) -> None:
    prompt = (
        f"This will delete all items, users, and events in namespace '{namespace}' "
        f"via {base_url}. Type 'reset' to continue: "
    )
    response = input(prompt).strip().lower()
    if response != "reset":
        print("Aborting reset.")
        sys.exit(1)


def post_json(session: requests.Session, url: str, payload: Dict, retries: int = 3) -> Dict:
    for attempt in range(retries):
        response = session.post(url, json=payload, timeout=30)
        if response.status_code in (200, 201, 202):
            if response.content:
                return response.json()
            return {"status": "ok"}
        if response.status_code >= 500 and attempt < retries - 1:
            continue
        raise RuntimeError(f"POST {url} failed: {response.status_code} {response.text}")
    raise AssertionError("unreachable")


def reset_namespace(base_url: str, namespace: str, org_id: str) -> List[Dict]:
    session = requests.Session()
    session.verify = False
    session.headers.update(
        {
            "Content-Type": "application/json",
            "X-Org-ID": org_id,
        }
    )

    endpoints = [
        ("events", "/v1/events:delete"),
        ("users", "/v1/users:delete"),
        ("items", "/v1/items:delete"),
    ]
    results: List[Dict] = []
    for label, path in endpoints:
        url = f"{base_url.rstrip('/')}{path}"
        payload = {"namespace": namespace}
        response = post_json(session, url, payload)
        deleted = response.get("deleted_count")
        results.append(
            {
                "resource": label,
                "endpoint": url,
                "deleted_count": deleted,
                "response": response,
            }
        )
        print(f"✓ Deleted {deleted} {label} (endpoint {path})")
    return results


def write_evidence(evidence_dir: str, metadata: Dict) -> str:
    Path(evidence_dir).mkdir(parents=True, exist_ok=True)
    timestamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    path = Path(evidence_dir) / f"reset_{timestamp}.json"
    with open(path, "w", encoding="utf-8") as fh:
        json.dump(metadata, fh, indent=2)
    return str(path)


def main() -> None:
    args = parse_args()
    if not args.force:
        confirm(args.namespace, args.base_url)

    operations = reset_namespace(args.base_url, args.namespace, args.org_id)
    evidence = {
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "base_url": args.base_url,
        "namespace": args.namespace,
        "org_id": args.org_id,
        "operations": operations,
    }
    artifact = write_evidence(args.evidence_dir, evidence)
    print(f"Reset evidence written to {artifact}")


if __name__ == "__main__":
    main()
