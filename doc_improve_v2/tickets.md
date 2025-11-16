Here’s a concrete backlog for documentation improvements.

I’ll group tickets by theme, each with:

* **ID**
* **Title**
* **Background**
* **Tasks**
* **Definition of Done**

---

## A. Onboarding & Navigation

### [×] DOCV2-101 – Add a single canonical “New here?” path to `README.md`

**Background**
Right now there are several “start here” candidates (`README.md`, `docs/overview.md`, `docs/business_overview.md`, `GETTING_STARTED.md`, `docs/quickstart_http.md`, `docs/onboarding_checklist.md`). Newcomers must choose a path before they understand anything, which is unnecessary friction.

**Tasks**

1. Edit `README.md` and add a **“New here? Start with this 4-step path”** box near the top.
2. Suggested steps (adapt if needed):

   1. **Business/value:** link to `docs/business_overview.md`
   2. **Core concepts & jargon:** link to `docs/concepts_and_metrics.md`
   3. **Hosted integration quickstart:** link to `docs/quickstart_http.md`
   4. **Deep API details:** link to `docs/api_reference.md`
3. Add one sentence per step clarifying the outcome (e.g. “Understand what RecSys does and who it’s for”).
4. Add optional link to persona-based view (`docs/overview.md`) and onboarding checklist (`docs/onboarding_checklist.md`) as *secondary* reading: “If you like structured checklists/persona views, see…”.

**Definition of Done**

* `README.md` has a clearly labeled, visually distinct “New here?” section.
* It includes exactly one recommended path for a newcomer.
* All links resolve and point to existing docs.
* No other doc claims to be “the main starting point” without pointing back to this path.

---

### [×] DOCV2-102 – Align `docs/overview.md` and `docs/onboarding_checklist.md` with the canonical path

**Background**
`docs/overview.md` and `docs/onboarding_checklist.md` currently behave like alternative starting points. They need to be reframed as helpers that *plug into* the canonical path, not compete with it.

**Tasks**

1. Update the intro of `docs/overview.md` to:

   * Explicitly say: “If you haven’t read the main onboarding path in `README.md`, start there first.”
   * Clarify this doc is a **persona map + lifecycle overview**, not the primary quickstart.
2. Update the intro of `docs/onboarding_checklist.md`:

   * Add a note that Phase 0 is “Do the main onboarding path in `README.md`”.
   * Make each persona checklist reference specific docs in the canonical path where relevant.
3. Remove any language from both docs that implies “this is the main entry point to RecSys”.

**Definition of Done**

* Both docs reference the canonical onboarding path in `README.md`.
* Neither doc claims to be the “primary” starting point.
* Persona checklists clearly point to the main four-step path where appropriate.

---

### [×] DOCV2-103 – Create “Zero-to-First-Recommendation” narrative

**Background**
Newcomers lack a story-driven walkthrough: from a fictional customer, through item/user/event ingestion, to a recommendations response they can read and understand. Current quickstarts are good but fragmented and jargon-heavy.

**Tasks**

1. Create a new doc: `docs/zero_to_first_recommendation.md`.
2. Structure it as a narrative, not a reference:

   1. **Scenario setup**

      * Introduce a fictional shop or app (e.g. “Acme Storefront”).
      * Show 1–3 sample items as JSON (titles, categories, price).
   2. **Ingest data**

      * Show example calls (cURL is fine) for:

        * `POST /v1/items:upsert`
        * `POST /v1/users:upsert`
        * `POST /v1/events:upsert` (view/add/purchase)
      * Keep payloads tiny and annotate key fields in comments.
   3. **Request recommendations**

      * Show a `POST /v1/recommendations` request using that data.
      * Include a short response and explain:

        * What `k`, `surface`, `namespace` mean in this example.
        * How to read `reasons` / trace fields, at a high level.
   4. **Connect to concepts**

      * Inline links to `docs/concepts_and_metrics.md` for any jargon.
      * Explicitly mention how guardrails / rules *would* affect this scenario (but don’t deep-dive).
