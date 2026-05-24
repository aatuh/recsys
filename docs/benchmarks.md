# Benchmarks

Use `scripts/loadtest.sh` to create reproducible serving evidence for a specific deployment and dataset.

```bash
BASE_URL=http://localhost:8000 \
TENANT_ID=demo \
SURFACE=home \
REQUESTS=1000 \
CONCURRENCY=25 \
CATALOG_SIZE=8 \
ARTIFACT_SIZE_BYTES="$(find tmp/commercial-proof-kit/pipelines/objectstore -type f -print0 2>/dev/null | xargs -0 stat -c '%s' 2>/dev/null | awk '{s+=$1} END {print s+0}')" \
REPORT_JSON=tmp/loadtest-report.json \
REPORT_MARKDOWN=tmp/loadtest-report.md \
bash scripts/loadtest.sh
```

Expected result: the command prints request totals, RPS, p50/p95/p99 latency, status codes, and writes JSON/Markdown
reports when report paths are set.

## Report Template

Keep benchmark reports with enough context to be useful:

- RecSys version, image tag, and commit.
- Dataset/catalog size, artifact size, user cardinality, request count, and concurrency.
- Deployment shape: replicas, CPU/memory requests and limits, database/object-store location, and cache TTLs.
- p50, p95, p99 latency, RPS, success/error counts, and status-code distribution.
- CPU, memory, and degradation notes.

Do not generalize from one fixture. A local ecommerce-mini report is useful as a smoke reference; production sizing
requires a report from the operator's own catalog, traffic shape, and infrastructure.
