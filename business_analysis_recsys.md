# RecSys Licensing Evaluation

## Executive Summary
- **Verdict:** The current RecSys build demonstrates a functioning ingestion → ranking → audit flow but fails to meet our bar for “world-class” recommendations. Core personalization leans on popularity with no evidence of collaborative or content-based candidates. Business controls exist in theory, yet the evaluated build did not honor a manual boost. We need deeper assurances before licensing.
- **Focus Areas:** Recommendation quality, configurability, experimentation readiness, governance/tenancy.
- **Recommendation:** Pause licensing discussions until the vendor (project maintainers) supplies evidence of richer candidate generation, fixes override application, and clarifies production hardening for multi-tenant use.

## Evaluation Methodology
- Imported sample catalog, users, and events via public API (`/v1/items:upsert`, `/v1/users:upsert`, `/v1/events:batch`) with org UUID `00000000-0000-0000-0000-000000000001`.
- Queried recommendation variants (`/v1/recommendations`, `/v1/items/{id}/similar`, `/v1/bandit/recommendations`) under different blends, constraints, and overrides.
- Inserted contextual bandit policies and inspected selection behavior (`/v1/bandit/policies:upsert`, `/v1/bandit/decide`).
- Reviewed audit trail (`/v1/audit/decisions`) and configuration documentation (`README.md`, `CONFIGURATION.md`).

## Key Findings

### Recommendation Quality
- All recommendation responses—despite ample event signals—returned `model_version: popularity_v1`. Candidate-source diagnostics reported `collaborative.count = 0` and `content.count = 0`, implying missing or inactive models for co-visitation beyond anchors and for embeddings.
- Personalized requests for power users (e.g., `runner_alex`) produced lists indistinguishable from popularity-only baselines. Reasons included “personalization,” yet the items were generic, suggesting minimal incremental lift from user profiles.
- Cold-start recommendations defaulted to the same popularity ordering, indicating no fallback diversity beyond popularity.

### Configurability and Overrides
- API supports blend weights, MMR, caps, price filters, tag filters, and exclusion lists. These knobs applied successfully to remove IDs or adjust tag-constrained results.
- Created a high-priority manual boost for `watch_gps_pro` via `/v1/admin/manual_overrides`; response confirmed active status (`override_id: d8f562a4-8c1d-4031-b1f1-d791386ad5bb`). Subsequent home-surface recommendations ignored the boost, showing no change in ranking. This undermines confidence in merchandising controls.
- Segments and rules endpoints return empty payloads. Documentation does not provide schema examples or confirm evaluation order, leaving uncertainty about practical use.

### Bandit & Experimentation
- Able to upsert “baseline” and “novelty” bandit policies and observe Thompson sampling choose the novelty arm for the specified context. Audit trail captured experiment metadata.
- Despite policy toggles, ranked items still derived from the same popularity candidate pool, limiting the impact of experimentation. No evidence that differing policies change upstream candidate generation.
- Reward reporting available but untested; no aggregate KPIs or dashboards shipped by default.

### Observability & Governance
- Audit API stored full decision traces with anchors, candidate-source timing, applied caps, and bandit metadata—useful for compliance and debugging.
- System requires `X-Org-ID` header (UUID). However, authentication is disabled by default, and README/CONFIGURATION lack guidance on production auth, rate limiting, or tenancy isolation enforcement.
- LLM explain endpoint depends on OpenAI; network-restricted environments may block it, and no cost/compliance mitigations are documented.

## Risk Assessment
- **Quality Risk (High):** Lack of collaborative/content candidates means the service cannot deliver differentiated personalization, especially for mature catalogs.
- **Merchandising Risk (High):** Manual boost failure suggests business-rule execution bugs. Without reliable overrides, stakeholders cannot guarantee campaign behavior.
- **Operational Risk (Medium):** Multi-tenant story hinges on org headers but lacks auth, monitoring, or throttling guidance. Potential data-leak or abuse vectors remain open.
- **Adoption Risk (Medium):** Heavy reliance on external LLM provider and undocumented embedding ingestion pipeline complicate enterprise rollout.

## Recommendations to Vendor
1. Provide live evidence (logs, metrics, or reproducible scripts) showing collaborative filtering and embedding similarity contributing non-zero candidates in production.
2. Fix manual override application and supply integration tests or trace excerpts proving boosts/pins affect rankings for scoped surfaces.
3. Document rule DSL, segment-profile format, and precedence so integrators can adopt them confidently.
4. Clarify authentication, rate limiting, and tenancy isolation best practices for deployments beyond the demo environment.
5. Offer evaluation tooling or aggregated quality metrics to track lift when experimenting with bandit policies.

## Open Questions
- Are embeddings supported in this build? If so, what ingestion endpoint/schema activates them, and how do we verify they influence ranking?
- What safeguards ensure overrides cannot be silently ignored? Is there monitoring/alerting when a boost fails to apply?
- Does the roadmap include advanced models (e.g., matrix factorization, sequence models) or only the documented heuristics?
- How is data governance handled for audit logs that contain hashed user IDs—are there retention policies or GDPR compliance features?

## Licensing Recommendation
- **Status:** Do not proceed until evidence gaps close.
- **Conditions for reconsideration:** Demonstrated multi-model candidate generation, reliable override rule execution, documented multi-tenant security posture, and validation metrics that meet enterprise expectations.
