#!/usr/bin/env python3
"""
Manage recommendation “env profiles” per namespace without rewriting api/.env.

Features:
- fetch: pull the current /v1/admin/recommendation/config snapshot and store it
         under analysis/env_profiles/<namespace>/<profile>.json
- apply: push a stored profile (or arbitrary JSON file) back to the API with author/notes
- list:  show locally stored profiles plus basic metadata
- delete: remove a stored profile (no API calls)
"""
from __future__ import annotations

import argparse
import json
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, List

import requests

try:  # optional TLS suppression
    import urllib3
except ImportError:  # pragma: no cover
    urllib3 = None


DEFAULT_PROFILES_DIR = Path("analysis/env_profiles")


def build_session(insecure: bool) -> requests.Session:
    session = requests.Session()
    if insecure:
        session.verify = False
        if urllib3 is not None:
            urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
    return session


def profile_path(base: Path, namespace: str, profile: str) -> Path:
    safe_namespace = namespace.replace("/", "_")
    safe_profile = profile.replace("/", "_")
    return base / safe_namespace / f"{safe_profile}.json"


def cmd_list(args: argparse.Namespace) -> None:
    base = Path(args.profiles_dir)
    if not base.exists():
        print("No profiles directory found.")
        return
    rows: List[str] = []
    for ns_dir in sorted(p for p in base.iterdir() if p.is_dir()):
        if args.namespace and ns_dir.name != args.namespace:
            continue
        for entry in sorted(ns_dir.glob("*.json")):
            meta = load_metadata(entry)
            rows.append(
                f"{ns_dir.name}/{entry.stem}"
                f" — updated {meta.get('updated_at', 'unknown')} by {meta.get('updated_by', 'n/a')}"
            )
    if rows:
        print("\n".join(rows))
    else:
        print("No profiles found for the given filter.")


def load_metadata(path: Path) -> Dict[str, Any]:
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {}
    meta = data.get("metadata")
    return meta if isinstance(meta, dict) else {}


def cmd_fetch(args: argparse.Namespace) -> None:
    session = build_session(args.insecure)
    url = f"{args.base_url.rstrip('/')}/v1/admin/recommendation/config"
    resp = session.get(
        url,
        headers={"X-Org-ID": args.org_id},
        params={"namespace": args.namespace},
        timeout=args.timeout,
    )
    resp.raise_for_status()
    payload = resp.json()
    destination = (
        Path(args.output)
        if args.output
        else profile_path(Path(args.profiles_dir), args.namespace, args.profile)
    )
    destination.parent.mkdir(parents=True, exist_ok=True)
    metadata = payload.get("metadata") or {}
    metadata["fetched_at"] = datetime.now(timezone.utc).isoformat()
    metadata["profile"] = args.profile
    payload["metadata"] = metadata
    destination.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    print(f"Profile saved to {destination}")


def load_profile_payload(path: Path) -> Dict[str, Any]:
    with path.open("r", encoding="utf-8") as fh:
        data = json.load(fh)
    if "config" not in data:
        raise ValueError(f"Profile file {path} missing 'config' key.")
    return data


def cmd_apply(args: argparse.Namespace) -> None:
    base = Path(args.profiles_dir)
    path = Path(args.input) if args.input else profile_path(base, args.namespace, args.profile)
    if not path.exists():
        raise SystemExit(f"Profile file {path} not found.")
    payload = load_profile_payload(path)
    namespace = args.namespace or payload.get("namespace") or "default"
    body = {
        "namespace": namespace,
        "config": payload["config"],
        "author": args.author,
        "notes": args.notes,
    }
    session = build_session(args.insecure)
    url = f"{args.base_url.rstrip('/')}/v1/admin/recommendation/config"
    resp = session.post(
        url,
        headers={"X-Org-ID": args.org_id},
        json=body,
        timeout=args.timeout,
    )
    resp.raise_for_status()
    print(f"Applied profile '{args.profile}' to namespace '{namespace}'.")


def cmd_delete(args: argparse.Namespace) -> None:
    base = Path(args.profiles_dir)
    path = profile_path(base, args.namespace, args.profile)
    if not path.exists():
        print(f"Profile {path} does not exist.")
        return
    path.unlink()
    # remove namespace dir if empty
    try:
        path.parent.rmdir()
    except OSError:
        pass
    print(f"Deleted profile {path}.")


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Manage recommendation env profiles via the admin API.")
    parser.add_argument("--profiles-dir", default=str(DEFAULT_PROFILES_DIR), help="Directory to store local profiles.")
    parser.add_argument("--namespace", default="default", help="Namespace to target (where applicable).")

    sub = parser.add_subparsers(dest="command", required=True)

    list_cmd = sub.add_parser("list", help="List stored profiles.")
    list_cmd.set_defaults(func=cmd_list)

    fetch_cmd = sub.add_parser("fetch", help="Fetch current config and store it as a profile.")
    fetch_cmd.add_argument("--base-url", required=True)
    fetch_cmd.add_argument("--org-id", required=True)
    fetch_cmd.add_argument("--profile", required=True, help="Profile name to store.")
    fetch_cmd.add_argument("--output", help="Optional explicit output path.")
    fetch_cmd.add_argument("--timeout", type=float, default=30.0)
    fetch_cmd.add_argument("--insecure", action="store_true")
    fetch_cmd.set_defaults(func=cmd_fetch)

    apply_cmd = sub.add_parser("apply", help="Apply a stored profile to the namespace via the API.")
    apply_cmd.add_argument("--base-url", required=True)
    apply_cmd.add_argument("--org-id", required=True)
    apply_cmd.add_argument("--profile", required=True, help="Profile name to apply.")
    apply_cmd.add_argument("--input", help="Optional explicit JSON file path instead of profiles dir.")
    apply_cmd.add_argument("--author", default="env-profile-cli")
    apply_cmd.add_argument("--notes", default="applied via env_profile_manager")
    apply_cmd.add_argument("--timeout", type=float, default=30.0)
    apply_cmd.add_argument("--insecure", action="store_true")
    apply_cmd.set_defaults(func=cmd_apply)

    delete_cmd = sub.add_parser("delete", help="Remove a stored profile (local only).")
    delete_cmd.add_argument("--profile", required=True)
    delete_cmd.set_defaults(func=cmd_delete)

    return parser


def main() -> None:
    parser = build_parser()
    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
