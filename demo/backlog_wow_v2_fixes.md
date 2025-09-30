Here’s a tough-love review + a crisp backlog aimed at the “exec demo” service (separate from the dev demo). I’ll be opinionated and concrete, with code-aware notes and acceptance tests.
What’s already working (nice!)

    One-click start: App.tsx creates a random namespace, seeds minimal data, and routes to the stage. Great “start here” flow. 

The stage: clear persona/surface/K, concise info text, “Explain this page,” “Decide + Recommend,” merchandising pinch-test (Pin/Block), and an adaptation modal. This already feels like a guided story.

Before/After: the diff overlay communicates adaptation well.

Bandit plumbing: generated client exposes decide/reward and “recommendations with bandit.” You’re wired for a winner banner and uplift copy.
Gaps that blunt the “wow”

    Mobile: the split layout won’t hold on phones; controls crowd the stage and the “Watch it adapt” button is easy to miss while scrolling. (Your HTML has the viewport meta, so you’re ready to go.) 

“Why this list” reads like engineer-speak: currently a list of item reasons with minimal executive framing. You do have an LLM explain endpoint and an Explain view in the other UI—neither is yet leveraged here.

Bandit outcome is hard to demo: no reliable way to steer the result for storytelling; you do have a reward endpoint to pre-prime a winner.

Aesthetic restraint: inline styles everywhere; spacing/contrast is close but not “tasteful default.” (The admin UI has a tokenized system you can mimic.)
P0 — Ship for the “3-minute wow” (exec demo service only)
FIX-EDS-06 · Mobile-first stage (hero list above the fold)

What

    Collapse left rail into a “Controls” sheet on ≤768px. Keep persona/surface/K; move “Apply” into the sheet footer; show a single compact bar above the list with the chosen persona/surface.

    Make the “Decide + Recommend” and “Watch it adapt” CTAs sticky at the top of the list (on mobile), and persistent in a right header cluster on desktop.

Acceptance

    On an iPhone-width viewport: first screen shows “Top-K” list with winner chip, and two sticky CTAs; controls are one tap away.

    No horizontal scroll; tap targets ≥44px.

Notes
You already have the stage split and modal plumbing in Stage.tsx; add a useIsMobile and conditional layout.
FIX-EDS-07 · Executive-grade “Why this list”

What

    Replace the current modal body with a two-part narrative:

        One-sentence headline (“Popular now + Balanced brands”).

        3 bullets: What we optimized (relevance/diversity), What mattered most (signals), Business guardrails honored (caps/rules).

    Under a “Details ▾” fold, show concise line-items per result.

    Option A (deterministic): extend recommend response with a compact structured explain (signal contributions, MMR lambda, cap hits, overlap).

    Option B (LLM): call /v1/explain/llm with target_type:"surface", surface, namespace, from/to “last 7 days”. Use the Summary / Key findings / Suggested fix sections you already standardized in the other app. Fallback to deterministic copy on timeout.

Acceptance

    Execs can read one sentence + three bullets and grasp why these items, without opening “Details.”

    LLM disabled? Modal still renders a great deterministic narrative.

Why this is easy

    You’ve documented exactly the numeric explain to add (pop/co-vis/emb contributions, overlap, MMR, caps) and the UI pattern to render it. Implement the “numeric”/“full” explain_level suggested in your notes and prefer that over tags.

FIX-EDS-08 · Rig-the-bandit (storytelling switch)

What
Add a small “Make Baseline win / Make Diverse win” dev-toggle (hidden behind “⋯”) that pre-primes rewards using the existing endpoint before you call decide. Implementation: call /v1/bandit/reward N times with reward=1 for the intended winner and a couple of 0s for the other policy in the same (namespace,surface,bucket); then run decide + recommendations.

Acceptance

    Toggling to “Diverse” -> “Winner: Diverse” ≥95% of runs for the demo namespace & surface.

    Resetting the demo (new namespace) clears the bias.

FIX-EDS-09 · Adaptation modal usability polish

What

    Move Replay into the modal header (top-right), keep keyboard focus on it after close to enable “repeat while scrolling.”

    Add a one-line explainer at the top: “We simulated view → view → click → add → purchase. The right column shows the list seconds later.” You already compute that sequence in buildSimulatedWeekEvents. 

Acceptance

    Without scrolling, an exec can hit Replay repeatedly.

    ESC to close; focus returns to the triggering button.

P1 — Taste, restraint, confidence
FIX-EDS-10 · Minimal design system for the exec demo

What

    Introduce tokens.css (spacing, radii, typography scale, subtle shadows) and replace inline styles in App.tsx, Stage.tsx, TopKCard.tsx, Modal.tsx.

    Prefer className over inline styles for consistency.

