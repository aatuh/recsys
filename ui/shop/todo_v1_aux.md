This document defines the canonical payloads we will send from the demo “Online Shop” to Recsys so it behaves like a real ecommerce integration. It covers items, users, events, event-type weights/decay, and standard metadata conventions. All JSON examples are exactly what the Recsys API expects.
0) Namespaces & endpoints

    Namespace: default (configurable)

    Core endpoints

        Upsert items: /v1/items:upsert

        Upsert users: /v1/users:upsert

        Batch events: /v1/events:batch

        Get recommendations: /v1/recommendations

        Get similar items: /v1/items/{item_id}/similar

        Upsert event-type config: /v1/event-types:upsert (name may appear as event_type_config:upsert in client)

1) Items (Products)

Purpose: catalog used for candidate retrieval and ranking; tags enable constraints & brand/category caps; available and price are first-class signals.

Schema
{
  "item_id": "<product.id>",
  "available": true,
  "price": 129.99,                    // optional but highly recommended
  "tags": [
    "brand:Umbrella",                 // brand prefix
    "category:Electronics",           // category prefix
    "cat:cameras",                    // optional short alias
    "color:black", "material:alloy"  // optional facets as tags
  ],
  "props": {
    "name": "Compact Bracelet",
    "sku": "SKU-38",
    "url": "/products/<product.id>",
    "image_url": "https://.../image.jpg",
    "brand": "Umbrella",
    "category": "Electronics",
    "currency": "USD",
    "description": "Sturdy, stylish bracelet.",
    "attributes": {                   // free-form detailed attributes
      "color": "Black",
      "material": "Alloy",
      "size": "S/M"
    }
  },
  "embedding": [ /* optional float[]; 384-d preferred */ ]
}

Required fields: item_id. Recommended: available, price, tags (at least brand/category), and props.name.

Tag conventions

    Brand tags: brand:<BrandName> (e.g., brand:Acme)

    Category tags: category:<TopLevel> and optionally cat:<lower-subcat> (e.g., category:Electronics, cat:headphones)

    Facets as tags: use key:value (lowercase key) for coarse filters (e.g., color:black, material:leather).

    These tags power brand/category caps and include_tags_any constraints, so use them consistently.

Availability

    available = stockCount > 0. Update promptly on stock changes and after checkout.

Embeddings (optional)

    Use a 384‑d text embedding of (title + description + brand + category) if available; omit if not. You can backfill later without breaking anything.

2) Users

Purpose: anchor personalization; traits drive profiling and (optionally) segment assignment.

Schema
{
  "user_id": "<user.id>",
  "traits": {
    "display_name": "Jane Doe",            // optional, non-PII preferred
    "locale": "en-US",
    "country": "US",
    "device": "mobile|desktop",
    "signup_ts": "2025-10-26T12:00:00Z",
    "last_seen_ts": "2025-10-27T14:58:21Z",
    "loyalty_tier": "bronze|silver|gold|vip",
    "newsletter": true,
    "preferred_categories": ["Electronics", "Toys"],
    "brand_affinity": { "Umbrella": 0.8, "Acme": 0.2 },
    "price_sensitivity": "low|mid|high",
    "lifetime_value_bucket": "L|M|H"
  }
}

Required: user_id. Keep PII out; stable pseudonymous IDs are ideal.
3) Events & Event Dictionary

Purpose: behavioral signals for popularity, co‑visitation, collaborative filtering, and bandits.

Event types

    0 = view — product impression/open

    1 = click — intent signals (e.g., click-through from a list/banner)

    2 = add — add-to-cart (stronger intent)

    3 = purchase — conversion signal (per line item)

    4 = custom — free-form events (order, pageview, search, etc.)

Common envelope
{
  "user_id": "<user.id>",
  "item_id": "<product.id>",          // omit for order-level events
  "type": 0,                           // see mapping above
  "value": 1,                          // see per-type semantics below
  "ts": "2025-10-27T14:49:02.942Z",
  "meta": { /* see below */ },
  "source_event_id": "<local_event_uuid>" // enables idempotency
}

