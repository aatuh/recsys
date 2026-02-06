---
tags:
  - project
  - security
---

# Security Policy

## Reporting a vulnerability

Please **do not** open public GitHub issues for security vulnerabilities.

Preferred reporting options:

1. **GitHub Security Advisories / Private vulnerability reporting** (if enabled for this repo)
2. Email: `security@recsys.app`

Include:

- A clear description of the issue and potential impact
- Steps to reproduce (PoC if possible)
- Affected versions/commit hashes
- Any suggested fixes or mitigations

## Coordinated disclosure

We follow coordinated disclosure:

- We will acknowledge receipt within **72 hours** (best effort)
- We will work with you on a disclosure timeline where feasible
- We will credit reporters in release notes unless you prefer anonymity

## Supported versions

- We aim to support the latest stable minor release line.
- Security fixes may be backported at our discretion, depending on severity and effort.

## Security hardening guidance (high level)

- Run the service with least privilege
- Keep secrets out of images; use environment variables or a secrets manager
- Restrict network access; prefer private networking for internal deployments
- Monitor logs and metrics for anomalous traffic
