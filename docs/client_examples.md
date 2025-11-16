# Client Examples

These minimal examples show how to call the RecSys HTTP API from real application code. They complement the cURL snippets in `docs/quickstart_http.md`. Each example assumes:

```text
BASE_URL = https://api.recsys.example.com
ORG_ID   = 00000000-0000-0000-0000-000000000001
NAMESPACE = retail_demo
```

Replace these with values from your environment.

---

## Python (requests)

```python
import requests

BASE_URL = "https://api.recsys.example.com"
ORG_ID = "00000000-0000-0000-0000-000000000001"
NS = "retail_demo"

headers = {
    "Content-Type": "application/json",
    "X-Org-ID": ORG_ID,
}

def upsert_item():
    payload = {
        "namespace": NS,
        "items": [
            {
                "item_id": "sku_123",
                "available": True,
                "price": 29.99,
                "tags": ["brand:acme", "category:fitness", "color:blue"],
                "props": {"title": "Acme Smart Bottle"}
            }
        ]
    }
    resp = requests.post(f"{BASE_URL}/v1/items:upsert", json=payload, headers=headers, timeout=10)
    resp.raise_for_status()

def fetch_recommendations():
    payload = {
        "namespace": NS,
        "user_id": "user_001",
        "k": 8,
        "include_reasons": True
    }
    resp = requests.post(f"{BASE_URL}/v1/recommendations", json=payload, headers=headers, timeout=10)
    resp.raise_for_status()
    return resp.json()

if __name__ == "__main__":
    upsert_item()
    recs = fetch_recommendations()
    for item in recs["items"]:
        print(item["item_id"], item.get("reasons"))
```

---

## JavaScript (Node.js + fetch)

Requires Node.js 18+ or `node-fetch`.

```javascript
const BASE_URL = "https://api.recsys.example.com";
const ORG_ID = "00000000-0000-0000-0000-000000000001";
const NS = "retail_demo";

const headers = {
  "Content-Type": "application/json",
  "X-Org-ID": ORG_ID,
};

async function upsertItem() {
  const payload = {
    namespace: NS,
    items: [
      {
        item_id: "sku_123",
        available: true,
        tags: ["brand:acme", "category:fitness"],
        props: { title: "Acme Smart Bottle" },
      },
    ],
  };
  const resp = await fetch(`${BASE_URL}/v1/items:upsert`, {
    method: "POST",
    headers,
    body: JSON.stringify(payload),
  });
  if (!resp.ok) throw new Error(`Upsert failed: ${resp.status}`);
}

async function fetchRecommendations() {
  const payload = {
    namespace: NS,
    user_id: "user_001",
    k: 8,
    include_reasons: true,
  };
  const resp = await fetch(`${BASE_URL}/v1/recommendations`, {
    method: "POST",
    headers,
    body: JSON.stringify(payload),
  });
  if (!resp.ok) throw new Error(`Recommendations failed: ${resp.status}`);
  return resp.json();
}

(async () => {
  await upsertItem();
  const recs = await fetchRecommendations();
  recs.items.forEach(item => {
    console.log(item.item_id, item.reasons);
  });
})();
```

---

## Where to go next

- `docs/quickstart_http.md` – hosted HTTP quickstart (ingestion, troubleshooting, error handling).
- `docs/api_reference.md` – full endpoint catalog and behavioral guarantees.
- `docs/api_errors_and_limits.md` – statuses, limits, retry/backoff guidance.
```*** End Patch