Per-type semantics

    view: value = 1. meta should include surface, widget, and whether item was recommended (see below).

    click: value = 1. meta.href, meta.text when applicable, plus recommendation context.

    add: value = quantity. meta.cart_id, meta.unit_price, meta.currency.

    purchase (recommended): emit one event per line item with value = quantity. meta.order_id, meta.unit_price, meta.currency, and optional meta.line_item_id.

    custom (order-level summary, optional): if you want an order summary, send as type = 4 (custom) without item_id to avoid double-counting conversions per item. Use meta.kind = "order", meta.order_id, meta.total, meta.currency, and meta.items (array of {item_id, qty, unit_price}).

Recommendation context (add to meta where applicable)
{
  "surface": "home|pdp|cart|checkout",
  "widget": "home_top_picks|pdp_similar|...",
  "recommended": true,                 // if came from a recommendation
  "request_id": "<uuid-from-recs-call>",
  "rank": 1,                           // position in the list
  "experiment": "exp-2025-10-27A",
  "ab_bucket": "B",
  "session_id": "<browser-session-id>",
  "referrer": "/products/..."
}

Examples

    View on PDP

{
  "user_id": "u123",
  "item_id": "itm_A",
  "type": 0,
  "value": 1,
  "ts": "2025-10-27T13:42:21Z",
  "meta": { "surface": "pdp" }
}

    Click from a recommended carousel

{
  "user_id": "u123",
  "item_id": "itm_B",
  "type": 1,
  "value": 1,
  "ts": "2025-10-27T14:48:58Z",
  "meta": {
    "surface": "home",
    "widget": "home_top_picks",
    "recommended": true,
    "request_id": "6f5f...",
    "rank": 3,
    "href": "/products/itm_B",
    "text": "Compact Bracelet"
  }
}

    Add to cart (2 units)

{
  "user_id": "u123",
  "item_id": "itm_C",
  "type": 2,
  "value": 2,
  "ts": "2025-10-27T14:49:03Z",
  "meta": { "cart_id": "c_9", "unit_price": 39.99, "currency": "USD" }
}

    Purchase (line item)

{
  "user_id": "u123",
  "item_id": "itm_C",
  "type": 3,
  "value": 2,
  "ts": "2025-10-27T14:58:24Z",
  "meta": { "order_id": "o_45", "unit_price": 39.99, "currency": "USD" }
}

    Order summary (custom)

{
  "user_id": "u123",
  "type": 4,
  "value": 477.37, // total
  "ts": "2025-10-27T14:58:24Z",
  "meta": {
    "kind": "order",
    "order_id": "o_45",
    "currency": "USD",
    "items": [
      { "item_id": "itm_A", "qty": 1, "unit_price": 199.99 },
      { "item_id": "itm_C", "qty": 2, "unit_price": 39.99 }
    ]
  }
}

Event hygiene

    Always set ts in ISO‑8601 UTC.

    Use unique source_event_id (your local events.id) to dedupe on retries.

    Ensure items & users are upserted before sending events that reference them.

4) Event-type weights & decay (per namespace)

Set weights and half-lives to tune recency and importance. Recommended ecommerce defaults:
type	code	weight	half_life_days	active
view	0	0.05	3	true
click	1	0.20	7	true
add	2	0.70	21	true
purchase	3	1.00	60	true
custom	4	0.10	3	true

Upsert payload
{
  "namespace": "default",
  "types": [
    { "type": 0, "name": "view", "weight": 0.05, "half_life_days": 3,  "is_active": true },
    { "type": 1, "name": "click", "weight": 0.20, "half_life_days": 7,  "is_active": true },
    { "type": 2, "name": "add",   "weight": 0.70, "half_life_days": 21, "is_active": true },
    { "type": 3, "name": "purchase", "weight": 1.00, "half_life_days": 60, "is_active": true },
    { "type": 4, "name": "custom", "weight": 0.10, "half_life_days": 3,  "is_active": true }
  ]
}

    Tune half-lives per category if you have fast/slow-moving inventory.

