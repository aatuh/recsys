At a high level: for *recsys-savvy engineers on your team*, this doc set is very strong. For *someone who doesn’t know what Recsys does or what “MMR / ALS / guardrails / S7” mean*, it’s still too dense and assumes too much context.

If I had to put numbers on it:

* **For internal senior engineers:** clarity 8.5–9/10, completeness 9/10
* **For an external-ish integrator or PM with little jargon:** clarity 6–7/10, learnability 6/10

Below I’ll break it down through the three lenses you asked for and then give you concrete “if I owned this repo, what I’d change first”.

---

## 1. As a technical writer

### 1.1 Clarity & Understandability

**What’s working**

* You clearly know your audience *when you choose to*:

  * `overview.md` starts with personas (Business, Integration Engineers, Developers/Ops) and “Who should read this?” which is exactly what good docs do. 
  * `database_schema.md` opens with a simple purpose statement and “Who should read this?” too. 
* `CONFIGURATION.md` is explicitly framed as *“written in plain English, with copy-paste payloads and a clear mapping between environment variables and per-request overrides”* and actually delivers on that fairly well. The “Ingestion → Signals → Blending → Personalization → Diversity/Caps → Rules → Response” mental model is one of the clearest conceptual summaries in the whole set. 
* The README “Need a map?” + persona/quick-links table is excellent IA work. It tells people what to read depending on their role. 
* The glossary at the end of README is solid, concrete, and written in accessible language (ALS, Anchors, Blended scoring, Candidate, Caps, Cold start, etc.). 

**Where it falls down for a low-context reader**

* **Jargon density is very high very early.**
  “MMR”, “guardrails”, “S7/S8/S9”, “segment lifts”, “ALS”, “bandit policies”, “scenario suite” appear all over the place before they’re really introduced in everyday language. Example: the Onboarding & Coverage checklist talks about scenario S7, cold-start guardrails, NDCG/MRR, coverage thresholds with no gentle conceptual ramp.
* **Key concepts are defined, but not where you need them.**
  The glossary is tucked near the end of README. A novice has to scroll a *lot* before they discover “oh, co-visitation is defined down here”.
* **A lot of content is tuned for evaluation-rubric power users, not first-time integrators.**
  README spends a ton of space on tuning harness, AI optimizer, simulation orchestrator, load & chaos harness, etc. (all great, but advanced).
  Somebody who just wants, “How do I send items/users/events and get recommendations back?” has to dig through that noise.
* **Metric jargon is assumed knowledge.**
  You use NDCG, MRR, “segment lifts”, “long-tail share”, “coverage ≥0.60” everywhere without a short conceptual explanation or link to a “Metrics” mini-primer.

Net: clarity is *good* at the sentence and paragraph level, but **too much expertise is assumed at the document level** for someone new to recommender systems.

### 1.2 Flow & Information Architecture

**What’s good**

* **Entry points are thoughtful.**

  * README → “Start Here / Recsys Tuning Playbook” → role-based mapping to deeper docs. 
  * `overview.md` repeats this with persona sections plus Quick Links and a lifecycle checklist (Seed → Configure → Run simulations → Deploy rules → Monitor). 
* Most individual docs have a clear internal structure:

  * `CONFIGURATION.md` → TL;DR → mental model → env vars → tuning. 
  * `api_reference.md` is a well-structured table-based reference with “Usage tips”.
  * `bespoke_simulations.md` is a clean step-by-step: Prepare env profiles → Build fixture → Run simulation → Interpret evidence. 
  * `rules-runbook.md` reads like an actual runbook (rule precedence, telemetry, guardrails, operational checklist).

**Where the flow is confusing as a *whole***

* **There’s no single “Hello World” path for a newcomer.**
  There is no one doc explicitly called “Getting Started” that takes a novice from:

  1. “What is this?” →
  2. “How to spin up the service or hit the demo base URL” →
  3. “Copy-paste items/users/events JSON” →
  4. “Copy-paste `/v1/recommendations` call and interpret the response”.
     The closest is `CONFIGURATION.md` + bits of README, but they’re still heavily slanted toward tuning and guardrails rather than making your *first successful call*.
* **Overlap & duplication blur the mental map.**

  * Env/config concepts live in **three** places: README config section, `CONFIGURATION.md`, and `env_reference.md`.
  * Guardrails + simulation show up in README, `bespoke_simulations.md`, `rules-runbook.md`, and `overview.md`.
    This makes it harder for a new reader to know “where is the canonical source?” and creates a risk of drift.
