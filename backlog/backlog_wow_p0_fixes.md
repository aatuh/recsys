# backlog_wow_p0_fixes.md

P0 fixes — what to change and why
1) Recommendations Playground
A. Persona/surface change doesn’t fetch (stale list shown)

Why it happens

    Persona buttons only set recUserId; there’s no fetch after the state change, and the old recOut stays on screen. (Persona setter exists; fetch isn’t tied to it.) 

The surface you select is not sent to the /recommendations API, so changing it can’t change results. (Calls pass {namespace,user_id,k,blend}; no surface/context.)

Fix

    Refactor a “single source of truth” fetch and call it from persona/surface changes. Also pass surface in context so rules/bandit-style scenarios can differ by placement.

    Clear the prior list immediately (optimistic) and guard against out-of-order responses.

// in RecommendationsSection.tsx
const [inflightKey, setInflightKey] = useState<string | null>(null);
const [autoUpdate] = useState(true); // default ON for demo wow

const runRecommendFor = useCallback(async (uid: string) => {
  const key = `${namespace}|${uid}|${selectedSurface.id}|${k}|${blend.pop}-${blend.cooc}-${blend.als}`;
  setInflightKey(key);
  setRecLoading(true);
  setRecOut([]); // clear stale results immediately

  const res = await recommend({
    namespace,
    user_id: uid,
    k,
    blend,
    // make surface meaningful for rules/bucketed logic:
    context: { surface: selectedSurface.id },
    include_reasons: true,
  });

  if (inflightKey && inflightKey !== key) return; // stale response
  setRecOut(res.items ?? []);
  setRecLoading(false);
}, [namespace, selectedSurface.id, k, blend, inflightKey]);

// Persona change -> set and fetch *immediately*
const onChoosePersona = (preset: Persona) => {
  setRecUserId(preset.userId);
  void runRecommendFor(preset.userId);
};

// Optional: auto‐update when surface changes
useEffect(() => {
  if (!autoUpdate) return;
  const uid = recUserId || exampleUser;
  void runRecommendFor(uid);
}, [selectedSurface.id]);

B. “After clicking Show recommendations the list is the same”

Why it happens

    If your seed produces weakly differentiated users or you don’t pass surface, the same list is plausible. (The call currently ignores surface.) 

Fix

    Pass context.surface as above.

    Also, make persona presets shape the request so they feel different instantly (no data-science rabbit hole): set per-persona overrides right before calling recommend (e.g., VIP → tighter caps; Trend Seeker → higher novelty). You already compute blends client-side elsewhere (bandit presets). Reuse that approach here for persona presets.

const personaOverrides: Record<string, Partial<typeof blend>> = {
  vip:   { pop: 1.0, cooc: 0.6, als: 0.1 },
  trend: { pop: 0.5, cooc: 0.6, als: 0.4 },
  new:   { pop: 0.9, cooc: 0.3, als: 0.0 },
};

// before calling recommend:
const effBlend = { ...blend, ...(personaOverrides[selectedPersona.id] ?? {}) };

C. Not obvious which user is selected

Fix

    Add a small “Selected user” pill right under the persona row, always visible (no need to open Advanced).

// near persona chips
<div style={{display:"flex",gap:8,alignItems:"center"}}>
  <span className="badge">User: {recUserId || exampleUser}</span>
  <button onClick={()=>navigator.clipboard.writeText(recUserId || exampleUser)}>
    Copy ID
  </button>
</div>

2) User Session Simulator
A. Too much scrolling to see results after simulation

Fix the layout

    Convert to a two-pane “workbench”: left = controls & logs, right = live recommendations (reuse the same Recommendations section component). Sticky right column on desktop, stacked on mobile.

// in UserSessionView.tsx layout wrapper
<div style={{
  display:"grid",
  gridTemplateColumns:"minmax(340px, 380px) 1fr",
  gap:24
}}>
  <LeftPane />   {/* Simulate Week, Session controls, Logs */}
  <RightPane>   {/* <RecommendationsSection compact /> */}
    <RecommendationsSection ... />
  </RightPane>
</div>

Wire it so it actually updates

    After batchEvents (your simulator), call the same runRecommendFor(currentUser) you use in the playground and render the list on the right. (You already call recommend in this view; reuse and write to the shared recommendationsPlayground state so both views stay in sync.) 

B. “View recommendations” button does nothing

Why it happens

    The button sets the view query param, which should flip the view, but that coupling is fragile across duplicates of App/Nav. Safer to both set the param and programmatically change the view/state. In Bandit we already set the param; replicate + add a fallback. 

Fix

// onClick handler
setActiveViewParam("recommendations-playground" as ViewType);
requestAnimationFrame(() => {
  // hard fallback if some builds don't listen to the URL:
  document.getElementById("recommendations-stage")
    ?.scrollIntoView({ behavior: "smooth" });
});

If you adopt the two-pane layout, just remove this button.
3) Bandit
A. “Decide + Recommend shows the same result”

Why it happens

    The demo locally compares two client-side presets by calling /recommendations twice with different blends and picking the higher average top-5 score. If the top lists happen to match, it looks “unchanged.”

Make the outcome visibly diverge

    Strengthen the presets so lists differ:

const policyPresets = [
  { id:"baseline", name:"Baseline", blend:{ pop:1.0, cooc:0.3, als:0.0 }},
  { id:"diverse",  name:"Diverse Explorer", blend:{ pop:0.4, cooc:0.7, als:0.5 }},
];

    Use surface (as in the playground fix) for different business caps via rules, or simply add synthetic caps in the request overrides for the “diverse” policy (lower brand cap / higher MMR lambda) to force a list shift.

    In the UI, keep showing the diff block (you already compute a compact diff; keep it prominent). 

B. “What are Baseline & Diverse Explorer? Are they saved?”

Why it happens

    They’re client-side presets (no DB save). 

Fix — make it concrete

    Add a “Save demo policies to service” button that upserts them via /v1/bandit/policies:upsert, and show a toast with the two policy IDs. Your docs already describe the payload.

await banditPoliciesUpsert({
  namespace,
  policies: availablePolicies.map(p => ({
    policy_id: p.policy_id, name: p.name, active: true,
    blend_alpha: p.blend_alpha, blend_beta: p.blend_beta,
    blend_gamma: p.blend_gamma, mmr_lambda: p.mmr_lambda ?? 0.6,
    brand_cap: p.brand_cap ?? 1, category_cap: p.category_cap ?? 2
  }))
});

C. “View recommendations” does nothing

    Keep the current behavior (update the recs state and switch the view). You already:

        write the winner’s items into the playground state (setRecommendationsPlayground) and

        set the view param to jump to the Playground.
        That’s correct; keep both to be robust. 

Small but high-leverage UX tweaks (exec-friendly)

    Status line above the list: “Showing VIP Loyalist on Homepage Hero · User user-0001 · K=20”. Make it a compact, always-visible breadcrumb so nobody needs to open Advanced.

    One big button: Rename to “Show recommendations” everywhere. Disable it while loading; show a skeleton grid.

    Reason chips, first 3 only: Keep reasons terse by default; “Show more” expands. (You already have the diff/reasons components.)

    Copy link in the top nav is great — include the current persona/surface in the URL so “what you see is what you share.” (The nav already sets view/namespace; include persona & surface too.) 

Why this hits the “wow” bar

    Zero dead ends. Every control either updates the list or tells you why it can’t. No stale UI.

    One mental model. Persona = user ID + intent; Surface = placement; Policy = tuning bundle. You can see and feel each.

    Instant differentiation. Personas and policies visibly change the list without explaining algorithms.

