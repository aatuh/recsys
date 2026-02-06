---
tags:
  - project
  - support
---

# Support

This project supports **self-serve adoption**. We keep support lightweight and mostly asynchronous.

## Community support (free)

- GitHub Issues for bugs and feature requests
- Discussions for questions
- RecSys Copilot (Custom GPT) for docs Q&A: [`chatgpt.com/g/.../recsys-copilot`](https://chatgpt.com/g/g-68c82a5c7704819185d0ff929b6fff11-recsys-copilot)

Do not paste secrets or customer data into external tools.

We do not guarantee response times for free support.

## Commercial support (paid)

Commercial customers get the support level defined in their agreement.

Typical support channels include:

- Private support email which will be provided upon commercial agreement
- Private issue tracker or GitHub private issues

**No-meetings policy (default):**

- Support is **async-first**
- Calls are only by exception and must be pre-agreed and time-boxed

## Support expectations by plan (typical)

Support terms are defined in your agreement. This table summarizes the typical differences between tiers:

| Plan | Response expectations | Channels | Escalation |
| --- | --- | --- | --- |
| Commercial Evaluation | Best-effort async (no SLA) | Email / private issues (provided during eval) | By exception |
| Starter | Best-effort async (no SLA) | Support email / private issues | By exception |
| Growth | Typically within 2 business days (async) | Support email / private issues | By exception |
| Enterprise | Custom (can include SLA/premium support) | Custom | Defined in agreement |

Notes:

- Premium support/SLA (8×5 or 24×7) is an optional add-on for Growth/Enterprise (see `pricing/index.md`).

## What we can help with

- Installation and upgrade guidance
- Reproducible bug investigation (with logs/configs)
- Security patch guidance

## What we do not provide by default

- 24/7 on-call
- Operating your infrastructure
- Unlimited custom development without a fixed scope

## Before opening an issue

Please include:

- Version/commit hash
- Deployment mode (docker compose / k8s / helm)
- Logs and minimal reproduction steps
