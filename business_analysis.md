# RecSys Shop Evaluation — Business & Technical Assessment

## 1. Engagement Setup
- **Scope.** Evaluate the RecSys-powered `/shop` surface as a prospective licensee. Only the web-shop API (`/shop`) may be used; no direct calls to the core RecSys API.
- **Environment.** Local docker-compose stack with empty databases at start. Primary tools: `curl` for API exploration plus read-only database access.
- **Success criteria.** Determine whether the solution can meet “world-class” quality expectations: configurable ranking, robust data tooling, advanced algorithms (embeddings, bandits, personalization), and decision analytics suitable for enterprise rollout.

## 2. Evaluation Activities
1. **Seeded baseline data.** Invoked `/api/seed`, `/api/admin/sync-items`, `/api/admin/sync-users`, `/api/events/seed`, and `/api/events/flush` to populate the demo catalog, users, and events.
2. **Behavioral probes.** Exercised `/api/recommendations` with constraints (`minPrice`, `maxPrice`, `includeTags`, `excludeTags`, `brandCap`, `categoryCap`, `surface`) to validate configurability.
3. **Feature inspection.** Reviewed server implementations for diversity, cold-start insertion, algorithm profiles, bandit integration, analytics, seeders, and data contracts.
4. **Custom persona test.** Created a synthetic persona via `/api/users`; re-synced to observe personalization impact.
5. **Diagnostics.** Pulled analytics via `/api/admin/analytics` and reviewed container logs.

## 3. Key Findings

### 3.1 Configurability & Controls
- Request-level filters and caps operate as advertised. Calls such as `GET /api/recommendations?includeTags=shoes&brandCap=1` enforced constraints and logged applied caps (`shop/app/api/recommendations/route.ts:120-220`).
- Algorithm profiles are mutable at runtime. `POST /api/admin/recommendation-settings` updated blend weights and surface defaults; subsequent recommendations reflected the new `profile_id`.
- Diversity enforcement consults live product metadata to cap brands/categories, returning audit data in the `filters` block (`shop/src/server/services/diversity.ts:17-85`).

### 3.2 Gaps Limiting “World-Class” Claims
- **Seeder failures.** `POST /api/users/seed` and `POST /api/products/seed` crash with `PrismaClientValidationError: Unknown argument 'skipDuplicates'` because sqlite does not support that option (`shop/app/api/users/seed/route.ts:224-227`, `shop/app/api/products/seed/route.ts:225-228`). This blocks generation of the richer catalogs/users described in README.
- **Bandit disabled.** Despite `SHOP_BANDIT_ENABLED=true`, `getBanditFeatureStatus` signals “Missing policies: manual_explore_default” because the namespace has no policies (`shop/src/server/services/recsys.ts:533-571`). The shop logs warnings and always falls back to static ranking.
- **Cold-start fillers dominate.** The route always appends `SHOP_FRESH_ITEM_SLOTS` zero-score items following diversity filtering, even when k is already satisfied (`shop/app/api/recommendations/route.ts:225-273`). Live calls routinely returned 30% cold-start filler, hurting list quality.
- **Analytics unusable.** `GET /api/admin/analytics` mislabels all events as `unknown` and reports `CTR=0`. Root cause: the code indexes a string-based event type into a numeric array (`shop/app/api/admin/analytics/route.ts:14-48`), so funnel metrics are invalid.
- **Data richness lacking.** Default seeding (`shop/app/api/seed/route.ts:22-37`) provides only five brands, coarse categories, no attributes, and no embeddings. `buildItemContract` never supplies embeddings (`shop/src/lib/contracts/item.ts:140-183`), so similarity relies purely on popularity/co-visitation.
- **Missing API parity.** Although the server-side Next page calls `getSimilar` (`shop/app/products/[id]/page.ts`), the `/shop` API omits equivalent endpoints, limiting third-party integration during evaluation.

### 3.3 Positive Signals
- Seen-item filtering works: after logging a view via `/api/events`, the next recommendation call excluded that item and documented the exclusion (`shop/app/api/recommendations/route.ts:153-183`).
- Tooling exists to sync deletions/updates to RecSys (`shop/src/server/services/itemSync.ts`, `shop/app/api/admin/tools/route.ts`), indicating awareness of catalog hygiene needs.
- Cold-start insertions emit structured metrics through `logColdStart` (`shop/src/server/logging/coldStart.ts`), providing observability hooks once thresholds are tuned.

## 4. Business Impact Assessment
- **Quality risk.** With broken seeders and no embeddings, the demo cannot exhibit semantic similarity or deep personalization—the core differentiators touted in the README. This undermines confidence in premium recommendation quality.
- **Experimentation credibility.** Bandit integration is effectively non-functional today. Enterprise buyers expect live experimentation; the current warning state signals production risk.
- **Analytics & ROI.** The analytics API’s incorrect output prevents measuring funnel lift, making it difficult to build a business case for licensing.
- **Cold-start experience.** Guaranteed low-score filler items are a red flag for UX metrics (CTR, revenue per impression). Without smarter gating they would drag performance below industry benchmarks.

## 5. Recommended Remediation Steps
1. **Fix Prisma seeders.** Remove `skipDuplicates` or switch sqlite to Postgres so `/api/users/seed` and `/api/products/seed` work, then reseed with the richer templates (attributes, embeddings) promised in documentation.
2. **Provide embeddings & traits.** Extend item contracts to include 384-d vectors and enhance user traits/segments so similarity and personalization features can be demonstrated.
3. **Resolve bandit configuration.** Provision the `manual_explore_default` policy in the namespace (or disable the flag) and expose API evidence for policy selection and rewards.
4. **Tame cold-start injections.** Only backfill when ranked items < k, or expose per-surface toggles for `SHOP_FRESH_ITEM_SLOTS` to prevent filler dominance.
5. **Repair analytics pipeline.** Correct event type mapping in `/api/admin/analytics` and surface CTR/ATC/conversion metrics that reflect actual behavior.
6. **Expand `/shop` API coverage.** Add endpoints mirroring `getSimilar` and decision traces so partner teams can integrate without touching the internal RecSys API.
7. **Document real evaluation workflow.** Provide a playbook (events to insert, sync scripts, validation reports) so buyers can replicate a realistic setup swiftly.

## 6. Conclusion
The architecture shows promising configurability and feature hooks, but critical execution gaps—broken data onboarding, disabled experimentation, inaccurate analytics, and cold-start noise—prevent us from validating “world-class” recommendation quality today. Addressing the remediation items above is essential before entering licensing negotiations or wider trials.