3. At the end, add a “Where to go next” with links:

   * `docs/quickstart_http.md` (full hosted quickstart)
   * `docs/api_reference.md`
   * `docs/simulations_and_guardrails.md` (for safety/guardrails)
4. Link this new doc:

   * From `README.md` “New here?” path as the main story doc after business overview.
   * From `docs/quickstart_http.md` as a recommended reading for people who prefer narratives.

**Definition of Done**

* `docs/zero_to_first_recommendation.md` exists and can be followed end-to-end with dummy data.
* It uses one coherent scenario instead of fragmented examples.
* New readers can explain what the system does in their own words after reading it.
* All links resolve and the examples are syntactically valid.

---

### [×] DOCV2-104 – Clarify `GETTING_STARTED.md` vs `docs/quickstart_http.md` roles

**Background**
You have both `GETTING_STARTED.md` (repo/local) and `docs/quickstart_http.md` (hosted/API). The division is implied but not explicit, which confuses newcomers choosing between “run locally” vs “just call the hosted API”.

**Tasks**

1. Update `GETTING_STARTED.md` intro to state:

   * This is for **running RecSys locally from source**.
   * It assumes you’re a contributor or want to inspect the system internals.
   * It complements, not replaces, the hosted HTTP quickstart.
2. Update `docs/quickstart_http.md` intro to state:

   * This is for **integrating your app with the hosted RecSys API**.
   * You do *not* need to run the stack locally to follow it.
3. Add a small decision section in `README.md`:
   “Do you want to…

   * Integrate with the hosted API? → `docs/quickstart_http.md`
   * Run the system locally? → `GETTING_STARTED.md`”
4. Cross-link both docs so each references the other as an alternative path.

**Definition of Done**

* Each doc clearly states its audience and use-case.
* Confusion between “local dev” vs “hosted integration” is addressed by an explicit decision point.
* Links are in place between the two docs and from `README.md`.

---

## B. Concepts & Jargon

### [×] DOCV2-201 – Refactor `docs/concepts_and_metrics.md` into “concept cards”

**Background**
The concepts primer is useful but reads like a glossary wall: ALS, MMR, coverage, guardrails, etc. For newcomers, it’s cognitively heavy and not clearly tied to where they’ll encounter each concept.

**Tasks**

1. In `docs/concepts_and_metrics.md`, reformat each major concept as a “card” with consistent subheadings:

   * **What it is (1–2 sentences, no math)**
   * **Where you’ll see it** (endpoints, dashboards, config)
   * **Why you should care** (business/engineering stakes)
   * **Advanced details** (optional section; include math, acronyms, link to external resources if needed)
2. Add a short **“How to use this doc”** section at the top:

   * Explain this is a reference you can skim.
   * Tell readers they can ignore the “Advanced details” on first pass.
3. Ensure all acronyms (ALS, MMR, etc.) have a plain-language explanation before any technical detail.

**Definition of Done**

* Every major term has “Where you’ll see it” and “Why you should care”.
* A non-ML backend engineer can read only the top part of each card and still feel oriented.
* The number of unexplained acronyms in the first half of the doc is effectively zero.

---

### [×] DOCV2-202 – Jargon-lite pass on `docs/business_overview.md` and `docs/overview.md`

**Background**
The business-facing docs drift into jargon (ALS, MMR, etc.) and internal names too quickly. Non-technical PMs and business stakeholders should be able to understand the value story without visiting the glossary.

**Tasks**

1. Edit `docs/business_overview.md`:

   * Remove or push heavy acronyms and algorithm names into a short “Advanced concepts (optional)” section at the end.
   * Replace jargon in the main body with plain language (“we re-rank lists to balance relevance and diversity”, “we use collaborative filtering models”, etc.).
