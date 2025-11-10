#!/usr/bin/env python3
"""
Patch api/.env (or any target file) using profile templates and inline overrides.

Examples:
  configure_env.py --profile dev
  configure_env.py --set PROFILE_BOOST=0.45 --set PROFILE_STARTER_BLEND_WEIGHT=0.55
  configure_env.py --profile prod --set API_PORT=8081 --note \"staging override\"
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, List, Optional, Tuple

import yaml

DEFAULT_ENV_FILE = "api/.env"
PROFILES_REGISTRY = "config/profiles.yml"
PROFILE_DIR = "api/env"
HISTORY_DIR = "analysis/env_history"


@dataclass
class EnvLine:
    kind: str  # kv/comment/blank
    raw: str
    key: Optional[str] = None
    value: Optional[str] = None
    suffix: str = ""


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Configure api/.env via profiles and overrides.")
    parser.add_argument("--env-file", default=DEFAULT_ENV_FILE, help="Target env file to rewrite.")
    parser.add_argument(
        "--profile",
        choices=["dev", "test", "prod", "ci"],
        help="Optional profile to load from api/env/<profile>.env before applying overrides.",
    )
    parser.add_argument(
        "--set",
        dest="overrides",
        action="append",
        default=[],
        help="Override in KEY=VALUE form. Can be repeated.",
    )
    parser.add_argument(
        "--note",
        help="Optional note stored in env history for auditing.",
    )
    parser.add_argument(
        "--profiles-file",
        default=PROFILES_REGISTRY,
        help="Optional registry describing namespace→profile mappings (default: config/profiles.yml).",
    )
    parser.add_argument(
        "--namespace",
        help="Namespace whose profile should be applied (looked up in profiles file).",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show the changes without rewriting the env file.",
    )
    return parser.parse_args()


def read_lines(path: str) -> List[str]:
    if not os.path.exists(path):
        return []
    with open(path, "r", encoding="utf-8") as fh:
        return fh.readlines()


def parse_env_lines(lines: List[str]) -> Tuple[List[EnvLine], Dict[str, EnvLine]]:
    parsed: List[EnvLine] = []
    index: Dict[str, EnvLine] = {}
    for original in lines:
        stripped = original.strip()
        if stripped == "":
            parsed.append(EnvLine(kind="blank", raw=original))
            continue
        if stripped.startswith("#"):
            parsed.append(EnvLine(kind="comment", raw=original))
            continue
        if "=" not in original:
            parsed.append(EnvLine(kind="other", raw=original))
            continue
        key, rest = original.split("=", 1)
        suffix = ""
        value = rest.rstrip("\n")
        if " #" in value:
            value, tail = value.split(" #", 1)
            suffix = " #" + tail
        line = EnvLine(kind="kv", raw=original, key=key.strip(), value=value.strip(), suffix=suffix)
        parsed.append(line)
        if line.key:
            index[line.key] = line
    return parsed, index


def apply_overrides(
    parsed: List[EnvLine],
    index: Dict[str, EnvLine],
    overrides: Dict[str, str],
) -> Tuple[List[EnvLine], Dict[str, str], Dict[str, str]]:
    before: Dict[str, str] = {}
    after: Dict[str, str] = {}
    remaining = overrides.copy()
    for key, new_val in overrides.items():
        if key in index:
            line = index[key]
            before[key] = line.value or ""
            line.value = new_val
            after[key] = new_val
            del remaining[key]
    for key, new_val in remaining.items():
        line = EnvLine(
            kind="kv",
            raw="",
            key=key,
            value=new_val,
            suffix="",
        )
        parsed.append(line)
        before.setdefault(key, "")
        after[key] = new_val
    return parsed, before, after


def render_lines(parsed: List[EnvLine]) -> str:
    rendered: List[str] = []
    for line in parsed:
        if line.kind == "kv" and line.key is not None:
            value = line.value or ""
            suffix = line.suffix
            rendered.append(f"{line.key}={value}{suffix}\n")
        else:
            rendered.append(line.raw if line.raw.endswith("\n") else line.raw + "\n")
    if not rendered or rendered[-1].endswith("\n"):
        return "".join(rendered)
    rendered.append("\n")
    return "".join(rendered)


def parse_overrides(pairs: List[str]) -> Dict[str, str]:
    overrides: Dict[str, str] = {}
    for pair in pairs:
        if "=" not in pair:
            raise ValueError(f"Override '{pair}' must be in KEY=VALUE format.")
        key, value = pair.split("=", 1)
        overrides[key.strip()] = value.strip()
    return overrides


def write_file(path: str, content: str) -> None:
    Path(path).parent.mkdir(parents=True, exist_ok=True)
    with open(path, "w", encoding="utf-8") as fh:
        fh.write(content)


def compute_hash(content: str) -> str:
    return hashlib.sha256(content.encode("utf-8")).hexdigest()


def write_history(
    env_file: str,
    profile: Optional[str],
    overrides: Dict[str, str],
    before: Dict[str, str],
    after: Dict[str, str],
    note: Optional[str],
    env_hash: str,
) -> str:
    Path(HISTORY_DIR).mkdir(parents=True, exist_ok=True)
    timestamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    history = {
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "env_file": env_file,
        "profile": profile,
        "overrides": overrides,
        "before": before,
        "after": after,
        "note": note,
        "env_hash": env_hash,
    }
    path = Path(HISTORY_DIR) / f"{timestamp}.json"
    with open(path, "w", encoding="utf-8") as fh:
        json.dump(history, fh, indent=2)
    return str(path)


def load_profiles(path: str) -> Dict[str, Dict]:
    registry_path = Path(path)
    if not registry_path.exists():
        return {}
    try:
        data = yaml.safe_load(registry_path.read_text())
    except yaml.YAMLError as exc:
        raise ValueError(f"Failed to parse profiles file {path}: {exc}") from exc
    if not isinstance(data, dict):
        return {}
    return data.get("profiles", {})


def find_profile_for_namespace(profiles: Dict[str, Dict], namespace: str) -> Optional[str]:
    key = namespace.strip().lower()
    for profile, cfg in profiles.items():
        namespaces = cfg.get("namespaces") or []
        normalized = [n.strip().lower() for n in namespaces]
        if key in normalized:
            return profile
    return None


def main() -> None:
    args = parse_args()
    profile_lines: List[str] = []
    profile_path = None
    resolved_profile = args.profile
    profiles_map = {}
    if args.namespace:
        profiles_map = load_profiles(args.profiles_file)
        if not resolved_profile:
            resolved_profile = find_profile_for_namespace(profiles_map, args.namespace)
            if resolved_profile:
                print(f"[info] resolved profile '{resolved_profile}' for namespace '{args.namespace}' via {args.profiles_file}")
            else:
                print(f"[info] no profile mapping found for namespace '{args.namespace}' in {args.profiles_file}")
    if resolved_profile:
        profile_path = os.path.join(PROFILE_DIR, f"{resolved_profile}.env")
        if not os.path.exists(profile_path):
            raise FileNotFoundError(f"Profile file {profile_path} not found.")
        profile_lines = read_lines(profile_path)

    target_exists = os.path.exists(args.env_file)
    if profile_lines:
        base_lines = profile_lines
    elif target_exists:
        base_lines = read_lines(args.env_file)
    else:
        base_lines = []

    parsed, index = parse_env_lines(base_lines)
    overrides = parse_overrides(args.overrides)

    if not overrides and not profile_lines:
        print("Nothing to change: no profile chosen and no overrides provided.")
        return

    parsed, before, after = apply_overrides(parsed, index, overrides)
    rendered = render_lines(parsed)

    print(f"Target env file: {args.env_file}")
    if resolved_profile:
        print(f"  Loaded profile: {resolved_profile}")
    if overrides:
        print("  Overrides:")
        for k, v in overrides.items():
            prev = before.get(k, "")
            print(f"    {k}: '{prev}' -> '{v}'")

    if args.dry_run:
        print("Dry-run mode: env file not written.")
        return

    write_file(args.env_file, rendered)
    history_path = write_history(
        env_file=args.env_file,
        profile=resolved_profile,
        overrides=overrides,
        before=before,
        after=after,
        note=args.note,
        env_hash=compute_hash(rendered),
    )
    print(f"Env file updated. History recorded in {history_path}")


if __name__ == "__main__":
    main()
