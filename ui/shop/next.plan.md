# Next.js Shop Demo connected to Recsys

## Scope

- Build `shop/` Next.js app (App Router, TS) to act as a simple ecommerce front-end connected to recsys. Anonymous admin-style CRUD for products, users, carts, orders; comprehensive client telemetry persisted to SQLite and forwarded to recsys.

## Tech stack

- Next.js 15 (App Router), TypeScript, Edge runtime where suitable.
- Styling: TailwindCSS + shadcn/ui (Radix).
- Data: SQLite via Prisma (typed schema + migrations).
- Validation: Zod (+ react-hook-form).
- Data fetching/state: next/app server components + route handlers; client-side for carts via React Query.
- Recsys API client: generated via existing codegen, targeting `shop/src/lib/api-client`.

## Database (Prisma, SQLite)

Tables (minimal):

- `products`: id, sku, name, description, price, currency, brand, category, imageUrl, stockCount, tags[], createdAt, updatedAt.
- `users`: id, displayName, traits JSON, createdAt.
- `carts`: id, userId, createdAt, updatedAt.
- `cart_items`: id, cartId, productId, qty, unitPrice.
- `orders`: id, userId, total, currency, createdAt.
- `order_items`: id, orderId, productId, qty, unitPrice.
- `events`: id (uuid), userId, productId, type ('view'|'click'|'add'|'purchase'|'custom'), value (float), ts, meta JSON, sourceEventId, recsysStatus ('pending'|'sent'|'failed'), sentAt.
Notes:
- Use `sourceEventId` to dedupe via recsys `EventsBatch`.
- Seeders create ~100 products, 10 users, some carts/orders.

## Recsys integration

- Endpoints used:
  - Ingest: `/v1/items:upsert`, `/v1/users:upsert`, `/v1/events:batch`.
  - List (optional tooling): `/v1/items`, `/v1/users`, `/v1/events`.
  - Recs: `/v1/recommendations`, `/v1/items/{item_id}/similar`.
- Env:
  - `RECSYS_API_BASE_URL` (default `http://localhost:8000`).
  - `RECSYS_NAMESPACE` (default `default`).
- Event type mapping to recsys `Event.Type`: view→0, click→1, add→2, purchase→3, custom→4.
- Sync rules:
  - On product create/update/delete → upsert/delete recsys items.
  - On user create/select → upsert recsys users.
  - On any tracked event → write to SQLite, attempt forward to recsys.
  - Retry failed events via a small admin UI action.

## App structure (key paths)

- `shop/app/(shop)/layout.tsx`, `shop/app/(shop)/page.tsx` (home).
- `shop/app/products/*` (list, filters/pagination, create/edit/view).
- `shop/app/products/[id]/page.tsx` (detail + similar/reco widgets).
- `shop/app/cart/page.tsx`, `shop/app/checkout/page.tsx`.
- `shop/app/users/*` (picker + CRUD).
- `shop/app/events/page.tsx` (event log with filters/pagination).
- `shop/app/api/*` (route handlers: products, users, carts, orders, events, seed, retry-forward).
- `shop/src/server/db/*` (Prisma client, schema init).
- `shop/src/server/repositories/*` (data access per aggregate).
- `shop/src/server/services/recsys.ts` (thin typed client wrapper).
- `shop/src/lib/telemetry/*` (client hooks + beacon/fetch failover).
- `shop/src/components/*` (shadcn-based UI primitives, DataTable).
- `shop/src/lib/api-client/*` (generated from Swagger).

## UX flows

- Login: header user selector (no auth). Persist in cookie/localStorage.
- Browse/Filters: server-rendered list with query params; client filters synced to URL; DataTable with column visibility, sort, export CSV.
- Product detail: large image, details, add-to-cart; log view click.
- Cart/Checkout: basic flow, then order creation + purchase events.
- Telemetry: track link/banner clicks, product opens, add-to-cart, purchases; mark `recommended=true` when coming from recommendation.
- Admin-lite: full CRUD pages (anon), events viewer with retry.

## Event pipeline

- Client hook `useTelemetry()` posts to `/api/events` using `navigator.sendBeacon` (fallback to `fetch`).
- API writes to SQLite and forwards via recsys client; failures marked `failed` and visible in UI.
- Batch forwarding for noisy interactions; drain on interval via `/api/events/flush` (invoked by user action or server action timer).

## Recommendations surfaces

- Home: "Top picks for you" (calls `/v1/recommendations`).
- PDP: "Similar items" (calls `/v1/items/{id}/similar`).
- Both include `include_reasons` optionally, render reasons tooltip.

