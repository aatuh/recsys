# Fixture Templates

The fixtures under `analysis/fixtures/` allow you to seed the API with
customer-specific catalog, user, and event data via
`analysis/scripts/seed_dataset.py --fixture-path <file>`.

## File layout & how they are used

- **`sample_customer.json`** — Minimal fixture used in README / GETTING_STARTED walkthroughs to seed data quickly. Consumed by `analysis/scripts/seed_dataset.py --fixture-path analysis/fixtures/sample_customer.json`.
- **`templates/marketplace.json`** — Multi-vertical marketplace mix (GMV vs. long-tail). Used by `seed_dataset.py` and `run_simulation.py` when you want a richer baseline.
- **`templates/media.json`** — Streaming/media catalog highlighting binge vs. casual cohorts. Used by `seed_dataset.py` and `run_simulation.py`.
- **`templates/retail.json`** — Retail assortment with inventory flags and repeat buyers. Used by `seed_dataset.py` and `run_simulation.py`.
- **`batch_simulations.yaml`** — Example batch manifest; each entry points to a fixture + overrides so you can replay multiple customers. Consumed via `analysis/scripts/run_simulation.py --batch-file analysis/fixtures/batch_simulations.yaml`.
- **`env_profiles/profile.example.json`** — Sample output of `env_profile_manager.py fetch`. Copy this into `analysis/env_profiles/<namespace>/<profile>.json` when you want a starting point for profile tuning without hitting production, then run `env_profile_manager.py apply` or `tuning_harness.py`.

Each template defines three lists:

```json
{
  "items": [
    {
      "item_id": "sku_demo_001",
      "category": "Home",
      "brand": "HavenCraft",
      "price": 89.0,
      "available": true,
      "tags": ["home", "decor", "brand:havencraft"],
      "props": {
        "margin": 0.32,
        "novelty": 0.18,
        "popularity_hint": 0.65,
        "popularity_rank_norm": 0.7
      }
    }
  ],
  "users": [
    {
      "user_id": "user_demo_001",
      "traits": {
        "segment": "home_refresh",
        "description": "Marketplace shoppers browsing seasonal décor",
        "lifetime_value_bucket": "high"
      }
    }
  ],
  "events": [
    {
      "user_id": "user_demo_001",
      "item_id": "sku_demo_001",
      "type": 3,
      "ts": "2025-09-25T12:00:00Z",
      "value": 1,
      "meta": {
        "surface": "home",
        "session_id": "sess_demo_001"
      }
    }
  ]
}
```

### Editing tips

1. **Copy a template** into a new file (e.g., `analysis/fixtures/customers/<customer>.json`).
2. Update the `items` array with your catalog IDs, `tags`, and `props`. The
   `props` map is optional but helps steer the ranking engine (`margin`,
   `novelty`, `popularity_hint`, `popularity_rank_norm`).
3. Set `traits.segment` to match the cohorts used in the evaluation suite (`new_users`,
   `power_users`, etc.) or introduce your own—scenario S7 will echo whatever segment you pick.
4. Generate events that mirror your production signals. The `type` field follows
   the API contract: 0=view, 1=click, 2=add-to-cart, 3=purchase, 4=custom.
5. Run `python analysis/scripts/seed_dataset.py --fixture-path <file>` and inspect
   `analysis/evidence/seed_segments.json` to confirm segment counts.

Need a reference env profile? Copy `analysis/fixtures/env_profiles/profile.example.json`
into `analysis/env_profiles/<namespace>/<profile>.json` and edit it before running
`analysis/scripts/env_profile_manager.py apply` or `analysis/scripts/tuning_harness.py`.
This keeps working files out of version control while still providing a known-good schema.

Need to test multiple customers? Create a manifest similar to
`analysis/fixtures/batch_simulations.yaml` and pass it to
`analysis/scripts/run_simulation.py --batch-file ...`. Each entry can point to a
different fixture and env overrides so you can replay bespoke datasets quickly.
