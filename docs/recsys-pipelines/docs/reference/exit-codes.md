---
diataxis: reference
tags:
  - recsys-pipelines
---
# Exit codes

This repo uses conventional exit codes:

- 0: success
- 1: runtime failure (pipeline step failed)
- 2: usage/config error (missing flags, invalid config)

Job binaries follow the same pattern.

## Read next

- Debug failures: [How-to: Debug a failed pipeline run](../how-to/debug-failures.md)
- Pipeline failed runbook: [Runbook: Pipeline failed](../operations/runbooks/pipeline-failed.md)
- CLI reference: [CLI reference](cli.md)