5) Recommendation-time constraints & caps

    Price filter: pass constraints.price_between: [min, max].

    Tag filter: pass constraints.include_tags_any: ["category:Electronics", "cat:cameras"].

    Brand/category caps: runtime overrides brand_cap and category_cap require consistent tagging (brand:*, category:*).

6) Data flow & ordering

    Upsert items (and mark availability) → Upsert users → Send events.

    Flush pending events in batches; provide source_event_id for idempotency.

    Re-upsert items on CRUD or stock/price change.

7) Quality & risk notes

    Cold start: seed with item popularity (views) and make sure tags/price are present so fallback retrieval works.

    Co‑visitation: ensure add and purchase events are per line item; this drives basket co-occurrence.

    Stale inventory: keep available=false for OOS to avoid recommending them; a nightly reconciliation is fine in addition to realtime updates.

    Privacy: avoid PII in traits and meta; pseudonymous user_id only.

8) Implementation checklist (shop app)


9) Sample batched requests

Upsert 2 items
{
  "namespace": "default",
  "items": [
    {
      "item_id": "p_101",
      "available": true,
      "price": 199.99,
      "tags": ["brand:Acme", "category:Electronics", "cat:headphones", "color:black"],
      "props": {"name": "Acme ANC Headphones", "sku": "SKU-101", "url": "/products/p_101"}
    },
    {
      "item_id": "p_102",
      "available": false,
      "price": 29.99,
      "tags": ["brand:Globex", "category:Books", "cat:nonfiction"],
      "props": {"name": "Globex Strategy", "sku": "SKU-102", "url": "/products/p_102"}
    }
  ]
}

Upsert users
{
  "namespace": "default",
  "users": [
    { "user_id": "u123", "traits": {"country": "US", "device": "mobile", "loyalty_tier": "silver"} },
    { "user_id": "u456", "traits": {"country": "FI", "device": "desktop", "newsletter": true} }
  ]
}

Batch events
{
  "namespace": "default",
  "events": [
    {"user_id": "u123", "item_id": "p_101", "type": 0, "value": 1, "ts": "2025-10-27T10:00:00Z"},
    {"user_id": "u123", "item_id": "p_101", "type": 1, "value": 1, "ts": "2025-10-27T10:00:02Z",
     "meta": {"surface": "home", "widget": "home_top_picks", "recommended": true, "request_id": "r-1", "rank": 1}},
    {"user_id": "u123", "item_id": "p_101", "type": 2, "value": 2, "ts": "2025-10-27T10:01:00Z",
     "meta": {"cart_id": "c123", "unit_price": 199.99, "currency": "USD"}},
    {"user_id": "u123", "item_id": "p_101", "type": 3, "value": 2, "ts": "2025-10-27T10:05:00Z",
     "meta": {"order_id": "o789", "unit_price": 199.99, "currency": "USD"}},
    {"user_id": "u123", "type": 4, "value": 399.98, "ts": "2025-10-27T10:05:00Z",
     "meta": {"kind": "order", "order_id": "o789", "currency": "USD",
               "items": [{"item_id": "p_101", "qty": 2, "unit_price": 199.99}]}}
  ]
}
10) Field-by-field mapping from shop DB

    products.id → item.item_id

    products.stockCount > 0 → item.available

    products.price → item.price; products.currency → item.props.currency

    products.brand → item.tags += ["brand:<brand>"]

    products.category → item.tags += ["category:<category>", optional "cat:<lower-subcat>"]

    products.name → item.props.name; products.sku → item.props.sku

    products.imageUrl → item.props.image_url

    products.description → item.props.description

    users.id → user.user_id

    users.traitsText (JSON) → user.traits

    events.userId → event.user_id

    events.productId → event.item_id

    events.type (view|click|add|purchase|custom) → event.type (0..4)

    events.value → event.value

    events.ts → event.ts

    events.metaText (JSON) → event.meta

    events.id → event.source_event_id

