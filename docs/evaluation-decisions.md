# Evaluation Decisions

Use this guide when RecSys evaluation output needs to become a ship, hold, or rollback decision.

## Decision order

Run the gates in this order. Do not interpret KPI movement until data integrity is credible.

| Step | Gate | Default decision when it fails |
| --- | --- | --- |
| 1 | Schemas validate for exposures, outcomes, assignments, and reports. | Hold. Fix instrumentation or dataset export. |
| 2 | Join integrity is acceptable for the surfaces and slices being evaluated. | Hold. Metrics are not trustworthy. |
| 3 | Guardrails hold: errors, latency, empty recommendations, and warning rates do not regress materially. | Roll back or hold with a time-boxed mitigation. |
| 4 | Primary KPI improves enough to matter and is stable across key slices. | Hold if inconclusive; roll back if meaningfully negative. |
| 5 | Rollback path is ready before rollout. | Hold. Do not ship a change that cannot be reversed. |

## Baseline thresholds

These are starting points, not universal targets. Tune them to the product, sample size, and business risk.

| Signal | Starting point |
| --- | --- |
| Join rate | Aim for at least 95% on the primary analysis slices before trusting KPI movement. |
| Error rate | No worse than 0.1-0.5 percentage points absolute unless explicitly accepted. |
| Latency | p95 no worse than 10-20% relative unless capacity testing supports it. |
| Empty recommendations | No worse than 0.2-1.0 percentage points absolute. |
| Primary KPI | Ship only when the improvement clears the pre-agreed minimum effect and guardrails hold. |

## Common decisions

| Finding | Decision | Next action |
| --- | --- | --- |
| KPI improved, join rate is low | Hold | Fix `request_id`, tenant, surface, or assignment joins before interpreting results. |
| KPI improved, guardrails regressed | Roll back by default | Reduce blast radius only if a safe mitigation is already known. |
| KPI is neutral and guardrails hold | Hold | Continue the experiment or close as inconclusive. |
| KPI regressed and guardrails hold | Roll back | Check slices only to explain the regression, not to excuse it. |
| Offline gate fails in CI | Hold | Fix the regression or update the baseline only after review. |

## Evidence to keep

- Input dataset names, schema versions, and generation time.
- Report output path and report hash when available.
- Primary KPI, guardrails, and join-rate summary.
- Slices reviewed and any excluded slices.
- Config, rules, algorithm, artifact, and manifest versions involved.
- Rollback lever chosen, if a rollback happened.

## Validation commands

Run the local proof path when checking the repository fixture:

```bash
make proof-kit-test
```

Expected result: the command prints `commercial proof kit smoke passed`.

For custom datasets, run the relevant `recsys-eval` schema validation and report commands from checked-in configs under
`recsys-eval/configs/eval/`.

## Read next

- [Integration and Evaluation](integration.md)
- [Data Contracts](reference/data-contracts.md)
- [Rollback Config and Rules](operations/runbooks/rollback-config-rules.md)
- [Stale Artifact Manifest](operations/runbooks/stale-artifact-manifest.md)
