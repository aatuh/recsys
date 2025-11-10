# Fixture Templates

The fixtures under `analysis/fixtures/` allow you to seed the API with
customer-specific catalog, user, and event data via
`analysis/scripts/seed_dataset.py --fixture-path <file>`.

## File layout

- `sample_customer.json` – minimal reference used in the README walkthrough.
- `templates/marketplace.json` – multi-vertical marketplace with GMV vs. long-tail items.
- `templates/media.json` – streaming/media catalog highlighting binge vs. casual cohorts.
- `templates/retail.json` – retail assortment with inventory flags and repeat buyers.
- `batch_simulations.yaml` – example manifest for `run_simulation.py --batch-file`.

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

Need to test multiple customers? Create a manifest similar to
`analysis/fixtures/batch_simulations.yaml` and pass it to
`analysis/scripts/run_simulation.py --batch-file ...`. Each entry can point to a
different fixture and env overrides so you can replay bespoke datasets quickly.
