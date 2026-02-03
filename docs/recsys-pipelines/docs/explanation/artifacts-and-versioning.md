
# Artifacts and versioning

## What is an artifact?

A file (JSON) that the online recommender uses to make decisions quickly.
Artifacts are precomputed offline so serving stays fast.

## Versioning

Artifacts are version-addressed:

- Compute the payload
- Remove volatile build metadata
- Hash the remaining JSON (SHA-256 hex)
- Embed the version into the final artifact

If the canonical input does not change, the version should not change.

## Publishing protocol (two-phase)

Publishing is ordered to keep serving safe:

1) Write the versioned blob
1) Validate the artifact (including version recompute)
1) Write the version record
1) Swap the manifest pointer last

This means a failed publish does not break "current".

## Rollback

Rollback is changing the manifest pointer to an older URI.
The older versioned blob remains available.

See `how-to/rollback-manifest.md`.
