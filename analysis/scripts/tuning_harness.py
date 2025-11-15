#!/usr/bin/env python3
"""
Automated tuning harness.

Iterates through a grid (or random sample) of blend/MMR/fanout values,
applies them via env_profile_manager + recommendation_config API, runs
seed + quality evaluation, and records metrics under analysis/results/tuning_runs/.
"""
from __future__ import annotations

import argparse
import itertools
import json
import math
import random
import shutil
import subprocess
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, Iterable, List, Tuple


DEFAULT_BASE_URL = "https://api.pepe.local"
DEFAULT_NAMESPACE = "default"
DEFAULT_ORG_ID = "00000000-0000-0000-0000-000000000001"
DEFAULT_RESULTS_DIR = Path("analysis/results/tuning_runs")
PROFILE_MANAGER = ["python", "analysis/scripts/env_profile_manager.py"]


def run_cmd(cmd: List[str]) -> None:
    print(f"$ {' '.join(cmd)}")
    subprocess.run(cmd, check=True)


def load_json(path: Path) -> Dict:
    with path.open("r", encoding="utf-8") as fh:
        return json.load(fh)


def write_json(path: Path, payload: Dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as fh:
        json.dump(payload, fh, indent=2)


def extract_metrics(quality_path: Path) -> Dict[str, float]:
    data = load_json(quality_path)
    overall = data.get("overall", {})
    lift = overall.get("lift", {})
    coverage = data.get("coverage", {})
    return {
        "ndcg_lift": lift.get("ndcg@10"),
        "mrr_lift": lift.get("mrr@10"),
        "recall_lift": lift.get("recall@20"),
        "catalog_coverage": coverage.get("system_catalog_coverage"),
        "long_tail_share": overall.get("system", {}).get("long_tail_share@20"),
        "user_count": data.get("user_count"),
    }


def quality_evidence_paths(namespace: str) -> Tuple[Path, Path]:
    base = Path("analysis")
    return base / "quality_metrics.json", Path("analysis/results") / f"{namespace}_warm_quality.json"


def build_install_overrides(alpha: float, beta: float, gamma: float, mmr: float, fanout: int) -> Dict[str, str]:
    return {
        "BLEND_ALPHA": str(alpha),
        "BLEND_BETA": str(beta),
        "BLEND_GAMMA": str(gamma),
        "MMR_LAMBDA": str(mmr),
        "POPULARITY_FANOUT": str(fanout),
    }


def apply_profile(base_url: str, org_id: str, namespace: str, profile: str, overrides: Dict[str, str]) -> None:
    fetch_cmd = PROFILE_MANAGER + [
        "fetch",
        "--base-url",
        base_url,
        "--org-id",
        org_id,
        "--namespace",
        namespace,
        "--profile",
        profile,
    ]
    run_cmd(fetch_cmd)
    profile_file = Path("analysis/env_profiles") / namespace / f"{profile}.json"
    payload = load_json(profile_file)
    config = payload.get("config", {})
    for key, value in overrides.items():
        config[key] = type_cast(value)
    payload["config"] = config
    profile_file.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    apply_cmd = PROFILE_MANAGER + [
        "apply",
        "--base-url",
        base_url,
        "--org-id",
        org_id,
        "--namespace",
        namespace,
        "--profile",
        profile,
        "--author",
        "tuning-harness",
        "--notes",
        f"auto tuning {profile}",
    ]
    run_cmd(apply_cmd)


def type_cast(value: str):
    try:
        if "." in value:
            return float(value)
        return int(value)
    except ValueError:
        return value


def run_seed(base_url: str, org_id: str, namespace: str, users: int, events: int) -> None:
    cmd = [
        "python",
        "analysis/scripts/seed_dataset.py",
        "--base-url",
        base_url,
        "--org-id",
        org_id,
        "--namespace",
        namespace,
        "--user-count",
        str(users),
        "--min-events",
        str(events),
    ]
    run_cmd(cmd)


def run_quality(base_url: str, org_id: str, namespace: str, env_file: str, sleep_ms: int) -> None:
    cmd = [
        "python",
        "analysis/scripts/run_quality_eval.py",
        "--base-url",
        base_url,
        "--org-id",
        org_id,
        "--namespace",
        namespace,
        "--env-file",
        env_file,
        "--sleep-ms",
        str(sleep_ms),
    ]
    run_cmd(cmd)


def build_grid(values: Dict[str, List]) -> Iterable[Dict[str, float]]:
    keys = sorted(values.keys())
    for combo in itertools.product(*(values[k] for k in keys)):
        yield dict(zip(keys, combo))


def build_random(space: Dict[str, Tuple[float, float]], samples: int, seed: int) -> Iterable[Dict[str, float]]:
    rng = random.Random(seed)
    for _ in range(samples):
        config = {}
        for key, (lo, hi) in space.items():
            config[key] = round(rng.uniform(lo, hi), 4)
        yield config


def main() -> None:
    parser = argparse.ArgumentParser(description="Automated tuning harness.")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL)
    parser.add_argument("--org-id", default=DEFAULT_ORG_ID)
    parser.add_argument("--namespace", default=DEFAULT_NAMESPACE)
    parser.add_argument("--env-file", default="api/.env")
    parser.add_argument("--profile-name", default="tuning")
    parser.add_argument("--sleep-ms", type=int, default=200)
    parser.add_argument("--user-count", type=int, default=600)
    parser.add_argument("--event-count", type=int, default=40000)
    parser.add_argument("--results-dir", default=str(DEFAULT_RESULTS_DIR))
    parser.add_argument("--grid", action="store_true", help="Use Cartesian grid search.")
    parser.add_argument("--samples", type=int, default=5, help="Number of random samples (if grid disabled).")
    parser.add_argument("--seed", type=int, default=1234, help="Random seed for sampling.")
    parser.add_argument("--alphas", default="0.2,0.3,0.4")
    parser.add_argument("--betas", default="0.3,0.5,0.7")
    parser.add_argument("--gammas", default="0.1,0.2,0.3")
    parser.add_argument("--mmrs", default="0.15,0.25")
    parser.add_argument("--fanouts", default="400,600,800")
    args = parser.parse_args()

    results_dir = Path(args.results_dir)
    timestamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    run_dir = results_dir / f"{args.namespace}_{timestamp}"
    run_dir.mkdir(parents=True, exist_ok=True)

    grid_values = {
        "alpha": [float(x) for x in args.alphas.split(",") if x],
        "beta": [float(x) for x in args.betas.split(",") if x],
        "gamma": [float(x) for x in args.gammas.split(",") if x],
        "mmr": [float(x) for x in args.mmrs.split(",") if x],
        "fanout": [int(x) for x in args.fanouts.split(",") if x],
    }

    if args.grid:
        iterator = build_grid(grid_values)
    else:
        space = {
            "alpha": (min(grid_values["alpha"]), max(grid_values["alpha"])),
            "beta": (min(grid_values["beta"]), max(grid_values["beta"])),
            "gamma": (min(grid_values["gamma"]), max(grid_values["gamma"])),
            "mmr": (min(grid_values["mmr"]), max(grid_values["mmr"])),
            "fanout": (min(grid_values["fanout"]), max(grid_values["fanout"])),
        }
        iterator = build_random(space, args.samples, args.seed)

    summary: List[Dict] = []
    for idx, config in enumerate(iterator, start=1):
        run_id = f"{idx:03d}"
        print(f"=== Run {run_id}: {config} ===")
        overrides = build_install_overrides(
            config["alpha"],
            config["beta"],
            config["gamma"],
            config["mmr"],
            int(config["fanout"]),
        )
        apply_profile(args.base_url, args.org_id, args.namespace, args.profile_name, overrides)
        run_seed(args.base_url, args.org_id, args.namespace, args.user_count, args.event_count)
        run_quality(args.base_url, args.org_id, args.namespace, args.env_file, args.sleep_ms)
        metrics = extract_metrics(quality_evidence_paths(args.namespace)[0])
        row = {
            "run": run_id,
            "parameters": overrides,
            "metrics": metrics,
        }
        summary.append(row)
        write_json(run_dir / f"run_{run_id}.json", row)

    write_json(run_dir / "summary.json", {"runs": summary})
    print(f"Tuning runs completed. Summary at {run_dir / 'summary.json'}")


if __name__ == "__main__":
    main()
