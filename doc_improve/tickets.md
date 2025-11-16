Here’s a structured backlog you can drop straight into Jira/Linear/etc.

I’ll group tickets by theme and give each one:

* **ID**
* **Title**
* **Background (plain language)**
* **Tasks (step-by-step)**
* **Definition of Done (checklist)**

I’ll assume current files:

* Root: `README.md`, `CONFIGURATION.md`
* `/docs`: `overview.md`, `api_reference.md`, `bespoke_simulations.md`, `database_schema.md`, `env_reference.md`, `rules-runbook.md`

---

## A. Core structure & new docs

### [x] DOC-001 – Create `GETTING_STARTED.md` (repo quickstart)

**Background**
Right now, there is no single “Hello World” doc that takes a newcomer from “cloned the repo” to “I got my first recommendations”. Pieces are spread across README and scripts.

**Tasks**

1. In repo root, create a new file: `GETTING_STARTED.md`.
2. Add sections:

   1. **Prerequisites**

      * List tools needed (Docker, Python version, `make` or equivalent).
   2. **Start the stack**

      * Add a simple command example (e.g. `make up` or `docker compose up`) based on how the project currently runs.
   3. **Seed sample data**

      * Find how sample data is currently seeded (there are scripts under `analysis/scripts/...`).
      * Add one recommended command path (just one, not all variants), with a short explanation of what it does.
   4. **Call the API for recommendations**

      * Add a working HTTP example (curl or similar) for `/v1/recommendations` against localhost.
      * Explain the important request fields in comments (e.g. `namespace`, `user_id`, `k`).
      * Show a truncated JSON response with a brief explanation of what the key fields mean.
   5. **Next steps**

      * Link to:

        * `docs/quickstart_http.md` (will be created in another ticket)
        * `docs/api_reference.md` (renamed `api_reference.md`)
        * `docs/concepts_and_metrics.md` (new)

**Definition of Done**

* `GETTING_STARTED.md` exists in the repo root.
* A new engineer can follow it step-by-step and:

  * Start the system locally.
  * Seed some data.
  * Call the recommendations endpoint and see non-empty results.
* All commands have been test-run once and are correct.
* Links to other docs resolve (no broken links).

---

### [x] DOC-002 – Add `docs/quickstart_http.md` (HTTP-only integration guide)

**Background**
External integrators may not use the repo or scripts. They need an HTTP-only guide showing how to ingest data and request recommendations using the public API.

**Tasks**

1. Create `docs/quickstart_http.md`.
2. Add sections:

   1. **Base URL, auth, and namespaces**

      * Explain where the API is hosted (configurable/base URL).
      * Explain how auth works (e.g. header name, token).
      * Explain what a `namespace` is in plain language.
   2. **Ingest minimal data**

      * Provide **copy-paste** HTTP examples (curl-style) for:

        * `POST /v1/items:upsert`
        * `POST /v1/users:upsert` (if used)
        * `POST /v1/events:batch` (if used)
      * Show minimal required fields only.
   3. **Fetch recommendations and similar items**

      * Provide example for:

        * `POST /v1/recommendations`
        * `GET /v1/items/{item_id}/similar` (if available)
      * Briefly explain the important request and response fields.
   4. **Common mistakes & error messages**

      * From the existing docs and code, identify the most common integration mistakes:

        * Missing/incorrect `namespace`
        * Missing/incorrect org or auth headers
        * Invalid JSON shape
        * Unsupported blend name, etc.
      * Add a short “Problem → Likely cause → What to check” table.
   5. **Next steps**

      * Link to `docs/api_reference.md` for full details.

**Definition of Done**

* `docs/quickstart_http.md` exists and contains:

  * A clear explanation of base URL/auth/namespaces.
  * At least one working example each for ingesting items and requesting recommendations.
  * A “Common mistakes” section with at least 3–5 items.
* Examples are syntactically valid (e.g. curl commands would run with appropriate env values filled).
* Links to other docs resolve.

---

### [x] DOC-003 – Create `docs/concepts_and_metrics.md` (concepts + glossary + metrics primer)

**Background**
Currently, definitions of key terms (ALS, MMR, coverage, guardrails, etc.) and metrics (NDCG, MRR, lift, long-tail share) are scattered across multiple files. Newcomers struggle with the jargon.

**Tasks**

1. Create `docs/concepts_and_metrics.md`.
2. Collect all existing short explanations and glossaries from:

   * `README.md` (glossary at the bottom).
   * `CONFIGURATION.md`.
   * `docs/env_reference.md`.
   * `docs/simulations_and_guardrails.md`.
   * Any other references to NDCG/MRR/coverage/guardrails.
3. Organize the content into four sections:

   1. **Core concepts**

      * Candidate, retrieval vs ranking, popularity, co-visitation, embeddings, personalization, rules/overrides, bandits.
   2. **Metrics primer**

      * NDCG, MRR, “segment lift”, “catalog coverage”, “long-tail share” – each in plain language.
      * Use small, simple examples (e.g. recommending 5 items and assessing which are relevant).
   3. **Guardrails – concept only**

      * Explain what guardrails do: protecting user experience from bad recommendations, ensuring coverage, etc.
      * Mention that implementation details live in `docs/simulations_and_guardrails.md`.
   4. **Glossary**

      * Move the existing glossary from `README.md` here, and expand if needed.
