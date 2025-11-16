# RecSys in Plain Language

This page explains RecSys without machine learning jargon. It focuses on what the system does for your business, the kinds of problems it solves, and what it is **not** meant to do.

---

## What RecSys is

RecSys is a service that helps you decide **which items to show which users, and in what order**.

- Your apps send RecSys a list of items, users, and events (views, clicks, purchases).
- RecSys returns **ranked recommendations** for each surface (home feed, PDP widget, email, etc.).
- You control the rules and safety rails so experiments cannot silently hurt your key business metrics.

You can think of it as a **recommendation control panel** that sits between your data and your customer experiences.

---

## Problems RecSys helps you solve

- **Stale or generic recommendations** – Replace hard-coded “top sellers” with up-to-date, personalized suggestions.
- **Expensive manual merchandising** – Reduce time spent hand-curating carousels, seasonal lists, and cross-sell bundles.
- **Inconsistent experiences across surfaces** – Use the same policies across home, PDP, search, email, and more.
- **Hidden risks in rule changes** – Catch bad changes before they reach customers using simulations and guardrails.
- **Lack of audit trail** – Answer “why did the system show this?” with stored decisions and evidence.

---

## KPIs RecSys typically affects

Exact impact depends on your business, but RecSys is usually introduced to improve:

- **Engagement** – Click-through rate (CTR), depth of session, time on site.
- **Conversion** – Add-to-cart rate, checkout completion, trial starts.
- **Revenue** – Revenue per session, average order value, attach rate for upsells.
- **Catalog health** – Exposure for long-tail items, balance across merchants or categories.

RecSys does not guarantee specific numbers, but it gives you a controlled way to **test** and **measure** these outcomes.

---

## What RecSys is *not*

RecSys focuses on ranking and guardrails for recommendations. It is **not** intended to be:

- A **data warehouse** or general analytics store.
- A **generic ML platform** for arbitrary models.
- A **feature store** for all your customer or item attributes.
- A replacement for your existing **BI dashboards** or experimentation tooling.

It integrates with those systems instead of replacing them.

---

## How it fits into your architecture

At a high level, RecSys sits between your apps and your data:

```text
Your apps and services
        │
        ▼
   RecSys HTTP API
   - ingest items/users/events
   - apply policies, rules, guardrails
   - rank and filter candidates
        │
        ▼
Ranked recommendations
shown in your UI (web, app, email)
```

- Your systems remain the **source of truth** for catalog, users, and events.
- RecSys consumes a copy of that data, makes ranking decisions, and returns ordered lists.
- You control when to roll out changes and how strictly guardrails should protect your KPIs.

If you want a deeper product narrative and rollout story, continue with `docs/business_overview.md`.