This contract is stable and safe to implement now. We can tune weights/half-lives and add embeddings later without breaking the pipeline.

Awesome — I turned your “what should we send to Recsys?” question into a full, copy-pasteable data contract you can implement right now.

You’ll find it in the canvas: “Recsys Shop — Data Contract (Items, Users, Events, Metadata)”. It defines canonical payloads, naming conventions, and example JSON for items, users, events, a clean event dictionary, event-type weights/half-life, and a field-by-field mapping from your shop DB → Recsys.

Here are the essentials so you can wire this up quickly:
TL;DR (what matters most)

    Items → send item_id, available, price, and tags like brand:Umbrella, category:Electronics (these power brand/category caps and request-time constraints). Put rich stuff into props and optional embedding. The Recsys item upsert supports available, price, tags, props, and embedding out of the box.

    Users → user_id plus a small traits object (non-PII): locale, device, loyalty_tier, preferred_categories, etc. The user upsert accepts a free-form traits object.

    Events → use the built-in mapping view=0, click=1, add=2, purchase=3, custom=4. Send line-item purchase events (value=qty), and if you also want an order summary, emit it as custom (type 4) without item_id to avoid double counting. Your event batch/flush already forwards meta when present.

Why this fits your stack

    Your Next.js shop already calls these exact endpoints: /v1/items:upsert, /v1/users:upsert, /v1/events:batch, /v1/recommendations, /v1/items/{id}/similar. The event type mapping is the same one we standardize on.

    The generated client models show the shapes:
    – Item supports available, price, tags, props, optional embedding.
    – User supports traits.
    – Event supports meta, source_event_id, optional item_id.

    Your service wrapper already exposes upsertItems, upsertUsers, and event batching, plus the mapEventTypeToCode(...) helper. Plug the contract straight into it.

Event metadata (do this!)

Add consistent meta across view/click/add/purchase so we can audit recommendation effectiveness and bandits later:

    surface, widget, recommended, request_id, rank, session_id, etc.
    Your flush route will pass meta through to Recsys if you populate metaText when inserting events.
    (Your plan already called out “mark recommended=true when coming from recommendation”. Keep that — it’s perfect. )

Important tweak based on your sample DB dump

I saw both line-item purchase events and an order-level purchase (no item_id, value≈477.37). Keep the line-item purchases as type 3; change the order-level one to type 4 (custom) with meta.kind="order" and totals. That avoids attributing a purchase to “no item” while still logging revenue. (The event model allows item_id to be omitted; using custom for order summaries is the cleanest pattern.)
Event-type weights & decay

Ship sensible ecommerce defaults (you can upsert them per-namespace):

    view: 0.05, half-life 3d

    click: 0.20, 7d

    add: 0.70, 21d

    purchase: 1.00, 60d

    custom: 0.10, 3d
    See the EventTypeConfig models to upsert; wire this into a tiny admin action if you like.

Where to put the changes (paths in your repo)

    Build item payloads with tags (brand/category/facets) before calling upsertItems(...). Use your wrapper in src/server/services/recsys.ts.

    Forward rich meta when creating events via /api/events — you already store metaText and your forwarder sends it along. Keep doing that.

    Checkout already emits per-line-item purchases; add a second custom order event if you want order totals for analytics.

Next steps (actionable)

    Adopt the contract (see canvas) for items/users/events in your API handlers.

    Normalize tags: brand:*, category:*, plus optional cat:* and simple facet tags (color:black). This unlocks brand/category caps & tag constraints at recommendation time.

    Switch order-level purchase to custom (no item_id) with order summary in meta. (You can keep the line-item purchases as is.)

    Upsert event-type config once per namespace using the recommended weights/half-lives.

    (Optional) Add embeddings to items when ready (384-d vector supported by the API); you can backfill later.
