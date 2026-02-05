
# Runbook: Limit exceeded

## Symptoms

- Error message includes "limit exceeded"

## Why this exists

Limits prevent resource blowups from pathological inputs.
Raising limits blindly can cause OOM or slowdowns.

## Triage

1) Identify which limit triggered (events/sessions/items/neighbors)
1) Inspect raw event volume for the window
1) Look for data bugs (duplicate events, runaway session ids)

## Recovery

- Fix upstream data if it's a bug
- For genuine growth, raise limits gradually and benchmark

See `reference/config.md` and `explanation/validation-and-guardrails.md`.

## Read next

- Config reference (limits): [`reference/config.md`](../../reference/config.md)
- Validation and guardrails: [`explanation/validation-and-guardrails.md`](../../explanation/validation-and-guardrails.md)
- Debug failures: [`how-to/debug-failures.md`](../../how-to/debug-failures.md)
- Operate pipelines daily: [`how-to/operate-daily.md`](../../how-to/operate-daily.md)
