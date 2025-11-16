#!/usr/bin/env python3
"""
Python-based load test for the /v1/recommendations endpoint.

This approximates the k6 script in analysis/load/recommendations_k6.js:
- Drives a configurable request-per-second profile across stages.
- Targets a user pool and surface, mirroring the JS payload.
- Writes a JSON summary with latency percentiles and failure rate.

Usage (similar knobs as the Makefile load-test target):

  python analysis/load/recommendations_load.py \
    --base-url http://localhost:8000 \
    --org-id 00000000-0000-0000-0000-000000000001 \
    --namespace default \
    --surface home \
    --user-pool load_user_0001,load_user_0002 \
    --rps-targets 10,100,1000 \
    --stage-duration 30s \
    --summary-path analysis/results/load_test_summary.json
"""

from __future__ import annotations

import argparse
import json
import math
import os
import threading
import time
from dataclasses import dataclass
from itertools import cycle
from typing import Iterable, List, Sequence

import requests


def parse_duration(raw: str) -> float:
    """Parse a duration like '30s', '1m', '2h' into seconds."""
    raw = raw.strip().lower()
    if not raw:
        return 0.0
    if raw.endswith("ms"):
        return float(raw[:-2]) / 1000.0
    if raw.endswith("s"):
        return float(raw[:-1])
    if raw.endswith("m"):
        return float(raw[:-1]) * 60.0
    if raw.endswith("h"):
        return float(raw[:-1]) * 3600.0
    return float(raw)


def parse_rps_targets(raw: str) -> List[int]:
    return [int(float(part.strip())) for part in raw.split(",") if part.strip()]


def parse_user_pool(raw: str) -> List[str]:
    users = [u.strip() for u in raw.split(",") if u.strip()]
    return users or ["load_user_0001"]


def percentile(values: Sequence[float], p: float) -> float:
    if not values:
        return 0.0
    if p <= 0:
        return values[0]
    if p >= 100:
        return values[-1]
    k = (len(values) - 1) * (p / 100.0)
    f = math.floor(k)
    c = math.ceil(k)
    if f == c:
        return values[int(k)]
    d0 = values[f] * (c - k)
    d1 = values[c] * (k - f)
    return d0 + d1


@dataclass
class Metrics:
    latencies_ms: List[float]
    failures: int
    count: int

    def __init__(self) -> None:
        self.latencies_ms = []
        self.failures = 0
        self.count = 0
        self._lock = threading.Lock()

    def record(self, latency_ms: float, success: bool) -> None:
        with self._lock:
            self.latencies_ms.append(latency_ms)
            self.count += 1
            if not success:
                self.failures += 1

    def to_summary(self) -> dict:
        latencies = sorted(self.latencies_ms)
        failure_rate = (self.failures / self.count) if self.count else 0.0
        return {
            "http_req_duration_ms": {
                "p50": percentile(latencies, 50.0),
                "p95": percentile(latencies, 95.0),
                "p99": percentile(latencies, 99.0),
            },
            "http_req_failed_rate": failure_rate,
            "iterations": self.count,
        }


def make_payload(namespace: str, user_id: str, surface: str, k: int, include_reasons: bool) -> dict:
    return {
        "namespace": namespace,
        "user_id": user_id,
        "k": k,
        "include_reasons": include_reasons,
        "context": {"surface": surface},
    }


def run_request(
    base_url: str,
    org_id: str,
    namespace: str,
    surface: str,
    k: int,
    include_reasons: bool,
    user_cycle: Iterable[str],
    user_lock: threading.Lock,
    timeout_s: float,
    metrics: Metrics,
) -> None:
    with user_lock:
        user_id = next(user_cycle)
    payload = make_payload(namespace, user_id, surface, k, include_reasons)
    start = time.monotonic()
    ok = False
    try:
        resp = requests.post(
            f"{base_url.rstrip('/')}/v1/recommendations",
            json=payload,
            headers={"X-Org-ID": org_id},
            timeout=timeout_s,
        )
        ok = resp.status_code == 200 and bool(resp.json().get("items"))
    except Exception:
        ok = False
    elapsed_ms = (time.monotonic() - start) * 1000.0
    metrics.record(elapsed_ms, ok)