4. Once this file is created, update other docs to:

   * Remove duplicate concept explanations where possible.
   * Add links to this new file where jargon appears (e.g. “See docs/concepts_and_metrics.md for definitions”).

**Definition of Done**

* `docs/concepts_and_metrics.md` exists and:

  * Contains all previously defined glossary items.
  * Contains at least one short paragraph for each major concept and metric used in the docs.
* Other docs are updated to link to this file instead of re-defining the same concepts.
* No references to the old glossary location remain in README.

---

### [x] DOC-004 – Create `docs/business_overview.md` (non-technical product overview)

**Background**
There’s no dedicated “business-facing” overview. PMs and non-technical stakeholders see a lot of tuning scripts and CI details instead of a simple value narrative.

**Tasks**

1. Create `docs/business_overview.md`.
2. Add sections:

   1. **What this system does (in business language)**

      * Explain that it’s a domain-agnostic recommendation engine, capable of powering “similar items”, personalized feeds, search reranking, etc.
   2. **Example use cases**

      * 3–5 concrete examples: e.g. e-commerce product recommendations, content feeds, etc.
   3. **Rollout story (timeline)**

      * What an organization can expect in:

        * Week 1 (initial integration and basic recs)
        * Week 4 (better tuning, initial guardrails, basic metrics)
        * Week 8+ (experimentation, bandits, advanced guardrails)
   4. **Safety and guardrails (business view)**

      * Explain why guardrails are important: avoiding dead ends, spammy content, unfair exposure.
      * Mention simulations as a way to test impact before deployment.
   5. **Evidence & auditability**

      * Summarize in plain language:

        * Decision traces/explanations.
        * Metrics dashboards/Prometheus.
        * Simulation reports and guardrail checks.
   6. **Where to dig deeper**

      * Link to:

        * `docs/overview.md` (personas & lifecycle)
        * `docs/simulations_and_guardrails.md`
        * `docs/concepts_and_metrics.md`
        * `docs/rules_runbook.md`
3. If the writer is not a subject matter expert, they should:

   * Draft this doc based on existing content.
   * Request review from someone familiar with the product to confirm correctness.

**Definition of Done**

* `docs/business_overview.md` exists, written in non-technical language.
* It can be read in under 10 minutes and gives a coherent picture of:

  * What the system does,
  * How it’s rolled out,
  * How safety/guardrails are handled.
* Links to other docs work and there are no “TODO” placeholders left.

---

## B. Renames, canonical references, and README cleanup

### [x] DOC-005 – Rename `docs/api_reference.md` → `docs/api_reference.md` and update links

**Background**
`api_reference.md` is effectively a reference for all endpoints. The new name `api_reference.md` is clearer.

**Tasks**

1. Rename file:

   * `docs/api_reference.md` → `docs/api_reference.md`.
2. Search the repo for references to `api_reference.md` and update them to `api_reference.md`:

   * In `README.md`
   * In `docs/overview.md`
   * In any other docs/scripts.
3. Update the top-level heading inside the file to something like “API Reference”.
4. Confirm that no file still refers to the old name.

**Definition of Done**

* Only `docs/api_reference.md` exists (no `api_reference.md` remains).
* All internal links have been updated and checked.
* The doc’s own heading matches the new name.

---

### [x] DOC-006 – Rename `docs/env_reference.md` → `docs/env_reference.md` and make it canonical

**Background**
Environment variable documentation is currently split between several places. We want a single canonical reference file.

**Tasks**

1. Rename file:

   * `docs/env_reference.md` → `docs/env_reference.md`.
2. Search for references to `env_reference.md` and update them to `env_reference.md`.
3. In `docs/env_reference.md`, ensure:

   * It clearly states at the top that this is the **canonical list** of env vars and configuration flags.
   * Sections are organized by theme (e.g. ingestion, diversity, personalization, rules, bandits, service).
4. In other docs (`README.md`, `CONFIGURATION.md`, etc.):

   * Remove detailed env var tables or lists where they duplicate this reference.
   * Replace them with short summaries and a link to `docs/env_reference.md`.

**Definition of Done**

* Only `docs/env_reference.md` exists (no `env_reference.md`).
* That file is clearly marked as canonical for env vars.
* No other doc contains large copied tables of env vars; they instead link to `docs/env_reference.md`.

---

### [x] DOC-007 – Simplify `README.md` into a front door + quickstart + persona map

**Background**
`README.md` currently mixes intro, tuning workflows, guardrails details, and glossaries. It overwhelms newcomers.

**Tasks**

1. Edit `README.md` to follow this structure:

   1. **Short product description** (3–5 bullets).
   2. **What you can do with it** (examples).
   3. **Quickstart (very high-level)**

      * Briefly mention “clone → run → seed → get first recommendations”.
      * Link to `GETTING_STARTED.md` for details.
   4. **Where to go next (persona map)**

      * Business/Product → `docs/business_overview.md`, `docs/overview.md`, `docs/rules_runbook.md`.
      * Integration Engineers → `GETTING_STARTED.md`, `docs/quickstart_http.md`, `docs/api_reference.md`, `docs/env_reference.md`, `docs/database_schema.md`.
      * Dev/Ops → `docs/overview.md`, `docs/tuning_playbook.md`, `docs/simulations_and_guardrails.md`, `docs/rules_runbook.md`.
   5. **Advanced topics**

      * Brief descriptions linking to tuning, simulations, etc.
