# Exec Demo Service — Usability‑First Backlog (RecSys)

**Vision**: A separate, ultra‑polished "Exec Demo" web app that delivers a one‑glance, one‑click WOW. It must tell a simple story in plain English, prove adaptivity in seconds, and be effortless to demo live. Tasteful defaults, minimal controls, zero dead ends.

---

## P0 — The 3‑minute WOW Path (Must‑ship)

# TODO: Cluttered, must be 100% mobile friendly

✅ **EDS‑00 · Create the new service scaffold**

* **What**: New `exec-demo` web app (Vite + React) in the monorepo; separate Docker service and route (e.g., `https://exec.local`). Minimal dependencies, same OpenAPI client.
* **Why**: Isolate the exec experience, keep demo‑ui as a playground.
* **Impl notes**: Add `packages/api-client` workspace (shared OpenAPI client). New `docker-compose` service + Caddy route. `.env` with `VITE_API_BASE_URL`.
* **Acceptance**: `make dev` spins up existing services + `exec-demo`; opening `exec.local` shows the hero.

✅ **EDS‑01 · Single‑button Hero: “Start the demo”**

* **What**: One primary CTA that (a) seeds a namespace, (b) autoselects persona+surface, (c) routes to the Stage.
* **Why**: One click to delight.
* **Impl notes**: Use shared seeding adapter; seed users/items/events + baseline segments; persist namespace to querystring.
* **Acceptance**: One click → populated recommendations with human reasons.

✅ **EDS‑02 · The Stage (split layout)**

* **What**: Left = compact controls (Persona, Surface, K). Right = large, elegant Top‑K cards with 3 badges each (e.g., Popular now · Similar to X · Diverse brands). Subheader shows a 1‑sentence explanation.
* **Why**: Executives focus on the result; controls are there but quiet.
* **Impl notes**: Reuse tokens; build `TopKCard`, `BadgesRow`, `HumanSummary` components. Keep advanced tucked away.
* **Acceptance**: Clean stage, scroll‑free on laptop, obvious “what I’m seeing.”

# TODO: "Why this list" explanation is not very useful or informative (from the pov of "executive")

✅ **EDS‑03 · “Explain this page” (plain English)**

* **What**: Button above the list opens a concise modal describing why these items ranked (no jargon). Link to details.
* **Why**: Tell first, numbers on demand.
* **Impl notes**: Summarize reason codes into 5 badge types + 1 paragraph. Details panel can show raw reasons.
* **Acceptance**: A readable 2–3 sentence summary that a non‑technical VP understands.

✅ **EDS‑04 · One‑tap Before/After (personalization moment)**

* **What**: Button runs a 5‑event “week” (view→view→click→add→purchase) and animates a side‑by‑side diff.
* **Why**: Visually proves adaptivity.
* **Impl notes**: Simulate events for the selected persona; recompute recs; animate changed ranks with subtle transitions. Provide “Replay for another persona.”
* **Acceptance**: Press once → clear, animated delta of the list.

TODO: "Watch it adapt" popup "replay button should be on top to allow constant scrolling down

✅ **EDS-05 · Bandit “chooses a winner” in 10 seconds**

* **What**: Preload two policies (Baseline vs Diverse). Single CTA: “Decide + Recommend.” Show winner chip + tiny “uplift this session”.
* **Why**: Demonstrate online learning without knobs.
* **Impl notes**: One‑shot endpoint (decide+rank) + mock reward. Keep copy to one sentence.
* **Acceptance**: One press, visibly different list, winner labeled.

TODO: Any way to affect bandit decision in demo (Similar as "watch it adapt")?

✅ **EDS-06 · Rule engine pinch-test**

* **What**: Tiny “Merchandising” inline demo: click “Pin this item to #1” or “Block brand X” → instant re‑rank snapshot.
* **Why**: Show business control without scary UI.
* **Impl notes**: Preload sample rules; use dry‑run or temporary rule for current surface; revert on close.
* **Acceptance**: One click shows rule effect; second click restores.

**EDS‑07 · Reset & Share**

* **What**: “Reset Demo” (reseed) + “Copy Link to this story” + back button to main page.
* **Why**: Zero dead ends; reproducible demo.
* **Impl notes**: Namespace in URL; reseed clears and reloads; copy writes a short URL (optional).
* **Acceptance**: Fresh start in one click; link opens same state on another machine.

**EDS‑08 · Error/Loading as narrative**