## Tooling & ops

- Add `shop/Dockerfile` and `shop/entrypoint.sh`.
- Extend `docker-compose.yml` with `shop` service (`3002:3002`).
- `shop/.env.example` documenting required envs (DB path, recsys vars).
- Extend codegen to populate `shop/src/lib/api-client`.
- NPM scripts: dev, build, start, lint, typecheck, prisma:migrate, seed.

## Testing

- Unit: event mapping, repositories (SQLite in-temp), recsys client map.
- ESLint + typecheck gates.

## Security & DX

- No auth by design; restrict destructive actions behind a UI toggle.
- Input validation (Zod), server-side constraints (Prisma).
- Clear logger, error boundaries, toast feedback.

## Acceptance

- Can browse, filter, CRUD products/users; add to cart and checkout.
- Events visible with filters and retriable forwarding.
- Recommendations render and clicks tracked, forwarded to recsys.
- Docker `make dev` runs API + shop + swagger; shop available at 3002.

---

## Epics and Tickets

### Epic E1: Foundation & Scaffolding

Epic description: Create a new Next.js app in `shop/` with TailwindCSS and shadcn/ui, establish linting/formatting, and baseline app layout.

- [x] SHOP-1 — Scaffold Next.js app
- Description: Create `shop/` with Next.js 15, TypeScript, App Router, and pnpm scripts (dev/build/start/lint/typecheck). Configure base tsconfig and `.gitignore`.
- Deliverables: Running dev server; `app/` directory; working TS.
- Dependencies: none.

- [x] SHOP-2 — TailwindCSS setup
- Description: Add Tailwind, configure `tailwind.config.ts`, `postcss.config.js`, global CSS with base/components/utilities; include a base theme and CSS reset.
- Deliverables: Utility classes available; sample page uses Tailwind.
- Dependencies: SHOP-1.

- [ ] SHOP-3 — Install shadcn/ui and base components
- Description: Add shadcn/ui CLI, install Button, Input, Select, Dialog, Toaster, Tooltip, Table, Sheet, Tabs; configure theme tokens.
- Deliverables: Components render; Toaster wired globally.
- Dependencies: SHOP-2.

- [ ] SHOP-4 — Lint/format/typecheck baseline
- Description: Configure ESLint (Next + TS rules), Prettier, and `pnpm` scripts. Add CI-friendly commands.
- Deliverables: `pnpm lint` and `pnpm typecheck` pass on scaffold.
- Dependencies: SHOP-1.

### Epic E2: Database & ORM (SQLite + Prisma)

Epic description: Introduce Prisma, model schema, migrations, and seed.

- [x] SHOP-10 — Add Prisma and SQLite configuration
- Description: Install Prisma, set `DATABASE_URL` to `file:./dev.db`, generate client. Create `src/server/db/client.ts` for a singleton.
- Deliverables: `prisma` folder present; client can connect.
- Dependencies: E1 done.

- [x] SHOP-11 — Define Prisma schema for core tables
- Description: Model products, users, carts, cart_items, orders, order_items, events with required fields and indexes for filters.
- Deliverables: `schema.prisma` with relations and constraints.
- Dependencies: SHOP-10.

- [ ] SHOP-12 — Migrations and client generation
- Description: Create initial migration, generate Prisma client, verify migrations apply on clean DB.
- Deliverables: Migration files; `pnpm prisma:migrate` works.
- Dependencies: SHOP-11.

- [x] SHOP-13 — Seed script for demo data
- Description: Add `scripts/seed.ts` to create ~100 products, 10 users, sample carts/orders, and a few events. Use realistic prices/categories/brands.
- Deliverables: `pnpm seed` populates DB idempotently.
- Dependencies: SHOP-12.

### Epic E3: Repository Layer

Epic description: Type-safe data access with pagination/filtering and transactions.

- [x] SHOP-20 — ProductRepository with filters/pagination
- Description: Methods: list(filter/sort/paginate), getById, create, update, delete. Filters: text, brand, category, tags, price range, availability.
- Deliverables: Unit-tested repository.
- Dependencies: SHOP-12.

- [x] SHOP-21 — UserRepository
- Description: Methods: list/paginate, getById, create, update, delete. Traits as JSON.
- Deliverables: Unit tests included.
- Dependencies: SHOP-12.

- [x] SHOP-22 — CartRepository
- Description: Methods: getOrCreate(userId), addItem(productId, qty), updateQty, removeItem, clear, computeTotals; validates stock.
- Deliverables: Unit tests (incl. stock edge cases).
- Dependencies: SHOP-12.

