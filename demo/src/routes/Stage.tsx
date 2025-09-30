import {
  forwardRef,
  useCallback,
  useEffect,
  useImperativeHandle,
  useMemo,
  useRef,
  useState,
} from "react";
import { useSearchParams } from "react-router-dom";
import TopKCard from "../components/TopKCard";
import type { ScoredItem } from "../types/recommendations";
import {
  BanditService,
  IngestionService,
  RankingService,
  RuleService,
  ExplainService,
} from "../lib/api-client";
import type {
  types_BanditPolicy as BanditPolicy,
  RuleResponse,
} from "../lib/api-client";
import { ensureApiBase } from "../lib/api";

const SIGNAL_COPY: Record<string, string> = {
  recent_popularity: "Popular now",
  co_visitation: "Viewed together",
  embedding: "Similar items",
  personalization: "Personalized fit",
  diversity: "Balanced brands",
};

type ItemMeta = {
  title?: string;
  brand?: string;
  price?: number;
  tags?: string[];
  available?: boolean;
};

type QueryResult = {
  items: ScoredItem[];
  summary: string;
  order: string[];
  meta: Record<string, ItemMeta>;
};

type RawRecommendationItem = {
  item_id?: string;
  itemId?: string;
  id?: string; // be lenient: some serializers use "id"
  score?: number;
  reasons?: string[];
};

type RulePreview =
  | { mode: "none" }
  | { mode: "pin"; itemId: string; ruleId?: string }
  | { mode: "block"; brand: string; ruleId?: string };

type ControlsHandle = {
  apply: () => void;
};

