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
from typing import Dict, Iterable, List, Optional, Tuple


DEFAULT_BASE_URL = "https://api.pepe.local"
DEFAULT_NAMESPACE = "default"
DEFAULT_ORG_ID = "00000000-0000-0000-0000-000000000001"
DEFAULT_RESULTS_DIR = Path("analysis/results/tuning_runs")
PROFILE_MANAGER = ["python", "analysis/scripts/env_profile_manager.py"]
CONFIG_FIELD_MAP = {
    "BLEND_ALPHA": "blend_alpha",
    "BLEND_BETA": "blend_beta",
    "BLEND_GAMMA": "blend_gamma",
    "MMR_LAMBDA": "mmr_lambda",
    "POPULARITY_FANOUT": "popularity_fanout",
    "PROFILE_BOOST": "profile_boost",
    "PROFILE_MIN_EVENTS_FOR_BOOST": "profile_min_events_for_boost",
    "PROFILE_STARTER_BLEND_WEIGHT": "profile_starter_blend_weight",
}


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


def extract_metrics(quality_path: Path, segment: Optional[str] = None) -> Dict[str, float]:
    data = load_json(quality_path)
    overall = data.get("overall", {})
    lift = overall.get("lift", {})
    coverage = data.get("coverage", {})
    result = {
        "ndcg_lift": lift.get("ndcg@10"),
        "mrr_lift": lift.get("mrr@10"),
        "recall_lift": lift.get("recall@20"),
        "catalog_coverage": coverage.get("system_catalog_coverage"),
        "long_tail_share": overall.get("system", {}).get("long_tail_share@20"),
        "user_count": data.get("user_count"),
    }
    if segment:
        segments = data.get("segments") or {}
        seg_key = segment.strip()
        seg_data = segments.get(seg_key)
        if seg_data is None:
            # try lowercase fallback (most segments stored lowercase)
            seg_data = segments.get(seg_key.lower())
        if seg_data:
            seg_lift = seg_data.get("lift", {})
            result["segment_ndcg_lift"] = seg_lift.get("ndcg@10")
            result["segment_mrr_lift"] = seg_lift.get("mrr@10")
            result["segment_user_count"] = seg_data.get("users")
        else:
            result["segment_ndcg_lift"] = None
            result["segment_mrr_lift"] = None
            result["segment_user_count"] = None
    return result


def quality_evidence_paths(namespace: str) -> Tuple[Path, Path]:
    base = Path("analysis")
    return base / "quality_metrics.json", Path("analysis/results") / f"{namespace}_warm_quality.json"


def build_install_overrides(
    alpha: float,
    beta: float,
    gamma: float,
    mmr: float,
    fanout: int,
    profile_boost: Optional[float] = None,
    profile_min_events: Optional[float] = None,
    starter_blend_weight: Optional[float] = None,
) -> Dict[str, str]:
    overrides = {
        "BLEND_ALPHA": str(alpha),
        "BLEND_BETA": str(beta),
        "BLEND_GAMMA": str(gamma),
        "MMR_LAMBDA": str(mmr),
        "POPULARITY_FANOUT": str(fanout),
    }
    if profile_boost is not None:
        overrides["PROFILE_BOOST"] = str(profile_boost)
    if profile_min_events is not None:
        overrides["PROFILE_MIN_EVENTS_FOR_BOOST"] = str(int(round(profile_min_events)))
    if starter_blend_weight is not None:
        overrides["PROFILE_STARTER_BLEND_WEIGHT"] = str(starter_blend_weight)
    return overrides


def apply_profile(
    base_url: str,
    org_id: str,
    namespace: str,
    profile: str,
    overrides: Dict[str, str],
    insecure: bool = False,
    segment: Optional[str] = None,
) -> None:
    fetch_cmd = PROFILE_MANAGER + (["--namespace", namespace] if namespace else [])
    fetch_cmd.extend([
        "fetch",
        "--base-url",
        base_url,
        "--org-id",
        org_id,
        "--profile",
        profile,
    ])
    if insecure:
        fetch_cmd.append("--insecure")
    run_cmd(fetch_cmd)
    profile_file = Path("analysis/env_profiles") / namespace / f"{profile}.json"
    payload = load_json(profile_file)
    config = payload.get("config", {})
    if segment:
        normalized = segment.strip().lower()
        if not normalized:
            raise ValueError("segment name cannot be empty when --segment is provided")
        segment_profiles = config.get("segment_profiles")
        if not isinstance(segment_profiles, dict):
            segment_profiles = {}
        entry = segment_profiles.get(normalized, {})
        for key, value in overrides.items():
            field = CONFIG_FIELD_MAP.get(key)
            if not field:
                continue
            entry[field] = type_cast(value)
        segment_profiles[normalized] = entry
        config["segment_profiles"] = segment_profiles
    else:
        for key, value in overrides.items():
            field = CONFIG_FIELD_MAP.get(key)
            if not field:
                continue
            config[field] = type_cast(value)
    payload["config"] = config
    profile_file.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    apply_cmd = PROFILE_MANAGER + (["--namespace", namespace] if namespace else [])
    apply_cmd.extend([
        "apply",
        "--base-url",
        base_url,
        "--org-id",
        org_id,
        "--profile",
        profile,
        "--author",
        "tuning-harness",
        "--notes",
        f"auto tuning {profile}",
    ])
    if insecure:
        apply_cmd.append("--insecure")
    run_cmd(apply_cmd)


def type_cast(value: str):
    try:
        if "." in value:
            return float(value)
        return int(value)
    except ValueError:
        return value


