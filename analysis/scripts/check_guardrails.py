#!/usr/bin/env python3
"""
Validate tuning run summaries against guardrail thresholds.

Scans analysis/results/tuning_runs/**/summary.json and ensures that:
- overall ndcg/mrr lifts meet minimums
- when a segment is specified in run metrics, its ndcg/mrr lift meets minimums
- catalog coverage and long-tail share meet minimums (if provided)
"""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Dict, List, Optional


DEFAULT_RESULTS_DIR = Path("analysis/results/tuning_runs")


def load_summary(path: Path) -> Dict:
    with path.open("r", encoding="utf-8") as fh:
        return json.load(fh)


def check_run(
    run: Dict,
    ndcg_min: float,
    mrr_min: float,
    coverage_min: float,
    segment_ndcg_min: Optional[float],
    segment_mrr_min: Optional[float],
) -> List[str]:
    metrics = run.get("metrics") or {}
    failures: List[str] = []

    def fmt(value: Optional[float]) -> str:
        return "n/a" if value is None else f"{value:.3f}"

    ndcg = metrics.get("ndcg_lift")
    if ndcg is None or ndcg < ndcg_min:
        failures.append(
            f"{run['run']}: overall ndcg_lift={fmt(ndcg)} < {ndcg_min:.3f}"
        )
    mrr = metrics.get("mrr_lift")
    if mrr is None or mrr < mrr_min:
        failures.append(
            f"{run['run']}: overall mrr_lift={fmt(mrr)} < {mrr_min:.3f}"
        )
    coverage = metrics.get("catalog_coverage")
    if coverage_min > 0 and (coverage is None or coverage < coverage_min):
        failures.append(
            f"{run['run']}: coverage={fmt(coverage)} < {coverage_min:.3f}"
        )

    segment_ndcg = metrics.get("segment_ndcg_lift")
    segment_mrr = metrics.get("segment_mrr_lift")
    if segment_ndcg_min is not None:
        if segment_ndcg is None or segment_ndcg < segment_ndcg_min:
            failures.append(
                f"{run['run']}: segment ndcg_lift={fmt(segment_ndcg)} < {segment_ndcg_min:.3f}"
            )
    if segment_mrr_min is not None:
        if segment_mrr is None or segment_mrr < segment_mrr_min:
            failures.append(
                f"{run['run']}: segment mrr_lift={fmt(segment_mrr)} < {segment_mrr_min:.3f}"
            )
    return failures


def collect_summaries(base_dir: Path, namespace_filter: Optional[str]) -> List[Path]:
    if not base_dir.exists():
        return []
    summaries: List[Path] = []
    for entry in sorted(base_dir.iterdir()):
        if not entry.is_dir():
            continue
        if namespace_filter and namespace_filter not in entry.name:
            continue
        candidate = entry / "summary.json"
        if candidate.exists():
            summaries.append(candidate)
    return summaries


def main() -> None:
    parser = argparse.ArgumentParser(description="Check tuning guardrails.")
    parser.add_argument("--results-dir", default=str(DEFAULT_RESULTS_DIR))
    parser.add_argument("--namespace", help="Optional namespace substring filter (e.g., tune_seg_).")
    parser.add_argument("--min-ndcg", type=float, default=0.1)
    parser.add_argument("--min-mrr", type=float, default=0.1)
    parser.add_argument("--min-coverage", type=float, default=0.0)
    parser.add_argument("--min-segment-ndcg", type=float, default=0.1)
    parser.add_argument("--min-segment-mrr", type=float, default=0.1)
    args = parser.parse_args()

    base_dir = Path(args.results_dir)
    summaries = collect_summaries(base_dir, args.namespace)
    if not summaries:
        raise SystemExit(f"No summary.json files found under {base_dir} (filter={args.namespace!r}).")

    failures: List[str] = []
    for summary_path in summaries:
        data = load_summary(summary_path)
        runs = data.get("runs") or []
        if not runs:
            failures.append(f"{summary_path}: no runs present.")
            continue
        for run in runs:
            failures.extend(
                check_run(
                    run,
                    args.min_ndcg,
                    args.min_mrr,
                    args.min_coverage,
                    args.min_segment_ndcg,
                    args.min_segment_mrr,
                )
            )

    if failures:
        print("Guardrail check failed:")
        for failure in failures:
            print(f"- {failure}")
        raise SystemExit(1)
    print(f"Guardrail check passed for {len(summaries)} summaries.")


if __name__ == "__main__":
    main()
