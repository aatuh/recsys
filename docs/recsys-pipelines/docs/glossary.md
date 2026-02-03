
# Glossary

**Artifact**
: A precomputed data product used by the online recommender, such as popularity
  lists or item neighbors.

**Canonical events**
: Events stored in a normalized format that the rest of the pipeline relies on.

**Tenant**
: A logical customer or environment namespace.

**Surface**
: A recommendation placement (e.g. "home", "checkout").

**Segment**
: Optional sub-grouping within a surface (e.g. "new_users").

**Window**
: A time range that a job processes. In v1, windows are daily UTC buckets.

**Version**
: A deterministic identifier (SHA-256 hex) of an artifact payload excluding
  volatile build metadata.

**Manifest**
: A small JSON document that points to the current artifact URIs for a
  (tenant, surface).

**Registry**
: Storage for artifact records and current manifests.

**Object store**
: Storage for artifact blobs. In local mode, this is the filesystem.

**Idempotent**
: Safe to run multiple times without changing the result.