def parse_number_list(raw: str, cast=float) -> List:
    values = []
    for part in raw.split(","):
        part = part.strip()
        if not part:
            continue
        values.append(cast(part))
    return values


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
        "--users",
        str(users),
        "--events",
        str(events),
    ]
    run_cmd(cmd)


def reset_namespace(base_url: str, org_id: str, namespace: str) -> None:
    cmd = [
        "python",
        "analysis/scripts/reset_namespace.py",
        "--base-url",
        base_url,
        "--org-id",
        org_id,
        "--namespace",
        namespace,
        "--force",
    ]
    run_cmd(cmd)


def run_quality(
    base_url: str,
    org_id: str,
    namespace: str,
    env_file: str,
    sleep_ms: int,
    min_ndcg: float,
    min_mrr: float,
    min_cov: float,
    min_long_tail: float,
    limit_users: int,
    request_timeout: float,
) -> None:
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
    if limit_users and limit_users > 0:
        cmd.extend(["--limit-users", str(limit_users)])
    if min_ndcg is not None:
        cmd.extend(["--min-segment-lift-ndcg", str(min_ndcg)])
    if min_mrr is not None:
        cmd.extend(["--min-segment-lift-mrr", str(min_mrr)])
    if min_cov is not None and min_cov > 0:
        cmd.extend(["--min-catalog-coverage", str(min_cov)])
    if min_long_tail is not None and min_long_tail > 0:
        cmd.extend(["--min-long-tail-share", str(min_long_tail)])
    if request_timeout and request_timeout > 0:
        cmd.extend(["--request-timeout", str(request_timeout)])
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
    parser.add_argument("--segment", help="Optional segment name to tune (updates segment_profiles entry).")
    parser.add_argument("--sleep-ms", type=int, default=200)
    parser.add_argument("--insecure", action="store_true", help="Disable TLS verification for API calls.")
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
    parser.add_argument("--profile-boosts", default="", help="Optional comma-separated profile boost values.")
    parser.add_argument(
        "--profile-min-events",
        default="",
        help="Optional comma-separated values for PROFILE_MIN_EVENTS_FOR_BOOST (integers).",
    )
    parser.add_argument(
        "--starter-blend-weights",
        default="",
        help="Optional comma-separated starter blend weights (PROFILE_STARTER_BLEND_WEIGHT).",
    )
    parser.add_argument("--reset-namespace", action="store_true", help="Reset namespace before each seed.")
    parser.add_argument("--quality-min-segment-lift-ndcg", type=float, default=0.1)
    parser.add_argument("--quality-min-segment-lift-mrr", type=float, default=0.1)
    parser.add_argument("--quality-min-catalog-coverage", type=float, default=0.0)
    parser.add_argument("--quality-min-long-tail-share", type=float, default=0.0)
    parser.add_argument("--quality-limit-users", type=int, default=0)
    parser.add_argument("--quality-request-timeout", type=float, default=60.0)
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

    profile_boost_values = parse_number_list(args.profile_boosts) if args.profile_boosts else []
    profile_min_event_values = parse_number_list(args.profile_min_events, cast=int) if args.profile_min_events else []
    starter_weight_values = parse_number_list(args.starter_blend_weights) if args.starter_blend_weights else []

    if profile_boost_values:
        grid_values["PROFILE_BOOST"] = profile_boost_values
    if profile_min_event_values:
        grid_values["PROFILE_MIN_EVENTS_FOR_BOOST"] = profile_min_event_values
    if starter_weight_values:
        grid_values["PROFILE_STARTER_BLEND_WEIGHT"] = starter_weight_values

    if args.grid:
        iterator = build_grid(grid_values)
    else:
        space = {key: (min(values), max(values)) for key, values in grid_values.items() if values}
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
            profile_boost=config.get("PROFILE_BOOST"),
            profile_min_events=config.get("PROFILE_MIN_EVENTS_FOR_BOOST"),
            starter_blend_weight=config.get("PROFILE_STARTER_BLEND_WEIGHT"),
        )
        if args.reset_namespace:
            reset_namespace(args.base_url, args.org_id, args.namespace)
        apply_profile(
            args.base_url,
            args.org_id,
            args.namespace,
            args.profile_name,
            overrides,
            insecure=args.insecure,
            segment=args.segment,
        )
        run_seed(args.base_url, args.org_id, args.namespace, args.user_count, args.event_count)
        run_quality(
            args.base_url,
            args.org_id,
            args.namespace,
            args.env_file,
            args.sleep_ms,
            args.quality_min_segment_lift_ndcg,
            args.quality_min_segment_lift_mrr,
            args.quality_min_catalog_coverage,
            args.quality_min_long_tail_share,
            args.quality_limit_users,
            args.quality_request_timeout,
        )
        quality_path, stored_path = quality_evidence_paths(args.namespace)
        metrics = extract_metrics(quality_path, args.segment)
        if args.segment and not metrics.get("segment_user_count"):
            raise RuntimeError(
                f"Segment '{args.segment}' missing from evaluation results. "
                "Ensure the dataset seeds users for this segment or retry with a different namespace."
            )
        row = {
            "run": run_id,
            "parameters": overrides,
            "metrics": metrics,
        }
        summary.append(row)
        write_json(run_dir / f"run_{run_id}.json", row)
        try:
            shutil.copy2(quality_path, run_dir / f"quality_{run_id}.json")
        except FileNotFoundError:
            pass

    write_json(run_dir / "summary.json", {"runs": summary})
    print(f"Tuning runs completed. Summary at {run_dir / 'summary.json'}")


if __name__ == "__main__":
    main()
