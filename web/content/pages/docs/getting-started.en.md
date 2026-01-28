---
title: "Getting started"
description: "Run the stack and verify the Foo workflow end to end."
layout: "default"
---

## 1. Boot the stack

```bash
make dev
cd api && make migrate-up
```

## 2. Start the web app

```bash
cd web
pnpm install
pnpm dev
```

## 3. Create your first Foo

Open `http://localhost:3000/foos`, add a name, and submit. The list will refresh using the generated API client.
