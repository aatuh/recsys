# FAQ & Troubleshooting

Use this guide when you hit errors or surprising results while calling the RecSys API. It focuses on common symptoms, likely causes, and concrete debugging steps.

> **Where this fits:** Observability & analysis.

---

## 1. No or unexpected recommendations

### 1.1 I get empty or very short recommendation lists

Typical symptoms:

- HTTP 200, but `items` is empty or much shorter than expected.

Things to check:

- **Catalog and availability**
  - Confirm you have upserted items with `available=true`.
  - Call `/v1/items` (or your internal tools) to verify items exist in the namespace.
- **Events and personalization**
  - Ensure you have sent `events:batch` data for the users you query.
  - Cold-start users may see popularity-based results until more history arrives.
- **Filters and overrides**
  - Remove aggressive filters (price bands, tags, hard exclusions) and retry.
  - Temporarily drop overrides to confirm the baseline behavior.

If lists are still empty:

- Try a known-good test namespace (for example from [`GETTING_STARTED.md`](../GETTING_STARTED.md)).
- Ask your deployment owner whether any guardrails or rules are blocking results; see [`docs/simulations_and_guardrails.md`](simulations_and_guardrails.md) and [`docs/rules_runbook.md`](rules_runbook.md).

### 1.2 Recommendations look random or low quality

Typical symptoms:

- Lists are non-empty but feel irrelevant or unstable.

Things to check:

- **Event coverage**
  - Verify enough events exist for key users and surfaces.
  - Check that event `type` codes and timestamps are correct.
- **Namespace mix**
  - Confirm you are querying the same namespace you seeded.
- **Configuration**
  - Check env profiles for extreme blend or MMR settings ([`docs/configuration.md`](configuration.md), [`docs/env_reference.md`](env_reference.md)).

If this persists:

- Run a small simulation or tuning pass and review guardrail metrics ([`docs/tuning_playbook.md`](tuning_playbook.md), [`docs/simulations_and_guardrails.md`](simulations_and_guardrails.md)).

---

## 2. I get 400/422 on ingestion

Typical symptoms:

- `400 missing_org_id` or `400 invalid_namespace`
- `422 invalid_override`, `422 invalid_event`, or similar schema errors

Things to check:

- **Headers**
  - Ensure `X-Org-ID` is present and a valid UUID.
  - Include `X-API-Key` or `Authorization` only if your deployment requires it.
- **Namespace**
  - Confirm the `namespace` string matches what your team configured.
- Use a single namespace per org/surface to keep guardrails and audits clean.
- **Payload shape**
  - Validate JSON with a linter or `jq` before sending.
  - Compare your payload against the examples in [`docs/api_reference.md`](api_reference.md) and [`docs/client_examples.md`](client_examples.md).

If the error persists:

- Capture the response body (including `code` and `trace_id`).
- Search for the `code` in [`docs/api_errors_and_limits.md`](api_errors_and_limits.md) for more details.

---

## 3. I get 401/403 (auth problems)

Typical symptoms:

- `401 unauthorized` or `403 forbidden` when calling endpoints that should be available.

Things to check:

- **API key or auth header**
  - Confirm whether your deployment uses `X-API-Key`, `Authorization: Bearer`, or another scheme.
  - Ensure the header is present on every request, including `GET /health` in locked-down environments.
- **Org and namespace**
  - Some deployments restrict which orgs/namespaces a given key can access. Verify with your ops team.

If you still see 401/403:

- Capture the full response (status code, `code`, `message`, and `trace_id`).
- Share these details with the team operating the deployment.

---

## 4. I get 429 or 5xx (limits and transient errors)

Typical symptoms:

- `429 too_many_requests`
- `500 internal_server_error` or other 5xx codes

Things to check:

- **Rate limits**
  - Respect any `Retry-After` headers on 429 responses.
  - Implement exponential backoff (e.g., 0.5s → 1s → 2s) for retryable calls.
- **Idempotency**
  - Make ingestion calls (`items:upsert`, `users:upsert`, `events:batch`) idempotent so retries do not create duplicates.

If errors persist over time:

- Collect a small set of failing `trace_id` values and timestamps.
- Share them with the team operating the service along with approximate request volume.

See [`docs/api_errors_and_limits.md`](api_errors_and_limits.md) for full status code semantics and payload limits.

---

## 5. I’m not sure which doc to read next

- **Hosted API consumers**
  - Start with [`docs/quickstart_http.md`](quickstart_http.md) for the full HTTP walkthrough.
  - Keep [`docs/api_reference.md`](api_reference.md) and this FAQ open while you integrate.
- **Teams running the stack**
- Use [`GETTING_STARTED.md`](../GETTING_STARTED.md) plus [`docs/tuning_playbook.md`](tuning_playbook.md) and [`docs/simulations_and_guardrails.md`](simulations_and_guardrails.md) for deeper tuning and guardrail workflows.

When in doubt, skim [`docs/overview.md`](overview.md) for persona-based navigation and [`docs/concepts_and_metrics.md`](concepts_and_metrics.md) to decode terminology.