2. Remove or move detailed sections from README:

   * Move tuning workflow and AI optimizer details to `docs/tuning_playbook.md`.
   * Move detailed onboarding/coverage checklist + guardrail scenario details to `docs/simulations_and_guardrails.md`.
   * Move glossary content to `docs/concepts_and_metrics.md`.
3. Ensure README is no more than ~2–3 screens of text.

**Definition of Done**

* `README.md` is short, high-level, and mainly links out to other docs.
* Detailed tuning commands, guardrail config details, and glossary are no longer in README.
* Persona-based “where to go next” table exists and references the new doc structure.

---

### [x] DOC-008 – Make `CONFIGURATION.md` a conceptual configuration & data guide (not a full reference)

**Background**
`CONFIGURATION.md` currently mixes conceptual explanations with detailed environment/reference information, which should live in `docs/env_reference.md` and `docs/api_reference.md`.

**Tasks**

1. Review `CONFIGURATION.md` and:

   * Identify purely conceptual sections (mental model, data flow, what knobs do).
   * Identify detailed env var listings or technical reference bits.
2. Keep and polish the conceptual sections:

   * Mental model of the recommendation flow (e.g. ingestion → signals → blending → personalization → rules/guardrails).
   * High-level explanation of the main configuration “themes” (diversity, personalization, bandits, etc.).
3. For detailed references:

   * Remove large env var tables.
   * Add links to:

     * `docs/env_reference.md` (for env vars).
     * `docs/api_reference.md` (for input/output schemas).
4. Add a short section at the top:

   * Explain who should read this file (people who want to understand how the system thinks / high-level behaviour).

**Definition of Done**

* `CONFIGURATION.md` reads like a conceptual guide, not a raw reference.
* There are no large env var or endpoint tables in this file.
* It clearly links to `docs/env_reference.md` and `docs/api_reference.md` for details.

---

## C. Tuning, simulations, and guardrails

### [x] DOC-009 – Create `docs/tuning_playbook.md` and move tuning workflow & AI optimizer content there

**Background**
Tuning workflows and AI optimizer details are currently buried in README and other docs, making README too heavy and making tuning content hard to discover.

**Tasks**

1. Create `docs/tuning_playbook.md`.
2. Move or copy from existing docs:

   * The complete “Tuning Workflow” section (commands, script usage).
   * Any AI optimizer-specific instructions.
3. Organize the file into:

   1. **When to use the tuning harness** (situations when tuning is worth doing).
   2. **How to run the tuning harness** (practical steps & commands).
   3. **Understanding tuning results** (where outputs are stored, how to interpret them).
   4. **Using AI optimizer results** (if applicable).
   5. **Relationship with guardrails and simulations** (link to `docs/simulations_and_guardrails.md`).
4. In README and any other doc that previously held this content:

   * Replace detailed tuning sections with a short summary and link to `docs/tuning_playbook.md`.

**Definition of Done**

* `docs/tuning_playbook.md` exists and captures all tuning/optimizer instructions.
* README and other docs no longer contain long tuning sections; they point to the playbook.
* A developer can follow this doc to run tuning without needing README.

---

### [x] DOC-010 – Merge `docs/simulations_and_guardrails.md` and guardrail documentation into `docs/simulations_and_guardrails.md`

**Background**
Simulation instructions are in `bespoke_simulations.md`, while guardrail explanations are scattered across README and `rules-runbook.md`. They belong together.

**Tasks**

1. Create new file `docs/simulations_and_guardrails.md`.
2. Copy content from `docs/simulations_and_guardrails.md` into this new file, structured as:

   * Why simulate.
   * How to build fixtures.
   * How to run simulations.
   * How to read simulation output.
3. From README and `docs/rules-runbook.md`, copy guardrail-related content:

   * Explanation of `guardrails.yml`.
   * Explanation of scenario names (S7, etc.).
   * How guardrail checks fit into the workflow (e.g. CI step).
4. Integrate guardrails into the new file:

   * Add a conceptual “what guardrails are” introduction (or link to `docs/concepts_and_metrics.md`).
   * Add a section that walks through `guardrails.yml` structure, field by field.
   * Add a section that explains how simulations and guardrails work together.
5. In `docs/simulations_and_guardrails.md`, README, and `docs/rules-runbook.md`:

   * Replace detailed simulation/guardrail sections with short summaries and links to `docs/simulations_and_guardrails.md`.
   * Optionally, mark `bespoke_simulations.md` as deprecated or remove it (depending on your preference).

**Definition of Done**

* `docs/simulations_and_guardrails.md` exists and combines:

  * Simulation instructions.
  * Guardrail configuration and workflow.
* `docs/simulations_and_guardrails.md` no longer contains unique content, or is clearly deprecated.
* README and `docs/rules-runbook.md` point to the new file instead of re-explaining the same details.

---

### [x] DOC-011 – Reduce guardrail details in `docs/rules-runbook.md` and focus on operational runbook

**Background**
`rules-runbook.md` should be an operational document for on-call/operations, not the main home for guardrail configuration details.

**Tasks**

1. Open `docs/rules-runbook.md` and:

   * Identify guardrail-specific configuration explanations (e.g. `guardrails.yml` field-by-field descriptions).