2. Edit `docs/overview.md`:

   * Make sure persona descriptions and lifecycle phases are described in business-friendly language first.
   * Any jargon in the first 2–3 sections must either be:

     * Inline explained, or
     * Linked to `docs/concepts_and_metrics.md`.
3. Ensure both docs consistently explain the system as a **recommendation control plane** and **safety-focused system**, not an “ML black box”.

**Definition of Done**

* A PM with no ML background can read `docs/business_overview.md` and explain:

  * What RecSys does.
  * Where it plugs into their product.
  * Why guardrails and simulations matter.
* All algorithm-specific jargon is either:

  * Explained in plain language; or
  * Confined to clearly labeled advanced sections.

---

## C. API Behavior, Errors, Limits & SLOs

### [×] DOCV2-301 – Centralized API error & limits model

**Background**
Error handling, limits and retry guidance are scattered across docs. Integrators need a single place to understand status codes, error bodies, rate limits, and safe retry patterns.

**Tasks**

1. Create a dedicated section in `docs/api_reference.md` or a new doc `docs/api_errors_and_limits.md`.
2. Document:

   * Standard status codes you use (200, 400, 401, 403, 404, 409, 429, 5xx, etc.).
   * Example error response body schema (fields like `code`, `message`, `details`, `trace_id`).
   * Which errors are **transient** (safe to retry with backoff) vs **permanent**.
   * Rate limiting behavior (e.g. 429 semantics, headers if any).
   * Hard limits:

     * Max items per batch.
     * Max body size (if applicable).
     * Practical max `k` and any related performance caveats.
3. Link this doc from:

   * `docs/quickstart_http.md` under a “Errors & limits” subsection.
   * `docs/api_reference.md` overview.

**Definition of Done**

* There is exactly one authoritative place that describes error codes, limits and retry guidance.
* Quickstart and reference docs link to it instead of re-describing errors.
* Integrators can implement robust retry/backoff and input validation from this info alone.

---

### [×] DOCV2-302 – Document idempotency, consistency and pagination behavior

**Background**
Senior integrators will want to know how `*:upsert` behaves with duplicates, how fast writes propagate to recommendations, and how `GET` endpoints paginate. Right now they’d have to guess or ask.

**Tasks**

1. In `docs/api_reference.md`, add a subsection “Behavioral guarantees” covering:

   * **Idempotency**:

     * Are all `*:upsert` endpoints idempotent?
     * What happens on exact duplicate event ingestion?
     * Any deduplication windows or semantics.
   * **Consistency / freshness**:

     * Typical lag between writes (items, users, events) and their effect on `POST /v1/recommendations`.
     * Separate behavior for near-real-time vs batch-updated components if applicable.
   * **Pagination**:

     * For `GET /v1/items`, `GET /v1/users`, etc:

       * Pagination params (cursor vs offset).
       * Stability guarantees (e.g. “results are sorted by created_at desc; repeated calls with the same cursor yield the same items”).
2. Add a short “How to design clients around this” bullet list with practical advice (e.g. “don’t assume immediate visibility of ingested events”).

**Definition of Done**

* Idempotency behavior is clearly documented for all relevant endpoints.
* Consistency and lag expectations are described in human language, not just “eventual”.
* Pagination semantics are unambiguous and match actual behavior.

---

### [×] DOCV2-303 – Add minimal security & data handling section

**Background**
You do mention auth and multi-tenant headers, but security/compliance questions (TLS, PII handling, data retention, deletion) will come up early. There should be a single, high-level answer in the docs.

**Tasks**

1. Add a new doc `docs/security_and_data_handling.md` or a section in `docs/business_overview.md` + `docs/api_reference.md`.
2. Cover, at a minimum:

   * **Transport security** (HTTPS/TLS assumptions).
   * **Authentication & authorization model** (API keys, org headers, tenant isolation at a high level).
   * **PII guidance** (what you expect/allow in user/item fields, recommended anonymization/pseudonymization).
   * **Data retention & deletion semantics** (high-level; how to delete users/items/events).
