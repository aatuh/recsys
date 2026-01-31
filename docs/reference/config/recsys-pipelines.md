# recsys-pipelines configuration (lean)

Note: the current CLI expects **JSON** configs.

- events.source: kafka|pubsub|db|files
- artifacts.object_store: s3|gcs|fs
- artifacts.registry_db_dsn: Postgres DSN
- quality.max_null_rate: basic data quality gate
