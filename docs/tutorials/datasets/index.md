---
diataxis: tutorial
tags:
  - tutorial
  - datasets
  - reference
---
# Datasets (tutorial fixtures)
In this tutorial you will follow a guided walkthrough and verify a working result.


## Who this is for

- Developers running tutorials locally
- Anyone who needs a **small, deterministic fixture** for smoke tests and docs examples

## What you will get

- A list of built-in datasets used by tutorials
- Guidance on when to use each dataset

## Available datasets

- **[Tiny dataset](tiny/README.md)**  
  Intentionally small and human-readable. Good for docs, smoke tests, and demos.
- **Ecommerce mini dataset**
  Synthetic commerce catalog, pipeline exposures, and eval fixtures for the commercial proof kit. Source:
  `examples/data/ecommerce-mini/`.

## When to use which

- Use **Tiny** when you want:
  - fast local runs
  - deterministic debugging
  - copy/paste-friendly examples
- Use **Ecommerce mini** when you want:
  - a buyer-facing proof bundle
  - a realistic home recommendation surface
  - served recommendations plus evaluation evidence

## Read next

- Tutorials index: [Tutorials](../index.md)
- Quickstart (10 minutes): [Quickstart (10 minutes)](../quickstart.md)
- Local end-to-end: [local end-to-end (service → logging → eval)](../local-end-to-end.md)