* **Advanced topics are mixed into the “front door”.**
  The README jumps from persona table to a full Tuning Workflow (multiple scripts, namespaces, profiles, AI optimizer, guardrail checks) and load/chaos harnesses.
  This sequencing is perfect for a power user but overwhelming for a newcomer.

### 1.3 Completeness

For an engineering audience, coverage is excellent:

* API surface, including health/version/docs/metrics, ingestion, ranking, bandit, rules, audit, data governance.
* Env/algorithm config in depth (`env_reference.md` + README).
* DB schema with precise table layouts. 
* Simulations, guardrails, CI workflows.
* Rules & overrides runbook including metrics and on-call workflow.

What’s *missing* for someone who doesn’t know the system/jargon:

* **A conceptual “What business problem does this solve?” page** aimed at PMs / non-ML folks. README hints at it (“domain-agnostic recommendation API… safe defaults, multi-tenant, etc.”) but then disappears into scripts and evaluation harnesses. 
* **A simple lifecycle example in one concrete domain.** E.g. “Retail shop: show related products on PDP”. Walk through fields and reasoning in that context.
* **Formal “Metrics & guardrails explained in human terms” doc.** Something that explains NDCG/MRR, “segment lift”, cold-start coverage, long-tail share in plain language, ideally with micro-examples, not just thresholds in YAML/scripts.
* **Error handling, status codes and common failure modes** are not really documented in `api_reference.md`. There’s auth, rate limits, versioning, but no table of typical error responses or “common mistakes” (missing namespace, bad org header, invalid blend, etc.).

---

## 2. As a business/product representative

From a buyer / PM perspective, your docs scream:

> “This thing is serious, audited, and battle-tested; we care about guardrails, telemetry, and evidence.”

That’s a *good* signal. The repeated emphasis on:

* guardrails (coverage, lifts, cold-start),
* simulation and evidence bundles,
* audit trails, explainability, rule engine,
* bandit experimentation,

all say “we’re not a toy, and we can prove it”.

But there are issues for non-technical stakeholders:

* **Value is implicit, not explicit.**
  The docs assume the reader already knows why recommendations, guardrails, bandits etc. matter. There’s no short page that says in plain language:

  * “Here’s what you get if you switch us on in your shop in week 1 vs week 8.”
  * “Here’s how we protect you from bad behaviour (spammy items, over-personalization, etc.).”
* **Scenario names and metrics are opaque.**
  “S7,” “≥+10% lift on NDCG/MRR,” “long-tail share ≥0.20” without a one-paragraph story of *what that means for revenue or UX* is fine for engineers, not fine for PMs.
* **There’s no executive one-pager living in the docs themselves.**
  You *hint* at “executive_summary_template.md” in README, which is nice. 
  But there isn’t a visible “Executive Summary” markdown someone can open and grok in 5 minutes.

So: the docs support a narrative of **rigour and safety**, but they don’t yet tell a crisp business story in their own words.

---

## 3. As a senior developer / integrator

From a dev perspective, you’ve done a lot of things right:

* **The surface area is well documented.** API endpoints, DB schema, env vars, and operational scripts are all there and cross-linked.
* **Docs are “literate”:** they don’t just list knobs, they explain *why* you’d change them (e.g., explanations around `PROFILE_BOOST`, interactions between blend weights and MMR, guardrail implications).
* **Runbooks and checklists are exactly what you’d want in production.** Onboarding & coverage checklist, rule runbook, operational checklists, etc.

But as someone inheriting this repo for the first time:

* **It’s hard to know what is *required* vs *nice to have*.**
  The docs blur the line between: “You *must* do X to get correct behaviour” and “Here’s extra evaluation / tuning tooling you can use if you care.” Almost every doc heavily references `analysis/scripts/...` which makes it feel like you *must* adopt the entire harness just to integrate.
* **Canonical sources are not obvious.**
  As mentioned, env config exists in multiple places; guardrail semantics are defined in README, rules runbook, bespoke simulations, overview, etc. If I add a new guardrail field, I have to update a lot of documents and hope I don’t miss one.
* **There isn’t a minimal working example for each persona.**
  For integration engineers you give endpoints + DB schema, but not a “minimal example using only HTTP calls” to seed & query without touching internal scripts or Make.

So as a senior dev, I’d be happy with the depth, but I’d also grumble about:

* “Where’s the smallest path to ‘it works’?”
* “Which doc is the source of truth for env/guardrails?”

---

## 4. Summary per dimension you asked about

