#!/usr/bin/env python3
"""
Orchestrate a full RecSys simulation run:
  1) Configure env (optional profile + overrides)
  2) Restart API container and wait for /health
  3) Reset namespace
  4) Seed dataset (optionally from fixture)
  5) Run quality evaluation
  6) Run scenario suite

All steps record metadata so runs are auditable.
"""

from __future__ import annotations

import argparse
import json
import shutil
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, List, Optional
from types import SimpleNamespace

import yaml

from configure_env import find_profile_for_namespace, load_profiles
from env_utils import compute_env_hash
from guardrails import load_guardrails, resolve_guardrails


DEFAULT_BASE_URL = "https://api.pepe.local"
DEFAULT_NAMESPACE = "default"
DEFAULT_ORG_ID = "00000000-0000-0000-0000-000000000001"
DEFAULT_ENV_FILE = "api/.env"
DEFAULT_PROFILES_FILE = "config/profiles.yml"
ARTIFACT_SOURCES = [
    ("quality_metrics.json", "analysis/quality_metrics.json"),
    ("scenario_summary.json", "analysis/evidence/scenario_summary.json"),
    ("scenarios.csv", "analysis/scenarios.csv"),
    ("seed_manifest.json", "analysis/evidence/seed_manifest.json"),
    ("seed_segments.json", "analysis/evidence/seed_segments.json"),
    ("recommendation_dump.json", "analysis/results/recommendation_dump.json"),
    ("exposure_dashboard.json", "analysis/results/exposure_dashboard.json"),
    ("rules_effect_sample.json", "analysis/results/rules_effect_sample.json"),
    ("rules_block_sample.json", "analysis/results/rules_block_sample.json"),
    ("load_test_summary.json", "analysis/results/load_test_summary.json"),
]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Run an end-to-end RecSys simulation.")
    parser.add_argument("--customer", help="Customer or experiment label.")
    parser.add_argument("--batch-file", help="YAML/JSON manifest describing multiple simulations.")
    parser.add_argument("--batch-name", help="Optional batch label (defaults to manifest filename).")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL, help="Recommendation API base URL.")
    parser.add_argument("--namespace", default=DEFAULT_NAMESPACE, help="Namespace to target.")
    parser.add_argument("--org-id", default=DEFAULT_ORG_ID, help="Org ID / tenant identifier.")
    parser.add_argument("--env-file", default=DEFAULT_ENV_FILE, help="Env file to rewrite (default: api/.env).")
    parser.add_argument(
        "--env-profile",
        help="Optional profile to load via configure_env.py before overrides (namespace mappings from --profiles-file take precedence).",
    )
    parser.add_argument(
        "--env-override",
        dest="env_overrides",
        action="append",
        default=[],
        help="Env override in KEY=VALUE form (can be repeated).",
    )
    parser.add_argument(
        "--profiles-file",
        default=DEFAULT_PROFILES_FILE,
        help="Profiles registry (YAML/JSON) describing namespace→profile mappings.",
    )
    parser.add_argument("--fixture-path", help="Optional JSON fixture for seeding.")
    parser.add_argument("--item-count", type=int, default=320, help="Synthetic item count when not using fixtures.")
    parser.add_argument("--user-count", type=int, default=120, help="Synthetic user count when not using fixtures.")
    parser.add_argument("--min-events", type=int, default=5000, help="Minimum events when generating synthetic data.")
    parser.add_argument("--quality-limit-users", type=int, default=0, help="Optional cap for quality eval users.")
    parser.add_argument("--quality-sleep-ms", type=int, default=120, help="Sleep between quality eval calls.")
    parser.add_argument("--min-segment-lift-ndcg", type=float, default=0.1, help="Segment NDCG lift guardrail.")
    parser.add_argument("--min-segment-lift-mrr", type=float, default=0.1, help="Segment MRR lift guardrail.")
    parser.add_argument("--min-catalog-coverage", type=float, default=0.0, help="Min catalog coverage guardrail.")
    parser.add_argument("--min-long-tail-share", type=float, default=0.0, help="Min long-tail share guardrail.")
    parser.add_argument("--s7-min-avg-mrr", type=float, default=0.2, help="Scenario S7 MRR guardrail.")
    parser.add_argument("--s7-min-avg-categories", type=float, default=4.0, help="Scenario S7 category guardrail.")
    parser.add_argument("--skip-reset", action="store_true", help="Skip namespace reset step.")
    parser.add_argument("--skip-seed", action="store_true", help="Skip dataset seeding.")
    parser.add_argument("--skip-quality", action="store_true", help="Skip quality evaluation.")
    parser.add_argument("--skip-scenarios", action="store_true", help="Skip scenario suite.")
    parser.add_argument("--dry-run", action="store_true", help="Print planned steps without executing.")
    parser.add_argument(
        "--reports-dir",
        default="analysis/reports",
        help="Directory to store simulation metadata (default: %(default)s).",
    )
    parser.add_argument(
        "--guardrails-file",
        default="guardrails.yml",
        help="Optional guardrails config file (YAML/JSON). Set to blank to disable.",
    )
    args = parser.parse_args()
    if not args.customer and not args.batch_file:
        parser.error("Either --customer or --batch-file must be provided.")
    return args


