# Glossary

## Who this is for

Developers, operators, evaluators, and reviewers who need shared RecSys terms without reading the old archived docs.

## What you will get

- A compact vocabulary for current canonical docs.
- Terms that affect integration, operations, evaluation, and commercial review.
- Pointers to the pages that own deeper procedures.

## Terms

| Term | Meaning |
| --- | --- |
| Artifact | Immutable, versioned output from offline computation, such as popularity or co-occurrence data. |
| Artifact mode | Service mode where recommendation signals are loaded from a current manifest and artifact blobs. |
| Assignment | Experiment membership event containing experiment ID, variant, request ID, and user identifier. |
| Backfill | Reprocessing historical windows to rebuild canonical data or artifacts. |
| Candidate | Item considered for ranking before final ordering and filtering. |
| Config version | ETag for the active tenant config document. |
| Current manifest | Mutable manifest path that tells the service which artifact URIs are live for a tenant and surface. |
| DB-only mode | Service mode where signals are read from database-backed stores rather than published artifacts. |
| Exposure | Event recording what recommendations were shown for one request. |
| Join rate | Share of evaluation events that can be joined through stable identifiers, especially `request_id`. |
| Manifest | JSON document mapping artifact types to artifact URIs for a tenant and surface. |
| Outcome | Event recording what the user did after an exposure, such as click or conversion. |
| Ranking | Scoring and ordering candidates into the final top-K response. |
| Request ID | Correlation key used across recommendation responses, exposure logs, outcomes, and incidents. |
| Rules version | ETag for the active tenant rules document. |
| Segment | Cohort or slice label, with `default` used when no narrower segment is supplied. |
| Surface | Product location where recommendations are shown, such as `home`, `product_detail`, or `cart`. |
| Tenant | Organization or customer boundary for config, data, and authorization. |

## Example context

```json
{
  "tenant": "demo",
  "surface": "home",
  "segment": "default",
  "request_id": "req_123",
  "config_version": "W/\"config-etag\"",
  "rules_version": "W/\"rules-etag\""
}
```

## Read next

- [Data Contracts](data-contracts.md)
- [Configuration](config.md)
- [Evaluation Decisions](../evaluation-decisions.md)
- [Artifacts and Pipelines](../artifacts-and-pipelines.md)