2. Ensure those detailed explanations have been moved to `docs/simulations_and_guardrails.md` (see DOC-010).
3. Tidy up `docs/rules-runbook.md` to focus on:

   * Rule precedence and override behaviour (high level).
   * What to check when something looks wrong (e.g. too many promoted items, odd results).
   * How to use telemetry/metrics to debug.
   * Operational checklists for common incidents (e.g. turning off a rule quickly, reverting a change).
4. Where guardrails are mentioned in `rules-runbook.md`:

   * Link to `docs/simulations_and_guardrails.md` for full details.

**Definition of Done**

* `docs/rules-runbook.md` is primarily an operational runbook.
* Detailed guardrail configuration descriptions live in `docs/simulations_and_guardrails.md`.
* The runbook links to the guardrail doc when needed.

---

## D. Overview and flow

### [x] DOC-012 – Update `docs/overview.md` personas and lifecycle to match new structure

**Background**
`docs/overview.md` currently has persona guidance and lifecycle steps, but links and structure need to be updated to match the new docs.

**Tasks**

1. Update the persona section:

   * Ensure each persona (Business, Integration Engineers, Developers/Ops) has a clear list of suggested docs that reflects the new structure:

     * Business → `docs/business_overview.md`, `docs/overview.md`, `docs/rules_runbook.md`, `docs/concepts_and_metrics.md`.
     * Integration Engineers → `GETTING_STARTED.md`, `docs/quickstart_http.md`, `docs/api_reference.md`, `docs/env_reference.md`, `docs/database_schema.md`.
     * Dev/Ops → `docs/tuning_playbook.md`, `docs/simulations_and_guardrails.md`, `docs/rules_runbook.md`.
2. Review the lifecycle checklist section:

   * Make sure each lifecycle step points to the right new/renamed doc.
   * Remove embedded long command sequences if they are better served in `GETTING_STARTED.md` or `docs/tuning_playbook.md`.
3. Check for any old references (like `api_reference.md` or `env_reference.md`) and update them to the new names.

**Definition of Done**

* Persona guidance in `docs/overview.md` is up to date and matches the new doc structure.
* Lifecycle steps point to the appropriate docs instead of duplicating information.
* No links in `docs/overview.md` point to deleted/renamed files.

---

## E. API reference & error handling

### [x] DOC-013 – Enhance `docs/api_reference.md` with error handling and common patterns

**Background**
The API reference is strong on endpoints but weak on error handling and common integration patterns.

**Tasks**

1. Open `docs/api_reference.md` (renamed from `api_reference.md`).
2. Add a new section: **Error handling & status codes**:

   * List common HTTP response codes the API returns (e.g. 200, 400, 401/403, 404, 429, 500).
   * For each, provide:

     * A plain-language explanation.
     * Typical causes.
3. Add a new section: **Common patterns**:

   * “Get recommendations vs rerank a list” – explain the difference and when to use each endpoint.
   * Any other patterns that are currently documented in README or elsewhere (move them here).
4. Add links to:

   * `docs/quickstart_http.md` for “beginner-friendly” examples.
   * `docs/concepts_and_metrics.md` where concepts are referenced.

**Definition of Done**

* `docs/api_reference.md` contains:

  * A clear Error Handling section.
  * A “Common patterns” section.
* Any “rerank vs recommendations” explanation that used to live in README or other docs has been consolidated here.
* Links to other docs are added where relevant.

---

## F. Cross-linking & cleanup

### [x] DOC-014 – Update cross-links and remove duplicated concept explanations

**Background**
After introducing new docs (concepts, simulations, tuning, etc.), some content will be duplicated and links may be inconsistent.

**Tasks**

1. Search through all markdown files for repeated, long explanations of:

   * ALS, MMR, coverage, guardrails, etc.
   * Tuning workflow steps.
   * Guardrail scenarios and `guardrails.yml`.
2. For each repeated block:

   * Keep the most complete explanation in the appropriate canonical doc:

     * Concepts → `docs/concepts_and_metrics.md`
     * Tuning → `docs/tuning_playbook.md`
     * Simulations & guardrails → `docs/simulations_and_guardrails.md`
   * Shorten other occurrences to 1–2 sentences and link to the canonical doc.
3. Check all references to:

   * `api_reference.md` → should be `api_reference.md`.
   * `env_reference.md` → should be `env_reference.md`.
   * Old locations of glossary → should now refer to `docs/concepts_and_metrics.md`.

**Definition of Done**

* Each major concept (e.g. ALS, MMR, guardrails, tuning harness) has one “home” doc.
* Other docs contain shorter references and links rather than duplicate long explanations.
* No stale filenames remain in links.

---

### [x] DOC-015 – Sanity pass: “Newcomer path” QA

**Background**
The goal of all these changes is to make the system understandable and learnable for someone new. We need to verify the flow works end-to-end.

**Tasks**

1. Ask someone who is **not** familiar with the system (but can read English and basic code) to:

   * Start at `README.md`.
   * Follow links intended for “Integration Engineer” persona.
   * Attempt to:

     * Run the system locally (via `GETTING_STARTED.md`).
     * Make at least one successful HTTP call using `docs/quickstart_http.md`.
   * Read `docs/concepts_and_metrics.md` and summarize the key concepts back, to test clarity.