def load_manifest(path: str) -> List[Dict]:
    file_path = Path(path)
    data = file_path.read_text(encoding="utf-8")
    if file_path.suffix.lower() in {".yaml", ".yml"}:
        manifest = yaml.safe_load(data)
    else:
        manifest = json.loads(data)
    if isinstance(manifest, dict) and "runs" in manifest:
        manifest = manifest["runs"]
    if not isinstance(manifest, list):
        raise ValueError("Manifest must be a list or contain a 'runs' list.")
    return manifest


def prepare_run_args(
    base_args: argparse.Namespace,
    overrides: Optional[Dict],
    manifest_dir: Optional[Path],
    guardrails_cfg: Optional[Dict],
) -> SimpleNamespace:
    base_config = vars(base_args).copy()
    overrides = overrides or {}
    base_config.update(overrides)
    for key in ("batch_file", "batch_name"):
        base_config.pop(key, None)
    env_overrides = overrides.get("env_overrides", base_config.get("env_overrides"))
    if env_overrides is None:
        env_overrides = []
    if isinstance(env_overrides, str):
        env_overrides = [env_overrides]
    base_config["env_overrides"] = list(env_overrides)
    fixture_path = base_config.get("fixture_path")
    if fixture_path and manifest_dir and not Path(fixture_path).is_absolute():
        base_config["fixture_path"] = str((manifest_dir / fixture_path).resolve())
    customer = base_config.get("customer")
    if not customer:
        raise ValueError("Each simulation run must set 'customer'.")
    sim_args = SimpleNamespace(**base_config)
    apply_guardrails(sim_args, guardrails_cfg)
    apply_profile_mapping(sim_args, base_args.profiles_file)
    return sim_args


def apply_guardrails(sim_args: SimpleNamespace, guardrails_cfg: Optional[Dict]) -> None:
    if not guardrails_cfg or not sim_args.customer:
        return
    resolved = resolve_guardrails(guardrails_cfg, sim_args.customer, sim_args.namespace)
    quality = resolved.get("quality", {})
    scenarios = resolved.get("scenarios", {})
    for attr, key in (
        ("min_segment_lift_ndcg", "min_segment_lift_ndcg"),
        ("min_segment_lift_mrr", "min_segment_lift_mrr"),
        ("min_catalog_coverage", "min_catalog_coverage"),
        ("min_long_tail_share", "min_long_tail_share"),
    ):
        if key in quality:
            setattr(sim_args, attr, quality[key])
    for attr, key in (("s7_min_avg_mrr", "s7_min_avg_mrr"), ("s7_min_avg_categories", "s7_min_avg_categories")):
        if key in scenarios:
            setattr(sim_args, attr, scenarios[key])


def apply_profile_mapping(sim_args: SimpleNamespace, profiles_file: str) -> None:
    if getattr(sim_args, "env_profile", None):
        return
    if not sim_args.namespace:
        return
    registry = load_profiles(profiles_file)
    if not registry:
        return
    profile = find_profile_for_namespace(registry, sim_args.namespace)
    if profile:
        sim_args.env_profile = profile


def run_cmd(cmd: List[str], env: Optional[Dict[str, str]] = None) -> None:
    print(f"$ {' '.join(cmd)}")
    subprocess.run(cmd, check=True, env=env)


def configure_env(args: SimpleNamespace, note: str, dry_run: bool = False) -> None:
    if not args.env_profile and not args.env_overrides:
        return
    cmd = [
        "python",
        "analysis/scripts/configure_env.py",
        "--env-file",
        args.env_file,
    ]
    if args.env_profile:
        cmd.extend(["--profile", args.env_profile])
    if args.namespace:
        cmd.extend(["--namespace", args.namespace])
    if args.profiles_file:
        cmd.extend(["--profiles-file", args.profiles_file])
    for override in args.env_overrides:
        cmd.extend(["--set", override])
    cmd.extend(["--note", note])
    run_cmd(cmd)


