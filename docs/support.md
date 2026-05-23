# Support

RecSys support is async-first. Public community support and commercial support have different expectations.

## Community support

- Use GitHub Issues for reproducible bugs and feature requests.
- Include version or commit hash, deployment mode, logs, configuration summary, and minimal reproduction steps.
- Do not include secrets, customer data, private tokens, or vulnerability details in public issues.

Community support has no guaranteed response time.

## Commercial support

Commercial support is defined by the signed order form. Defaults from the recovered commercial references are:

| Plan | Default expectation |
| --- | --- |
| Commercial Evaluation | Best-effort async, no SLA. |
| Starter | Best-effort async, no SLA. |
| Growth | Async first-response target within 2 business days, no SLA unless purchased. |
| Enterprise | Defined in the signed order form. |

Premium 8x5 or 24x7 support can be captured as a custom order-form term for Growth or Enterprise.

## Before requesting help

```bash
git rev-parse --short HEAD
docker compose ps
docker compose logs --tail=100 api
make docs-check
```

Expected result: the support request includes the running version, service status, relevant logs, and whether docs and
local checks pass.

## Contact

- Public licensing questions: open a GitHub issue titled `Licensing question`.
- Confidential commercial inquiries: message [Aatu Harju on LinkedIn](https://www.linkedin.com/in/aatu-harju/).
- Vulnerabilities: follow [Security](security.md), not normal support channels.
