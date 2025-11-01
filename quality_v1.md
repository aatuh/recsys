# API Service Quality Backlog v1

- [x] **Ticket:** Introduce application composition root with dependency injection  
  **Category:** Architecture / SOLID  
  **Problem:** `cmd/api/main.go:42-259` wires logging, config, DB, migrations, workers, and HTTP handlers inline, leading to a monolithic `main` that is hard to test and extend. Panics on configuration errors (`cmd/api/main.go:49-55`) further couple bootstrap to environment specifics.  
  **Proposal:** Create an `internal/app` package (or adopt a lightweight DI tool like `fx`/`wire`) that encapsulates construction of logger, config providers, DB pool, rule manager, explain service, and HTTP router. Return structured errors instead of panics, and let `main` focus on lifecycle management.  
  **Impact:** Clear separation of composition vs. runtime logic, improved testability (components can be instantiated independently), and easier onboarding for new dependencies.

- [x] **Ticket:** Replace panic-driven env loading with typed config provider and defaults  
  **Category:** Robustness / DX  
  **Problem:** `internal/http/config/config.go:17-209` and `cmd/api/main.go:265-268` rely on `util.MustGetEnv`, causing panics when variables are missing and forcing exhaustive env setup for tests. Validation errors are returned one-at-a-time and defaults are not centrally documented.  
  **Proposal:** Introduce a typed config loader (e.g., `envconfig` or custom decoder) that supports defaults, required markers, aggregation of validation errors, and pluggable sources (env, file, flags). Emit structured diagnostics and allow injecting config during tests without environment churn.  
  **Impact:** More resilient startup, faster developer feedback, and simpler configuration management in different environments.

- [x] **Ticket:** Split god `Handler` into focused HTTP adapters backed by interface-driven services  
  **Category:** Architecture / SOLID  
  **Problem:** `internal/http/handlers/ingest.go:24-52` stores dozens of fields (store, rules, explain, tuning knobs) and every endpoint manipulates core logic directly (`recommend.go:67-108`). This violates single responsibility and makes unit testing handlers difficult.  
  **Proposal:** Introduce domain services (`IngestionService`, `RecommendationService`, `AdminService`, etc.) defined via interfaces in `internal/types` and implemented in `internal/services`. Keep HTTP handlers thin: decode → validate → call service → encode. Provide dependency injection through the composition root.  
  **Impact:** Better separation of concerns, easier mocking in tests, and clearer change boundaries when evolving business logic.

- [x] **Ticket:** Enforce strict tenant isolation instead of silent fallback to default org  
  **Category:** Security / Correctness  
  **Problem:** `internal/http/handlers/eventtypes.go:82-88` and similar call sites silently fall back to `DefaultOrg` if `X-Org-ID` is missing/invalid, risking cross-tenant data leakage.  
  **Proposal:** Require explicit tenant context—reject requests lacking a valid org ID (or authenticated tenant mapping). Centralize this in middleware that validates headers or auth tokens, and propagate a typed `TenantContext`.  
  **Impact:** Prevents accidental data mix-ups, satisfies multi-tenant isolation expectations, and clarifies contract for clients.

- [x] **Ticket:** Add authentication, authorization, and rate limiting guardrails  
  **Category:** Security / Reliability  
  **Problem:** The router (`cmd/api/main.go:87-235`) exposes all endpoints without auth, throttling, or abuse protection. Admin routes (rules, segments, bandit) can be invoked anonymously.  
  **Proposal:** Introduce API key or JWT authentication middleware, enforce per-tenant scopes, and add rate limiting (e.g., token bucket keyed by org ID + IP). Provide audit logging for privileged operations.  
  **Impact:** Stronger security posture, compliance readiness, and protection against replay or brute-force ingestion.

- [x] **Ticket:** Harden request handling with size limits, strict decoders, and validation  
  **Category:** Robustness / Security  
  **Problem:** Handlers such as `ItemsUpsert` (`internal/http/handlers/ingest.go:65-112`) and `Recommend` (`internal/http/handlers/recommend.go:45-99`) stream JSON directly into structs without `MaxBytesReader`, `Decode.DisallowUnknownFields`, or schema validation. Malicious or malformed payloads can exhaust memory or silently drop fields.  
  **Proposal:** Wrap handlers with `http.MaxBytesReader`, enable `DisallowUnknownFields`, and adopt a validation library (e.g., `go-playground/validator`) or generated schema checks. Return consistent 4xx errors for validation failures.  
  **Impact:** Reduced DoS surface, clearer API contracts, and less time spent chasing production data quality issues.