### Clarity

* **Micro-level (sentences, tables):** clear, consistent tone, mostly excellent.
* **Macro-level (what’s important, what’s advanced):** muddled for newcomers. Advanced tooling crowds out core concepts.

### Understandability for someone new to the system/jargon

* Strong conceptual pieces exist (`CONFIGURATION.md` mental model, README glossary), but they are:

  * scattered,
  * not front-loaded,
  * and overshadowed by advanced content and jargon.

I wouldn’t call the current set “excellent learnability” for a low-context newcomer yet.

### Flow from one topic to another

* Within individual docs: flows are generally good (intros, who-should-read-this, checklists).
* Across the whole set: there’s no clean, *progressive* reading path: “Intro → Basic Integration → Configuration → Advanced Tuning → Ops”. Instead, readers bounce between overlapping documents with repeated topics.

### Completeness

* For engineers and ops: very complete, arguably more than most SaaS APIs.
* For PMs / less technical stakeholders: missing a clear business-focused overview, metrics primer, and “read this first” story.
* For integrators: missing a simple, self-contained quickstart based on pure HTTP requests, not the evaluation harness.

---

## 5. If I owned this, what I’d change first (to reach “excellent learnability”)

Not what you asked directly, but it’s the practical conclusion of the evaluation:

1. **Add a single, non-scary “Getting Started” doc (or section) for newcomers**

   * 1-page flow: what Recsys does, how to hit a demo base URL, copy-paste examples for `/v1/items:upsert`, `/v1/users:upsert`, `/v1/events:batch`, `/v1/recommendations`, with *annotated* responses.
   * Link to this first from README before any tuning/simulation stuff.

2. **Refactor env/config docs into one canonical chain**

   * Make `env_reference.md` the *only* detailed env reference and drastically slim down env tables in README. README should just point to it.

3. **Surface the glossary + metrics primer early**

   * Move (or duplicate) glossary higher in README or into a separate `concepts.md` linked from every doc.
   * Create a small “Metrics & guardrails explained” doc written in PM-friendly language; link from everywhere you mention NDCG/MRR/S7.

4. **Split README into two logical tiers**

   * Top: high-level overview, quickstart, persona map.
   * Bottom / separate docs: tuning harness, AI optimizer, load/chaos testing, simulation orchestrator.

5. **Reduce duplication around guardrails & simulations**

   * Decide whether `CONFIGURATION.md` + `bespoke_simulations.md` + README guardrail section + `rules-runbook.md` each own a slice of the story, and make clear cross-links (“for guardrail details, see X; this doc only covers Y”).

If you do those five things, the core content you already have is strong enough that you *will* get to “excellent understandability” for newcomers without rewriting everything from scratch.

# Reorganization Plan

Here is a sketch for a concrete new structure (file list, section order, and where to move/trim content) tailored to the personas you care about most.

I’ll give you:

1. **New file structure + roles**
2. **Outline for each doc (what it contains, who it’s for)**
3. **A migration plan: where current content should move/merge/vanish**

---

## 1. Proposed doc structure

### Root

* `README.md` – **Front door + quickstart** (no deep tuning, no CI workflows)
* `GETTING_STARTED.md` – **Step-by-step “Hello Recsys”** for low-context integrators
* `CONFIGURATION.md` – **Configuration & data ingestion *conceptual* guide** (keeps the mental model you already have)

### `/docs` directory

* `docs/overview.md` – **Personas & lifecycle** (your current overview, slightly trimmed)
* `docs/concepts_and_metrics.md` – **New:** key concepts, glossary, metrics/guardrails primer
* `docs/quickstart_http.md` – **New:** HTTP-only quickstart (cURL/HTTPie/postman) for APIs, separate from repo-driven `GETTING_STARTED.md`
* `docs/api_reference.md` – **Rename from** `api_reference.md`; keep as authoritative endpoint ref
* `docs/env_reference.md` – **Rename from** `env_reference.md`; canonical env/algorithm reference
* `docs/database_schema.md` – keep, maybe add examples 
* `docs/tuning_playbook.md` – **New:** move tuning harness, AI optimizer, guardrail scripts out of README into here
* `docs/simulations_and_guardrails.md` – **Merge** `bespoke_simulations.md` + guardrails.yml explanations
* `docs/rules_runbook.md` – keep as is or lightly trim; remains the ops runbook for rules/overrides
* `docs/business_overview.md` – **New:** 1–2 page product/exec overview in plain biz language

That’s basically:

* Root = front-door + repo-ops + first happy path
* `/docs` = deeper material, clearly themed

---

## 2. What each file should look like

### `README.md` – front door + quickstart

**Audience:** anybody landing on the repo for the first time.

**Keep it short and non-scary. Target ~2–3 screens.**

**Sections:**

1. **What Recsys Is (high-level)**

   * 3–5 bullets: domain-agnostic recommendation API, opaque IDs, safe defaults, guardrails, multi-tenant. 

2. **What You Can Do With It (in normal words)**

   * 3–4 example use cases (“similar items on PDP”, “personalized home feed”, “rerank search results”, etc.).

3. **Quickstart (Your First Recommendations in 5 Steps)**
   Link to `GETTING_STARTED.md` but give a *tiny* inline version:

   * Clone + `make up` (or however you boot)
   * Seed sample data (`seed_dataset.py` or sample fixture) 
   * Hit `/v1/recommendations` with a copy-paste request
   * See JSON response + link to `docs/api_reference.md` for fields

4. **Where To Go Next (Persona Map)**
   Keep your “Need a map?” table but simplify: point into the new docs:

   * Business/Product → `docs/business_overview.md`, `docs/overview.md` (Business section), `docs/rules_runbook.md`
   * Integration Eng → `GETTING_STARTED.md`, `docs/quickstart_http.md`, `docs/api_reference.md`, `docs/env_reference.md`, `docs/database_schema.md`
   * Dev/Ops → `docs/overview.md` (Dev/Ops), `docs/tuning_playbook.md`, `docs/simulations_and_guardrails.md`, `docs/rules_runbook.md`

5. **Pointers to Advanced Stuff**

   * One small section listing: tuning harness, AI optimizer, load/chaos workflows, CI guardrails – each just a one-liner pointing to the right doc (`docs/tuning_playbook.md`, `docs/simulations_and_guardrails.md`, etc.).

**What to remove from README:**

* Full **Tuning Workflow** commands; move to `docs/tuning_playbook.md` and leave only a link.
* Detailed **Onboarding & Coverage checklist + cold-start guardrails**; condense to a 3-bullet summary and link to `docs/simulations_and_guardrails.md`.
* Huge **configuration/feature flag section**; move to `docs/env_reference.md` and `CONFIGURATION.md` (conceptually).

---

### `GETTING_STARTED.md` – repo-centric quickstart

**Audience:** integration engineer with repo access; okay with Python/Make, not with your internal jargon yet.

**Goal:** get them from zero to “I ran the stack locally and got recommendations” *once*.

**Sections:**

1. **Prereqs**

   * Docker, Python version, `make`, etc.

2. **Start the Stack**

   * `make up` / `docker compose up` snippet.

3. **Seed Sample Data**

   * Single recommended path; minimal footguns:

     ```bash
     python analysis/scripts/seed_dataset.py \
       --base-url http://localhost:8000 \
       --org-id "$RECSYS_ORG_ID" \
       --namespace demo
     ```

     plus optional `--fixture-path` example.

4. **Call the API**

   * One example for `/v1/items:upsert` (if needed), but focus on **one** `/v1/recommendations` request with comments explaining:

     * `namespace`
     * `user_id`
     * `k`
     * `context.surface`
   * Show how to inspect the trace fields (`trace.extras`, reasons) without drowning them.

5. **What’s Next**

   * “Want to call it from your app?” → `docs/quickstart_http.md`
   * “Want to tune algo knobs?” → `docs/tuning_playbook.md`
   * “Want to understand jargon?” → `docs/concepts_and_metrics.md`

---

### `CONFIGURATION.md` – conceptual configuration & data guide

You already have a strong “mental model & data flow” and a reasonably clear TL;DR. Keep that, but treat env var details as *examples*, not as the canonical reference (that’s `docs/env_reference.md`).

**Audience:** integrators & devs who now have a basic integration and want to understand how the engine thinks.

**Sections:**

1. **What Configuration Controls (business language)**

   * Popularity, co-visitation, embeddings, personalization, diversity, rules, bandits – each as a short paragraph.

2. **Mental Model & Data Flow** (keep your existing diagram & explanation)

3. **Data Requirements & Recommended Shapes**

   * What fields items/users/events *must* and *should* have.
   * Link to `docs/api_reference.md` and `docs/database_schema.md` for exact schemas, don’t duplicate.