3. Link this doc from:

   * `README.md` (small “Security & data handling” link).
   * `docs/business_overview.md` (for stakeholders).
   * `docs/api_reference.md` (for engineers).

**Definition of Done**

* A security-conscious engineer or PM can get a reasonable first answer from the docs without a meeting.
* There is one clearly linked doc/section covering security and data handling basics.
* Claims in the doc are verified against actual system behavior/policies.

---

## D. Business / PM-Facing Docs

### [×] DOCV2-401 – Add a “PM / Product FAQ” to `docs/business_overview.md`

**Background**
PMs and other non-technical stakeholders need answers to recurring questions: ownership, rollout, risk, success metrics. Right now they’d need to piece this together from multiple docs.

**Tasks**

1. In `docs/business_overview.md`, add a new section “Product & PM FAQ”.
2. Include concise Q&A entries, for example:

   * “How do we measure success of RecSys?” (with common KPIs).
   * “Who owns rules vs algorithm knobs?” (clarify ownership boundaries).
   * “How long before we see impact after launch?” (ballpark, clearly flagged as approximate).
   * “What happens if recommendations look wrong?” (who to contact, guardrails/simulations role).
   * “How do we roll out safely?” (tie to simulations, bandit experiments, A/B).
3. Link relevant answers to deeper docs:

   * Simulations/guardrails → `docs/simulations_and_guardrails.md`
   * Bandits/experiments → analytics/bandit docs.
   * Rules changes → `docs/rules_runbook.md`.

**Definition of Done**

* `docs/business_overview.md` has a PM FAQ section with 5–10 focused questions.
* Each answer is <5 sentences and links to deeper docs where warranted.
* A PM can read this section and feel they know “who does what” and “how risky is this”.

---

### [×] DOCV2-402 – Add a before/after case study or scenario

**Background**
You talk about value (revenue, CTR, guardrails, etc.) but don’t show a simple “before vs after RecSys” example. Storytelling makes the product real for business people.

**Tasks**

1. Either:

   * Add a new section in `docs/business_overview.md` (“Example rollout at Acme Retail”), or
   * Create `docs/example_rollout_case_study.md`.
2. Describe a simple, realistic scenario:

   * Initial state (static carousels, manual merchandising, pain points).
   * Integration steps (surfaces adopted, rules/guardrails put in place, experiments run).
   * Outcomes (even if hypothetical ranges, clearly labeled as such).
3. Keep it narrative, but short (1–2 screens of reading).
4. Link it from:

   * `docs/business_overview.md` near the top.
   * `docs/overview.md` under lifecycle.

**Definition of Done**

* There is at least one end-to-end narrative of a customer adopting RecSys.
* It clearly articulates starting pain, rollout steps, and resulting benefits.
* It avoids technical jargon unless explicitly marked as an aside.

---

## E. Examples & Tooling

### [×] DOCV2-501 – Add minimal client code samples (Python + one other language)

**Background**
You currently rely on cURL and Swagger. Many engineers want to copy–paste working snippets in a real language (Python, Node, etc.), especially for basic ingestion and recommendation calls.

**Tasks**

1. Create `docs/client_examples.md` or add a “Client examples” section in `docs/quickstart_http.md`.
2. For each chosen language (Python + 1 other that matches your user base, e.g. Node.js):

   * Provide a minimal, complete example that:

     * Sets up auth headers.
     * Calls `POST /v1/items:upsert` with a single item.
     * Calls `POST /v1/recommendations` for a user.
     * Prints the results.
3. Add comments explaining the important fields, but don’t drown the code in commentary.
4. Make sure all examples are tested against a dev/hosted environment and actually work.

**Definition of Done**

* There are at least two language examples available, each end-to-end.
* `docs/quickstart_http.md` links to the client examples.
* Engineers can integrate by copy–pasting and adjusting these snippets.

---

### [×] DOCV2-502 – Add doc CI: link checker + example tests