- [x] SHOP-23 — OrderRepository (transactional)
- Description: Create order from cart in a transaction; write order_items, reduce stock, clear cart, return order summary.
- Deliverables: Unit tests with transaction rollback cases.
- Dependencies: SHOP-22.

- [x] SHOP-24 — EventRepository
- Description: Methods: create (single/batch), list with filters, mark status (pending/sent/failed), resend batch selection.
- Deliverables: Unit tests; pagination verified.
- Dependencies: SHOP-12.

### Epic E4: API Routes (Next.js Route Handlers)

Epic description: Implement `/app/api/*` endpoints for CRUD, cart, checkout, events, and seed.

- [x] SHOP-30 — Products API
- Description: Routes: GET `/api/products` (filters/sort/paginate), POST create; GET `/api/products/[id]`, PUT/PATCH, DELETE.
- Deliverables: Zod validation; error responses with helpful messages.
- Dependencies: SHOP-20.

- [x] SHOP-31 — Users API
- Description: Routes: GET `/api/users`, POST; GET `/api/users/[id]`, PATCH, DELETE.
- Deliverables: Zod validation; consistent JSON shapes.
- Dependencies: SHOP-21.

- [x] SHOP-32 — Cart API
- Description: Routes: GET `/api/cart?userId=...`; POST `/api/cart/items` add; PATCH `/api/cart/items/[productId]` qty; DELETE item/clear.
- Deliverables: Returns totals; handles stock/validation errors.
- Dependencies: SHOP-22.

- [x] SHOP-33 — Checkout API
- Description: POST `/api/checkout` to create order from cart; record purchase events.
- Deliverables: Transactional; returns order summary + id.
- Dependencies: SHOP-23, SHOP-24.

- [x] SHOP-34 — Orders API
- Description: Routes: GET `/api/orders` (paginate), GET `/api/orders/[id]`.
- Deliverables: Simple list/detail.
- Dependencies: SHOP-23.

- [x] SHOP-35 — Events API (ingest/list/flush)
- Description: POST `/api/events` (single/batch), GET `/api/events` (filters), POST `/api/events/flush` to forward pending.
- Deliverables: Status tracking; batch size limits.
- Dependencies: SHOP-24.

- [x] SHOP-36 — Seed API
- Description: POST `/api/seed` to reinitialize data for demos.
- Deliverables: Idempotent; optional `force` param.
- Dependencies: SHOP-13.

### Epic E5: Recsys API Client

Epic description: Provide a typed client and light wrapper for recsys operations.

- [x] SHOP-40 — Generate recsys API client (OpenAPI)
- Description: Extend existing codegen to emit `shop/src/lib/api-client`. Ensure it is not manually edited.
- Deliverables: Generated client committed; script to refresh.
- Dependencies: E1 done.

- [x] SHOP-41 — Recsys service wrapper
- Description: `src/server/services/recsys.ts` exposing: upsertItems, upsertUsers, batchEvents, recommendations, similarItems; centralized error handling and base URL/namespace injection.
- Deliverables: Unit tests mocking client.
- Dependencies: SHOP-40.

- [x] SHOP-42 — Environment/config wiring
- Description: Read `RECSYS_API_BASE_URL` and `RECSYS_NAMESPACE` from env; provide sane defaults; document in `.env.example`.
- Deliverables: Config module with validation.
- Dependencies: SHOP-40.

### Epic E6: Data Sync to Recsys

Epic description: Keep recsys in sync with local shop data and events.

- [x] SHOP-50 — Sync products to recsys on CRUD
- Description: On create/update/delete, call recsys items upsert/delete. Map fields (price/tags) to recsys `Item`.
- Deliverables: Integration tests or manual verifications.
- Dependencies: SHOP-30, SHOP-41.

- [x] SHOP-51 — Upsert users on create/select
- Description: On new user or when a user is selected in UI, upsert to recsys; include traits if present.
- Deliverables: Verified via API logs.
- Dependencies: SHOP-31, SHOP-41.

- [x] SHOP-52 — Forward events with idempotency
- Description: Batch forward `events` table entries with status `pending`; set `sourceEventId` to local `events.id` to dedupe.
- Deliverables: Retry strategy; status transitions updated.
- Dependencies: SHOP-35, SHOP-41.

- [x] SHOP-53 — Admin retry action for failed events
- Description: Endpoint + UI button to retry forwarding `failed` events.
- Deliverables: Works for selected rows and bulk.
- Dependencies: SHOP-52.

### Epic E7: Telemetry (Client)

Epic description: Capture user interactions and send to backend reliably.