def restart_api(args: SimpleNamespace) -> None:
    cmd = [
        "python",
        "analysis/scripts/restart_api.py",
        "--base-url",
        args.base_url,
    ]
    run_cmd(cmd)


def reset_namespace(args: SimpleNamespace) -> None:
    if args.skip_reset:
        print("Skipping namespace reset (--skip-reset).")
        return
    cmd = [
        "python",
        "analysis/scripts/reset_namespace.py",
        "--base-url",
        args.base_url,
        "--org-id",
        args.org_id,
        "--namespace",
        args.namespace,
        "--force",
    ]
    run_cmd(cmd)


def seed_dataset(args: SimpleNamespace) -> None:
    if args.skip_seed:
        print("Skipping dataset seeding (--skip-seed).")
        return
    cmd = [
        "python",
        "analysis/scripts/seed_dataset.py",
        "--base-url",
        args.base_url,
        "--namespace",
        args.namespace,
        "--org-id",
        args.org_id,
        "--item-count",
        str(args.item_count),
        "--user-count",
        str(args.user_count),
        "--min-events",
        str(args.min_events),
    ]
    if args.fixture_path:
        cmd.extend(["--fixture-path", args.fixture_path])
    run_cmd(cmd)


def run_quality(args: SimpleNamespace) -> None:
    if args.skip_quality:
        print("Skipping quality evaluation (--skip-quality).")
        return
    cmd = [
        "python",
        "analysis/scripts/run_quality_eval.py",
        "--base-url",
        args.base_url,
        "--namespace",
        args.namespace,
        "--org-id",
        args.org_id,
        "--env-file",
        args.env_file,
        "--limit-users",
        str(args.quality_limit_users),
        "--sleep-ms",
        str(args.quality_sleep_ms),
        "--min-segment-lift-ndcg",
        str(args.min_segment_lift_ndcg),
        "--min-segment-lift-mrr",
        str(args.min_segment_lift_mrr),
    ]
    if args.min_catalog_coverage > 0:
        cmd.extend(["--min-catalog-coverage", str(args.min_catalog_coverage)])
    if args.min_long_tail_share > 0:
        cmd.extend(["--min-long-tail-share", str(args.min_long_tail_share)])
    run_cmd(cmd)


def run_scenarios(args: SimpleNamespace) -> None:
    if args.skip_scenarios:
        print("Skipping scenario suite (--skip-scenarios).")
        return
    cmd = [
        "python",
        "analysis/scripts/run_scenarios.py",
        "--base-url",
        args.base_url,
        "--namespace",
        args.namespace,
        "--org-id",
        args.org_id,
        "--env-file",
        args.env_file,
        "--s7-min-avg-mrr",
        str(args.s7_min_avg_mrr),
        "--s7-min-avg-categories",
        str(args.s7_min_avg_categories),
    ]
    run_cmd(cmd)


def copy_artifacts(report_dir: Path) -> List[Dict[str, str]]:
    artifacts_dir = report_dir / "artifacts"
    copied: List[Dict[str, str]] = []
    for name, src in ARTIFACT_SOURCES:
        src_path = Path(src)
        if not src_path.exists():
            continue
        dest_path = artifacts_dir / name
        dest_path.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src_path, dest_path)
        copied.append({"name": name, "path": str(dest_path.relative_to(report_dir))})
    return copied


def load_json_safe(path: Optional[Path]) -> Optional[Dict]:
    if path is None or not path.exists():
        return None
    try:
        with open(path, "r", encoding="utf-8") as fh:
            return json.load(fh)
    except json.JSONDecodeError:
        return None