2. Observe or ask for feedback:

   * Where did they get confused?
   * Which docs felt too advanced?
   * Which questions were not answered?
3. Based on feedback, make small adjustments:

   * Add missing links.
   * Clarify confusing sentences.
   * Add micro-examples where necessary.

**Definition of Done**

* At least one newcomer has walked through the docs and reported:

  * They could get a local instance running.
  * They could get a non-empty recommendations response.
  * They understood, at a high level, what key concepts (e.g. guardrails, coverage, personalization) mean.
* At least 1–2 small improvements are applied based on that feedback.

Here’s a small, focused backlog just for the **remaining issues** I called out after your rewrite. These assume the *current* structure is mostly correct and we’re just tightening the last loose ends.

---

## [x] DOC-201 – Add / Fix `GETTING_STARTED.md` (repo-level quickstart)

**Background**
Several docs now **reference** a file named `GETTING_STARTED.md` in the project root. That file does *not* currently exist in the uploaded docs. This creates broken links and leaves a gap for people who have cloned the repo and want a simple “run it locally and get your first recommendations” guide.

**Goal**
Provide a **single, concrete guide** that explains how to run the system from the repo, seed some data, and call the API locally, without assuming prior knowledge of your internal tooling.

**Tasks**

1. **Create the file**

   * In the repository root, add a new file: `GETTING_STARTED.md`.

2. **Add prerequisites section**

   * List required tools:

     * Docker / Docker Compose (if used).
     * `make` (if your commands use Makefiles).
     * Python version (if Python scripts are used).
     * Any other mandatory tools (e.g. `pnpm`, `node`, etc.).
   * For each tool, include a short note:

     * “If you don’t know what this is, ask your team or see [link to general install guide].”

3. **Add “Run the stack locally” section**

   * Provide a **single recommended** way to start everything, for example:

     * `make dev` or `docker compose up` (whatever is correct for this repo).
   * Describe what this command does in 1–2 sentences (“Starts the API service, database, and supporting components on your machine,” etc.).

4. **Add “Seed sample data” section**

   * Show **one** simple, recommended command that:

     * Seeds a small example dataset into the system (e.g. a script that posts items/users/events).
   * Explain:

     * Where the script lives (e.g. `analysis/scripts/seed_dataset.py`).
     * What environment variables or flags the user must set (e.g. namespace, base URL, org ID).
   * Provide a minimal example where all flags are filled in with obvious placeholder values (`demo`, `http://localhost:XXXX`, etc.).

5. **Add “Call the API” section**

   * Provide a `curl` example (or HTTPie) calling the **local** `/v1/recommendations` endpoint (or whatever your primary endpoint is).
   * Clearly label:

     * Path and method.
     * Required headers (auth, org-id, etc.).
     * Key JSON fields (`namespace`, `user_id`, `k`, etc.).
   * Show a **truncated** sample response and annotate:

     * Where the recommended IDs are.
     * Where metadata lives (e.g. reasons, trace, scores).

6. **Add “If you get stuck” section**

   * Link to:

     * `docs/quickstart_http.md` (for pure HTTP integration examples).
     * `docs/api_reference.md` (for full details).
     * `docs/concepts_and_metrics.md` (for understanding jargon).
   * Suggest what to check if the user sees:

     * “Connection refused”.
     * 401/403.
     * 500 errors.

7. **Update existing references**

   * Search all `.md` files for `GETTING_STARTED.md`.
   * Make sure the context is still correct (e.g. “For repo-based setup, see GETTING_STARTED.md”).
   * Fix any wording that assumed this file already existed.

**Definition of Done**

* `GETTING_STARTED.md` exists in the repo root.
* A developer who has **never** used this repo can:

  * Clone the repo.
  * Follow only `GETTING_STARTED.md`.
  * Bring the stack up locally.
  * Seed at least a small dataset.
  * Successfully call a local recommendations endpoint.
* All references to `GETTING_STARTED.md` in other docs:

  * Point to this file.
  * Make sense in context (e.g. “start here if you’re running locally”).

---

## [x] DOC-202 – Cross-link audit and stale filename cleanup

**Background**
You’ve renamed and reorganized several docs (e.g. to `api_reference.md`, `env_reference.md`, `business_overview.md`, `concepts_and_metrics.md`). After this kind of refactor, it’s very easy for some links to still point to old filenames or wrong paths.

**Goal**
Ensure **all** internal links in markdown files are correct and that there are no references to old filenames or missing docs.

**Tasks**

1. **Create a list of canonical docs and filenames**

   * In a scratch file or issue comment, list the intended filenames:

     * `GETTING_STARTED.md`
     * `CONFIGURATION.md`
     * `docs/overview.md`
     * `docs/concepts_and_metrics.md`
     * `docs/business_overview.md`
     * `docs/quickstart_http.md`
     * `docs/api_reference.md`
     * `docs/env_reference.md`
     * `docs/database_schema.md`
     * `docs/tuning_playbook.md`
     * `docs/simulations_and_guardrails.md`
     * `docs/rules-runbook.md`
   * Confirm these are the actual names on disk.

2. **Search for references to old filenames**

   * Grep / search for:

     * `api_endpoints.md`
     * `env_vars.md`
     * Any other old names you used before the refactor.
   * For each occurrence:

     * Replace it with the new filename/path.
     * Adjust link text if needed (e.g. “API Endpoints” → “API Reference”).

