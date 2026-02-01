
# CLI reference

## recsys-pipelines

Commands:
- `run` : ingest + validate + compute + publish
- `version`

### run

```bash
recsys-pipelines run \
  --config <path> \
  --tenant <tenant> \
  --surface <surface> \
  --segment <segment optional> \
  --start YYYY-MM-DD \
  --end YYYY-MM-DD
```

Notes:
- end date is inclusive
- windows are daily UTC

Incremental (checkpointed) example:

```bash
recsys-pipelines run \
  --config <path> \
  --tenant <tenant> \
  --surface <surface> \
  --end YYYY-MM-DD \
  --incremental
```

## Job binaries

All jobs take `--config --tenant --surface --start --end`.
Some also take `--segment`.

- `job_ingest`
- `job_validate`
- `job_popularity` (segment optional)
- `job_cooc` (segment optional)
- `job_publish` (segment optional)
