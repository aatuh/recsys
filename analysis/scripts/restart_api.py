#!/usr/bin/env python3
"""
Restart the API service via docker compose and wait for /health to pass.

Designed to chain after configure_env.py so scripted experiments can safely
apply env overrides, bounce the container, and verify readiness before seeding.
"""

from __future__ import annotations

import argparse
import subprocess
import sys
import time
from datetime import datetime

import requests
import urllib3

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

DEFAULT_BASE_URL = "http://localhost:8000"
DEFAULT_SERVICE = "api"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Restart API via docker compose and wait for health.")
    parser.add_argument("--compose-file", help="Optional docker-compose.yml path.")
    parser.add_argument("--project-name", help="Optional docker compose project name.")
    parser.add_argument("--service", default=DEFAULT_SERVICE, help="Service name to recreate (default: api).")
    parser.add_argument(
        "--base-url",
        default=DEFAULT_BASE_URL,
        help="Base URL used for health checks (default: http://localhost:8000).",
    )
    parser.add_argument(
        "--health-path",
        default="/health",
        help="Relative path for health endpoint (default: /health).",
    )
    parser.add_argument(
        "--timeout",
        type=int,
        default=120,
        help="Seconds to wait for health before failing (default: 120).",
    )
    parser.add_argument(
        "--interval",
        type=int,
        default=3,
        help="Seconds between health probes (default: 3).",
    )
    parser.add_argument(
        "--skip-health",
        action="store_true",
        help="Skip the health check (not recommended).",
    )
    return parser.parse_args()


def run_compose(args: argparse.Namespace) -> None:
    cmd = ["docker", "compose"]
    if args.compose_file:
        cmd.extend(["-f", args.compose_file])
    if args.project_name:
        cmd.extend(["-p", args.project_name])
    cmd.extend(["up", "-d", "--force-recreate", args.service])
    print(f"[{datetime.utcnow().isoformat()}] Running: {' '.join(cmd)}")
    subprocess.run(cmd, check=True)


def wait_for_health(url: str, timeout: int, interval: int) -> None:
    deadline = time.time() + timeout
    while time.time() < deadline:
        try:
            resp = requests.get(url, timeout=5, verify=False)
            if resp.status_code == 200:
                print(f"[{datetime.utcnow().isoformat()}] Health check passed ({url}).")
                return
            print(f"[{datetime.utcnow().isoformat()}] Health endpoint returned {resp.status_code}; retrying...")
        except requests.RequestException as exc:
            print(f"[{datetime.utcnow().isoformat()}] Health probe failed: {exc}; retrying...")
        time.sleep(interval)
    raise TimeoutError(f"Service did not become healthy before timeout ({timeout}s).")


def main() -> None:
    args = parse_args()
    try:
        run_compose(args)
    except subprocess.CalledProcessError as exc:
        print(f"docker compose failed with exit code {exc.returncode}", file=sys.stderr)
        sys.exit(exc.returncode)

    if args.skip_health:
        print("Skipping health check per --skip-health.")
        return

    health_url = args.base_url.rstrip("/") + args.health_path
    try:
        wait_for_health(health_url, args.timeout, args.interval)
    except TimeoutError as exc:
        print(str(exc), file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