* **What**: Skeletons and helpful copy (“Fetching items trending in the last 30 days…”). Friendly error with Reset action.
* **Why**: Keep the vibe premium under failure.
* **Impl notes**: Global error boundary; toasts for success; consistent language.
* **Acceptance**: No raw stack traces; all states feel intentional.

**EDS‑09 · Taste kit**

* **What**: Type scale, whitespace, soft motion (stage only). Color system, shadows, spacing tokens.
* **Why**: Premium feel.
* **Impl notes**: `ui/tokens` in the new app; card elevation, rounded‑2xl, tight rhythm; 60fps animations.
* **Acceptance**: Visual polish passes “screenshare at 150% zoom.”

---

## P1 — Productized Polish (Should‑ship)

**EDS‑10 · Curated Personas**

* Ship 3: VIP, Casual, New. One‑click persona switch that re‑runs recs and adjusts the human summary.

**EDS‑11 · Story Mode (guided 60‑sec tour)**

* 3 stops: (1) See your list, (2) Why it’s ranked, (3) Watch it adapt. Overlay coach‑marks; skippable.

**EDS‑12 · PDF “Leave‑behind”**

* Export one page: pipeline diagram + current settings + Top‑K with reasons. Give it after the meeting.

**EDS‑13 · Keyboard shortcuts**

* `P` persona cycle, `S` surface, `B` bandit, `R` reset, `?` help.

**EDS‑14 · Deep links**

* Everything stateful encoded in URL (namespace, persona, surface, k, policy). Open on another device and it recreates the scene.

**EDS‑15 · Kiosk Mode**

* Auto‑seed on load, hide controls, loop Story Mode; ideal for expo booths.

---

## P2 — Demo Resilience & Delight (Nice‑to‑ship)

**EDS‑16 · Offline/Flaky‑net grace**

* Cache last results; show “snapshot mode” if API unreachable; let Reset attempt reconnection.

**EDS‑17 · Micro‑metrics overlay**

* Tiny hint of impact: “Diverse policy lifted CTR +6% (sampled).” Keep it obviously illustrative.

**EDS‑18 · Theme switch (Light/Dim)**

* Subtle dark theme for theatre rooms.

**EDS‑19 · Sample domains switcher**

* Toggle catalog flavor: Retail · Content · Lottery. Same mechanics, different copy/items.

**EDS‑20 · Presenter Mode**

* Hide cursor prompts, reduce animations for screen‑share; single‑key progression of the Story.

---

## Architecture & Repo Tasks

**EDS‑A1 · Monorepo wiring**

* Add `packages/api-client` (generated), import from both UIs. Add `exec-demo/` with its own vite config, tokens, and build.

**EDS‑A2 · Docker & Proxy**

* New service in `docker-compose.yml`; Caddy route `exec.local` (TLS via mkcert in dev). Healthcheck endpoint.

**EDS‑A3 · Config schema**

* Mirror env validation in `exec-demo` (`VITE_API_BASE_URL`, `VITE_SWAGGER_UI_URL`, logging, analytics flag).

**EDS‑A4 · Shared utilities**

* Extract `useAsync`, `useSafeQuerySync`, storage helpers into `packages/ui-utils` if reuse is meaningful; otherwise copy minimal code.

**EDS‑A5 · CI**

* Lint, typecheck, build `exec-demo`; upload Playwright screenshots on PR.

---

## Copy Deck (sprinkle across UI)

* **Hero**: “One click. A living, explainable list.”
* **Stage subtitle**: “Blended relevance with guardrails.”
* **Explain modal**: “Heavy recent demand, similar to what this persona likes, and balanced brands so you’re not stuck with one label.”
* **Before/After button**: “Watch it adapt.”
* **Bandit**: “We try two strategies and shift traffic to the winner—automatically.”
* **Reset**: “Start fresh. Same magic.”

---

## Acceptance Criteria (end‑to‑end)

1. From a cold open, **one click** seeds and shows a populated Top‑K with a one‑sentence human explanation.
2. Press **Watch it adapt** → a clear, animated before/after diff.
3. Press **Decide + Recommend** → list changes and the winning policy is labeled.
4. Click **Pin to #1** on an item → immediate re‑rank snapshot, then revert.
5. **Copy link** opens the same state on another machine.

---

## Nice Implementation Hints

* Keep all numbers behind a “Details” drawer.
* Animations: 150–250ms ease‑out; only on rank changes and coach‑marks.
* Cards: generous white space, bold title, 3 badges, reason snippet; nothing else.
* Mobile: single column, sticky control bar; no horizontal scroll.
