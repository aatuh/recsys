# Exposure logging and attribution (lean)

Log what you showed (exposure) and what the user did (outcome).
Join by request_id when possible.

Service support:

- Enable JSONL logging with `EXPOSURE_LOG_ENABLED=true` and `EXPOSURE_LOG_PATH`.
- Set `EXPOSURE_LOG_FORMAT=eval_v1` to emit recsys-eval compatible exposure

  records (`exposure.v1` schema). Default is `service_v1`.

- Retention is controlled via `EXPOSURE_LOG_RETENTION_DAYS` (file-based logs).

Attribution reminder:

- Outcomes must carry the **same request_id** returned by `/v1/recommend` to

  join in recsys-eval.
