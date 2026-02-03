# recsys-pipelines configuration

Note: the current CLI expects **JSON** configs.

Core knobs you will likely set:

- raw_source.type: fs|s3|minio|postgres|kafka
- object_store.type: fs|s3|minio
- registry_dir: manifest registry location (fs)
- db.dsn: Postgres DSN (optional, for DB-backed signals)
- limits.max_users_per_run, limits.max_items_per_user (implicit/session jobs)
- limits.max_items_per_artifact, limits.max_neighbors_per_item

Artifacts produced (v1):

- popularity
- cooc
- implicit (collaborative)
- content_sim
- session_seq
