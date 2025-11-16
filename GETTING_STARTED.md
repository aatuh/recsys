# Getting Started (Local Repo Workflow)

Follow this guide after cloning the repo to bring the RecSys stack up locally, seed sample data, and call `/v1/recommendations`. It focuses on **running from source**. If you only need hosted HTTP examples, skip this file and read `docs/quickstart_http.md`.

---

## 1. Prerequisites

Install the following before continuing:

- **Docker + Docker Compose v2** – required to run Postgres, the API, proxy, and supporting containers. [Install guide](https://docs.docker.com/get-docker/).
- **GNU Make** – the Makefile wraps the common Docker commands. Available by default on macOS/Linux; Windows users can install via WSL or [GNUWin32](http://gnuwin32.sourceforge.net/packages/make.htm).
- **Python 3.10+** – helper scripts (seed, tuning, guardrails) use Python. Install from [python.org](https://www.python.org/downloads/) and ensure `python3` is on your PATH.
- **pip packages:** `requests`, `urllib3`. Install once via `python3 -m pip install --upgrade pip && python3 -m pip install requests urllib3`.
- **curl + jq** (optional but handy) – used for quick API checks.

> If you’re unsure how to install these tools, ask your team or follow the linked guides before proceeding. Hosted API consumers can ignore this doc and use `docs/quickstart_http.md`.

---

## 2. Start the local stack

From the repo root:

```bash
make env PROFILE=dev   # first-time setup: copies api/env/dev.env -> api/.env
make dev               # builds and runs API, Postgres, proxy, UI containers
```

`make dev` leaves containers running in the background. Check readiness with:

```bash
curl -s http://localhost:8000/health | jq
```

You should see `{"status":"ok"}`. Use `docker compose logs -f api` (or `make logs`) if the health check fails. Shut everything down later with `make down` (this also clears Docker volumes).

---

## 3. Seed sample data

With the stack running, seed a demo namespace using the helper script. Replace `--org-id` only if you know your tenant UUID; the default below matches the seeded fixtures.

```bash
python3 analysis/scripts/seed_dataset.py \
  --base-url http://localhost:8000 \
  --org-id 00000000-0000-0000-0000-000000000001 \
  --namespace demo \
  --users 600 \
  --events 40000
```

What this does:

1. Upserts ~320 catalog items with tags/props/embeddings.
2. Adds ≥120 sample users with behavioral segments.
3. Batches ~40k events so personalization, guardrails, and traces have signal.
4. Writes evidence to `analysis/evidence/seed_segments.json` so you can inspect segment counts later.

> Need a clean slate? Run `python3 analysis/scripts/reset_namespace.py --base-url http://localhost:8000 --org-id 00000000-0000-0000-0000-000000000001 --namespace demo --force` before seeding again.

---

## 4. Call `/v1/recommendations`

Once the namespace has data, hit the API locally:

```bash
curl -s -X POST http://localhost:8000/v1/recommendations \
  -H "Content-Type: application/json" \
  -H "X-Org-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
        "namespace": "demo",
        "user_id": "user_0001",
        "k": 8,
        "include_reasons": true,
        "context": { "surface": "home" }
      }' | jq
```

You should receive a payload similar to:

```json
{
  "items": [
    {
      "item_id": "sku_1017",
      "score": 0.87,
      "reasons": ["popularity","co_visitation","personalization"]
    }
  ],
  "trace": {
    "namespace": "demo",
    "policy": "blend:v1_mmr",
    "sourceMetrics": { "popularity": 600, "collaborative": 220, "content": 140 }
  }
}
```

- `items[]` contains the ranked recommendations (ID + score + reason codes).
- `trace` includes the policy name, candidate source counts, and other metadata useful for debugging or guardrail evidence.

Feel free to modify `overrides` (e.g., blend weights) or `k` to experiment—no redeploy required.

---

## 5. Troubleshooting & next steps

Common issues:

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| Connection refused | Containers not running or still starting | `make dev`, then re-run after logs show API ready |
| 400 `missing_org_id` | `X-Org-ID` header omitted | Include the UUID header on every request |
| 401/403 | API auth enabled but key missing | Ask your admin for API key/headers |
| Empty list returned | Namespace not seeded or `available=false` items | Rerun the seed script and inspect `/v1/items` / `analysis/evidence/seed_segments.json` |

Where to go next:

- **Hosted-only examples:** `docs/quickstart_http.md`
- **Terminology & metrics:** `docs/concepts_and_metrics.md`
- **Persona/lifecycle map:** `docs/overview.md`
- **Full endpoint reference:** `docs/api_reference.md`
- **Tuning & guardrails:** `docs/tuning_playbook.md`, `docs/simulations_and_guardrails.md`

Raise any gaps or confusion by filing an issue—this file should make onboarding painless for anyone running RecSys locally from source.
