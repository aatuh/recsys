# backlog_to_api_mapping_v1

Below is the endpoint-to-backlog mapping based on `recsys_svc_api_v1.md` and `backlog_recsys_svc_v1.md`.

## Endpoint mapping to existing backlog items

### `POST /v1/recommend`

**Maps to (core contract + behavior)**

* **SVC-10** Define API resources + versioning (endpoint exists in v1 list). 
* **SVC-11** Request validation + normalization (schema, limits, defaults, tag
  normalization; also drives `400/413/415/422`).  
* **SVC-12** Deterministic response rules (stable ordering/tie-break). 
* **SVC-13** Error model + `application/problem+json` (Problem Details).  

**Maps to (security & tenancy middleware / non-business goals)**

* **SVC-20** Authn: OIDC/JWT and/or API keys (drives `401`).  (no api keys)
* **SVC-21** Authz tenant scoping (drives `403` tenant mismatch).  
* **SVC-23** Rate limiting + quotas + payload caps (drives `429`).  
* **SVC-24** Input safety / outbound policy (if any external calls). 
* **SVC-25** Secrets management (JWT verification keys, store creds). 

**Maps to (observability / non-business goals)**

* **SVC-30** OTel tracing (traceparent spans for auth/config/store/algo).  
* **SVC-31** Metrics (RED + stage metrics used by `meta.timings_ms`).  
* **SVC-32** Structured logs with correlation (`X-Request-Id`).  

**Maps to (reliability/perf / non-business goals)**

* **SVC-42** Resilience patterns (timeouts, retries, circuit breakers; drives
  `503/504`).  
* **SVC-50** Profiling / allocation budget. 
* **SVC-51** Caching strategy (config/rules/signals). 
* **SVC-52** Backpressure (fast-fail overload, aligns with `429/503`).  

**Maps to (exposure logging / experimentation)**

* **SVC-60** Exposure event schema (request context + returned list + versions).  
* **SVC-61** Experiment assignment hook (request carries experiment id/variant).  
* **SVC-62** Privacy-aware logging. 

**Maps to (testing/docs)**

* **SVC-80..83** Contract/security/load/dependency simulation tests. 
* **SVC-90** README quickstart + sample requests/responses. 

---

### `POST /v1/similar`

**Maps to**

* Same cross-cutting items as `/v1/recommend`: **SVC-10..13**, **SVC-20..25**,
  **SVC-30..32**, **SVC-42**, **SVC-50..52**, **SVC-80..83**, **SVC-90**.  
* Endpoint is listed in v1 backlog (“optional”), so it still maps to **SVC-10**. 

---

### `POST /v1/recommend/validate`

**Maps to**

* **SVC-11** validation/normalization
* **SVC-13** Problem Details errors
* **SVC-23** rate limit/payload caps (still should be protected)
* **SVC-30..32** trace/log/metrics for debugging validation calls  

**Gap:** the backlog doesn’t explicitly mention this endpoint (it’s in the API
spec but not in SVC-10’s endpoint list).  
➡️ I’m adding a new backlog item below.

---

### `GET /healthz`

**Maps to**

* **SVC-10** endpoint list includes `/healthz`. 
* **SVC-1** bootstrap/layout (wiring the route)
* **SVC-2** containerization/runtime hardening (used for health endpoints)  

---

### `GET /readyz`

**Maps to**

* **SVC-10** endpoint list includes `/readyz`. 
* **SVC-42** dependency readiness semantics (deps reachable, etc.)
* **SVC-51** caching (optional warm cache check)  

---

### `GET /metrics`

**Maps to**

* **SVC-10** endpoint list includes `/metrics`. 
* **SVC-31** Prometheus scrape + dashboards
* **SVC-33** alerting/runbooks (metrics are inputs)  

---

### `GET /version`

**Maps to**

* **SVC-1** service bootstrap (route)
* (loosely) **SVC-90** docs (version endpoint described)

**Gap:** backlog does not explicitly include `/version`.  
➡️ I’m adding a new backlog item below.

---

## Admin/control-plane endpoint mapping

### `GET /v1/admin/tenants/{tenant_id}/config`

**Maps to**

* **SVC-70** tenant config management (feature)
* **SVC-22** function-level authorization (admin scopes)
* **SVC-21** tenant scoping enforcement (tenant_id must be authorized)  

**Gap:** backlog has “tenant config management” but not explicit GET endpoint
implementation.  
➡️ I’m adding explicit endpoint items below.

---

### `PUT /v1/admin/tenants/{tenant_id}/config`

**Maps to**

* **SVC-70** tenant config management
* **SVC-22** admin authz
* **SVC-26** audit logging for admin actions
* **SVC-51** caching (config cache invalidation)
* **SVC-13** standardized errors (`409` for version mismatch, `422` invalid)  

**Gap:** no explicit “If-Match / optimistic concurrency” item. 
➡️ Included in new items below.

---

### `GET /v1/admin/tenants/{tenant_id}/rules`

**Maps to**

* **SVC-71** rules publishing/cache invalidation (feature area)
* **SVC-22** admin authz
* **SVC-21** tenant scoping  

**Gap:** backlog doesn’t explicitly define the GET endpoint.  

---

### `PUT /v1/admin/tenants/{tenant_id}/rules`

**Maps to**

* **SVC-71** rules publishing
* **SVC-22** admin authz
* **SVC-26** audit logging
* **SVC-51** cache invalidation
* **SVC-13** standardized errors (`409/422`)  

**Gap:** backlog doesn’t explicitly define the PUT endpoint / concurrency behavior.  

---

### `POST /v1/admin/tenants/{tenant_id}/cache/invalidate`

**Maps to**

* **SVC-71** rules publishing/cache invalidation
* **SVC-51** caching strategy (invalidate targets)
* **SVC-22** admin authz
* **SVC-21** tenant scoping
* **SVC-30..32** observe invalidation requests  

**Gap:** backlog mentions invalidation conceptually but not explicit endpoint.  
