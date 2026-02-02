
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