3. **Search for broken or missing references**

   * Search for strings like:

     * `GETTING_STARTED.md`
     * `business_overview.md`
     * `concepts_and_metrics.md`
     * `quickstart_http.md`
     * etc.
   * For each link:

     * Check that the target file actually exists.
     * Check that the **relative path** is correct (e.g. from README to `docs/...`).
   * Fix any incorrect paths (for example `docs\file` vs `docs/file` if path separators are wrong).

4. **Manual click-through**

   * Open each main doc in a markdown viewer (or GitHub web UI).
   * Click through:

     * All links in `README.md`.
     * All links in `docs/overview.md`.
     * All links in `docs/business_overview.md`.
     * All links in `docs/quickstart_http.md`.
   * Fix any broken links you encounter.

**Definition of Done**

* No markdown file contains links to:

  * `api_endpoints.md`
  * `env_vars.md`
  * Any other removed or renamed file.
* All main docs’ internal links (README, overview, business overview, quickstart HTTP) successfully navigate to the correct targets.
* There are **no 404s** when clicking internal markdown links in your usual viewer (GitHub, IDE preview, etc.).

---

## [x] DOC-203 – Clarify “hosted vs local” and doc scope around dev tooling commands

**Background**
Some docs use commands like `make`, Docker, `pnpm`, and internal scripts. These are appropriate for people running the full stack locally or working on the codebase itself, but **not** for external integrators who only call the hosted HTTP API.

Right now, it’s possible for a low-context reader to end up in a doc with `make dev` and think “Do I need to do this to use the API?”, which causes confusion.

**Goal**
Make it explicit in each relevant doc who it is for and whether it’s about **running the system locally** or just about **using the hosted API**.

**Tasks**

1. **Identify docs that mention dev tooling**

   * Search the markdown files for:

     * `make `
     * `docker`
     * `pnpm`
     * `analysis/scripts/`
   * List all docs where these appear (for example: README, tuning playbook, simulations doc, etc.).

2. **At the top of each such doc, add or clarify “Who should read this”**

   * Ensure each doc that includes dev-only commands has a line like:

     * “This doc is for developers running the RecSys stack locally or contributing to the codebase.”
   * Make sure it also includes a pointer for integrators, e.g.:

     * “If you just need to call the hosted API, see `docs/quickstart_http.md` instead.”

3. **Add short inline notes near dev commands**

   * For each dev-command-heavy section (e.g. long `make` or script invocations), add a short note before the first command:

     * Example:

       > “You only need to run these commands if you’re running RecSys locally from source. Hosted API users can skip this and use `docs/quickstart_http.md`.”
   * Keep the note short but explicit.

4. **Check HTTP-only docs**

   * Verify that `docs/quickstart_http.md` and `docs/business_overview.md` do **not** contain dev-only commands (Docker, make, pnpm).
   * If they do, either:

     * Move those commands into dev-oriented docs, or
     * Clearly label them as “advanced / local setup (optional)”.

**Definition of Done**

* Every doc that uses dev commands (make, Docker, pnpm, internal scripts):

  * Clearly states at the top who it’s for (local dev / contributors).
  * Contains a small inline note near the first dev command explaining that hosted API users can ignore it.
* HTTP-integration docs and business docs:

  * Don’t misleadingly suggest you must run Docker/Make to use the hosted API.
  * Remain clean and approachable for non-dev stakeholders.

---

## [x] DOC-204 – Micro-copy pass: introduce acronyms and jargon with links to concepts doc

**Background**
You’ve centralized definitions in `docs/concepts_and_metrics.md`, which is great. However, some docs still **introduce acronyms or jargon (e.g. MMR, ALS, bandits)** without an inline explanation or link on first mention. For a new reader, this is a small but noticeable usability hit.

**Goal**
Ensure that the **first time** an acronym or specialized term appears in each doc, it either:

* is written out in full (with the acronym), or
* links to `docs/concepts_and_metrics.md`.

**Tasks**

1. **Identify key acronyms and jargon**

   * Use `docs/concepts_and_metrics.md` as the source of truth for important terms:

     * Examples: ALS, MMR, bandit, guardrail, segment lift, coverage, long-tail share, scenario S7, etc.
   * Build a simple list of these terms.

2. **Scan each major doc**

   * For each of the following docs:

     * `README.md`
     * `CONFIGURATION.md`
     * `docs/overview.md`
     * `docs/business_overview.md`
     * `docs/tuning_playbook.md`
     * `docs/simulations_and_guardrails.md`
     * `docs/rules-runbook.md`
   * Check where each term in your list first appears.

3. **Introduce or link terms on first mention**

   * For each first occurrence:

     * Either:

       * Write it out with the acronym, e.g.
         “Maximal Marginal Relevance (MMR)”
       * Or add a parenthetical link, e.g.
         “MMR (see `docs/concepts_and_metrics.md` for a plain-language explanation).”
   * Choose whichever keeps the sentence clean in that context.

4. **Avoid over-linking**

   * Only link on the **first** occurrence per doc, not every time the word appears.
   * If a doc is meant for a non-technical audience (e.g. `business_overview.md`), prefer plain-language phrasing + one link, rather than heavy jargon.

5. **Re-run a quick pass over `docs/concepts_and_metrics.md`**

   * Make sure all terms you’re linking from other docs are actually defined there.
   * If a term is missing, add a short definition.

