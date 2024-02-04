# Docs versioning

## Who this is for

- Maintainers who publish releases
- Customers who need docs that match what they are running

## What you will get

- What “dev” and “latest” mean on the hosted site
- How to publish docs for a specific release tag
- How to preview older versions locally

## Versioning model

This MkDocs site is versioned using `mike` (versions are stored on the `gh-pages` branch).

Hosted paths:

- `/dev/` = docs built from `master` (can change every merge)
- `/latest/` = alias to the newest released version
- `/vX.Y.Z/` = docs for a specific release tag

## Release tags

Docs versions are published from git tags.

Recommended suite tag format:

- `recsys-suite/vX.Y.Z` (SemVer)

The Pages workflow strips the `recsys-suite/` prefix, so the version label becomes `vX.Y.Z`.

## Publishing a new version

1. Create a tag:

   ```bash
   git tag recsys-suite/v0.1.0
   ```

2. Push the tag:

   ```bash
   git push origin recsys-suite/v0.1.0
   ```

3. GitHub Actions deploys:
   - `/v0.1.0/`
   - updates `/latest/` to point to `v0.1.0`

## Local preview (optional)

To preview multiple versions locally, install `mike` and serve the `gh-pages` branch:

```bash
python -m venv .venv
. .venv/bin/activate
python -m pip install mkdocs mkdocs-material mkdocs-swagger-ui-tag pymdown-extensions mike

mike serve
```

## Read next

- Docs update policy: [`project/docs-per-release.md`](docs-per-release.md)
- What’s new: [`whats-new/index.md`](../whats-new/index.md)
