# Commercial Use & Licensing

This page explains how to purchase and use a **commercial license** for parts of this repository that are otherwise
licensed under **AGPLv3**.

This page is informational and describes our commercial offering at a high level.

## Why a commercial license?

The AGPLv3 is designed for software used over a network. If you modify AGPL-covered code and provide network access to users,
AGPLv3 requires offering those users access to the Corresponding Source of your modified version (see Section 13).

A commercial license allows you to use the covered components under alternative terms, typically enabling:

- Internal or external deployment without AGPL source-offer obligations (subject to the commercial agreement)
- Keeping modifications private
- Using the software in proprietary stacks

## What is covered?

Commercial licensing applies to the components that are AGPLv3 in this repository, including typically:

- `recsys-service` (serving API; `docker compose` service name: `api`)
- `recsys-algo` (algorithms used by the service)
- `recsys-pipelines` (batch pipelines and artifact generation)

`recsys-eval` remains Apache-2.0.

## What you get when you buy

A typical commercial purchase includes:

- A signed commercial license grant (agreement + order form)
- A license token/file for bookkeeping (optional, **not DRM**)
- Access to **commercial release artifacts** (e.g., signed container images) if you offer those
- Security and patch releases according to the purchased tier
- Optional support terms (if purchased)

See [`pricing.md`](pricing.md) for tier definitions.

## How to buy

Recommended low-friction flow:

1. Choose a tier in [`pricing.md`](pricing.md)
2. Request a commercial license (public or private inquiry)
3. Receive:
   - commercial license paperwork,
   - delivery instructions for artifacts (if applicable),
   - optional support contact

How to request a commercial license:

- Open a GitHub issue titled **"Commercial licensing inquiry"** (public), or
- If your inquiry must be confidential, say so in the issue and we will move to a private channel before exchanging
  details.

## Evaluation licenses (optional)

If you offer evaluation terms, define them clearly:

- Duration (e.g., 30 days)
- Limits (e.g., 1 deployment, non-production)
- What’s included (e.g., private artifact access)

Document details in `docs/licensing/eval_license.md` if you provide this.

## Where are the legal terms?

Commercial terms live in:

- `docs/licensing/commercial_license.md`
- `docs/licensing/order_form.md`