- [x] SHOP-60 — Telemetry hook and provider
- Description: `useTelemetry()` + context to emit view/click/add/purchase/custom with `meta` and `recommended` flag.
- Deliverables: Hook API docs; sample usage in components.
- Dependencies: SHOP-35.

- [x] SHOP-61 — Beacon/fetch fallback + backoff
- Description: Use `navigator.sendBeacon` when available; fallback to fetch with retry/backoff; queue transiently if offline.
- Deliverables: Tested in devtools offline mode.
- Dependencies: SHOP-60.

- [x] SHOP-62 — Recommendation context tagging
- Description: When navigating via a recommendation, tag outbound links with a query param; telemetry reads it and marks events accordingly.
- Deliverables: Badge in UI indicating "Recommended".
- Dependencies: SHOP-100/101.

- [x] SHOP-63 — Zod validation for event payloads
- Description: Validate client-side before POST; drop obviously invalid payloads; log warnings.
- Deliverables: Schemas and tests.
- Dependencies: SHOP-60.

### Epic E8: UI Shell & Navigation

Epic description: Global layout, navigation, user selector, theme and toasts.

- [x] SHOP-70 — App shell and navigation
- Description: Header, sidebar/topbar navigation to Home, Catalog, Cart, Users, Events; breadcrumb component.
- Deliverables: Responsive nav; active route styles.
- Dependencies: SHOP-3.

- [x] SHOP-71 — Theme toggle, toasts, error boundary
- Description: Implement dark/light toggle, Toast provider, and top-level error boundary.
- Deliverables: Visible theme state persisted; errors render friendly UI.
- Dependencies: SHOP-3.

- [x] SHOP-72 — User selector login (no auth)
- Description: Header control to pick current user; persist in localStorage/cookie; create new user inline.
- Deliverables: Current user available via context.
- Dependencies: SHOP-31.

### Epic E9: Catalog UI

Epic description: Product browsing, filters, CRUD, and detail page.

- [x] SHOP-80 — Product list with filters/pagination
- Description: DataTable with columns (name, price, brand, category, stock); filters (text, brand, category, tags, price range, availability); server-backed pagination and sort.
- Deliverables: URL-synced query params; CSV export.
- Dependencies: SHOP-30, SHOP-70.

- [x] SHOP-81 — Product create/edit forms
- Description: RHF + Zod forms for create/edit; image URL field; client/server validation; success toasts.
- Deliverables: Modal or page forms; errors shown inline.
- Dependencies: SHOP-30, SHOP-71.

- [x] SHOP-82 — Product detail (PDP)
- Description: Display product info, large image, add-to-cart; fire view event on mount; track add-to-cart clicks.
- Deliverables: Works with telemetry; handles out-of-stock.
- Dependencies: SHOP-32, SHOP-60.

- [x] SHOP-83 — Image handling and placeholders
- Description: Fallback placeholder images; error handling for broken URLs; consistent aspect ratios.
- Deliverables: Polished grid/list visuals.
- Dependencies: SHOP-80.

### Epic E10: Cart & Checkout

Epic description: Simple cart and purchase transactions.

- [x] SHOP-90 — Cart page UI
- Description: View/update cart; change quantities; remove items; show totals and CTA to checkout.
- Deliverables: Telemetry for add/remove/update.
- Dependencies: SHOP-32, SHOP-70.

- [x] SHOP-91 — Checkout flow
- Description: Review order, confirm purchase -> calls Checkout API; transitions to success page; stock decremented.
- Deliverables: Purchase events recorded.
- Dependencies: SHOP-33, SHOP-60.

- [x] SHOP-92 — Orders list/history (optional)
- Description: List recent orders for the selected user; link to order detail.
- Deliverables: Basic, read-only.
- Dependencies: SHOP-34.

### Epic E11: Recommendation Surfaces

Epic description: Show recommendations and similar items, with reasons.

- [x] SHOP-100 — Home: Top picks for you
- Description: Call `/v1/recommendations` for current user; render grid; include reasons tooltip when available; track clicks.
- Deliverables: Works with user selector and telemetry.
- Dependencies: SHOP-41, SHOP-72.

- [x] SHOP-101 — PDP: Similar items
- Description: Call `/v1/items/{id}/similar`; render horizontally scrollable list; track clicks with `recommended=true`.
- Deliverables: PDP shows relevant items.
- Dependencies: SHOP-41, SHOP-82.

- [x] SHOP-102 — Reasons UI and badges
- Description: Tooltip or popover listing reasons; badge indicating recommended origin on tiles/links.
- Deliverables: Usable on Home and PDP.
- Dependencies: SHOP-100/101.

