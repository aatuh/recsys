#!/usr/bin/env python3
"""
Verify that every env profile (api/env/*.env) contains the same keys as api/.env.
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Check env profile parity.")
    parser.add_argument("--base", default="api/.env", help="Path to the canonical env file.")
    parser.add_argument(
        "--profiles",
        nargs="*",
        default=None,
        help="Explicit list of profile files (default: api/env/*.env).",
    )
    parser.add_argument("--strict", action="store_true", help="Exit 1 if any profile is missing keys.")
    return parser.parse_args()


def parse_env_keys(path: Path) -> set[str]:
    keys: set[str] = set()
    for line in path.read_text().splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key = line.split("=", 1)[0].strip()
        if key:
            keys.add(key)
    return keys


def main() -> None:
    args = parse_args()
    base_path = Path(args.base)
    if not base_path.exists():
        raise FileNotFoundError(f"Base env file '{base_path}' not found.")
    base_keys = parse_env_keys(base_path)

    profile_paths = (
        [Path(p) for p in args.profiles]
        if args.profiles
        else sorted(Path("api/env").glob("*.env"))
    )

    missing_total = {}
    for profile in profile_paths:
        if not profile.exists():
            print(f"[warn] profile '{profile}' missing")
            continue
        profile_keys = parse_env_keys(profile)
        missing = sorted(base_keys - profile_keys)
        if missing:
            missing_total[profile] = missing
            print(f"[fail] {profile}: missing {len(missing)} keys")
            for key in missing:
                print(f"  - {key}")
        else:
            print(f"[ok] {profile}")

    if missing_total and args.strict:
        sys.exit(1)


if __name__ == "__main__":
    main()
