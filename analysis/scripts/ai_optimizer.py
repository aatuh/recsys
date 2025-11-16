#!/usr/bin/env python3
"""
AI-assisted optimizer for recommendation tuning.

Reads historical tuning runs (analysis/results/tuning_runs/**/summary.json),
fits a lightweight surrogate model, and proposes the next parameter sets.
Falls back to weighted random exploration if scikit-learn is unavailable.
"""

from __future__ import annotations

import argparse
import json
import math
import random
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional, Sequence, Tuple

import numpy as np

try:
    from sklearn.gaussian_process import GaussianProcessRegressor
    from sklearn.gaussian_process.kernels import Matern, WhiteKernel

    HAS_SKLEARN = True
except Exception:  # pragma: no cover - optional dependency
    HAS_SKLEARN = False


DEFAULT_RESULTS_DIR = Path("analysis/results/tuning_runs")
PARAM_KEYS = ["BLEND_ALPHA", "BLEND_BETA", "BLEND_GAMMA", "MMR_LAMBDA", "POPULARITY_FANOUT"]


@dataclass
class Sample:
    params: Dict[str, float]
    target: float
    metadata: Dict[str, float]


def load_samples(results_dir: Path, namespace_filter: Optional[str], objective: str) -> List[Sample]:
    summaries = []
    for entry in results_dir.iterdir():
        if not entry.is_dir():
            continue
        if namespace_filter and namespace_filter not in entry.name:
            continue
        summary_path = entry / "summary.json"
        if summary_path.exists():
            summaries.append(summary_path)
    samples: List[Sample] = []
    for summary in summaries:
        with summary.open("r", encoding="utf-8") as fh:
            data = json.load(fh)
        for run in data.get("runs", []):
            params = {k: float(v) for k, v in run.get("parameters", {}).items() if k in PARAM_KEYS}
            if len(params) != len(PARAM_KEYS):
                continue
            metrics = run.get("metrics") or {}
            value = metrics.get(objective)
            if value is None or math.isnan(value):
                continue
            samples.append(Sample(params=params, target=float(value), metadata=metrics))
    return samples


def random_within(bounds: Dict[str, Tuple[float, float]], rng: random.Random) -> Dict[str, float]:
    suggestion: Dict[str, float] = {}
    for key, (lo, hi) in bounds.items():
        if lo == hi:
            suggestion[key] = lo
        else:
            suggestion[key] = round(rng.uniform(lo, hi), 4)
    return suggestion


def parse_bounds(args: argparse.Namespace) -> Dict[str, Tuple[float, float]]:
    return {
        "BLEND_ALPHA": tuple(map(float, args.alpha_range)),
        "BLEND_BETA": tuple(map(float, args.beta_range)),
        "BLEND_GAMMA": tuple(map(float, args.gamma_range)),
        "MMR_LAMBDA": tuple(map(float, args.mmr_range)),
        "POPULARITY_FANOUT": tuple(map(float, args.fanout_range)),
    }


def fit_surrogate(samples: List[Sample]) -> Optional[GaussianProcessRegressor]:
    if not HAS_SKLEARN or len(samples) < 5:
        return None
    X = np.array([[s.params[k] for k in PARAM_KEYS] for s in samples])
    y = np.array([s.target for s in samples])
    kernel = Matern(length_scale=[0.1, 0.1, 0.05, 0.05, 50], length_scale_bounds=(1e-2, 1e2), nu=2.5)
    kernel += WhiteKernel(noise_level=1e-3)
    gp = GaussianProcessRegressor(kernel=kernel, normalize_y=True, n_restarts_optimizer=3)
    gp.fit(X, y)
    return gp


def score_candidates(
    candidates: List[Dict[str, float]],
    surrogate: Optional[GaussianProcessRegressor],
    samples: List[Sample],
) -> List[Tuple[Dict[str, float], float]]:
    if surrogate is None:
        # simple exploitation/exploration: blend top historical + random noise
        sorted_samples = sorted(samples, key=lambda s: s.target, reverse=True)
        top = sorted_samples[: max(1, len(sorted_samples) // 5 or 1)]
        scored: List[Tuple[Dict[str, float], float]] = []
        for cand in candidates:
            jitter = random.random() * 0.02
            base = top[random.randrange(len(top))].target if top else 0.0
            scored.append((cand, base + jitter))
        return scored

    X = np.array([[cand[k] for k in PARAM_KEYS] for cand in candidates])
    means, stds = surrogate.predict(X, return_std=True)
    scores = means + 0.5 * stds  # exploration bonus
    return list(zip(candidates, scores))


def choose_candidates(
    bounds: Dict[str, Tuple[float, float]],
    samples: List[Sample],
    suggestions: int,
    rng: random.Random,
) -> List[Dict[str, float]]:
    surrogate = fit_surrogate(samples)
    raw_candidates: List[Dict[str, float]] = []
    for _ in range(suggestions * 8):
        raw_candidates.append(random_within(bounds, rng))
    scored = score_candidates(raw_candidates, surrogate, samples)
    scored.sort(key=lambda item: item[1], reverse=True)
    deduped: List[Dict[str, float]] = []
    seen = set()
    for cand, _score in scored:
        key = tuple(round(cand[k], 5) for k in PARAM_KEYS)
        if key in seen:
            continue
        seen.add(key)
        deduped.append(cand)
        if len(deduped) >= suggestions:
            break
    return deduped


def main() -> None:
    parser = argparse.ArgumentParser(description="AI-assisted optimizer for tuning harness.")
    parser.add_argument("--results-dir", default=str(DEFAULT_RESULTS_DIR))
    parser.add_argument("--namespace", help="Optional namespace substring filter (e.g., tune_seg_).")
    parser.add_argument("--objective", default="segment_ndcg_lift", help="Metric key to maximize.")
    parser.add_argument("--suggestions", type=int, default=5, help="Number of parameter sets to propose.")
    parser.add_argument("--seed", type=int, default=2025)
    parser.add_argument("--alpha-range", nargs=2, type=float, default=[0.2, 0.5])
    parser.add_argument("--beta-range", nargs=2, type=float, default=[0.3, 0.6])
    parser.add_argument("--gamma-range", nargs=2, type=float, default=[0.1, 0.4])
    parser.add_argument("--mmr-range", nargs=2, type=float, default=[0.1, 0.4])
    parser.add_argument("--fanout-range", nargs=2, type=float, default=[300, 900])
    parser.add_argument("--output", type=str, help="Optional path to write JSON suggestions.")
    args = parser.parse_args()

    results_dir = Path(args.results_dir)
    if not results_dir.exists():
        raise SystemExit(f"Results directory {results_dir} does not exist.")

    samples = load_samples(results_dir, args.namespace, args.objective)
    if not samples:
        raise SystemExit("No tuning samples found; run the harness first.")

    bounds = parse_bounds(args)
    rng = random.Random(args.seed)
    candidates = choose_candidates(bounds, samples, args.suggestions, rng)

    payload = {
        "objective": args.objective,
        "samples_used": len(samples),
        "suggestions": candidates,
        "notes": "Apply via tuning_harness.py --segment ... using these parameter sets.",
    }
    if args.output:
        Path(args.output).write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
        print(f"Wrote suggestions to {args.output}")
    else:
        print(json.dumps(payload, indent=2))


if __name__ == "__main__":
    main()
