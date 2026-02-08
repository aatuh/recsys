---
diataxis: reference
tags:
  - recsys-pipelines
---
# Output layout (local filesystem)

With default local config, outputs go under `.out/`.

## Canonical

`.out/canonical/<tenant>/<surface>/exposures/YYYY-MM-DD.jsonl`

## Staged artifacts

`.out/artifacts/<tenant>/<surface>/<segment>/<type>/<start>_<end>/`

- `<version>.json`
- `current.version`

## Object store

`.out/objectstore/<tenant>/<surface>/<kind>/<version>.json`

## Registry

Current manifest:

- `.out/registry/current/<tenant>/<surface>/manifest.json`

Version records:

- `.out/registry/records/<tenant>/<surface>/<type>/<version>.json`

## Notes

- Records are append-only and version-addressed.
- Manifest points to URIs (`file://...` in local mode).

## Read next

- Start here: [Start here](../start-here.md)
- Operate pipelines daily: [How-to: Operate pipelines daily](../how-to/operate-daily.md)
- Artifacts and versioning: [Artifacts and versioning](../explanation/artifacts-and-versioning.md)
- Roll back the manifest: [How-to: Roll back to a previous artifact version](../how-to/rollback-manifest.md)
