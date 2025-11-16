#!/usr/bin/env python3
"""
Seed Recsys evaluation dataset via public API.

Creates:
  - Items (>= 320) across 8 categories with long-tail distribution.
  - Users (>= 120) mapped to behavioral segments.
  - Events (>= 5k) spanning multiple weeks with segment-specific patterns.

All writes happen via HTTPS to BASE_URL using X-Org-ID header.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import random
import time
from dataclasses import dataclass
from collections import Counter
from datetime import datetime, timedelta, timezone
from typing import Dict, Iterable, List, Optional, Tuple

import requests
import urllib3

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

DEFAULT_BASE_URL = "http://localhost:8000"
DEFAULT_NAMESPACE = "default"
DEFAULT_ORG_ID = "00000000-0000-0000-0000-000000000001"

BATCH_SIZE_ITEMS = 50
BATCH_SIZE_USERS = 100
BATCH_SIZE_EVENTS = 500

SEED = 20251103
EMBEDDING_DIMS = 384


@dataclass
class SegmentProfile:
    name: str
    description: str
    category_weights: Dict[str, float]
    activity_level: float  # relative number of interactions per user
    purchase_bias: float  # probability modifier for purchase events
    diversity_bias: float  # lower value -> more niche focus


CATEGORIES = [
    "Electronics",
    "Books",
    "Home",
    "Fitness",
    "Fashion",
    "Beauty",
    "Gourmet",
    "Outdoors",
]

BRANDS: Dict[str, List[str]] = {
    "Electronics": ["AcmeTech", "Voltify", "Nimbus", "CircuitLab"],
    "Books": ["LeafPress", "Quanta", "PioneerWords", "AtlasReads"],
    "Home": ["CozyNest", "HavenCraft", "UrbanDwelling", "BrightLiving"],
    "Fitness": ["PulseGear", "Stride", "IronWorks", "ZenMotion"],
    "Fashion": ["AuraThreads", "MidnightLane", "CoutureX", "LoomLab"],
    "Beauty": ["LuxeGlow", "PureBloom", "RadiantHue", "VelvetSky"],
    "Gourmet": ["SavoryCo", "Epicurean", "HarvestBite", "SpiceTrail"],
    "Outdoors": ["SummitPeak", "Trailblaze", "Skyline", "TerraQuest"],
}

TAGS_BY_CATEGORY = {
    "Electronics": ["electronics", "gadgets", "smart"],
    "Books": ["books", "reading", "literature"],
    "Home": ["home", "decor", "living"],
    "Fitness": ["fitness", "health", "active"],
    "Fashion": ["fashion", "style", "apparel"],
    "Beauty": ["beauty", "skincare", "wellness"],
    "Gourmet": ["food", "gourmet", "kitchen"],
    "Outdoors": ["outdoors", "adventure", "travel"],
}

SEGMENTS: List[SegmentProfile] = [
    SegmentProfile(
        name="power_users",
        description="High-activity shoppers across electronics and fitness with broad interests.",
        category_weights={"Electronics": 0.35, "Fitness": 0.2,
                          "Home": 0.15, "Fashion": 0.1, "Outdoors": 0.2},
        activity_level=1.4,
        purchase_bias=1.2,
        diversity_bias=0.7,
    ),
    SegmentProfile(
        name="new_users",
        description="Recently onboarded users exploring popular categories with limited actions.",
        category_weights={"Electronics": 0.25, "Books": 0.2,
                          "Home": 0.2, "Fashion": 0.2, "Beauty": 0.15},
        activity_level=0.6,
        purchase_bias=0.4,
        diversity_bias=1.0,
    ),
    SegmentProfile(
        name="niche_readers",
        description="Book-centric users with niche gourmet interests.",
        category_weights={"Books": 0.6, "Gourmet": 0.25, "Home": 0.15},
        activity_level=1.0,
        purchase_bias=0.8,
        diversity_bias=0.4,
    ),
    SegmentProfile(
        name="trend_seekers",
        description="Fashion and beauty enthusiasts with high novelty appetite.",
        category_weights={"Fashion": 0.4, "Beauty": 0.35,
                          "Electronics": 0.1, "Home": 0.15},
        activity_level=1.1,
        purchase_bias=0.9,
        diversity_bias=1.2,
    ),
    SegmentProfile(
        name="weekend_adventurers",
        description="Outdoor-focused users with taste for fitness and gourmet gear.",
        category_weights={"Outdoors": 0.5, "Fitness": 0.25,
                          "Gourmet": 0.15, "Electronics": 0.1},
        activity_level=0.9,
        purchase_bias=0.7,
        diversity_bias=0.8,
    ),
]

ITEM_DESCRIPTORS = [
    "signature",
    "curated",
    "smart",
    "limited",
    "artisan",
    "premium",
    "versatile",
]


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def weighted_choice(weights: Dict[str, float]) -> str:
    total = sum(weights.values())
    r = random.random() * total
    upto = 0.0
    for k, w in weights.items():
        upto += w
        if upto >= r:
            return k
    return next(iter(weights))


def long_tail_index(idx: int, total: int) -> float:
    """Zipf-like weighting so earlier items are more popular."""
    return 1.0 / math.pow(idx + 1, 0.85) / sum(1.0 / math.pow(i + 1, 0.85) for i in range(total))


def deterministic_embedding(text: str, dims: int = EMBEDDING_DIMS) -> List[float]:
    seed = text.strip().lower().encode("utf-8")
    if not seed:
        return []
    block = hashlib.sha256(seed).digest()
    vec: List[float] = []
    for i in range(dims):
        if i and i % len(block) == 0:
            counter = bytes([i // len(block)])
            block = hashlib.sha256(seed + counter).digest()
        b = block[i % len(block)]
        vec.append(float(b) / 127.5 - 1.0)
    return vec


def ensure_item_embeddings(items: List[Dict]) -> None:
    for item in items:
        if item.get("embedding"):
            continue
        text_parts: List[str] = []
        for key in ("brand", "category", "description"):
            value = item.get(key)
            if value:
                text_parts.append(str(value))
        tags = item.get("tags") or []
        if not text_parts and tags:
            text_parts.append(" ".join(tags))
        text = " ".join(text_parts).strip()
        if not text:
            continue
        item["embedding"] = deterministic_embedding(text)


def chunked(iterable: Iterable, n: int) -> Iterable[List]:
    batch: List = []
    for item in iterable:
        batch.append(item)
        if len(batch) == n:
            yield batch
            batch = []
    if batch:
        yield batch


def post_json(session: requests.Session, url: str, payload: Dict, retries: int = 3) -> Dict:
    for attempt in range(retries):
        response = session.post(url, json=payload, timeout=30)
        if response.status_code in (200, 201, 202):
            return response.json() if response.content else {"status": "ok"}
        if response.status_code >= 500 and attempt < retries - 1:
            time.sleep(1 + attempt)
            continue
        raise RuntimeError(
            f"POST {url} failed: {response.status_code} {response.text}")
    raise AssertionError("unreachable")


# ---------------------------------------------------------------------------
# Fixture helpers & generators
# ---------------------------------------------------------------------------

def load_fixture(path: str) -> Dict[str, List[Dict]]:
    with open(path, "r", encoding="utf-8") as fh:
        data = json.load(fh)
    if not isinstance(data, dict):
        raise ValueError(f"Fixture file {path} must contain a JSON object.")
    fixture: Dict[str, List[Dict]] = {}
    for key in ("items", "users", "events"):
        value = data.get(key)
        if value is not None and not isinstance(value, list):
            raise ValueError(f"Fixture field '{key}' must be a list.")
        fixture[key] = value
    return fixture


def generate_items(count: int) -> List[Dict]:
    items: List[Dict] = []
    popularity_values: List[float] = []

    for idx in range(count):
        category = weighted_choice({c: 1.0 for c in CATEGORIES})
        brand = random.choice(BRANDS[category])
        base_price = random.uniform(12, 240)
        margin = random.uniform(0.1, 0.6)
        novelty = random.random()
        availability = random.random() > 0.05
        tags = list({*TAGS_BY_CATEGORY[category], brand.lower()})
        if margin > 0.45:
            tags.append("high_margin")
        if novelty > 0.8:
            tags.append("new_arrival")
        if not availability:
            tags.append("backorder")

        popularity = long_tail_index(idx, count)
        popularity_values.append(popularity)

        item = {
            "item_id": f"item_{idx+1:04d}",
            "category": category,
            "brand": brand,
            "price": round(base_price, 2),
            "available": availability,
            "tags": tags,
            "description": f"{random.choice(ITEM_DESCRIPTORS)} {category.lower()} pick by {brand}",
            "props": {
                "margin": round(margin, 3),
                "novelty": round(novelty, 3),
                "popularity_hint": round(popularity, 6),
            },
        }
        items.append(item)

    # Normalize popularity hints to ensure coverage metadata
    max_pop = max(popularity_values) if popularity_values else 1.0
    for item, pop_val in zip(items, popularity_values):
        item["props"]["popularity_rank_norm"] = round(pop_val / max_pop, 6)

    return items


def generate_users(total_users: int) -> List[Dict]:
    users: List[Dict] = []
    for idx in range(total_users):
        segment = SEGMENTS[idx % len(SEGMENTS)]
        join_delta = random.randint(0, 45)
        join_date = datetime(2025, 9, 1, tzinfo=timezone.utc) + \
            timedelta(days=join_delta)
        traits = {
            "segment": segment.name,
            "description": segment.description,
            "join_date": join_date.isoformat(),
            "activity_level": round(segment.activity_level, 2),
        }
        users.append({"user_id": f"user_{idx+1:04d}", "traits": traits})
    return users


def generate_events(users: List[Dict], items: List[Dict], min_events: int) -> List[Dict]:
    events: List[Dict] = []
    item_by_category: Dict[str, List[Dict]] = {c: [] for c in CATEGORIES}
    for item in items:
        item_by_category[item["category"]].append(item)

    # Precompute popularity weights
    pop_weights = [item["props"]["popularity_rank_norm"] for item in items]
    pop_total = sum(pop_weights)
    pop_norm = [w / pop_total for w in pop_weights]

    base_date = datetime(2025, 9, 1, tzinfo=timezone.utc)
    final_date = base_date + timedelta(days=60)

    while len(events) < min_events:
        user = random.choice(users)
        segment = next(seg for seg in SEGMENTS if seg.name ==
                       user["traits"]["segment"])
        interactions = max(3, int(random.gauss(8 * segment.activity_level, 3)))

        for _ in range(interactions):
            event_ts = base_date + timedelta(
                seconds=random.randint(
                    0, int((final_date - base_date).total_seconds()))
            )
            category = weighted_choice(segment.category_weights)
            category_items = item_by_category.get(category, items)
            if random.random() < 0.15 * segment.diversity_bias:
                # occasional exploration outside preferred categories
                item = random.choices(items, weights=pop_norm, k=1)[0]
            else:
                item = random.choice(category_items)

            event_type = random.choices(
                population=[0, 1, 2, 3],
                weights=[
                    0.55,
                    0.25,
                    0.12,
                    0.08 * segment.purchase_bias,
                ],
            )[0]
            value = 1 if event_type != 2 else random.uniform(1, 3)

            events.append(
                {
                    "user_id": user["user_id"],
                    "item_id": item["item_id"],
                    "type": event_type,
                    "ts": event_ts.isoformat(),
                    "value": round(value, 2),
                    "meta": {
                        "segment": segment.name,
                        "category": category,
                        "price": item["price"],
                        "margin": item["props"]["margin"],
                    },
                }
            )

            if len(events) >= min_events:
                break

    # Keep events sorted for deterministic ingestion order
    events.sort(key=lambda e: e["ts"])
    return events


# ---------------------------------------------------------------------------
# Main workflow
# ---------------------------------------------------------------------------

def seed_dataset(
    base_url: str,
    namespace: str,
    org_id: str,
    item_count: int,
    user_count: int,
    min_events: int,
    fixture: Optional[Dict[str, List[Dict]]] = None,
    fixture_path: Optional[str] = None,
) -> Dict:
    random.seed(SEED)

    if fixture and fixture.get("items"):
        items = list(fixture["items"])
    else:
        items = generate_items(item_count)
    ensure_item_embeddings(items)
    if fixture and fixture.get("users"):
        users = list(fixture["users"])
    else:
        users = generate_users(user_count)
    if fixture and fixture.get("events"):
        events = list(fixture["events"])
    else:
        events = generate_events(users, items, min_events)

    session = requests.Session()
    session.verify = False
    session.headers.update(
        {
            "Content-Type": "application/json",
            "X-Org-ID": org_id,
        }
    )

    evidence_dir = "analysis/evidence"
    payload_manifest = {
        "items_payloads": [],
        "users_payloads": [],
        "events_payloads": [],
    }

    # Items
    url_items = f"{base_url.rstrip('/')}/v1/items:upsert"
    for batch in chunked(items, BATCH_SIZE_ITEMS):
        payload = {"namespace": namespace, "items": batch}
        response = post_json(session, url_items, payload)
        payload_manifest["items_payloads"].append(
            {"request": payload, "response": response}
        )

    # Users
    url_users = f"{base_url.rstrip('/')}/v1/users:upsert"
    for batch in chunked(users, BATCH_SIZE_USERS):
        payload = {"namespace": namespace, "users": batch}
        response = post_json(session, url_users, payload)
        payload_manifest["users_payloads"].append(
            {"request": payload, "response": response}
        )

    # Events
    url_events = f"{base_url.rstrip('/')}/v1/events:batch"
    for batch in chunked(events, BATCH_SIZE_EVENTS):
        payload = {"namespace": namespace, "events": batch}
        response = post_json(session, url_events, payload)
        payload_manifest["events_payloads"].append(
            {
                "request_summary": {
                    "namespace": namespace,
                    "events_count": len(batch),
                    "ts_start": batch[0]["ts"],
                    "ts_end": batch[-1]["ts"],
                },
                "response": response,
            }
        )

    manifest_path = f"{evidence_dir}/seed_manifest.json"
    with open(manifest_path, "w", encoding="utf-8") as fh:
        json.dump(
            {
                "base_url": base_url,
                "namespace": namespace,
                "org_id": org_id,
                "item_count": len(items),
                "user_count": len(users),
                "event_count": len(events),
                "seed": SEED,
                "fixture": fixture_path,
                "batches": payload_manifest,
            },
            fh,
            indent=2,
        )

    snapshot = {
        "items": items[:5],
        "users": users[:5],
        "events": events[:5],
    }
    with open(f"{evidence_dir}/seed_samples.json", "w", encoding="utf-8") as fh:
        json.dump(snapshot, fh, indent=2)

    segment_counts = Counter(
        (user.get("traits") or {}).get("segment", "unknown") for user in users
    )
    segment_samples: Dict[str, Dict] = {}
    for user in users:
        segment = (user.get("traits") or {}).get("segment", "unknown")
        if segment not in segment_samples:
            segment_samples[segment] = user.get("traits", {})

    stats = {
        "items": len(items),
        "users": len(users),
        "events": len(events),
        "categories": sorted({item["category"] for item in items}),
        "segments": sorted({user["traits"]["segment"] for user in users}),
        "segment_counts": dict(sorted(segment_counts.items())),
        "time_span_days": (
            (
                datetime.fromisoformat(
                    events[-1]["ts"]) - datetime.fromisoformat(events[0]["ts"])
            ).days
            if events
            else 0
        ),
    }
    with open(f"{evidence_dir}/seed_stats.json", "w", encoding="utf-8") as fh:
        json.dump(stats, fh, indent=2)
    with open(f"{evidence_dir}/seed_segments.json", "w", encoding="utf-8") as fh:
        json.dump(
            {
                "segment_counts": stats["segment_counts"],
                "segment_samples": segment_samples,
            },
            fh,
            indent=2,
        )

    return stats


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Seed evaluation dataset via API.")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL)
    parser.add_argument("--org-id", default=DEFAULT_ORG_ID)
    parser.add_argument("--namespace", default=DEFAULT_NAMESPACE)
    parser.add_argument("--items", type=int, default=320)
    parser.add_argument("--users", type=int, default=120)
    parser.add_argument("--events", type=int, default=5200)
    parser.add_argument(
        "--fixture-path",
        type=str,
        default=None,
        help="Path to bespoke fixture JSON with items/users/events (optional).",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    fixture_data = load_fixture(
        args.fixture_path) if args.fixture_path else None
    stats = seed_dataset(
        base_url=args.base_url,
        namespace=args.namespace,
        org_id=args.org_id,
        item_count=args.items,
        user_count=args.users,
        min_events=args.events,
        fixture=fixture_data,
        fixture_path=args.fixture_path,
    )
    print(json.dumps({"status": "seeded", "stats": stats}, indent=2))


if __name__ == "__main__":
    main()
