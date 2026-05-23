# Integration and Evaluation

Use this guide when a web, mobile, desktop, or backend client is ready to call RecSys and produce evaluation data.

## Integration sequence

1. Choose the recommendation surface, such as `home`, `product_detail`, or `cart`.
2. Decide the tenant source: tenant claim in JWT/API key context, or the configured tenant header.
3. Send pseudonymous user/session identifiers in recommendation requests.
4. Persist response metadata, especially `request_id`, `tenant_id`, `surface`, algorithm version, config version, and
   rules version.
5. Log exposure events when recommendations are shown.
6. Log outcome events when the user clicks, converts, or otherwise responds.
7. Run `recsys-eval` against joined exposure/outcome data before shipping ranking changes.

## Request identity

Production integrations should use JWT or API key auth. Local development can use dev headers from `api/.env.example`.

| Header or field | Purpose |
| --- | --- |
| `X-Org-Id` | Default tenant header in local config. |
| `X-Request-Id` | Optional caller-provided request ID. The API also returns request metadata. |
| `user.anonymous_id` | Pseudonymous user identifier for personalization and evaluation joins. |
| `user.session_id` | Session-level identifier for session signals and attribution. |
| `context.device` | Client class such as `web`, `mobile`, or `desktop`. |

Do not send raw names, email addresses, phone numbers, or other direct PII in recommendation payloads. The EU-baseline
posture is pseudonymous identifiers, documented retention, and minimal context fields.

## Minimal request

```bash
curl -sS http://localhost:8000/v1/recommend \
  -H 'Content-Type: application/json' \
  -H 'X-Org-Id: demo' \
  -H 'X-Dev-User-Id: local-dev' \
  -H 'X-Dev-Org-Id: demo' \
  -d '{
    "surface": "home",
    "k": 10,
    "user": {
      "anonymous_id": "anon-123",
      "session_id": "sess-456"
    },
    "context": {
      "device": "mobile",
      "country": "FI"
    }
  }'
```

## Exposure and outcome join keys

Use a stable join shape:

| Event | Required join fields |
| --- | --- |
| Recommendation response | `request_id`, `tenant_id`, `surface`, item IDs, rank. |
| Exposure | `request_id`, `user_id` or pseudonymous equivalent, timestamp, displayed items. |
| Outcome | `request_id`, `user_id` or pseudonymous equivalent, `item_id`, event type, timestamp. |
| Assignment | `experiment_id`, `variant`, `request_id`, user identifier, timestamp. |

The schema sources are listed in [Data contracts](reference/data-contracts.md).

## Evaluation loop

The smallest reliable loop is:

```bash
cd recsys-eval
make test
go run ./cmd/recsys-eval --help
```

Then run the relevant checked-in config from `recsys-eval/configs/eval/` against your dataset. Treat reports as
decision support, not as automatic ship approval. A good ship decision also checks data quality, sample ratio mismatch,
guardrail metrics, and operational rollback readiness.

## Ecommerce mini proof kit

Use this path when you want a compact evaluator-facing proof that the checked-in fixture, offline report, and pipelines
still work together:

```bash
make proof-kit-test
```

Expected result: the command prints `commercial proof kit smoke passed`.

What it validates:

- `examples/data/ecommerce-mini/eval/exposures.jsonl`, `outcomes.jsonl`, and `assignments.jsonl` pass
  `recsys-eval` schema validation.
- `recsys-eval` writes JSON and Markdown offline reports under `tmp/commercial-proof-kit/eval/`.
- `recsys-pipelines` runs against `examples/data/ecommerce-mini/pipelines/exposure.jsonl`.
- The pipeline writes a manifest at `tmp/commercial-proof-kit/pipelines/registry/current/demo/home/manifest.json`.
- Published object-store artifacts exist under `tmp/commercial-proof-kit/pipelines/objectstore/`.

Primary evidence files:

- `scripts/test_commercial_proof_kit.sh`
- `examples/data/ecommerce-mini/README.md`
- `recsys-eval/configs/eval/offline.ecommerce-mini.yaml`

## Failure handling

- Return user-safe fallback UI when the API returns 4xx, 429, or 5xx.
- Preserve request IDs in client logs and server logs for reconstruction.
- Do not expose internal validation details or stack traces to end users.
- For empty recommendation sets, use merchandising fallback content and review the [operations guide](operations.md).
