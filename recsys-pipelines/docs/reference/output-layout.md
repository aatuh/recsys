
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