- [x] **Ticket:** Instrument API with structured metrics and distributed tracing  
  **Category:** Observability  
  **Problem:** Beyond error tallies (`cmd/api/main.go:94-100`), there is no request/latency instrumentation, trace context propagation, or standardized logging correlation.  
  **Proposal:** Adopt OpenTelemetry for HTTP/server/db spans, expose Prometheus metrics (request counts, latencies, DB usage), and enrich zap logs with trace/span IDs. Provide dashboards and alerts for key SLOs.  
  **Impact:** Faster incident response, measurable performance baselines, and easier capacity planning.

- [x] **Ticket:** Make DB pool sizing, retry, and observability configurable  
  **Category:** Reliability / Performance  
  **Problem:** `internal/http/db/pool.go:10-27` hard-codes connection limits and lacks query instrumentation or retry semantics. Long-running queries inherit request contexts without deadlines (`internal/store/*.go`).  
  **Proposal:** Move pool sizing and timeouts into config, wrap store calls with context deadlines, add query logging/metrics, and introduce safe retry policies for transient errors. Consider using PGX `BeforeAcquire` hooks for connection health.  
  **Impact:** Better database resilience under load, clearer capacity tuning knobs, and reduced risk of hung requests.

- [x] **Ticket:** Propagate background worker failures to health checks and shutdown logic  
  **Category:** Reliability  
  **Problem:** The decision recorder (`cmd/api/main.go:150-164`) runs in a goroutine; failures only log warnings on shutdown. If the writer encounters persistent errors, the API keeps serving without signaling degraded state.  
  **Proposal:** Expose a health component for background workers, emit metrics, and trigger graceful degradation (e.g., disable recording with alerts). Include worker lifecycle management in the composition root.  
  **Impact:** Detect silent failures sooner and avoid data loss in audit pipelines.

- [x] **Ticket:** Abstract LLM explain clients behind provider interfaces with circuit breakers  
  **Category:** Robustness / SOLID  
  **Problem:** `cmd/api/main.go:122-149` instantiates OpenAI client inline, with minimal timeout handling and no fallbacks beyond a null client. Provider choice logic is tied to env strings.  
  **Proposal:** Define an `ExplainClient` interface with provider-specific implementations registered via config. Add circuit breaker/timeout defaults, response caching, and structured error mapping. Support dependency injection for testing/mocking.  
  **Impact:** Safer integration with external LLMs, easier to add new providers, and reduced risk of cascading latency.

- [x] **Ticket:** Strengthen CORS configuration safety  
  **Category:** Security  
  **Problem:** `internal/http/middleware/cors.go:19-78` treats missing `CORS_ALLOWED_ORIGINS` as `allowAll`, effectively enabling any origin.  
  **Proposal:** Require explicit allowlists, fail closed by default, and support environment-specific presets. Add logging when wildcards are used and provide tests covering regex conversion.  
  **Impact:** Prevents unintended cross-origin access and aligns with least-privilege defaults.

- [x] **Ticket:** Expand automated coverage with API integration and contract tests  
  **Category:** Quality / DX  
  **Problem:** Existing tests focus on lower-level packages; there are limited end-to-end checks that ingest data and assert recommendation responses or admin workflows through HTTP (`api/test` covers selected paths only).  
  **Proposal:** Add `httptest`-backed suites (or dockerized integration tests) that exercise ingestion → recommendation → audit flows, validate Swagger schema compliance, and guard against regression in pagination/filters. Automate them via `make test`.  
  **Impact:** Higher confidence in refactors, better regression detection, and living documentation of API behavior.

- [x] **Ticket:** Provide developer-friendly local profiles and feature flags  
  **Category:** DX / Maintainability  
  **Problem:** Developers must mirror production-like env vars to run the API; toggles such as rule engine or decision tracing require manual env editing.  
  **Proposal:** Introduce profile-based config (`dev`, `test`, `prod`) with sane defaults, CLI to generate `.env` templates, and feature flags exposed via config structures. Document workflows in `README.md`.  
  **Impact:** Faster onboarding, fewer "panic: missing env" issues, and safer experimentation with optional subsystems.