### Epic E12: Events Console

Epic description: Visualize and manage local telemetry events.

- [x] SHOP-110 — Events list with filters/pagination
- Description: Table with event type, user, product, ts, status; filters by date range, type, user, product; paginate server-side.
- Deliverables: Copy-to-clipboard of JSON meta.
- Dependencies: SHOP-35, SHOP-70.

- [x] SHOP-111 — Retry failed forwarding
- Description: Row/bulk action to retry forwarding failed events; shows toast outcomes.
- Deliverables: Integrates with `/api/events/flush`.
- Dependencies: SHOP-52, SHOP-110.

- [x] SHOP-112 — Event detail modal
- Description: Prettified JSON view of event `meta` and delivery history.
- Deliverables: Accessible dialog.
- Dependencies: SHOP-110.

### Epic E13: Docker & DevOps

Epic description: Containerize and run alongside existing services.

- [x] SHOP-120 — Dockerfile and entrypoint
- Description: Multi-stage Dockerfile; production serve; development entrypoint.
- Deliverables: Image builds locally.
- Dependencies: E1 done.

- [x] SHOP-121 — docker-compose service
- Description: Add `shop` service (port 3002), volumes for source and node_modules; depends on api and swagger.
- Deliverables: `make dev` starts the shop.
- Dependencies: SHOP-120.

- [x] SHOP-122 — .env.example and config loader
- Description: Document and load `DATABASE_URL`, `RECSYS_API_BASE_URL`, `RECSYS_NAMESPACE`; validate at startup.
- Deliverables: Running app with defaults.
- Dependencies: SHOP-10, SHOP-42.

### Epic E14: Quality & Testing (not needed)

Epic description: Linting, unit tests, and end-to-end smoke.

- [ ] SHOP-130 — ESLint + TypeScript checks in CI
- Description: Ensure `pnpm lint` and `pnpm typecheck` gate commits/PRs.
- Deliverables: CI or pre-commit hooks configured.
- Dependencies: SHOP-4.

- [ ] SHOP-131 — Unit tests for repositories/services
- Description: Use Vitest/Jest to test repositories and recsys wrapper; SQLite temp DB for tests.
- Deliverables: Passing tests with coverage for happy/edge paths.
- Dependencies: SHOP-20..24, SHOP-41.

- [ ] SHOP-132 — Playwright E2E smoke tests
- Description: Scenarios: browse catalog; add to cart; checkout; click a recommendation; verify events recorded.
- Deliverables: CI-friendly runner and docs.
- Dependencies: UI epics done.

- [ ] SHOP-133 — Accessibility checks
- Description: Add basic a11y lint and runtime checks (axe in dev only); keyboard navigation verified.
- Deliverables: No critical issues.
- Dependencies: UI epics done.

- [ ] SHOP-134 — Error/logging utilities
- Description: Structured server logging; client error boundary logs to console; redaction for sensitive data.
- Deliverables: Clear logs for failures.
- Dependencies: SHOP-71.

### Epic E15: Documentation (not needed)

Epic description: Developer and operator documentation for the shop.

- [x] SHOP-140 — README and quick start
- Description: How to run via docker-compose and locally; env vars; scripts; seeding.
- Deliverables: `shop/README.md` complete.
- Dependencies: E1–E13.

- [ ] SHOP-141 — Architecture overview
- Description: Document layers (UI, API routes, repositories, DB), data flow, telemetry pipeline, recsys integration.
- Deliverables: `shop/docs/architecture.md`.
- Dependencies: E2–E6.

- [ ] SHOP-142 — API route reference
- Description: List major `/api/*` endpoints, payload shapes, and example requests.
- Deliverables: `shop/docs/api.md`.
- Dependencies: E4.

### Epic E16: Accessibility & UX polish (not needed)

Epic description: Ensure the demo is usable and pleasant on desktop and mobile.

- [ ] SHOP-150 — Keyboard navigation and focus states
- Description: Ensure all interactive elements are reachable; visible focus styles.
- Deliverables: Verified by keyboard-only pass.
- Dependencies: UI epics done.

- [ ] SHOP-151 — Empty/loading states & skeletons
- Description: Add skeletons, spinners, and empty-state copy for major views.
- Deliverables: Consistent UX across network states.
- Dependencies: UI epics done.

- [ ] SHOP-152 — Responsive layout
- Description: Test on small screens; adjust grids, typography, and navigation for mobile.
- Deliverables: No horizontal scroll; readable content.
- Dependencies: UI epics done.