Acceptance

    Lighthouse “Best Practices” & “Accessibility” ≥ 95 (mobile).

    No inline style blocks in these four files. 

FIX-EDS-11 · Always-on share & reset

What

    Add “Copy demo link” (current querystring) and “Reset demo” to the header cluster. Reset = new namespace + reseed via existing seeding path. You already stash itemMeta in localStorage per namespace—clear that too. 

Acceptance

    Fresh namespace, fresh seed, same one-click route to the stage.

FIX-EDS-12 · Polished empty/loading states

What

    Skeleton rows for Top-K; copy like “Fetching items popular in the last 30 days…”

    Friendly error card with “Reset demo.”

Acceptance

    No spinner walls; each state narrates what’s happening.

P2 — Feels productized
FIX-EDS-13 · “Executive handout” (PDF)

What

    One-pager: headline, top-K with badges, winner policy + uplift, persona/surface, and your “Why this list” headline. Export from the browser (no server) using print CSS.

Acceptance

    Prints to one A4/Letter page with margins and your brand.

FIX-EDS-14 · Keyboard + screen reader polish

What

    Focus traps in modals, labelled buttons, ARIA live region for winner change (bandit).

Acceptance

    Tabbing cycles correctly; VoiceOver/JAWS announces “Winner: Diverse, +6% uplift this session.”

Code review notes (targeted)
Web (exec demo)

    Typed clients vs fetch: Stage.tsx relies on generated clients elsewhere; seedMinimal still uses raw fetch. Move to your generated IngestionService to standardize error handling and CORS config. 

Local state + URL: Great that you param-sync via useSearchParams. Add input validation (e.g., clamp k to sane values) before sending to the API.

Storage hygiene: you cache demo:itemMeta:${ns}; clear it on “Reset demo.” (You already have rich storage patterns in the other UI—consider a tiny wrapper here.)

Accessibility: ensure modal close buttons have aria-label, and that focus returns to the triggering control after modal close in AdaptDiffOverlay.
API

    LLM explain endpoint is well-validated (RFC3339 window, required fields). Expose X-Explain-Cache (already set) in CORS so the UI can surface cache hits. 

Event types list: great convenience endpoint; surface it in the exec demo’s “Explain this page → Details” so the weights are visible without opening Swagger.

Bandit priming: Use your existing /v1/bandit/reward to rig outcomes in demo mode (P0). Keep the code path identical to real flow—only the pre-rewards differ.
Concrete fixes for your current TODOs

    “Cluttered, must be 100% mobile friendly” → FIX-EDS-06 implements a mobile sheet for controls, sticky CTAs, and no horizontal scroll.

    “Why this list is not useful to execs” → FIX-EDS-07 gives a one-sentence headline + 3 bullets; plug in /v1/explain/llm for narrative (with deterministic fallback). 

“Replay button should be on top” → FIX-EDS-09 moves Replay to the modal header and preserves focus for quick repeats.

“Any way to affect bandit decision?” → FIX-EDS-08 adds a hidden “Make X win” switch that pre-rewards via /v1/bandit/reward before decide.
Micro-impl notes (drop-in)

    Rig bandit (demo-only)

// call before decide
async function primeBanditWinner(ns: string, surface: string, winner: "baseline"|"diverse") {
  const loser = winner === "baseline" ? "diverse" : "baseline";
  const bucket = "exec-demo"; // constant so rewards aggregate
  // 6 wins for winner, 1 loss for loser — tweak as needed
  for (let i=0;i<6;i++) await BanditService.postV1BanditReward({namespace: ns, surface, policy_id: winner, reward: 1, bucket_key: bucket, request_id: `prime-${winner}-${i}`});
  for (let i=0;i<2;i++) await BanditService.postV1BanditReward({namespace: ns, surface, policy_id: loser,  reward: 0, bucket_key: bucket, request_id: `prime-${loser}-${i}`});
}

API shape is already exposed in your generated client.

    Explain (LLM) from the stage

// minimal payload for surface-level narrative
await ExplainService.postV1ExplainLlm({
  namespace: ns, target_type: "surface", target_id: surface,
  surface, from: new Date(Date.now()-7*864e5).toISOString(), to: new Date().toISOString()
});

The handler already returns {markdown, model, cache, warnings} with X-Explain-Cache header.
TL;DR plan to “wow” an exec

    Open: giant “Start the demo” → seeded list (done). 

Read one sentence (FIX-EDS-07): “Popular now · Balanced brands” with 3 bullets.

Press “Watch it adapt” (FIX-EDS-09): before/after pops, replay stays on top.

Press “Decide + Recommend” (FIX-EDS-08): show winner chip; if needed, flip the hidden “Make Diverse win” to tell the story.
This keeps the demo one glance, one click, with copy and motion that sell the idea without exposing knobs.