**Definition of Done**

* In each major doc, the first occurrence of important jargon/acronyms:

  * Is either spelled out with the acronym, or
  * Clearly links to `docs/concepts_and_metrics.md`.
* `docs/concepts_and_metrics.md` contains entries for all terms being referenced.
* A new reader can follow any doc and:

  * Not get stuck on an unexplained acronym.
  * Always have a clear, single place to click for a definition.

Here you go — tickets just for the “nice-to-have” extra work I mentioned.

I’ll continue IDs from the last batch (`DOC-201+`).

---

## [x] DOC-205 – Add TL;DR blocks to heavy, advanced docs

**Background**
Some advanced docs (`tuning_playbook`, `simulations_and_guardrails`, `database_schema`, etc.) are dense by necessity. They’re good, but a quick “What is this? When do I need it?” summary at the top would help newcomers decide whether to dive in or not.

**Goal**
Add a small, high-signal **TL;DR** section at the top of heavy docs so a reader can decide in a few seconds if this document is relevant to them right now.

**Docs in scope (minimum)**

* `docs/tuning_playbook.md`
* `docs/simulations_and_guardrails.md`
* `docs/database_schema.md`
* Optionally: any other doc that feels “long and deep” (use judgment).

**Tasks**

1. **Identify heavy docs**

   * Open each `.md` in `docs/`.
   * Flag those that:

     * Are long, and
     * Are mainly for advanced / niche workflows (not basic intro).

2. **For each chosen doc, add a TL;DR near the top**

   * Insert right below the main title and “Who should read this?” / intro.
   * TL;DR format (recommended):

     * A heading: `### TL;DR`
     * 3–5 bullets:

       * **What this doc is for** (“How to tune RecSys using the evaluation harness.”)
       * **When you need it** (“Use this when onboarding a new org or changing signal weights.”)
       * **What you get out of it** (“A set of suggested env profiles with evidence from scenario suites.”)
       * **What it’s not** (“Not required for basic HTTP integration.”)

3. **Make sure TL;DR is plain language**

   * Use minimal jargon; if you must reference jargon, link to `docs/concepts_and_metrics.md`.
   * The TL;DR should be understandable by someone who’s just read README + Business Overview.

4. **Verify consistency**

   * Ensure TL;DR sections are visually consistent across docs:

     * Same heading style (`### TL;DR`).
     * Similar bullet structure.

**Definition of Done**

* All flagged “heavy” docs have a `### TL;DR` section directly under their title/intro.
* Each TL;DR has 3–5 short bullets clearly explaining:

  * Purpose,
  * When to use the doc,
  * What benefit it gives,
  * What it is *not*.
* A new reader can skim the TL;DR and correctly decide whether they need to read the rest right now.

---

## [x] DOC-206 – Add simple diagrams for mental model and architecture

**Background**
The docs describe the RecSys mental model and architecture in text (ingestion → signals → blending → personalization → diversity/caps → rules → response; client → API → DB, features, metrics). A simple diagram or two would help visual thinkers understand this pipeline faster.

**Goal**
Add **at least two simple diagrams** to the docs:

1. One for the **recommendation pipeline / mental model**.
2. One for the **high-level system architecture**.

You don’t need fancy tooling — a PNG/SVG exported from draw.io, Excalidraw, etc. is fine.

**Tasks**

1. **Design the “pipeline” diagram**

   * Show the conceptual flow:

     * Items/users/events in → “Signals” (popularity, co-visitation, embeddings) → “Blending” → “Personalization & cold-start handling” → “Diversity/MMR & caps” → “Rules & guardrails” → “Final ranked list”.
   * Keep shapes and labels simple; avoid implementation details.
   * Export as `docs/images/recsys_pipeline.png` (or `.svg`).

2. **Design the “architecture” diagram**

   * At minimum, show:

     * Client(s): web/app/backend.
     * RecSys API service(s).
     * Storage components: main DB, feature store / cache (if applicable).
     * Evaluation/metrics: Prometheus/Grafana, simulation harness, guardrail CI.
   * Optional: separate “operational” vs “evaluation/tuning” paths.
   * Export as `docs/images/recsys_architecture.png` (or `.svg`).

3. **Add diagrams to appropriate docs**

   * Pipeline diagram:

     * Embed in `CONFIGURATION.md` near the “Mental model & data flow” section.
     * Optionally also in `docs/concepts_and_metrics.md` near the top, as a visual anchor.
   * Architecture diagram:

     * Embed in `docs/overview.md` in the “Lifecycle / architecture” section.
     * Optionally referenced in `docs/business_overview.md` as “full system picture”.

4. **Add short captions**

   * Under each image, add a one-sentence caption:

     * Example: “High-level RecSys pipeline from raw events to final ranked recommendations.”
     * Example: “High-level RecSys architecture showing API, storage, and evaluation components.”

5. **Check relative paths**

   * Ensure image paths work from each doc (e.g. `![caption](images/recsys_pipeline.png)` from inside `docs/`).
   * Preview markdown in GitHub or your IDE to confirm images render.

**Definition of Done**

* There is at least:

  * 1 pipeline diagram image in the repo.
  * 1 architecture diagram image in the repo.
