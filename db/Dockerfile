# Extends pgvector and bakes init SQL into the image.
FROM pgvector/pgvector:pg16

# Any .sql or .sh here runs once on fresh data dir initialization.
COPY initdb/ /docker-entrypoint-initdb.d/