**Background**
As the docs expand (more cross-links, more code samples), the risk of broken links and stale snippets goes up. You want basic guardrails in CI.

**Tasks**

1. Add a simple script (Python/Go/whatever fits repo) that:

   * Scans `.md` files.
   * Verifies internal markdown links are valid (files exist, headings present).
2. Add a small test harness for examples:

   * At minimum, run the main cURL or code examples in a dry-run mode or unit-test wrapper (where feasible) to ensure they are syntactically correct and reference valid endpoints.
3. Wire these checks into your existing CI pipeline:

   * New PRs that modify docs must pass link and example checks.

**Definition of Done**

* CI fails if a markdown link points to a non-existent file or heading.
* CI fails if core examples are broken (syntax, path errors).
* There is a documented “How we test docs” note in `README.md` or a contributing doc.

---

## F. Analytics & Ops Docs

### [×] DOCV2-601 – Add an analytics docs index & tie to product questions

**Background**
You have several analytics docs under `analytics/` (bandit, dashboards, backfill plans, etc.), but there’s no single navigation or mapping to product questions (“How do we measure X?”, “Where do I look for Y?”).

**Tasks**

1. Create `analytics/README.md` or `docs/analytics_overview.md`.
2. For each existing analytics doc (bandit, dashboards, pipelines, etc.):

   * Add a short entry:

     * **What question it answers** (e.g. “How is bandit exploration configured?”, “How do I validate diversity?”).
     * **Who should read it** (Analyst, PM, Data Scientist, etc.).
     * **When to use it** (before launch, during rollout, ongoing optimization).
3. Link this overview:

   * From `docs/business_overview.md` in the PM/FAQ or metrics sections.
   * From `docs/onboarding_checklist.md` for relevant personas (analytics, data).

**Definition of Done**

* There is a single index doc that lists and explains all analytics docs.
* Each analytics doc has a one-line summary and target audience.
* Onboarding checklists refer to this index rather than raw scattered files.

---

## G. Quality & Polish

### [×] DOCV2-701 – Run a quality pass to remove truncated sentences / placeholder ellipses

**Background**
The current snapshot shows a lot of `...` mid-sentence. If this reflects the real docs (not just redaction here), it severely hurts readability and looks unprofessional.

**Tasks**

1. Search the repo for `...` in `.md` files.
2. For each occurrence:

   * Decide whether it is:

     * Intentional stylistic usage (e.g. speech-like).
     * Placeholder / truncation that should be expanded or removed.
   * Fix or rewrite incomplete sentences.
3. While touching those sections, quickly scan for obvious grammar/typo issues and fix them.

**Definition of Done**

* All docs read as complete, polished prose (no accidental-truncation ellipses).
* There are no obviously incomplete sentences left in user-facing docs.
* Optional: run a spellchecker or grammar check across docs once.

---

### [×] DOCV2-702 – Create a short doc style & terminology guide

**Background**
You already do some things well (persona labels, “Who is this for?”). As docs grow, consistency will decay without a small style guide (terminology, capitalization, tone).

**Tasks**

1. Add `docs/doc_style_guide.md` with:

   * Preferred product name capitalization (e.g. RecSys vs Recsys).
   * Rules for:

     * Including “Who should read this / Not for”.
     * Tone (direct, concise, low fluff).
     * Acronym usage (must be expanded on first usage; link to concepts doc).
   * Linking conventions:

     * Link to relative paths; avoid bare URLs.
     * Use explicit text instead of “here”.
2. Reference this style guide in:

   * `README.md` (small “Documentation style guide” note).
   * Any contributing docs.

**Definition of Done**

* Style guide exists and covers naming, tone, and basic structure conventions.
* New docs and edits follow this guide (at least for a few trial PRs).
* Future contributors have a clear reference for how to write docs.

---

If you want, next step I can do is pick one of these (e.g. DOC-103 Zero-to-First-Recommendation) and draft the actual doc content so you can drop it in with minimal editing.
