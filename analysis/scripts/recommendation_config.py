#!/usr/bin/env python3
"""Export/apply recommendation config snapshots via the admin API."""
from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any

import requests

try:  # optional dependency for TLS suppression
    import urllib3
except ImportError:  # pragma: no cover
    urllib3 = None


def build_session(insecure: bool) -> requests.Session:
    session = requests.Session()
    if insecure:
        session.verify = False
        if urllib3 is not None:
            urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    return session


def export_config(args: argparse.Namespace) -> None:
    session = build_session(args.insecure)
    params = {}
    if args.namespace:
        params["namespace"] = args.namespace
    resp = session.get(
        f"{args.base_url.rstrip('/')}/v1/admin/recommendation/config",
        headers={"X-Org-ID": args.org_id},
        params=params,
        timeout=args.timeout,
    )
    resp.raise_for_status()
    data = resp.json()
    output_path = Path(args.output)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
    print(f"Saved config snapshot to {output_path}")


def apply_config(args: argparse.Namespace) -> None:
    session = build_session(args.insecure)
    payload = build_apply_payload(args)
    resp = session.post(
        f"{args.base_url.rstrip('/')}/v1/admin/recommendation/config",
        headers={"X-Org-ID": args.org_id},
        json=payload,
        timeout=args.timeout,
    )
    resp.raise_for_status()
    print("Config updated via API")


def build_apply_payload(args: argparse.Namespace) -> dict[str, Any]:
    path = Path(args.input)
    with path.open("r", encoding="utf-8") as fh:
        data = json.load(fh)
    namespace = args.namespace or data.get("namespace") or "default"
    config_body: Any
    if "config" in data:
        config_body = data["config"]
    else:
        config_body = data
    return {
        "namespace": namespace,
        "config": config_body,
        "author": args.author,
        "notes": args.notes,
    }


def main() -> None:
    parser = argparse.ArgumentParser(description="Manage recommendation config snapshots via the API.")
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--org-id", required=True)
    parser.add_argument("--namespace", help="Namespace scope", default="default")
    parser.add_argument("--timeout", type=float, default=30.0)
    parser.add_argument("--insecure", action="store_true", help="Disable TLS verification")

    sub = parser.add_subparsers(dest="command", required=True)

    export_cmd = sub.add_parser("export", help="Export current config to a file")
    export_cmd.add_argument("--output", required=True)
    export_cmd.set_defaults(func=export_config)

    apply_cmd = sub.add_parser("apply", help="Apply config from a file")
    apply_cmd.add_argument("--input", required=True)
    apply_cmd.add_argument("--author", default="cli")
    apply_cmd.add_argument("--notes", default="applied via script")
    apply_cmd.set_defaults(func=apply_config)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