* `CONFIGURATION.md` and `docs/overview.md` embed the diagrams and render them correctly.
* Each diagram has a short, accurate caption.
* No broken image links in the markdown.

---

## [x] DOC-207 – Tone & micro-copy polish for non-technical readers

**Background**
Overall tone is good and direct. In some places, wording is quite brisk and assumes a “senior engineer” mindset. For non-technical stakeholders (PMs, biz), softening a few edges and clarifying intent can make the docs feel more welcoming without dumbing them down.

This is not a rewrite; it’s a **micro-copy pass**.

**Goal**
Make sure the tone in business-facing and entry-point docs is:

* clear,
* direct,
* but not unnecessarily sharp or intimidating for non-devs.

**Docs in scope**

Prioritize:

* `README.md`
* `docs/business_overview.md`
* `docs/concepts_and_metrics.md`
* `docs/overview.md`
* `docs/rules-runbook.md` (sections aimed at merch/PM)

**Tasks**

1. **Read each in-scope doc in “PM mode”**

   * Pretend you’re a product manager who:

     * Knows basic tech,
     * Does not write code,
     * Wants to feel comfortable asking questions.

2. **Flag lines that are:**

   * Overly brusque or snarky.
   * Heavy on internal shorthand (“we just…”, “obviously…”).
   * Slightly dismissive of less advanced usage (“if you don’t care about X, ignore this”).

3. **Rewrite flagged lines to be:**

   * Still honest and direct, but:

     * Neutral in tone.
     * Explanatory instead of judgemental.
   * Examples:

     * “You probably don’t want to do this” → “You typically don’t need this unless you’re doing X; if you are, here’s how.”
     * “This doc is not for you” → “This doc is mainly for developers running RecSys from source; if you just call the hosted API, you can skip it.”

4. **Add gentle context where needed**

   * If a sentence drops a concept or tool abruptly (e.g. “Use bandits”), consider:

     * Adding “(see `docs/concepts_and_metrics.md` for a short explanation)” or
     * Adding a 3–5 word parenthetical explanation.

5. **Avoid over-sanitising dev-only docs**

   * Do **not** touch the bluntness of deeply technical sections if they’re clearly labeled for devs (e.g. tuning harness docs).
   * Focus on business-facing and entry docs.

**Definition of Done**

* Business-facing and entry-point docs (README, Business Overview, Concepts/Metrics, Overview, rules runbook intro) read:

  * Clear and direct,
  * Technically accurate,
  * Non-intimidating for non-devs.
* No obvious “this sounds like an irritated senior engineer shouting into Slack” moments remain in those docs.
* Dev-only docs are left largely as-is, except where they accidentally talk to non-dev readers.

---

## [x] DOC-208 – Add a “New hire onboarding checklist” doc

**Background**
The docs are now well-structured and role-based. You can turn them into a concrete onboarding path for new team members (especially engineers and PMs) by writing a short “Day 1 / Day 2 …” style checklist. This is low effort and high impact for internal onboarding.

**Goal**
Create a short checklist document that you can hand to a new hire saying: “Follow this sequence to ramp up on RecSys.”

**Tasks**

1. **Create the doc**

   * Add `docs/onboarding_checklist.md`.

2. **Define 2–3 personas to onboard**

   * For example:

     * “New backend/integration engineer”
     * “New ML/recs engineer”
     * “New PM/product owner”
   * For each persona, you’ll write a small checklist.

3. **Write an onboarding sequence per persona**

   * For each persona, define ~10–15 steps (bullets), grouped roughly by day or phase:

     * **Phase 1 / Day 1–2 – Understanding**

       * “Read README.”
       * “Read docs/business_overview.md.”
       * “Read docs/concepts_and_metrics.md.”
     * **Phase 2 / Day 3–4 – Integration basics**

       * Engineer: “Follow docs/quickstart_http.md and call the hosted API with a sample payload.”
       * PM: “Review docs/api_reference.md at a high level and note any unclear fields.”
     * **Phase 3 / Day 5+ – Advanced / role-specific**

       * Backend/ML engineer: “Run GETTING_STARTED.md locally, skim tuning_playbook, run one simulation.”
       * PM: “Read rules-runbook intro and discuss rule strategy with team.”

4. **Cross-link to existing docs**

   * Don’t repeat content; just reference:

     * README, Business Overview, Concepts/Metrics, Quickstart HTTP, GETTING_STARTED, API Reference, Env Reference, Tuning Playbook, Simulations & Guardrails, Rules Runbook, Database Schema (if relevant).

5. **Add a short intro**

   * At the top of `onboarding_checklist.md`, explain:

     * “This is a suggested ramp-up path for new team members.”
     * “Timelines are indicative; go at your own pace.”
     * “Focus on the persona closest to your role; you don’t need to do everything.”

6. **Link this doc from README / overview**

   * In README’s persona map or in `docs/overview.md`, add:

     * A short line: “New team member? See docs/onboarding_checklist.md for a suggested ramp-up path.”

**Definition of Done**

* `docs/onboarding_checklist.md` exists with:

  * At least 2 personas,
  * 10–15 concrete steps per persona, grouped by phases/days.
* Steps are phrased as actionable tasks (“read X”, “run Y”, “try Z”), not vague suggestions.
* README or `docs/overview.md` links to this new doc so it’s discoverable.
* A new hire could realistically follow it as a week-1/2 plan.

---
