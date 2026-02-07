# Event join logic (exposures ↔ outcomes ↔ assignments)

## Who this is for

- Data engineers and analysts building an evaluation dataset
- Integrators wiring `request_id` propagation end-to-end
- Recommendation engineers validating offline evaluation quality

## What you will get

- The exact join key used by `recsys-eval`
- The invariants your logging must satisfy for valid attribution
- A checklist to debug low join rates

## Used by

- Minimum instrumentation spec: [`reference/minimum-instrumentation.md`](../minimum-instrumentation.md)
- Decision playbook (ship/hold/rollback): [`recsys-eval/docs/decision-playbook.md`](../../recsys-eval/docs/decision-playbook.md)
- Run eval and ship decisions: [`how-to/run-eval-and-ship.md`](../../how-to/run-eval-and-ship.md)

## Mental model

Think of each recommendation response as a “case”:

- **Exposure**: the ranked list you showed
- **Outcomes**: what the user did after seeing it (click, conversion)
- **Assignment** (optional): the experiment bucket for that request/user

`recsys-eval` attributes outcomes to exposures by joining on **`request_id`**.

## Join key and invariants

### Required: `request_id`

For evaluation, `request_id` must be present in:

- `exposure.v1` (`request_id` on the exposure record)
- `outcome.v1` (`request_id` on every outcome record you want to attribute)
- `assignment.v1` (if you analyze experiments)

Invariants to enforce:

- **Uniqueness:** one exposure list per `request_id` (do not reuse IDs across requests).
- **Propagation:** the same `request_id` flows serving → outcome event.
- **Stability:** do not change the `request_id` after you render a list (or you will split attribution).

### Strongly recommended: stable `user_id`

`recsys-eval` joins by `request_id`, but you should still ensure `user_id` is stable and consistent across exposures and
outcomes:

- it improves slice quality and sanity checks
- it helps detect “wrong request_id” bugs early
- it enables user-level analyses outside of strict request attribution

Do not log raw PII; use pseudonymous IDs.

## How `recsys-eval` joins

At a high level, `recsys-eval`:

1. groups all outcomes by `request_id`
2. attaches that outcome list to the exposure with the same `request_id`

This is a many-to-one join: one exposure → many outcomes.

Important implication: if your exposure stream contains multiple exposure records with the same `request_id`, later
records can overwrite earlier ones (so treat duplicates as a data quality bug).

## Join integrity: what to measure

`recsys-eval` reports join integrity as part of “Data Quality”:

- **Exposure join rate:** fraction of exposures that have at least one matching outcome
- **Outcome join rate:** fraction of outcomes that match an exposure
- **Assignment join rate:** fraction of assignments that match an exposure (when analyzing experiments)

In addition, compute a simple join rate in your warehouse (by surface and platform) to catch integration issues early.

Pseudo-SQL pattern:

```sql
select
  surface,
  count(*) as exposures,
  count(*) filter (
    where exists (
      select 1 from outcomes o
      where o.request_id = e.request_id
    )
  ) as exposures_with_outcomes
from exposures e
group by surface;
```

## Common failure patterns (and fixes)

- **Outcomes missing `request_id`**
  - Fix: propagate the ID you used when calling `/v1/recommend`, or store `meta.request_id` from the response.
- **Request IDs generated twice**
  - Symptom: exposure log uses one ID, outcome uses another.
  - Fix: centralize request ID generation; add an automated test that asserts “same request_id everywhere”.
- **Reusing the same `request_id` for multiple lists**
  - Symptom: attribution is “smeared” across requests; debugging becomes impossible.
  - Fix: generate a fresh ID per rendered list.
- **Exposure logged but list never rendered**
  - Symptom: exposure join rate drops even though outcomes are correct.
  - Fix: if you log exposures server-side, ensure the request corresponds to an actual render (or log exposures client-side).

## Read next

- Data contracts hub: [`index.md`](index.md)
- Exposure logging and attribution: [`explanation/exposure-logging-and-attribution.md`](../../explanation/exposure-logging-and-attribution.md)
- Eval events schemas: [`eval-events.md`](eval-events.md)