def write_summary(report_dir: Path, metadata: Dict, artifacts: List[Dict[str, str]]) -> None:
    artifacts_map = {art["name"]: report_dir / art["path"] for art in artifacts}
    quality = load_json_safe(artifacts_map.get("quality_metrics.json", Path()))
    scenarios = load_json_safe(artifacts_map.get("scenario_summary.json", Path()))
    exposure = load_json_safe(artifacts_map.get("exposure_dashboard.json", Path()))
    load_summary = load_json_safe(artifacts_map.get("load_test_summary.json", Path()))
    rules_effect = load_json_safe(artifacts_map.get("rules_effect_sample.json", Path()))

    lines = [
        "# Simulation Report",
        "",
        f"- Customer: `{metadata['customer']}`",
        f"- Timestamp: {metadata['timestamp']}",
        f"- Base URL: {metadata['base_url']}",
        f"- Namespace: {metadata['namespace']}",
        f"- Org ID: {metadata['org_id']}",
        f"- Env Hash: `{metadata['env_hash']}`",
        "",
    ]

    if quality:
        overall = quality.get("overall", {})
        system = overall.get("system", {})
        baseline = overall.get("baseline", {})
        lifts = overall.get("lift", {})
        lines.extend(
            [
                "## Quality Highlights",
                "",
                f"- System NDCG@10: {system.get('ndcg@10'):.4f} (baseline {baseline.get('ndcg@10'):.4f}, lift {lifts.get('ndcg@10'):.2f})"
                if system and baseline and lifts
                else "- Quality data available (see artifacts).",
            ]
        )
        segments = quality.get("segments", {})
        if segments:
            lines.append("")
            lines.append("| Segment | Users | NDCG lift | MRR lift |")
            lines.append("|---------|-------|-----------|----------|")
            for name, seg in segments.items():
                lift = seg.get("lift", {})
                lines.append(
                    f"| {name} | {seg.get('users', 0)} | {lift.get('ndcg@10', 0):.2f} | {lift.get('mrr@10', 0):.2f} |"
                )
            lines.append("")
    else:
        lines.append("## Quality Highlights")
        lines.append("")
        lines.append("- No quality metrics captured.")
        lines.append("")

    if scenarios:
        results = scenarios.get("results", [])
        passes = sum(1 for r in results if r.get("passed"))
        lines.append("## Scenario Suite")
        lines.append("")
        lines.append(f"- {passes}/{len(results)} scenarios passed.")
        lines.append("")
        lines.append("| ID | Name | Result |")
        lines.append("|----|------|--------|")
        for r in results:
            lines.append(f"| {r.get('id')} | {r.get('name')} | {'PASS' if r.get('passed') else 'FAIL'} |")
        lines.append("")
    else:
        lines.append("## Scenario Suite")
        lines.append("")
        lines.append("- Scenario summary not available.")
        lines.append("")

    if exposure:
        names = exposure.get("namespaces", {})
        lines.append("## Exposure Dashboard")
        lines.append("")
        if names:
            lines.append("| Namespace | Ratio | Status | Max/Mean |")
            lines.append("|-----------|-------|--------|----------|")
            for ns, stats in names.items():
                ratio = stats.get("exposure_ratio") or stats.get("brand_ratio") or 0
                status = stats.get("status", "unknown")
                lines.append(
                    f"| {ns} | {ratio:.2f} | {status.upper()} | "
                    f"{stats.get('max_brand_exposure', 0):.0f} / {stats.get('mean_brand_exposure', 0):.1f} |"
                )
            lines.append("")
        else:
            lines.append("- Exposure dashboard file present but no namespace data found.")
            lines.append("")

    if load_summary:
        metrics = load_summary.get("metrics", {})
        durations = metrics.get("http_req_duration_ms", {})
        lines.append("## Load Test Summary")
        lines.append("")
        lines.append(
            f"- Iterations: {metrics.get('iterations', 'n/a')} | "
            f"Failure rate: {metrics.get('http_req_failed_rate', 0):.4f}"
        )
        if durations:
            lines.append(
                f"- Latency p50/p95/p99 (ms): "
                f"{durations.get('p50', 0):.1f} / {durations.get('p95', 0):.1f} / {durations.get('p99', 0):.1f}"
            )
        lines.append("")

    if rules_effect:
        policy = (
            rules_effect.get("trace", {})
            .get("extras", {})
            .get("policy", {})
        )
        if policy:
            lines.append("## Rules Evidence")
            lines.append("")
            lines.append(
                "- Rule hits: "
                f"block={policy.get('rule_block_count', 0)}, "
                f"boost={policy.get('rule_boost_count', 0)}, "
                f"pin={policy.get('rule_pin_count', 0)}; "
                f"boost exposure={policy.get('rule_boost_exposure', 0)}, "
                f"pin exposure={policy.get('rule_pin_exposure', 0)}"
            )
            lines.append("")

    lines.append("## Artifacts")
    lines.append("")
    for art in artifacts:
        lines.append(f"- {art['name']}: `{art['path']}`")

    with open(report_dir / "README.md", "w", encoding="utf-8") as fh:
        fh.write("\n".join(lines))


