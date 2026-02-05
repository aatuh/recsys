
# Artifact schemas

Artifacts are JSON documents intended for serving systems.

Currently:

- Popularity artifact v1
- Co-occurrence artifact v1
- Implicit artifact v1 (collaborative)
- Content similarity artifact v1
- Session sequence artifact v1
- Manifest v1

Schemas:

- `schemas/artifacts/manifest.v1.json`
- (recommended) `schemas/artifacts/popularity.v1.json`
- (recommended) `schemas/artifacts/cooc.v1.json`
- (recommended) `schemas/artifacts/implicit.v1.json`
- (recommended) `schemas/artifacts/content_sim.v1.json`
- (recommended) `schemas/artifacts/session_seq.v1.json`

The builtin validator enforces structural rules.
See `explanation/artifacts-and-versioning.md`.

## Read next

- Start here: [`start-here.md`](../start-here.md)
- Add artifact type: [`how-to/add-artifact-type.md`](../how-to/add-artifact-type.md)
- Output layout: [`reference/output-layout.md`](output-layout.md)
- Artifacts and versioning: [`explanation/artifacts-and-versioning.md`](../explanation/artifacts-and-versioning.md)
