# Ecommerce Mini Local Load-Test Reference

This is a smoke reference result, not a sizing guarantee. It was generated from the local Docker Compose API/Postgres
stack after running the ecommerce-mini proof kit.

## Environment

| Field | Value |
| --- | --- |
| Generated at | `2026-05-24T08:34:46Z` |
| Target | `http://localhost:8000/v1/recommend` |
| Deployment | Local Docker Compose API and Postgres |
| Host CPU | AMD Ryzen 7 3700X, 8 cores / 16 threads |
| Host memory | 31 GiB |
| Host OS | Linux 7.0.7-arch1-1 x86_64 |
| Tenant / surface | `demo` / `home` |
| Catalog size | `8` items |
| Artifact size | `10276` bytes |
| Requests / concurrency | `20` / `5` |
| User cardinality | `20` |

## Result

| Metric | Value |
| --- | ---: |
| Successful responses | 20 |
| Client/network errors | 0 |
| Elapsed | 14 ms |
| RPS | 1400.03 |
| p50 latency | 1.05 ms |
| p95 latency | 11.42 ms |
| p99 latency | 11.42 ms |
| HTTP 200 | 20 |

## Degradation Notes

No non-2xx responses were observed in the low-concurrency smoke run. A separate 100-request local run at concurrency 10
hit the default local rate limiter and returned 429 responses, which is expected unless rate-limit settings are adjusted
for a capacity test.

## Reproduction Command

```bash
BASE_URL=http://localhost:8000 \
TENANT_ID=demo \
SURFACE=home \
REQUESTS=20 \
CONCURRENCY=5 \
USER_CARDINALITY=20 \
CATALOG_SIZE=8 \
ARTIFACT_SIZE_BYTES=10276 \
REPORT_JSON=tmp/reference-loadtest-ecommerce-mini.json \
REPORT_MARKDOWN=tmp/reference-loadtest-ecommerce-mini.md \
CPU_NOTES="local Docker Compose API and Postgres on this workstation" \
MEMORY_NOTES="no container memory pressure observed during smoke run" \
DEGRADATION="no non-2xx responses observed in the low-concurrency smoke run" \
bash scripts/loadtest.sh
```