def write_report(args: argparse.Namespace, steps: List[Dict[str, str]], env_hash: str) -> str:
    timestamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    report_dir = Path(args.reports_dir) / args.customer / timestamp
    report_dir.mkdir(parents=True, exist_ok=True)
    artifacts = copy_artifacts(report_dir)
    metadata = {
        "customer": args.customer,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "base_url": args.base_url,
        "namespace": args.namespace,
        "org_id": args.org_id,
        "env_file": args.env_file,
        "env_hash": env_hash,
        "steps": steps,
        "artifacts": artifacts,
    }
    with open(report_dir / "simulation_metadata.json", "w", encoding="utf-8") as fh:
        json.dump(metadata, fh, indent=2)
    write_summary(report_dir, metadata, artifacts)
    return str(report_dir)


def run_single(sim_args: SimpleNamespace) -> Dict:
    steps: List[Dict[str, str]] = []
    note = f"simulation:{sim_args.customer}:{datetime.now(timezone.utc).isoformat()}"

    if sim_args.dry_run:
        print(f"[dry-run] Would execute simulation for {sim_args.customer}")
        return {"customer": sim_args.customer, "report_path": "", "env_hash": ""}

    configure_env(sim_args, note, dry_run=sim_args.dry_run)
    steps.append(
        {
            "step": "configure_env",
            "note": note,
            "profile": sim_args.env_profile or "",
            "overrides": ",".join(sim_args.env_overrides),
        }
    )

    restart_api(sim_args)
    steps.append({"step": "restart_api"})

    reset_namespace(sim_args)
    steps.append({"step": "reset_namespace", "skipped": str(sim_args.skip_reset)})

    seed_dataset(sim_args)
    steps.append({"step": "seed_dataset", "fixture": sim_args.fixture_path or "", "skipped": str(sim_args.skip_seed)})

    run_quality(sim_args)
    steps.append({"step": "run_quality", "skipped": str(sim_args.skip_quality)})

    run_scenarios(sim_args)
    steps.append({"step": "run_scenarios", "skipped": str(sim_args.skip_scenarios)})

    env_hash = compute_env_hash(sim_args.env_file)
    report_path = write_report(sim_args, steps, env_hash)
    scenario_file = Path(report_path) / "artifacts" / "scenario_summary.json"
    scenario_data = load_json_safe(scenario_file)
    scenario_status = "unknown"
    if scenario_data:
        results = scenario_data.get("results", [])
        if results:
            scenario_status = "pass" if all(r.get("passed") for r in results) else "fail"
    return {
        "customer": sim_args.customer,
        "report_path": report_path,
        "env_hash": env_hash,
        "scenario_status": scenario_status,
    }


def main() -> None:
    args = parse_args()
    if args.batch_file:
        manifest = load_manifest(args.batch_file)
        manifest_dir = Path(args.batch_file).resolve().parent
        batch_results: List[Dict[str, str]] = []
        guardrails_cfg = load_guardrails(args.guardrails_file) if args.guardrails_file else None
        for idx, entry in enumerate(manifest, start=1):
            run_args = prepare_run_args(args, entry, manifest_dir, guardrails_cfg)
            print(f"=== [{idx}/{len(manifest)}] Running simulation for {run_args.customer} ===")
            result = run_single(run_args)
            batch_results.append(result)
        summary_path = write_batch_summary(args, batch_results)
        print(f"Batch summary written to {summary_path}")
    else:
        guardrails_cfg = load_guardrails(args.guardrails_file) if args.guardrails_file else None
        run_args = prepare_run_args(args, {}, None, guardrails_cfg)
        result = run_single(run_args)
        if result["report_path"]:
            print(f"Simulation metadata written to {result['report_path']}")


if __name__ == "__main__":
    main()
def write_batch_summary(args: argparse.Namespace, runs: List[Dict[str, str]]) -> str:
    batch_name = args.batch_name or (Path(args.batch_file).stem if args.batch_file else "batch")
    timestamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    payload = {
        "batch": batch_name,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "manifest": args.batch_file,
        "runs": runs,
    }
    summary_dir = Path(args.reports_dir) / "batches"
    summary_dir.mkdir(parents=True, exist_ok=True)
    path = summary_dir / f"{batch_name}_{timestamp}.json"
    with open(path, "w", encoding="utf-8") as fh:
        json.dump(payload, fh, indent=2)
    return str(path)
