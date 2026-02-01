
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
Some also take `--segment` or extra inputs.

- `job_ingest`
- `job_validate`
- `job_popularity` (segment optional)
- `job_cooc` (segment optional)
- `job_implicit` (segment optional)
- `job_content_sim` (segment optional, requires `--input`)
- `job_session_seq` (segment optional)
- `job_publish` (segment optional)

`job_content_sim` usage:

```bash
job_content_sim \
  --config <path> \
  --tenant <tenant> \
  --surface <surface> \
  --input <catalog.csv|catalog.jsonl> \
  --start YYYY-MM-DD \
  --end YYYY-MM-DD
```