export default function Stage() {
  const [params, setParams] = useSearchParams();
  const namespace = params.get("ns") || "default";
  const userId = params.get("u") || "user-1";
  const rawK = Number(params.get("k") || 10);
  const surface = params.get("s") || "home_top";

  const k = useMemo(() => clampTopK(rawK), [rawK]);
  const personaLabel = useMemo(() => userIdToPersona(userId), [userId]);

  const [items, setItems] = useState<ScoredItem[]>([]);

  // Keep ref in sync with items
  useEffect(() => {
    itemsRef.current = items;
  }, [items]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [adaptationResults, setAdaptationResults] = useState<{
    before: ScoredItem[];
    after: ScoredItem[];
    persona: string;
  } | null>(null);
  const [previousPositions, setPreviousPositions] = useState<
    Record<string, number>
  >({});
  const itemsRef = useRef<ScoredItem[]>([]);
  const [adaptBusy, setAdaptBusy] = useState(false);
  const [itemMeta, setItemMeta] = useState<Record<string, ItemMeta>>({});
  const [rulePreview, setRulePreview] = useState<RulePreview>({ mode: "none" });
  const [policies, setPolicies] = useState<BanditPolicy[]>([]);
  const [decideBusy, setDecideBusy] = useState(false);
  const [winnerPolicyId, setWinnerPolicyId] = useState<string | null>(null);
  const [policiesLoading, setPoliciesLoading] = useState<boolean>(true);
  const autoLoadCancelRef = useRef<null | (() => void)>(null);
  const banditActiveRef = useRef<boolean>(false);
  const ignorePlainUntilRef = useRef<number>(0);
  const applySeqRef = useRef<number>(0);
  const [listEpoch, setListEpoch] = useState<number>(0);
  const [banditItems, setBanditItems] = useState<ScoredItem[] | null>(null);
  const [listMode, setListMode] = useState<"auto" | "bandit">("auto");
  const [refreshTrigger, setRefreshTrigger] = useState<number>(0);
  const [forceRefresh, setForceRefresh] = useState<boolean>(false);
  const [activeRules, setActiveRules] = useState<RuleResponse[]>([]);
  const [ruleBusy, setRuleBusy] = useState(false);
  const [showControlsSection, setShowControlsSection] = useState(false);
  const [showAdaptSection, setShowAdaptSection] = useState(false);
  const [showMerchSection, setShowMerchSection] = useState(false);
  const [showBanditSection, setShowBanditSection] = useState(false);
  const [showExplainSection, setShowExplainSection] = useState(false);
  const [explainLoading, setExplainLoading] = useState(false);
  const [explainError, setExplainError] = useState<string | null>(null);
  const [explainMarkdown, setExplainMarkdown] = useState<string>("");
  const controlsRef = useRef<ControlsHandle | null>(null);
  const lastOrderRef = useRef<string[]>([]);
  const pendingReplayRef = useRef(false);

  // Form state for enhanced summary
  const [formUser, setFormUser] = useState<string>(userId);
  const [formSurface, setFormSurface] = useState<string>(surface);
  const [formK, setFormK] = useState<string>(String(k));
  const [formPersona, setFormPersona] = useState<string>(
    userIdToPersona(userId)
  );

  // Update form state when URL parameters change
  useEffect(() => {
    setFormUser(userId);
    setFormSurface(surface);
    setFormK(String(k));
    setFormPersona(userIdToPersona(userId));
    try {
      localStorage.setItem("demo:lastUser", userId);
      localStorage.setItem("demo:ns", namespace);
    } catch {
      // ignore storage errors in demo context
    }
  }, [userId, surface, k, namespace]);

  // Enhanced summary that includes context about the recommendations
  const enhancedSummary = useMemo(() => {
    // Use form state for immediate updates, fallback to URL params
    const currentPersona = formPersona || personaLabel;
    const currentSurface = formatSurfaceLabel(formSurface || surface);

    const parts: string[] = [];

    // Base context - use form state for immediate updates
    parts.push(`${currentPersona} persona · ${currentSurface}`);

    // Add bandit context if active
    if (listMode === "bandit" && winnerPolicyId) {
      const policyName = policyNameFromId(policies, winnerPolicyId);
      parts.push(`Bandit: ${policyName}`);
    }

    // Add rules context if any active rules
    if (activeRules.length > 0) {
      const ruleTypes = activeRules.map((rule) => {
        if (rule.action === "PIN") return "Pin rules";
        if (rule.action === "BLOCK") return "Block rules";
        return "Custom rules";
      });
      const uniqueTypes = [...new Set(ruleTypes)];
      parts.push(`${uniqueTypes.join(", ")} active`);
    }

    // Add adaptation context if in adaptation mode
    if (adaptationResults) {
      parts.push("Adapted for persona");
    }

    return parts.join(" · ");
  }, [
    formPersona,
    formSurface,
    personaLabel,
    surface,
    listMode,
    winnerPolicyId,
    policies,
    activeRules,
    adaptationResults,
  ]);

  useEffect(() => {
    if (Number.isNaN(rawK)) return;
    if (rawK === k) return;
    params.set("k", String(k));
    setParams(params, { replace: true });
  }, [rawK, k, params, setParams]);

  const apiBase = useMemo(() => {
    return import.meta.env.VITE_API_BASE_URL || "http://localhost:8081";
  }, []);

  useEffect(() => {
    ensureApiBase(apiBase);
  }, [apiBase]);

  // Clear bandit items when core params change to avoid stale lists
  useEffect(() => {
    setBanditItems(null);
    setWinnerPolicyId(null);
    setListMode("auto");
  }, [namespace, userId, surface, k]);
  // Load bandit policies for the namespace
  useEffect(() => {
    let cancelled = false;
    setPoliciesLoading(true);
    BanditService.getV1BanditPolicies(namespace)
      .then((list) => {
        if (cancelled) return;
        const arr = Array.isArray(list) ? list : [];
        if (arr.length > 0) {
          setPolicies(arr);
          setPoliciesLoading(false);
          return;
        }
        // Fallback to local demo hints if API returns empty
        const hinted = readLocalPolicyHints(namespace);
        if (hinted.length) {
          setPolicies(hinted);
        } else {
          setPolicies([]);
        }
        setPoliciesLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        const hinted = readLocalPolicyHints(namespace);
        if (hinted.length) {
          setPolicies(hinted);
        } else {
          setPolicies([]);
        }
        setPoliciesLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [namespace]);

  // Load active rules for the namespace and surface
  useEffect(() => {
    let cancelled = false;
    RuleService.getV1AdminRules({
      namespace,
      surface,
      enabled: true,
    })
      .then((response) => {
        if (cancelled) return;
        setActiveRules(response.rules || []);
      })
      .catch((error) => {
        if (cancelled) return;
        console.warn("Failed to load rules:", error);
        setActiveRules([]);
      });
    return () => {
      cancelled = true;
    };
  }, [namespace, surface]);

  useEffect(() => {
    return () => {};
  }, []);

  const merchPinCandidate = useMemo(() => {
    if (!items.length) return null;
    for (const preferredIndex of [2, 1, 0]) {
      const candidate = items[preferredIndex];
      if (candidate) return candidate;
    }
    return items[0] ?? null;
  }, [items]);

  const merchBlockBrand = useMemo(() => {
    const counts: Record<string, number> = {};
    for (const item of items) {
      const brand = itemMeta[item.id]?.brand;
      if (!brand) continue;
      counts[brand] = (counts[brand] || 0) + 1;
    }
    let winner: string | undefined;
    let winnerCount = 0;
    for (const [brand, count] of Object.entries(counts)) {
      if (count > winnerCount) {
        winner = brand;
        winnerCount = count;
      }
    }
    return winnerCount > 0 ? winner : undefined;
  }, [items, itemMeta]);

  const showSkeleton = loading && !items.length && !error;
  const showEmptyState =
    !loading &&
    (listMode === "bandit"
      ? !(banditItems && banditItems.length)
      : !items.length) &&
    !error &&
    !winnerPolicyId;
  const adaptLabel =
    adaptBusy || pendingReplayRef.current ? "Replaying…" : "Replay persona";

  useEffect(() => {
    if (
      rulePreview.mode === "pin" &&
      merchPinCandidate &&
      rulePreview.itemId !== merchPinCandidate.id
    ) {
      setRulePreview({ mode: "none" });
    }
    if (
      rulePreview.mode === "block" &&
      merchBlockBrand &&
      rulePreview.brand !== merchBlockBrand
    ) {
      setRulePreview({ mode: "none" });
    }
    if (rulePreview.mode !== "none" && !items.length) {
      setRulePreview({ mode: "none" });
    }
  }, [items, merchPinCandidate, merchBlockBrand, rulePreview]);

  const displayItems = useMemo(() => {
    // Prefer adaptation results. Otherwise, show bandit items when in bandit
    // mode, or plain items when in auto mode.
    const baseItems =
      adaptationResults?.after ||
      (listMode === "bandit" ? banditItems || [] : items);

    console.log("Display items calculation:", {
      listMode,
      hasBanditItems: !!banditItems,
      banditItemsLength: banditItems?.length,
      itemsLength: items.length,
      baseItemsLength: baseItems.length,
      baseItemsIds: baseItems.map((item) => item.id),
      activeRulesCount: activeRules.length,
    });

    // Rules are now applied by the backend, so we just return the items as-is
    return baseItems;
  }, [items, adaptationResults, banditItems, listMode, activeRules]);

  const queryRecommendations = useCallback(
    (opts?: { previousOrder?: string[] }) => {
      const request = RankingService.postV1Recommendations({
        namespace,
        user_id: userId,
        context: { surface },
        k,
        include_reasons: true,
      });
      return {
        promise: request.then((data) => {
          const payload = data as { items: RawRecommendationItem[] };
          if (!payload.items.length) {
            console.warn("Recommend response missing items", data);
          }
          const meta = safeReadItemMeta(namespace);
          const prevOrder = opts?.previousOrder ?? lastOrderRef.current;
          const normalized = normalizeItems(payload.items, prevOrder, meta);
          return {
            items: normalized.items,
            summary: humanizeReasons(normalized.items),
            order: normalized.order,
            meta,
          } satisfies QueryResult;
        }),
        cancel: () => request.cancel(),
      };
    },
    [namespace, userId, k, surface]
  );

  const applyRecommendations = useCallback((result: QueryResult) => {
    // Store current positions as previous before updating
    const currentPositions: Record<string, number> = {};
    itemsRef.current.forEach((item, idx) => {
      currentPositions[item.id] = idx;
    });
    console.log("Storing previous positions:", currentPositions);
    setPreviousPositions(currentPositions);

    setItems(result.items);
    lastOrderRef.current = result.order;
    setItemMeta(result.meta);
    setRulePreview({ mode: "none" });
    setListEpoch((v) => v + 1);
  }, []);
  const decideAndRecommend = useCallback(async () => {
    if (decideBusy) return;
    setDecideBusy(true);
    setLoading(true);
    setWinnerPolicyId(null);
    // Clear any adaptation snapshot so the list reflects fresh bandit results
    setAdaptationResults(null);
    banditActiveRef.current = true;
    ignorePlainUntilRef.current = Date.now() + 5000; // ignore plain results for 5s
    try {
      const applyId = ++applySeqRef.current;
      // Cancel any in-flight auto-load to avoid overwriting bandit results
      try {
        autoLoadCancelRef.current?.();
      } catch {
        /* ignore */
      }
      autoLoadCancelRef.current = null;
      let candidateIds = (Array.isArray(policies) ? policies : [])
        .filter((p) => p.active !== false)
        .map((p) => p.policy_id || "")
        .filter(Boolean);
      if (!candidateIds.length) {
        candidateIds = readLocalPolicyHints(namespace).map(
          (p) => p.policy_id || ""
        );
      }
      const resp = await RankingService.postV1BanditRecommendations({
        namespace,
        user_id: userId,
        surface,
        k,
        include_reasons: true,
        candidate_policy_ids: candidateIds.length ? candidateIds : undefined,
      });

      console.log("Bandit response received:", {
        fullResponse: resp,
        itemsField: resp.items,
        itemsLength: resp.items?.length,
        itemsType: typeof resp.items,
        isArray: Array.isArray(resp.items),
      });

      // Try to manually parse the response if the API client isn't working
      let rawItems = resp.items || [];
      // Helper to extract chosen policy id robustly
      const extractChosenPolicyId = (r: unknown): string | null => {
        if (!r) return null;
        try {
          const anyResp = r as Record<string, unknown>;
          const directSnake = anyResp["chosen_policy_id"];
          if (typeof directSnake === "string" && directSnake.trim()) {
            return directSnake.trim();
          }
          const directCamel = (anyResp as Record<string, unknown>)[
            "chosenPolicyId"
          ];
          if (typeof directCamel === "string" && directCamel.trim()) {
            return directCamel.trim();
          }
          // Some shapes may nest under `chosen`
          const chosen = anyResp["chosen"] as
            | Record<string, unknown>
            | undefined;
          if (chosen) {
            const idSnake = chosen["policy_id"];
            if (typeof idSnake === "string" && idSnake.trim()) {
              return idSnake.trim();
            }
            const idCamel = chosen["policyId"];
            if (typeof idCamel === "string" && idCamel.trim()) {
              return idCamel.trim();
            }
          }
        } catch {
          // ignore
        }
        return null;
      };

      // Helper to extract bucket key robustly
      const extractBucketKey = (r: unknown): string => {
        if (!r) return "";
        try {
          const anyResp = r as Record<string, unknown>;
          const snake = anyResp["bandit_bucket"];
          if (typeof snake === "string" && snake.trim()) return snake.trim();
          const camel = anyResp["banditBucket"];
          if (typeof camel === "string" && camel.trim()) return camel.trim();
        } catch {
          // ignore
        }
        return "";
      };

      const extractRequestId = (r: unknown): string => {
        if (!r) return "";
        try {
          const m = r as Record<string, unknown>;
          const snake = m["request_id"];
          if (typeof snake === "string" && snake.trim()) return snake.trim();
          const camel = m["requestId"];
          if (typeof camel === "string" && camel.trim()) return camel.trim();
        } catch {
          // ignore
        }
        return "";
      };

      const extractAlgorithm = (r: unknown): string | undefined => {
        if (!r) return undefined;
        try {
          const m = r as Record<string, unknown>;
          const algo = m["algorithm"];
          if (typeof algo === "string" && algo.trim()) return algo.trim();
        } catch {
          // ignore
        }
        return undefined;
      };

      if (!rawItems || rawItems.length === 0) {
        console.log(
          "API client didn't parse items correctly, trying manual parsing..."
        );
        // The response might be a string that needs to be parsed
        if (typeof resp === "string") {
          try {
            const parsed = JSON.parse(resp);
            rawItems = parsed.items || [];
            console.log("Manually parsed response:", { parsed, rawItems });
            // If we had to manually parse, also ensure we can read chosen id
            const manualChosen = extractChosenPolicyId(parsed);
            if (manualChosen) {
              setWinnerPolicyId(manualChosen);
            }
          } catch (e) {
            console.error("Failed to parse response as JSON:", e);
          }
        }
      }

      // Store current positions as previous before updating (same as applyRecommendations)
      const currentPositions: Record<string, number> = {};
      itemsRef.current.forEach((item, idx) => {
        currentPositions[item.id] = idx;
      });
      console.log("Storing previous positions for bandit:", currentPositions);
      setPreviousPositions(currentPositions);

      const meta = safeReadItemMeta(namespace);
      const prevOrder = lastOrderRef.current;
      const normalized = normalizeItems(
        rawItems as unknown as RawRecommendationItem[],
        prevOrder,
        meta
      );
      if (applyId !== applySeqRef.current) {
        return;
      }
      if (normalized.items.length > 0) {
        console.log("Bandit response received:", {
          rawItems: rawItems,
          normalizedItems: normalized.items,
          order: normalized.order,
          listMode: "bandit",
        });
        setBanditItems(normalized.items.slice(0, k));
        setListMode("bandit");
        // Update bandit-specific state without calling applyRecommendations
        lastOrderRef.current = normalized.order;
        setItemMeta(meta);
        setRulePreview({ mode: "none" });
        setListEpoch((v) => v + 1);
      } else {
        // Fallback only if we have no items yet; otherwise keep current list
        if (itemsRef.current.length === 0) {
          try {
            const fallbackReq = queryRecommendations();
            const fallback = await fallbackReq.promise;
            if (applyId === applySeqRef.current && fallback.items.length > 0) {
              setBanditItems(null);
              setListMode("auto");
              applyRecommendations(fallback);
            }
          } catch {
            // ignore; keep empty list
          }
        }
      }
      const chosen = extractChosenPolicyId(resp);
      console.log("Setting winner policy:", {
        chosen,
        allPolicies: policies.map((p) => ({ id: p.policy_id, name: p.name })),
        rawResponse: resp,
      });

      // If no chosen policy found, use the first active policy as fallback
      const finalChosen =
        chosen || (policies.length > 0 ? policies[0].policy_id : null) || null;
      console.log("Final chosen policy:", { chosen, finalChosen });
      setWinnerPolicyId(finalChosen);
      // Mock an "uplift this session" number
      const uplift =
        finalChosen === "diverse"
          ? 4 + Math.floor(Math.random() * 7)
          : 2 + Math.floor(Math.random() * 5);
      // Fire-and-forget a mock reward to teach the bandit
      try {
        // Validate required fields before sending
        const rewardData = {
          namespace,
          request_id: extractRequestId(resp),
          policy_id: finalChosen || "",
          bucket_key: extractBucketKey(resp),
          surface,
          reward: uplift >= 5,
          algorithm: extractAlgorithm(resp),
        };

        console.log("Sending bandit reward with:", rewardData);

        // Check if required fields are present
        if (
          !rewardData.namespace ||
          !rewardData.surface ||
          !rewardData.policy_id ||
          !rewardData.bucket_key
        ) {
          console.warn("Skipping bandit reward - missing required fields:", {
            hasNamespace: !!rewardData.namespace,
            hasSurface: !!rewardData.surface,
            hasPolicyId: !!rewardData.policy_id,
            hasBucketKey: !!rewardData.bucket_key,
            response: resp,
          });
          return;
        }

        await BanditService.postV1BanditReward(rewardData);
        console.log("Bandit reward sent successfully");
      } catch (error) {
        console.warn("Bandit reward failed:", error);
        // ignore reward errors in demo
      }
    } catch (e) {
      console.error(e);
    } finally {
      setDecideBusy(false);
      setLoading(false);
      // Allow plain applies again after the ignore window
      window.setTimeout(() => {
        banditActiveRef.current = false;
      }, 5000);
    }
  }, [
    decideBusy,
    policies,
    namespace,
    userId,
    surface,
    k,
    applyRecommendations,
    queryRecommendations,
  ]);

  const runAdaptation = useCallback(
    async (baseline: ScoredItem[], persona: string, user: string) => {
      if (!baseline.length) return;
      const snapshot = cloneItems(baseline);
      setAdaptBusy(true);
      try {
        const events = buildSimulatedWeekEvents(user, snapshot);
        await IngestionService.batchEvents({ namespace, events });

        const result = await queryRecommendations({
          previousOrder: snapshot.map((it) => it.id),
        });
        const queryResult = await result.promise;
        applyRecommendations(queryResult);
        setAdaptationResults({
          before: snapshot,
          after: queryResult.items,
          persona,
        });
      } catch (e: unknown) {
        console.error("Adaptation failed:", e);
        setAdaptationResults(null);
      } finally {
        setAdaptBusy(false);
      }
    },
    [namespace, queryRecommendations, applyRecommendations]
  );

  useEffect(() => {
    console.log("Main effect triggered with:", {
      userId,
      surface,
      k,
      namespace,
      refreshTrigger,
      forceRefresh,
      banditActive: banditActiveRef.current,
      ignoreUntil: ignorePlainUntilRef.current,
      now: Date.now(),
    });
    let active = true;
    setLoading(true);
    setError(null);

    // Store current positions before loading new recommendations
    const currentPositions: Record<string, number> = {};
    itemsRef.current.forEach((item, idx) => {
      currentPositions[item.id] = idx;
    });
    console.log("Storing previous positions for refresh:", currentPositions);
    setPreviousPositions(currentPositions);

    const applyId = ++applySeqRef.current;
    const request = queryRecommendations();
    autoLoadCancelRef.current = request.cancel;
    request.promise
      .then((result) => {
        if (!active) return;
        if (banditActiveRef.current && !forceRefresh) return;
        if (applyId !== applySeqRef.current) return;
        if (Date.now() < ignorePlainUntilRef.current && !forceRefresh) return;
        applyRecommendations(result);
        // Reset to auto mode when force refreshing to show regular items
        if (forceRefresh) {
          console.log("Force refresh: switching from bandit to auto mode");
          setListMode("auto");
          setBanditItems(null);
        }
        setForceRefresh(false); // Reset force refresh after applying
        if (pendingReplayRef.current) {
          pendingReplayRef.current = false;
          const baseline = cloneItems(result.items);
          window.setTimeout(() => {
            if (!active) return;
            runAdaptation(baseline, userIdToPersona(userId), userId);
          }, 120);
        }
      })
      .catch((e: unknown) => {
        if (!active) return;
        const msg = e instanceof Error ? e.message : "Failed to load";
        setError(msg);
        pendingReplayRef.current = false;
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
      request.cancel();
      autoLoadCancelRef.current = null;
    };
  }, [
    queryRecommendations,
    applyRecommendations,
    runAdaptation,
    userId,
    surface,
    k,
    namespace,
    refreshTrigger,
    forceRefresh,
  ]);

  function handleAdaptClick() {
    if (adaptBusy || pendingReplayRef.current || !items.length) return;
    setRulePreview({ mode: "none" });
    runAdaptation(items, personaLabel, userId);
  }

  // Explain the leading signals in plain English under the summary
  const topReasonKeys = useMemo(() => deriveTopReasonKeys(items, 3), [items]);

  return (
    <div className="stage-shell">
      <header className="demo-title-header">
        <h1 className="demo-title">RecSys Demo</h1>
      </header>

      <div className="stage-grid">
        <main className="stage-list-area">
          <div className="stage-nav-buttons">
            <button
              className={`btn btn-ghost btn-small ${
                showControlsSection ? "active" : ""
              }`}
              type="button"
              onClick={() => setShowControlsSection(!showControlsSection)}
            >
              Recommendations
            </button>
            <button
              className={`btn btn-ghost btn-small ${
                showAdaptSection ? "active" : ""
              }`}
              type="button"
              onClick={() => setShowAdaptSection(!showAdaptSection)}
            >
              Watch it adapt
            </button>
            <button
              className={`btn btn-ghost btn-small ${
                showMerchSection ? "active" : ""
              }`}
              type="button"
              onClick={() => setShowMerchSection(!showMerchSection)}
            >
              Merchandising rules
            </button>
            <button
              className={`btn btn-ghost btn-small ${
                showBanditSection ? "active" : ""
              }`}
              type="button"
              onClick={() => setShowBanditSection(!showBanditSection)}
            >
              Bandit chooses a policy
            </button>
            <button
              className={`btn btn-ghost btn-small ${
                showExplainSection ? "active" : ""
              }`}
              type="button"
              onClick={() => setShowExplainSection(!showExplainSection)}
            >
              Explain recommendations
            </button>
          </div>

          {showControlsSection && (
            <section className="adapt-card">
              <header>Recommendations</header>
              <div className="adapt-content">
                <p className="adapt-description">
                  Configure the recommendation parameters and generate fresh
                  recommendations.
                </p>
                <div className="controls-form-container">
                  <Controls
                    ref={controlsRef}
                    k={k}
                    surface={surface}
                    userId={userId}
                    showInlineApply={false}
                    formUser={formUser}
                    formSurface={formSurface}
                    formK={formK}
                    formPersona={formPersona}
                    onFormUserChange={setFormUser}
                    onFormSurfaceChange={setFormSurface}
                    onFormKChange={setFormK}
                    onFormPersonaChange={setFormPersona}
                    onRefresh={() => {
                      console.log(
                        "onRefresh called from recommendations section"
                      );
                      setRefreshTrigger((v) => v + 1);
                    }}
                  />
                </div>
                <div className="adapt-actions">
                  <button
                    className="btn btn-primary"
                    type="button"
                    onClick={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                      console.log("Recommend items clicked");
                      setRefreshTrigger((v) => v + 1);
                    }}
                    disabled={loading}
                  >
                    {loading ? "Loading..." : "Recommend items"}
                  </button>
                </div>
              </div>
            </section>
          )}

          {showAdaptSection && (
            <section className="adapt-card">
              <header>Watch it adapt</header>
              <div className="adapt-content">
                <p className="adapt-description">
                  Simulate user activity and see how recommendations change in
                  real-time.
                </p>
                <div className="adapt-actions">
                  <button
                    className="btn btn-secondary btn-small"
                    onClick={handleAdaptClick}
                    disabled={
                      adaptBusy || pendingReplayRef.current || !items.length
                    }
                    type="button"
                  >
                    {adaptLabel}
                  </button>
                </div>
                {adaptationResults && (
                  <div className="adapt-results">
                    <div className="adapt-results-header">
                      <span className="adapt-results-title">
                        Adapted for {adaptationResults.persona} persona
                      </span>
                      <span className="adapt-results-subtitle">
                        Simulated 5 events: view → view → click → add → purchase
                      </span>
                    </div>
                  </div>
                )}
              </div>
            </section>
          )}

          {showMerchSection && (
            <section className="merch-card">
              <header>Merchandising rules</header>
              {activeRules.length > 0 ? (
                <div className="rules-list">
                  {activeRules.map((rule) => (
                    <div key={rule.rule_id} className="rule-item">
                      <div className="rule-info">
                        <div className="rule-header">
                          <span className="rule-action">
                            {rule.action === "PIN" && "📌"}
                            {rule.action === "BLOCK" && "🚫"}
                            {rule.action === "BOOST" && "⚡"}
                          </span>
                          <span className="rule-name">{rule.name}</span>
                          <span className="rule-priority">
                            Priority: {rule.priority}
                          </span>
                        </div>
                        <div className="rule-details">
                          {rule.action === "PIN" && rule.item_ids && (
                            <span className="rule-target">
                              Items: {rule.item_ids.slice(0, 3).join(", ")}
                              {rule.item_ids.length > 3 && "..."}
                            </span>
                          )}
                          {rule.action === "BLOCK" && rule.target_key && (
                            <span className="rule-target">
                              Blocking: {rule.target_type} "{rule.target_key}"
                            </span>
                          )}
                          {rule.action === "BOOST" && rule.target_key && (
                            <span className="rule-target">
                              Boosting: {rule.target_type} "{rule.target_key}"
                              by {rule.boost_value || 1}x
                            </span>
                          )}
                        </div>
                      </div>
                      <button
                        className="btn btn-ghost btn-tiny rule-disable-btn"
                        onClick={async () => {
                          if (ruleBusy) return;
                          setRuleBusy(true);
                          try {
                            await RuleService.putV1AdminRulesRuleId(
                              rule.rule_id,
                              {
                                ...rule,
                                enabled: false,
                              }
                            );
                            setActiveRules((prev) =>
                              prev.filter((r) => r.rule_id !== rule.rule_id)
                            );
                            setRefreshTrigger((v) => v + 1);
                          } catch (e) {
                            console.warn("Failed to disable rule:", e);
                          } finally {
                            setRuleBusy(false);
                          }
                        }}
                        disabled={ruleBusy}
                        title="Disable this rule"
                      >
                        ✕
                      </button>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="merch-message">
                  Use the pin and block buttons on items to create rules. Rules
                  are applied to recommendations in real-time.
                </div>
              )}
            </section>
          )}

          {showBanditSection && (
            <section className="adapt-card">
              <header>Bandit chooses a policy</header>
              <div className="adapt-content">
                <p className="adapt-description">
                  We try two strategies and shift traffic to the policy —
                  automatically.
                </p>
                {Array.isArray(policies) && policies.length ? (
                  <ul className="adapt-policy-list">
                    {(() => {
                      const normalizeId = (v?: string | null) =>
                        (v ?? "").toString().trim();
                      const chosenId = normalizeId(winnerPolicyId);
                      const allPolicies = Array.isArray(policies)
                        ? policies
                        : [];
                      const ordered = chosenId
                        ? [...allPolicies].sort((a, b) => {
                            const aId = normalizeId(a.policy_id);
                            const bId = normalizeId(b.policy_id);
                            if (aId === chosenId && bId !== chosenId) return -1;
                            if (bId === chosenId && aId !== chosenId) return 1;
                            return 0;
                          })
                        : allPolicies;
                      const visible = ordered.slice(0, 4);
                      return visible.map((p) => {
                        const isWinner =
                          normalizeId(p.policy_id) === chosenId && !!chosenId;

                        console.log("Rendering policy:", {
                          policyId: p.policy_id,
                          policyName: p.name,
                          winnerPolicyId,
                          isWinner,
                          comparison: `${normalizeId(
                            p.policy_id
                          )} === ${chosenId}`,
                          strictEqual:
                            normalizeId(p.policy_id) ===
                            normalizeId(winnerPolicyId),
                          truthy: !!winnerPolicyId,
                        });

                        return (
                          <li
                            key={p.policy_id || p.name}
                            className={`adapt-policy ${
                              isWinner ? "adapt-policy-winner" : ""
                            }`}
                          >
                            <span
                              className={`meta-pill ${
                                isWinner ? "meta-pill-winner" : ""
                              }`}
                            >
                              {isWinner && (
                                <span className="winner-label">CHOSEN</span>
                              )}
                              {p.name || p.policy_id}
                            </span>
                            {p.notes ? (
                              <span className="policy-notes">{p.notes}</span>
                            ) : null}
                          </li>
                        );
                      });
                    })()}
                  </ul>
                ) : (
                  <div className="adapt-description">
                    {policiesLoading ? "Loading policies…" : "No policies"}
                  </div>
                )}
                <div className="adapt-actions">
                  <button
                    className="btn btn-primary btn-small"
                    onClick={decideAndRecommend}
                    disabled={decideBusy}
                    type="button"
                  >
                    {decideBusy ? "Deciding…" : "Decide Policy + Recommend"}
                  </button>
                </div>
              </div>
            </section>
          )}

          {showExplainSection && (
            <section className="adapt-card">
              <header>Explain recommendations</header>
              <div className="adapt-content">
                <p className="adapt-description">
                  Generate a natural language narrative for this list using the
                  Explain API.
                </p>
                <div className="signals-explainer">
                  <div className="signals-title">
                    💡 For the best explain results
                  </div>
                  <ul className="signals-list">
                    <li className="signals-item">
                      <span className="signals-name">
                        Generate recommendations
                      </span>
                      <span className="signals-sep"> · </span>
                      <span className="signals-desc">
                        Start with a fresh recommendation list
                      </span>
                    </li>
                    <li className="signals-item">
                      <span className="signals-name">Pin/block some items</span>
                      <span className="signals-sep"> · </span>
                      <span className="signals-desc">
                        Create user preferences and constraints
                      </span>
                    </li>
                    <li className="signals-item">
                      <span className="signals-name">Click event buttons</span>
                      <span className="signals-sep"> · </span>
                      <span className="signals-desc">
                        Use 👁️ View, 👆 Click, 🛒 Buy to simulate user activity
                      </span>
                    </li>
                    <li className="signals-item">
                      <span className="signals-name">Generate explanation</span>
                      <span className="signals-sep"> · </span>
                      <span className="signals-desc">
                        See how the system interprets your behavior
                      </span>
                    </li>
                  </ul>
                </div>
                <div className="adapt-actions">
                  <button
                    className="btn btn-primary btn-small"
                    type="button"
                    disabled={explainLoading}
                    onClick={async () => {
                      try {
                        setExplainLoading(true);
                        setExplainError(null);
                        const from = new Date(
                          Date.now() - 7 * 86400_000
                        ).toISOString();
                        const to = new Date().toISOString();
                        const res = await ExplainService.postV1ExplainLlm({
                          namespace,
                          surface,
                          target_type: "surface",
                          target_id: surface,
                          from,
                          to,
                        });
                        setExplainMarkdown(res.markdown || "");
                      } catch (err) {
                        const msg =
                          err instanceof Error
                            ? err.message
                            : "Failed to fetch explanation.";
                        setExplainError(msg);
                      } finally {
                        setExplainLoading(false);
                      }
                    }}
                  >
                    {explainLoading ? "Generating…" : "Generate explanation"}
                  </button>
                </div>

                {explainError ? (
                  <p className="diff-error">{explainError}</p>
                ) : null}

                {explainMarkdown ? (
                  <div className="explain-markdown" role="article">
                    <pre className="explain-pre">{explainMarkdown}</pre>
                  </div>
                ) : null}
              </div>
            </section>
          )}

          {error ? (
            <div className="stage-error-card" role="alert">
              <div>
                <h3>We hit a snag loading recommendations.</h3>
                <p>{error}</p>
              </div>
            </div>
          ) : null}

          {showEmptyState ? (
            <div className="stage-empty-card">
              <p>
                Nothing yet. Run "Decide Policy + Recommend" to fetch a fresh
                list.
              </p>
            </div>
          ) : null}

          {!showEmptyState && items.length > 0 && (
            <div className="recommendations-summary">
              <div className="recommendations-context">
                Criteria: {enhancedSummary}
              </div>
            </div>
          )}

          {!showEmptyState && items.length > 0 && topReasonKeys.length > 0 ? (
            <div
              className="signals-explainer"
              role="note"
              aria-label="Why these items"
            >
              <div className="signals-title">Why these items</div>
              <ul className="signals-list">
                {topReasonKeys.map((key) => (
                  <li key={key} className="signals-item">
                    <span className="signals-name">{formatReasonTag(key)}</span>
                    <span className="signals-sep"> · </span>
                    <span className="signals-desc">
                      {REASON_DESCRIPTIONS[key]}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          ) : null}

          <div className="topk-grid">
            {showSkeleton
              ? Array.from({ length: Math.min(k, 10) }, (_, idx) => (
                  <SkeletonCard key={`skeleton-${idx}`} />
                ))
              : displayItems.map((it, idx) => {
                  // Calculate delta for any list changes
                  let delta = it.delta;
                  if (adaptationResults) {
                    // For adaptation results, compare with before state
                    const beforeIndex = adaptationResults.before.findIndex(
                      (item) => item.id === it.id
                    );
                    delta = beforeIndex >= 0 ? beforeIndex - idx : 0;
                  } else if (
                    listMode === "bandit" &&
                    previousPositions[it.id] !== undefined
                  ) {
                    // For bandit recommendations, use stored previous positions
                    delta = previousPositions[it.id] - idx;
                  } else if (lastOrderRef.current.length > 0) {
                    // For regular list changes, compare with previous order
                    const previousIndex = lastOrderRef.current.indexOf(it.id);
                    delta = previousIndex >= 0 ? previousIndex - idx : 0;
                  }

                  // Debug logging
                  console.log(
                    `Rendering ${it.id}: prev=${
                      previousPositions[it.id]
                    }, curr=${idx}`
                  );

                  return (
                    <div
                      key={`${it.id}-${idx}-${listEpoch}`}
                      style={{
                        animationDelay: `${idx * 50}ms`,
                      }}
                    >
                      <TopKCard
                        id={it.id}
                        title={it.title || it.id}
                        reasons={it.reasons}
                        delta={delta}
                        annotation={it.annotation}
                        muted={it.muted}
                        previousPosition={previousPositions[it.id]}
                        currentPosition={idx}
                        brand={itemMeta[it.id]?.brand}
                        busy={ruleBusy}
                        score={it.score}
                        price={it.price}
                        tags={it.tags}
                        available={it.available}
                        position={idx}
                        onPin={async (id) => {
                          if (ruleBusy) return;
                          // Create a PIN rule for this item
                          setRuleBusy(true);
                          try {
                            const existing = activeRules.find(
                              (r) =>
                                r.action === "PIN" &&
                                r.target_type === "ITEM" &&
                                r.item_ids?.includes(id)
                            );
                            if (existing) {
                              await RuleService.putV1AdminRulesRuleId(
                                existing.rule_id,
                                {
                                  ...existing,
                                  enabled: false,
                                }
                              );
                              setActiveRules((prev) =>
                                prev.filter(
                                  (r) => r.rule_id !== existing.rule_id
                                )
                              );
                            } else {
                              const created =
                                await RuleService.postV1AdminRules({
                                  namespace,
                                  surface,
                                  name: `Pin ${id} - Demo`,
                                  description: "Demo pin from card action",
                                  action: "PIN",
                                  target_type: "ITEM",
                                  item_ids: [id],
                                  enabled: true,
                                  priority: 100,
                                });
                              setActiveRules((prev) => [...prev, created]);
                            }
                            setRefreshTrigger((v) => v + 1);
                          } catch (e) {
                            console.warn("PIN from card failed:", e);
                          } finally {
                            setRuleBusy(false);
                          }
                        }}
                        onBlockBrand={async (brand) => {
                          if (ruleBusy) return;
                          setRuleBusy(true);
                          try {
                            const existing = activeRules.find(
                              (r) =>
                                r.action === "BLOCK" &&
                                r.target_type === "BRAND" &&
                                r.target_key === brand
                            );
                            if (existing) {
                              await RuleService.putV1AdminRulesRuleId(
                                existing.rule_id,
                                {
                                  ...existing,
                                  enabled: false,
                                }
                              );
                              setActiveRules((prev) =>
                                prev.filter(
                                  (r) => r.rule_id !== existing.rule_id
                                )
                              );
                            } else {
                              const created =
                                await RuleService.postV1AdminRules({
                                  namespace,
                                  surface,
                                  name: `Block ${brand} - Demo`,
                                  description: "Demo block from card action",
                                  action: "BLOCK",
                                  target_type: "BRAND",
                                  target_key: brand,
                                  enabled: true,
                                  priority: 100,
                                });
                              setActiveRules((prev) => [...prev, created]);
                            }
                            setRefreshTrigger((v) => v + 1);
                          } catch (e) {
                            console.warn("BLOCK from card failed:", e);
                          } finally {
                            setRuleBusy(false);
                          }
                        }}
                      />
                    </div>
                  );
                })}
          </div>
        </main>
      </div>
    </div>
  );
}

type ControlsProps = {
  k: number;
  surface: string;
  userId: string;
  showInlineApply?: boolean;
  onRefresh?: () => void;
  // Form state props
  formUser: string;
  formSurface: string;
  formK: string;
  formPersona: string;
  onFormUserChange: (user: string) => void;
  onFormSurfaceChange: (surface: string) => void;
  onFormKChange: (k: string) => void;
  onFormPersonaChange: (persona: string) => void;
};

const Controls = forwardRef<ControlsHandle, ControlsProps>(function Controls(
  {
    k,
    surface,
    userId,
    showInlineApply = true,
    onRefresh,
    formUser,
    formSurface,
    formK,
    formPersona,
    onFormUserChange,
    onFormSurfaceChange,
    onFormKChange,
    onFormPersonaChange,
  },
  ref
) {
  const [params, setParams] = useSearchParams();

  const applyChanges = useCallback(() => {
    console.log("applyChanges called with:", {
      formUser,
      formSurface,
      formK,
      currentUserId: userId,
      currentSurface: surface,
      currentK: k,
    });
    const nextK = clampTopK(Number(formK));

    // Check if parameters actually changed
    const paramsChanged =
      formUser !== userId ||
      formSurface !== surface ||
      String(nextK) !== String(k);

    if (paramsChanged) {
      params.set("u", formUser);
      params.set("s", formSurface);
      params.set("k", String(nextK));
      setParams(params, { replace: true });
    }

    // Only trigger refresh if parameters changed
    if (paramsChanged) {
      console.log("Calling onRefresh due to parameter changes...");
      onRefresh?.();
    }
  }, [
    formUser,
    formSurface,
    formK,
    params,
    setParams,
    onRefresh,
    userId,
    surface,
    k,
  ]);

  useImperativeHandle(
    ref,
    () => ({
      apply: applyChanges,
    }),
    [applyChanges]
  );

  function handlePersonaChange(nextPersona: string) {
    onFormPersonaChange(nextPersona);
    const suggested = personaToUserId(nextPersona);
    onFormUserChange(suggested);
  }

  return (
    <form
      className="controls-form"
      onSubmit={(event) => {
        event.preventDefault();
        applyChanges();
      }}
    >
      <label className="controls-field">
        <span className="controls-label">Persona</span>
        <select
          className="controls-select"
          value={formPersona}
          onChange={(e) => handlePersonaChange(e.target.value)}
        >
          <option value="VIP">VIP</option>
          <option value="Casual">Casual</option>
          <option value="New">New</option>
        </select>
      </label>

      <label className="controls-field">
        <span className="controls-label">User</span>
        <select
          className="controls-select"
          value={formUser}
          onChange={(e) => onFormUserChange(e.target.value)}
        >
          <option value="user-1">user-1</option>
          <option value="user-2">user-2</option>
          <option value="user-3">user-3</option>
          <option value="user-4">user-4</option>
          <option value="user-5">user-5</option>
        </select>
      </label>

      <label className="controls-field">
        <span className="controls-label">Surface</span>
        <select
          className="controls-select"
          value={formSurface}
          onChange={(e) => onFormSurfaceChange(e.target.value)}
        >
          <option value="home_top">Home · Top</option>
          <option value="home_personalized">Home · Personalized</option>
          <option value="browse_shoes">Browse · Shoes</option>
        </select>
      </label>

      <label className="controls-field">
        <span className="controls-label">Top-K</span>
        <input
          className="controls-input"
          type="number"
          min={5}
          max={30}
          value={formK}
          onChange={(e) => onFormKChange(e.target.value)}
        />
      </label>

      {showInlineApply ? (
        <button
          className="btn btn-primary btn-small controls-apply"
          type="submit"
        >
          Apply
        </button>
      ) : null}
    </form>
  );
});

function safeReadItemMeta(ns: string): Record<string, ItemMeta> {
  try {
    const rawMeta = localStorage.getItem(`demo:itemMeta:${ns}`);
    if (rawMeta) {
      return JSON.parse(rawMeta) as Record<string, ItemMeta>;
    }
  } catch {
    // fallthrough to legacy lookup
  }
  try {
    const legacyRaw = localStorage.getItem(`demo:items:${ns}`);
    if (!legacyRaw) return {};
    const legacy = JSON.parse(legacyRaw) as Record<string, string>;
    const mapped: Record<string, ItemMeta> = {};
    for (const [id, title] of Object.entries(legacy)) {
      mapped[id] = { title };
    }
    return mapped;
  } catch {
    return {};
  }
}

function normalizeItems(
  raw: RawRecommendationItem[],
  previousOrder: string[],
  meta: Record<string, ItemMeta>
): { items: ScoredItem[]; order: string[] } {
  console.log("normalizeItems called with:", {
    rawLength: raw.length,
    rawItems: raw.map((r) => ({
      id: r.item_id || r.itemId || r.id,
      score: r.score,
    })),
    previousOrder,
    previousOrderLength: previousOrder.length,
  });

  const nextOrder: string[] = [];
  const items = raw.map((it, idx) => {
    const camelId = (it as { itemId?: string | undefined }).itemId;
    const itemId = it.item_id || camelId || "";
    nextOrder.push(itemId);
    const previousIndex = previousOrder.indexOf(itemId);
    const delta = previousIndex >= 0 ? previousIndex - idx : 0;
    const info = meta[itemId] || {};
    return {
      id: itemId,
      title: info.title || itemId,
      score: it.score,
      reasons: it.reasons,
      delta,
      brand: info.brand,
      price: info.price,
      tags: info.tags,
      available: info.available,
    };
  });

  console.log("normalizeItems result:", {
    itemsLength: items.length,
    itemsIds: items.map((i) => i.id),
    order: nextOrder,
    deltas: items.map((i) => i.delta),
  });

  return { items, order: nextOrder };
}

function humanizeReasons(items: ScoredItem[]): string {
  const counts: Record<string, number> = {};
  for (const it of items) {
    for (const r of it.reasons || []) {
      counts[r] = (counts[r] || 0) + 1;
    }
  }
  const order = [
    "recent_popularity",
    "co_visitation",
    "embedding",
    "personalization",
    "diversity",
  ];
  const picked = order.filter((k) => counts[k]).slice(0, 3);
  if (!picked.length) return "";
  return picked
    .map((k) =>
      k === "recent_popularity"
        ? "Popular now"
        : k === "co_visitation"
        ? "Similar to what this persona likes"
        : k === "diversity"
        ? "Balanced brands"
        : k.charAt(0).toUpperCase() + k.slice(1)
    )
    .join(" · ");
}

function clampTopK(value: number): number {
  if (!Number.isFinite(value)) return 10;
  return Math.min(30, Math.max(5, Math.round(value)));
}

function formatSurfaceLabel(surface: string): string {
  if (!surface) return "Home";
  return surface
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" · ");
}

function formatReasonTag(reason: string): string {
  return (
    SIGNAL_COPY[reason] ||
    reason.replaceAll("_", " ").replace(/\b\w/g, (c) => c.toUpperCase())
  );
}

// Plain-English descriptions for leading signals
const REASON_DESCRIPTIONS: Record<string, string> = {
  recent_popularity:
    "Trending now based on many recent views, clicks, and purchases.",
  co_visitation:
    "Often viewed or bought together with items this persona likes.",
  personalization:
    "Tailored to this persona’s recent behavior and long‑term interests.",
  diversity: "Balances the mix so one brand or type doesn’t dominate the list.",
  embedding:
    "Similar attributes and feel (learned from item features and content).",
};

// Return top N raw reason keys by frequency across the list
function deriveTopReasonKeys(items: ScoredItem[], limit = 3): string[] {
  const counts: Record<string, number> = {};
  for (const item of items) {
    for (const reason of item.reasons || []) {
      counts[reason] = (counts[reason] || 0) + 1;
    }
  }
  return Object.entries(counts)
    .sort((a, b) => b[1] - a[1])
    .slice(0, limit)
    .map(([key]) => key);
}

function readLocalPolicyHints(ns: string): BanditPolicy[] {
  try {
    const raw = localStorage.getItem(`demo:bandit:candidates:${ns}`);
    if (!raw) return [];
    const ids = JSON.parse(raw) as string[];
    return ids.map((id) => ({
      policy_id: id,
      name: id.charAt(0).toUpperCase() + id.slice(1),
      active: true,
      notes:
        id === "diverse"
          ? "Emphasize exploration and brand variety."
          : "Balanced relevance with light diversity. Good default.",
    }));
  } catch {
    return [];
  }
}

function policyNameFromId(policies: BanditPolicy[], policyId: string): string {
  const found = policies.find((p) => (p.policy_id || "") === policyId);
  return found?.name || policyId;
}

function cloneItems(list: ScoredItem[]): ScoredItem[] {
  return list.map((item) => ({
    ...item,
    reasons: item.reasons ? [...item.reasons] : [],
  }));
}

type DemoEvent = {
  user_id: string;
  item_id: string;
  type: number;
  ts: string;
  value: number;
};

function buildSimulatedWeekEvents(
  user: string,
  baseline: ScoredItem[]
): DemoEvent[] {
  // Use past timestamps so explain window (last 7 days) reliably captures them
  const baseTime = Date.now() - 2 * 86400_000; // 2 days ago
  const preferred = baseline.map((it) => it.id).filter(Boolean);
  const padding = ["item-1", "item-2", "item-3", "item-4", "item-5"];
  const pool = Array.from(new Set([...preferred, ...padding]));
  const picks = [
    pool[0] || "item-1",
    pool[1] || pool[0] || "item-2",
    pool[1] || pool[0] || "item-2",
    pool[2] || pool[1] || "item-3",
    pool[3] || pool[2] || "item-4",
  ];
  const typeSequence = [1, 1, 2, 3, 3]; // view → view → click → add → purchase
  return typeSequence.map((type, idx) => ({
    user_id: user,
    item_id: picks[idx] || picks[picks.length - 1],
    type,
    ts: new Date(baseTime + idx * 60000).toISOString(),
    value: 1,
  }));
}

function userIdToPersona(user: string): string {
  if (user === "user-1") return "VIP";
  if (user === "user-2") return "Casual";
  if (user === "user-3") return "New";
  return "VIP";
}

function personaToUserId(persona: string): string {
  if (persona === "Casual") return "user-2";
  if (persona === "New") return "user-3";
  return "user-1";
}

function SkeletonCard(): React.ReactElement {
  return (
    <div className="topk-card skeleton-card" aria-hidden>
      <div className="skeleton-line skeleton-line-lg" />
      <div className="skeleton-line skeleton-line-sm" />
      <div className="skeleton-pill-row">
        <span className="skeleton-pill" />
        <span className="skeleton-pill" />
        <span className="skeleton-pill" />
      </div>
    </div>
  );
}
