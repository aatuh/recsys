#!/usr/bin/env python3
"""Helper to pause/unpause docker compose services to simulate failures."""
from __future__ import annotations

import argparse
import subprocess
import time
from pathlib import Path


def run(cmd: list[str]) -> None:
    subprocess.run(cmd, check=True)


def main() -> None:
    parser = argparse.ArgumentParser(description="Pause/unpause docker compose services for chaos testing.")
    parser.add_argument("service", help="Service name as defined in docker-compose.yml (e.g., db, api)")
    parser.add_argument("action", choices=["pause", "stop"], help="Chaos action to perform")
    parser.add_argument("duration", type=float, help="Duration in seconds before recovery")
    parser.add_argument("--compose-file", default="docker-compose.yml", help="Path to docker compose file")
    parser.add_argument("--project-directory", default=str(Path(__file__).resolve().parents[2]), help="Working directory for docker compose")
    args = parser.parse_args()

    compose_cmd = ["docker", "compose", "-f", args.compose_file]
    if args.action == "pause":
        run(compose_cmd + ["pause", args.service])
        try:
            time.sleep(args.duration)
        finally:
            run(compose_cmd + ["unpause", args.service])
    else:  # stop
        run(compose_cmd + ["stop", args.service])
        try:
            time.sleep(args.duration)
        finally:
            run(compose_cmd + ["start", args.service])


if __name__ == "__main__":
    main()
