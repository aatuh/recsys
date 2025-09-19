# Segment Setup Example

To see the segment dry-run tool in action, you need to create segments and profiles first. Here are example API calls:

## 1. Create a Profile

```bash
curl -X POST http://localhost:8080/v1/segment-profiles:upsert \
  -H "Content-Type: application/json" \
  -d '{
    "profiles": [{
      "profile_id": "vip-high-novelty",
      "description": "VIP users with high novelty preference",
      "blend_alpha": 0.8,
      "blend_beta": 0.2,
      "blend_gamma": 0.4,
      "mmr_lambda": 0.6,
      "brand_cap": 2,
      "category_cap": 3,
      "profile_boost": 0.25,
      "profile_window_days": 30,
      "profile_top_n": 12,
      "half_life_days": 3,
      "co_vis_window_days": 14,
      "purchased_window_days": 7,
      "rule_exclude_events": true
    }]
  }'
```

## 2. Create a Segment with Rules

```bash
curl -X POST http://localhost:8080/v1/segments:upsert \
  -H "Content-Type: application/json" \
  -d '{
    "segments": [{
      "segment_id": "vip",
      "name": "VIP Users",
      "description": "High-value VIP users",
      "priority": 100,
      "active": true,
      "profile_id": "vip-high-novelty",
      "rules": [{
        "enabled": true,
        "rule_expr_json": {
          "any": [
            {
              "all": [
                {"eq": ["user.tier", "VIP"]},
                {"gte": ["user.ltv_eur", 500]}
              ]
            },
            {
              "all": [
                {"eq": ["ctx.surface", "homepage"]},
                {"eq": ["ctx.device", "mobile"]},
                {"gte_days_since": ["user.last_play_ts", 7]}
              ]
            }
          ]
        }
      }]
    }]
  }'
```

## 3. Test the Dry-Run Tool

Now you can test in the demo-ui:

1. Go to "Segment Tools" tab
2. Set User Tier to "VIP"
3. Set Surface to "homepage"
4. Set Device to "mobile"
5. Click "Run Dry-Run"

You should see:
- Matched: Yes
- Segment: vip
- Profile: vip-high-novelty
- Effective configuration showing all the profile parameters

## Rule Expression Examples

The `rule_expr_json` field uses a simple DSL:

```json
{
  "any": [
    {"eq": ["user.tier", "VIP"]},
    {"gte": ["user.ltv_eur", 500]},
    {"in": ["ctx.surface", ["homepage", "casino"]]}
  ]
}
```

Available operators:
- `eq`: equals
- `gte`: greater than or equal
- `lte`: less than or equal
- `in`: value in array
- `gte_days_since`: days since timestamp
- `any`: OR logic
- `all`: AND logic
