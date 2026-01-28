---
title: "Architecture"
description: "How adapters, services, and UI stay decoupled."
layout: "default"
---

## Layers

1. **API** exposes DTOs and swagger types.
2. **API client** calls HTTP and returns DTOs.
3. **Domain adapters** map DTOs into domain models.
4. **Domain services** enforce rules and naming.
5. **UI** consumes domain data through hooks.

## Why it matters

- Replace the API without touching UI logic.
- Centralize mapping and validation in one place.
- Keep feature code clean and testable.
