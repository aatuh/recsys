
# Architecture

## Code organization

- `internal/domain`: pure, deterministic domain logic
- `internal/app/usecase`: orchestration of domain logic through ports
- `internal/ports`: interfaces the app depends on
- `internal/adapters`: IO implementations (filesystem, logging, etc.)
- `cmd/*`: binaries (all-in-one CLI and job-per-step)

## System interactions (C4-inspired, ASCII)

Level 1: context

+---------------------+           +------------------+
|  Offline scheduler  |  runs     | recsys-pipelines |
| (cron/airflow/k8s)  +---------->+ (this repo)      |
+---------------------+           +---------+--------+
                                            |
                                            | publishes
                                            v
                                    +-------+--------+
                                    | Artifact store  |
                                    | + Registry      |
                                    +-------+--------+
                                            |
                                            | consumed by
                                            v
                                    +-------+--------+
                                    | Online service  |
                                    | (recsys-service)|
                                    +-----------------+

Level 2: containers within this repo

- CLI and jobs in `cmd/*`
- Filesystem adapters
- Usecases (ingest/validate/compute/publish)

## Why ports/adapters

- Keeps domain logic deterministic and testable
- Makes storage pluggable (filesystem now, S3/GCS later)
- Makes validation pluggable (builtin now, GE/dbt later)

## Read next

- Start here: [`start-here.md`](../start-here.md)
- Data lifecycle (raw → canonical → publish): [`explanation/data-lifecycle.md`](data-lifecycle.md)
- Operate pipelines daily: [`how-to/operate-daily.md`](../how-to/operate-daily.md)
- Config reference: [`reference/config.md`](../reference/config.md)
- Glossary: [`glossary.md`](../glossary.md)