def run_stage(
    rps: int,
    duration_s: float,
    base_url: str,
    org_id: str,
    namespace: str,
    surface: str,
    k: int,
    include_reasons: bool,
    user_cycle: Iterable[str],
    user_lock: threading.Lock,
    timeout_s: float,
    metrics: Metrics,
    max_concurrency: int,
) -> None:
    if rps <= 0 or duration_s <= 0:
        return
    total_requests = int(rps * duration_s)
    if total_requests <= 0:
        return

    import concurrent.futures

    start = time.monotonic()
    with concurrent.futures.ThreadPoolExecutor(max_workers=max_concurrency) as executor:
        futures = []
        for i in range(total_requests):
            scheduled = start + (i / float(rps))
            delay = scheduled - time.monotonic()
            if delay > 0:
                time.sleep(delay)
            futures.append(
                executor.submit(
                    run_request,
                    base_url,
                    org_id,
                    namespace,
                    surface,
                    k,
                    include_reasons,
                    user_cycle,
                    user_lock,
                    timeout_s,
                    metrics,
                )
            )
        for f in futures:
            try:
                f.result()
            except Exception:
                # Errors are already reflected in metrics; ignore here.
                pass


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Python load test for /v1/recommendations.")
    parser.add_argument("--base-url", default=os.environ.get("BASE_URL", "http://localhost:8000"))
    parser.add_argument(
        "--org-id",
        default=os.environ.get("ORG_ID", "00000000-0000-0000-0000-000000000001"),
    )
    parser.add_argument("--namespace", default=os.environ.get("NAMESPACE", "default"))
    parser.add_argument("--surface", default=os.environ.get("SURFACE", "home"))
    parser.add_argument("--user-pool", default=os.environ.get("USER_POOL", "load_user_0001,load_user_0002,load_user_0003,load_user_0004,load_user_0005"))
    parser.add_argument("--k", type=int, default=int(os.environ.get("K", "20")))
    parser.add_argument(
        "--include-reasons",
        type=str,
        default=os.environ.get("INCLUDE_REASONS", "true"),
        help="Whether to request reasons (true/false).",
    )
    parser.add_argument(
        "--summary-path",
        default=os.environ.get("SUMMARY_PATH", "analysis/results/load_test_summary.json"),
    )
    parser.add_argument(
        "--stage-duration",
        default=os.environ.get("STAGE_DURATION", "30s"),
        help="Stage duration (e.g., 30s, 1m).",
    )
    parser.add_argument(
        "--rps-targets",
        default=os.environ.get("RPS_TARGETS", "10,100,1000"),
        help="Comma-separated RPS targets (e.g., 10,100,1000).",
    )
    parser.add_argument(
        "--http-timeout",
        default=os.environ.get("HTTP_TIMEOUT", "30s"),
        help="Per-request timeout (e.g., 30s).",
    )
    parser.add_argument(
        "--max-concurrency",
        type=int,
        default=int(os.environ.get("MAX_CONCURRENCY", "200")),
        help="Maximum concurrent requests.",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    include_reasons = str(args.include_reasons).strip().lower() == "true"
    user_pool = parse_user_pool(args.user_pool)
    stages_rps = parse_rps_targets(args.rps_targets)
    stage_duration_s = parse_duration(args.stage_duration)
    timeout_s = parse_duration(args.http_timeout)

    user_cycle = cycle(user_pool)
    user_lock = threading.Lock()
    metrics = Metrics()

    for rps in stages_rps:
        run_stage(
            rps=rps,
            duration_s=stage_duration_s,
            base_url=args.base_url,
            org_id=args.org_id,
            namespace=args.namespace,
            surface=args.surface,
            k=args.k,
            include_reasons=include_reasons,
            user_cycle=user_cycle,
            user_lock=user_lock,
            timeout_s=timeout_s,
            metrics=metrics,
            max_concurrency=args.max_concurrency,
        )

    summary = {
        "created_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "base_url": args.base_url,
        "namespace": args.namespace,
        "surface": args.surface,
        "stages": [
            {"target_rps": rps, "duration": args.stage_duration}
            for rps in stages_rps
        ],
        "metrics": metrics.to_summary(),
    }

    summary_json = json.dumps(summary, indent=2)
    print(summary_json)

    summary_path = args.summary_path
    os.makedirs(os.path.dirname(summary_path), exist_ok=True)
    with open(summary_path, "w", encoding="utf-8") as fh:
        fh.write(summary_json + "\n")


if __name__ == "__main__":
    main()