4. **Key Knobs by Theme (with cross-links)**

   * Brief subsections:

     * Diversity & caps
     * Personalization & cold start
     * Purchase suppression
     * Bandits / experiments
   * For each: *conceptual* explanation + “For full env var list, see `docs/env_reference.md`”.

5. **Typical Config Recipes**

   * “We want more diversity on home”
   * “We want more long-tail exposure”
     Each recipe is high-level plus links to env vars in `docs/env_reference.md`.

---

### `docs/overview.md` – personas & lifecycle

You can mostly keep this as-is, but remove some script-heavy bits and instead link to the dedicated docs.

**Audience:** everyone; especially good as a “second read” after README.

**Sections:** (you basically have these already)

1. **Personas (Business, Integration, Dev/Ops)** – but each bullet should now point to *the new docs* instead of repeating content:

   * Business → `docs/business_overview.md`, `docs/rules_runbook.md`
   * Integration → `GETTING_STARTED.md`, `docs/quickstart_http.md`, `docs/api_reference.md`, `docs/env_reference.md`, `docs/database_schema.md`
   * Dev/Ops → `docs/tuning_playbook.md`, `docs/simulations_and_guardrails.md`, `docs/rules_runbook.md`

2. **Lifecycle Checklist** – keep, but strip command spam; show the steps and link to the right doc/section for scripts. 

---

### `docs/concepts_and_metrics.md` – new “brain” doc

**Audience:** PMs, business stakeholders, and engineers who don’t live in recommender math.

**Sections:**

1. **Core Concepts (non-jargony)**

   * Candidate, retrieval vs ranking
   * Popularity vs co-visitation vs embeddings
   * Personalization & starter profiles
   * Rules & overrides
     You already explain most of these across README + CONFIGURATION + env docs; bring them together here.

2. **Metrics Primer**

   * NDCG, MRR, “segment lifts”, “catalog coverage”, “long-tail share” explained in *plain language* with tiny examples (like 5 items).
   * Explain why S7 cold-start matters, what “≥+10% lift” actually means for a cohort.

3. **Guardrails – What They Are, Business-wise**

   * Describe guardrails.yml at a high level.
   * Explain scenario suite S1–S10 conceptually, with 1–2 sentence summaries; link to `docs/simulations_and_guardrails.md` for implementation.

4. **Glossary**

   * Move the glossary from README here, and link to it early. 

Then in README & overview, when you drop jargon like “ALS”, “MMR”, “segment lifts”, you link here instead of redefining.

---

### `docs/quickstart_http.md` – HTTP-only quickstart

**Audience:** external integration engineer or partner who never wants to touch your analysis scripts.

**Sections:**

1. **Base URL, Auth, Namespaces**

   * Explain base URL, auth header, `namespace` expectations.

2. **Ingest Minimal Data via HTTP**

   * Example `curl` for `/v1/items:upsert`, `/v1/users:upsert`, `/v1/events:batch`.

3. **Fetch Recommendations and Similar Items**

   * `/v1/recommendations` example plus annotated response.
   * `/v1/items/{item_id}/similar` example.

4. **Common Mistakes**

   * Missing namespace, mismatch of org-id, invalid blend, etc. (this is missing from your docs currently and will save you support time).

5. **Next Steps**

   * Point to `docs/api_reference.md` and `docs/concepts_and_metrics.md`.

---

### `docs/api_reference.md` – rename of `api_reference.md`

Keep the core structure; it’s solid. But treat it as *reference*, not tutorial.

**Key tweaks:**

* Add **Error Handling & Status Codes** section:

  * Common 4xx/5xx, what they mean, and typical causes.
* Add **Common Patterns**:

  * “Rerank vs recommendations” (you already have this in README; move it here and keep a short copy/link in README).

---

### `docs/env_reference.md` – rename of `env_reference.md` (canonical)

This becomes the **only** authoritative list of env vars and their interactions.

**Sections:**

* Ingestion & windows
* Diversity, MMR, coverage
* Personalization & starter profiles
* Rules & overrides
* Bandits
* Service metadata

Anywhere else env vars are mentioned, they should be **conceptual** and link here instead of re-listing tables.

---

### `docs/tuning_playbook.md` – new home for advanced tuning

Move your **Tuning Workflow** and **AI optimizer** content here. Keep it explicit that this is **advanced**.

**Sections:**

1. When to use the tuning harness / AI optimizer
2. How to run the harness (your existing step-by-step)
3. Reading tuning outputs (`analysis/results/tuning_runs/...`)
4. Using AI optimizer results
5. Guardrail checks & best practices (link to `docs/simulations_and_guardrails.md`)

README should then only say: *“For tuning the algorithm, see docs/tuning_playbook.md.”*

---

### `docs/simulations_and_guardrails.md` – merge bespoke simulations + guardrails

Combine `bespoke_simulations.md` with the guardrails.yml explanation parts from `rules-runbook.md`/README.

**Sections:**

1. Why we simulate (business language + link to `docs/concepts_and_metrics.md`)
2. How to build a fixture (existing `bespoke_simulations` content). 
3. How to run simulations (`run_simulation.py`, batch simulations).
4. Guardrails.yml semantics & workflow (copied from rules-runbook). 
5. How to read simulation outputs (quality metrics, scenario summaries, seed manifests).

This becomes the one-stop shop for “are we safe to ship?” from an engineering perspective.

---

### `docs/rules_runbook.md` – keep focused on ops

Keep 80–90% of what you have (it’s good), but:

* Remove/trim the guardrails.yml details (move to `docs/simulations_and_guardrails.md`). 
* Keep focus on:

  * Rule precedence refresher
  * Telemetry & Prometheus counters
  * Day-to-day troubleshooting and operational checklists

This is the doc an on-call person grabs.

---

### `docs/business_overview.md` – new executive/PM overview

**Audience:** PMs, execs, pre-sales; basically “no code, just outcomes”.

**Sections:**

1. What Recsys does in business terms
2. Typical rollout story (week 1 vs week 4 vs week 8)
3. What guardrails guarantee (no dead ends, no spammy content, fairness/coverage) in plain language
4. Evidence & auditability story:

   * Simulations, guardrails, decision traces, Prometheus metrics, `/version` etc.
5. How to read the simulation/guardrail reports at a business level (high-level, link to `docs/simulations_and_guardrails.md`)

This is where you sell “we are safe and explainable”.

---

## 3. Migration plan (what moves where)

Here’s how to get from current state to that structure with minimal thrash:

1. **Create new skeleton docs first**

   * Add empty `GETTING_STARTED.md`, `docs/concepts_and_metrics.md`, `docs/quickstart_http.md`, `docs/tuning_playbook.md`, `docs/simulations_and_guardrails.md`, `docs/business_overview.md`.

2. **Refactor README.md**

   * Cut/paste:

     * Move **Tuning Workflow** and **AI optimizer** commands into `docs/tuning_playbook.md`.
     * Move detailed **Onboarding & Coverage checklist** + **Cold-start guardrails details** into `docs/simulations_and_guardrails.md`.
     * Move the **glossary** to `docs/concepts_and_metrics.md`.
   * Replace them with short 1–2 line summaries + links.

3. **Canonicalize env vars**

   * Rename `env_reference.md` → `env_reference.md`.
   * In `CONFIGURATION.md` and README, replace any env tables with:

     * A short explanation + link: *“See docs/env_reference.md for full env var list.”*

4. **Merge simulations & guardrails**

   * Copy the entire content of `bespoke_simulations.md` into `docs/simulations_and_guardrails.md` under appropriate headings. 
   * Copy the `guardrails.yml` section from `rules-runbook.md` and guardrail bits from README here.
   * Leave behind links in `rules_runbook.md` and README.

5. **Split getting started vs tuning**

   * Build `GETTING_STARTED.md` out of:

     * The simpler “Reset → Seed → Run scenarios” bits in README & overview.
   * Keep `docs/tuning_playbook.md` purely about advanced harness + AI optimizer.

6. **Rename `api_reference.md` → `api_reference.md`**

   * Update links in README, overview, CONFIGURATION, etc.

7. **Build the concepts & metrics doc by centralizing definitions**

   * Scrape all small explanations of ALS, MMR, coverage, segment lifts, S7, etc. from: README, CONFIGURATION, env docs, bespoke simulations.
   * Consolidate into `docs/concepts_and_metrics.md` and strip duplicated conceptual explanations from other docs, replacing them with links.

---

If you execute that, you end up with:

* **One obvious path for a newcomer**: `README → GETTING_STARTED → quickstart_http` / `api_reference`.
* **One place to understand ideas**: `docs/concepts_and_metrics.md`.
* **One place for env knobs**: `docs/env_reference.md`.
* **One place for simulations/guardrails**: `docs/simulations_and_guardrails.md`.

Everything else gets lighter, more focused, and easier to maintain.

If you want, I can next draft the actual text skeletons (headings + first paragraphs) for a couple of the new files so you can just fill in the blanks instead of starting from zero.
