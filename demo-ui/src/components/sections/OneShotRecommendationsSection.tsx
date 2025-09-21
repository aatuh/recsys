import React, { useState, useMemo } from "react";
import { Section, Row, Label, Button } from "../primitives/UIComponents";
// import { spacing, text, color } from "../../ui/tokens";
import { useToast } from "../../ui/Toast";
import {
  RankingService,
  type types_RecommendWithBanditRequest,
  type types_BanditPolicy,
  AuditService,
} from "../../lib/api-client";
import {
  useViewState,
  type BanditDecisionEntry,
} from "../../contexts/ViewStateContext";
import { DecisionTraceDrawer } from "../primitives/DecisionTraceDrawer";

interface OneShotRecommendationsSectionProps {
  namespace: string;
  availablePolicies: types_BanditPolicy[];
  generatedUsers: string[];
}

export function OneShotRecommendationsSection({
  namespace,
  availablePolicies,
  generatedUsers,
}: OneShotRecommendationsSectionProps) {
  const { banditPlayground, setBanditPlayground } = useViewState();
  const toast = useToast();

  // Use context state for consistency with Decision Section
  const { surface, context, candidatePolicyIds, algorithm } = banditPlayground;

  // Get example user for placeholder
  const exampleUser = useMemo(() => {
    return generatedUsers[0] || "user-0001";
  }, [generatedUsers.length]);

  // Local state for recommendations
  const [userId, setUserId] = useState("");
  const [k, setK] = useState(20);
  const [includeReasons, setIncludeReasons] = useState(true);
  const [lastRequestId, setLastRequestId] = useState<string | null>(null);
  const [showDecisionTrace, setShowDecisionTrace] = useState(false);
  const [currentDecisionId, setCurrentDecisionId] = useState<string | null>(
    null
  );

  // Use context state for recommendations
  const { recommendationResult, recommendationLoading, recommendationError } =
    banditPlayground;

  const handleGetRecommendations = async () => {
    const id = userId.trim() || exampleUser;

    setBanditPlayground((prev) => ({
      ...prev,
      recommendationLoading: true,
      recommendationError: null,
      recommendationResult: null,
    }));

    try {
      const reqId = `req_${Date.now()}_${Math.random()
        .toString(36)
        .slice(2, 8)}`;
      setLastRequestId(reqId);

      const request: types_RecommendWithBanditRequest = {
        namespace,
        surface,
        context,
        user_id: id,
        candidate_policy_ids: candidatePolicyIds,
        algorithm,
        k,
        include_reasons: includeReasons,
        request_id: reqId,
      };

      console.log("Getting bandit recommendations with request:", request);
      const response = await RankingService.postV1BanditRecommendations(
        request
      );
      console.log("Bandit recommendations response:", response);
      console.log("Response type:", typeof response);

      // Parse the response if it's a string
      let parsedResponse = response;
      if (typeof response === "string") {
        try {
          parsedResponse = JSON.parse(response);
          console.log("Parsed response:", parsedResponse);
        } catch (e) {
          console.error("Failed to parse response:", e);
          setBanditPlayground((prev) => ({
            ...prev,
            recommendationError: "Failed to parse recommendations response",
          }));
          return;
        }
      }

      console.log("Response keys:", Object.keys(parsedResponse || {}));
      console.log("Items count:", parsedResponse?.items?.length || 0);
      console.log("Items:", parsedResponse?.items);

      // Append a decision entry so dashboards update from recommendations too
      const explain = (parsedResponse as any)?.bandit_explain || {};
      const empBest = (
        explain && typeof explain === "object"
          ? (explain as Record<string, string>)["emp_best"]
          : undefined
      ) as string | undefined;
      const decisionEntry: BanditDecisionEntry = {
        id: `dec-${Date.now()}`,
        timestamp: new Date(),
        requestId: undefined,
        policyId: String((parsedResponse as any)?.chosen_policy_id ?? ""),
        surface: String(surface || ""),
        bucketKey: String((parsedResponse as any)?.bandit_bucket ?? ""),
        algorithm: String((parsedResponse as any)?.algorithm ?? ""),
        explore: Boolean((parsedResponse as any)?.explore),
        empBestPolicyId: empBest,
        context: (context as any) || {},
      };
      setBanditPlayground((prev) => ({
        ...prev,
        recommendationResult: parsedResponse,
        decisionHistory: [decisionEntry, ...prev.decisionHistory].slice(0, 500),
      }));
      toast.success(
        `Got ${String((parsedResponse as any)?.items?.length ?? 0)} items`,
        "Bandit recommendations"
      );
    } catch (err) {
      console.error("Failed to get bandit recommendations:", err);
      // Try to extract API error details if available
      let errorMessage = "Failed to get recommendations";
      const anyErr = err as any;
      if (anyErr && typeof anyErr === "object") {
        if (typeof anyErr.message === "string" && anyErr.message) {
          errorMessage = anyErr.message;
        }
        if (anyErr.body && typeof anyErr.body === "object") {
          const apiMsg = anyErr.body?.error?.message || anyErr.body?.message;
          if (typeof apiMsg === "string" && apiMsg) {
            errorMessage = apiMsg;
          }
        }
        if (anyErr.status) {
          errorMessage = `[${anyErr.status}] ${errorMessage}`;
        }
      }
      setBanditPlayground((prev) => ({
        ...prev,
        recommendationError: errorMessage,
      }));
      toast.error(errorMessage);
    } finally {
      setBanditPlayground((prev) => ({
        ...prev,
        recommendationLoading: false,
      }));
    }
  };

  const handleViewDecisionTrace = async () => {
    if (!lastRequestId) return;
    const delay = (ms: number) => new Promise((r) => setTimeout(r, ms));
    const maxAttempts = 6;
    try {
      let found: string | null = null;
      for (let i = 0; i < maxAttempts; i++) {
        // Try by request_id first
        const list = await AuditService.getV1AuditDecisions(
          namespace,
          undefined,
          undefined,
          undefined,
          lastRequestId,
          1
        );
        let dec = list.decisions?.[0];
        if (!dec) {
          // Fallback: most recent decision in namespace
          const fallback = await AuditService.getV1AuditDecisions(
            namespace,
            undefined,
            undefined,
            undefined,
            undefined,
            1
          );
          dec = fallback.decisions?.[0];
        }
        if (dec?.decision_id) {
          found = dec.decision_id;
          break;
        }
        await delay(250 * (i + 1));
      }
      if (found) {
        setCurrentDecisionId(found);
        setShowDecisionTrace(true);
      } else {
        setBanditPlayground((prev) => ({
          ...prev,
          recommendationError:
            "Decision trace not yet available. Try again in a moment.",
        }));
      }
    } catch {
      setBanditPlayground((prev) => ({
        ...prev,
        recommendationError: "Failed to load decision trace index",
      }));
    }
  };

  const getPolicyById = (policyId: string) => {
    return availablePolicies.find((p) => p.policy_id === policyId);
  };

  const getPolicyDisplayName = (policyId: string) => {
    const policy = getPolicyById(policyId);
    return policy?.name || policyId || "Unknown Policy";
  };

  const getEmpiricalBestPolicy = () => {
    return availablePolicies.find((p) => p.active) || availablePolicies[0];
  };

  const isExploration = (chosenPolicyId: string) => {
    const empiricalBest = getEmpiricalBestPolicy();
    return empiricalBest && chosenPolicyId !== empiricalBest.policy_id;
  };

  return (
    <Section title="One-Shot Recommendations">
      <div style={{ marginBottom: 16 }}>
        <p style={{ color: "#666", fontSize: 14, marginBottom: 16 }}>
          Get personalized recommendations using bandit-selected policies. The
          system will choose a policy and generate recommendations with ranking
          explanations.
        </p>

        {/* Input Fields */}
        <div style={{ marginBottom: 20 }}>
          <Row>
            <div style={{ flex: 1 }}>
              <label
                style={{
                  display: "block",
                  marginBottom: 4,
                  fontSize: 12,
                  fontWeight: "bold",
                }}
              >
                User ID (leave blank to use first generated)
              </label>
              <input
                type="text"
                value={userId}
                onChange={(e) => setUserId(e.target.value)}
                placeholder={exampleUser}
                style={{
                  width: "100%",
                  padding: "8px 12px",
                  border: "1px solid #ddd",
                  borderRadius: 4,
                  fontSize: 14,
                }}
              />
            </div>
            <div style={{ flex: 1 }}>
              <label
                style={{
                  display: "block",
                  marginBottom: 4,
                  fontSize: 12,
                  fontWeight: "bold",
                }}
              >
                Number of Recommendations (k)
              </label>
              <input
                type="number"
                value={k}
                onChange={(e) => setK(parseInt(e.target.value) || 20)}
                min="1"
                max="100"
                style={{
                  width: "100%",
                  padding: "8px 12px",
                  border: "1px solid #ddd",
                  borderRadius: 4,
                  fontSize: 14,
                }}
              />
            </div>
          </Row>

          <div style={{ marginTop: 12 }}>
            <label style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <input
                type="checkbox"
                checked={includeReasons}
                onChange={(e) => setIncludeReasons(e.target.checked)}
              />
              <span style={{ fontSize: 14 }}>Include ranking reasons</span>
            </label>
          </div>
        </div>

        {/* Action Button */}
        <div style={{ marginBottom: 20 }}>
          <Button
            onClick={handleGetRecommendations}
            disabled={recommendationLoading}
            style={{
              backgroundColor: recommendationLoading ? "#ccc" : "#28a745",
              color: "white",
              border: "none",
              padding: "12px 24px",
              borderRadius: 4,
              fontSize: 14,
              fontWeight: "bold",
              cursor: recommendationLoading ? "not-allowed" : "pointer",
            }}
          >
            {recommendationLoading
              ? "Getting Recommendations..."
              : "Get Recommendations"}
          </Button>
        </div>

        {/* Error Display */}
        {recommendationError && (
          <div
            style={{
              backgroundColor: "#f8d7da",
              color: "#721c24",
              border: "1px solid #f5c6cb",
              borderRadius: 4,
              padding: 12,
              marginBottom: 20,
            }}
          >
            <strong>Error:</strong> {recommendationError}
          </div>
        )}

        {/* Results Display */}
        {recommendationResult && (
          <div style={{ marginBottom: 20 }}>
            <Label text="Recommendation Results">
              <div
                style={{ display: "flex", flexDirection: "column", gap: 16 }}
              >
                {/* Decision Trace Viewer */}
                <div>
                  <Button
                    onClick={handleViewDecisionTrace}
                    disabled={!lastRequestId}
                    style={{
                      backgroundColor: lastRequestId ? "#007acc" : "#ccc",
                      color: "white",
                      border: "none",
                      padding: "8px 16px",
                      borderRadius: 4,
                      cursor: lastRequestId ? "pointer" : "not-allowed",
                      fontSize: 14,
                      alignSelf: "flex-start",
                    }}
                  >
                    View Decision Trace
                  </Button>
                </div>
                {/* Policy Information */}
                <div
                  style={{
                    backgroundColor: "#e3f2fd",
                    border: "1px solid #bbdefb",
                    borderRadius: 4,
                    padding: 12,
                  }}
                >
                  <div
                    style={{
                      fontSize: 12,
                      fontWeight: "bold",
                      marginBottom: 8,
                    }}
                  >
                    Bandit Policy Information
                  </div>
                  <div
                    style={{
                      display: "grid",
                      gridTemplateColumns: "1fr 1fr",
                      gap: 12,
                    }}
                  >
                    <div>
                      <div
                        style={{
                          fontSize: 12,
                          fontWeight: "bold",
                          marginBottom: 4,
                        }}
                      >
                        Chosen Policy
                      </div>
                      <div style={{ fontSize: 14 }}>
                        {getPolicyDisplayName(
                          recommendationResult.chosen_policy_id || ""
                        )}
                      </div>
                    </div>
                    <div>
                      <div
                        style={{
                          fontSize: 12,
                          fontWeight: "bold",
                          marginBottom: 4,
                        }}
                      >
                        Algorithm
                      </div>
                      <div
                        style={{ fontSize: 14, textTransform: "capitalize" }}
                      >
                        {recommendationResult.algorithm || "Unknown"}
                      </div>
                    </div>
                    <div>
                      <div
                        style={{
                          fontSize: 12,
                          fontWeight: "bold",
                          marginBottom: 4,
                        }}
                      >
                        Bucket Key
                      </div>
                      <div style={{ fontSize: 14, fontFamily: "monospace" }}>
                        {recommendationResult.bandit_bucket || "N/A"}
                      </div>
                    </div>
                    <div>
                      <div
                        style={{
                          fontSize: 12,
                          fontWeight: "bold",
                          marginBottom: 4,
                        }}
                      >
                        Exploration
                      </div>
                      <div
                        style={{
                          fontSize: 14,
                          color: recommendationResult.explore
                            ? "#ff9800"
                            : "#4caf50",
                          fontWeight: "bold",
                        }}
                      >
                        {recommendationResult.explore ? "YES" : "NO"}
                      </div>
                    </div>
                  </div>

                  {/* Exploration vs Exploitation Analysis */}
                  <div
                    style={{
                      marginTop: 12,
                      padding: 8,
                      backgroundColor: "rgba(255, 255, 255, 0.7)",
                      borderRadius: 4,
                    }}
                  >
                    <div
                      style={{
                        fontSize: 12,
                        fontWeight: "bold",
                        marginBottom: 8,
                      }}
                    >
                      Exploration vs Exploitation Analysis
                    </div>
                    <div
                      style={{
                        fontSize: 12,
                        color: "#6c757d",
                        marginBottom: 8,
                      }}
                    >
                      <strong>Exploration</strong> means the bandit chose a
                      different policy than the current empirical best to learn
                      more about it.
                      <br /> <strong>Exploitation</strong> means it chose the
                      current empirical best policy to maximize expected reward.
                    </div>
                    <div style={{ fontSize: 12 }}>
                      <div style={{ marginBottom: 4 }}>
                        <strong>Empirical Best Policy:</strong>{" "}
                        {getEmpiricalBestPolicy()?.name || "None available"}
                      </div>
                      <div style={{ marginBottom: 4 }}>
                        <strong>Chosen Policy:</strong>{" "}
                        {getPolicyDisplayName(
                          recommendationResult.chosen_policy_id || ""
                        )}
                      </div>
                      <div>
                        <strong>Decision Type:</strong>{" "}
                        <span
                          style={{
                            color: isExploration(
                              recommendationResult.chosen_policy_id || ""
                            )
                              ? "#ff9800"
                              : "#4caf50",
                            fontWeight: "bold",
                          }}
                        >
                          {isExploration(
                            recommendationResult.chosen_policy_id || ""
                          )
                            ? "EXPLORATION"
                            : "EXPLOITATION"}
                        </span>
                        {isExploration(
                          recommendationResult.chosen_policy_id || ""
                        ) && (
                          <span
                            style={{
                              fontSize: 12,
                              color: "#666",
                              marginLeft: 8,
                            }}
                          >
                            (Chose different policy than empirical best)
                          </span>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Bandit Explanation */}
                  {recommendationResult.bandit_explain &&
                    Object.keys(recommendationResult.bandit_explain).length >
                      0 && (
                      <div
                        style={{
                          marginTop: 12,
                          padding: 8,
                          backgroundColor: "rgba(255, 255, 255, 0.7)",
                          borderRadius: 4,
                        }}
                      >
                        <div
                          style={{
                            fontSize: 12,
                            fontWeight: "bold",
                            marginBottom: 8,
                          }}
                        >
                          Bandit Algorithm Explanation
                        </div>
                        <div style={{ fontSize: 12, fontFamily: "monospace" }}>
                          {Object.entries(
                            recommendationResult.bandit_explain
                          ).map(([key, value]) => (
                            <div key={key} style={{ marginBottom: 2 }}>
                              <strong>{key}:</strong> {value}
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                </div>

                {/* Recommendations List */}
                <div>
                  <div
                    style={{
                      fontSize: 12,
                      fontWeight: "bold",
                      marginBottom: 8,
                    }}
                  >
                    Recommended Items ({recommendationResult.items?.length || 0}
                    )
                  </div>
                  {recommendationResult.items &&
                  recommendationResult.items.length > 0 ? (
                    <div
                      style={{
                        display: "flex",
                        flexDirection: "column",
                        gap: 8,
                      }}
                    >
                      {recommendationResult.items.map((item, index) => {
                        // Helper function to detect rule-related reasons
                        const getRuleDecorations = (reasons: any) => {
                          const decorations = [];
                          const reasonStrings = Array.isArray(reasons)
                            ? reasons
                            : Object.values(reasons || {});

                          for (const reason of reasonStrings) {
                            const reasonStr = String(reason);
                            if (reasonStr.startsWith("rule.pin")) {
                              decorations.push({
                                type: "pinned",
                                icon: "⭐",
                                label: "Pinned",
                                color: "#ffc107",
                                backgroundColor: "#fff3cd",
                              });
                            } else if (reasonStr.startsWith("rule.block")) {
                              decorations.push({
                                type: "blocked",
                                icon: "🚫",
                                label: "Blocked",
                                color: "#dc3545",
                                backgroundColor: "#f8d7da",
                              });
                            } else if (reasonStr.startsWith("rule.boost")) {
                              const boostMatch = reasonStr.match(
                                /rule\.boost:([+-]?\d*\.?\d+)/
                              );
                              if (boostMatch) {
                                decorations.push({
                                  type: "boosted",
                                  icon: "⬆️",
                                  label: `+${boostMatch[1]}`,
                                  color: "#28a745",
                                  backgroundColor: "#d4edda",
                                });
                              }
                            }
                          }

                          return decorations;
                        };

                        const ruleDecorations = getRuleDecorations(
                          item.reasons
                        );
                        const isPinned = ruleDecorations.some(
                          (d) => d.type === "pinned"
                        );
                        const isBlocked = ruleDecorations.some(
                          (d) => d.type === "blocked"
                        );

                        return (
                          <div
                            key={item.item_id}
                            style={{
                              border: "1px solid #ddd",
                              borderRadius: 4,
                              padding: 12,
                              backgroundColor: isPinned
                                ? "#fff3cd"
                                : isBlocked
                                ? "#f8d7da"
                                : "#fff",
                              borderLeft: isPinned
                                ? "4px solid #ffc107"
                                : isBlocked
                                ? "4px solid #dc3545"
                                : undefined,
                            }}
                          >
                            <div
                              style={{
                                display: "flex",
                                justifyContent: "space-between",
                                alignItems: "flex-start",
                                marginBottom: 8,
                              }}
                            >
                              <div>
                                <div
                                  style={{
                                    fontSize: 14,
                                    fontWeight: "bold",
                                    marginBottom: 4,
                                    display: "flex",
                                    alignItems: "center",
                                    gap: 8,
                                  }}
                                >
                                  <span>
                                    #{index + 1} {item.item_id}
                                  </span>
                                  {isPinned && (
                                    <span
                                      style={{ fontSize: "16px" }}
                                      title="Pinned item"
                                    >
                                      ⭐
                                    </span>
                                  )}
                                  {isBlocked && (
                                    <span
                                      style={{ fontSize: "16px" }}
                                      title="Blocked item"
                                    >
                                      🚫
                                    </span>
                                  )}
                                </div>
                                <div
                                  style={{
                                    fontSize: 12,
                                    color: "#666",
                                    marginBottom: 4,
                                    display: "flex",
                                    alignItems: "center",
                                    gap: 8,
                                  }}
                                >
                                  <span>
                                    Score: {item.score?.toFixed(4) || "N/A"}
                                  </span>
                                  {ruleDecorations
                                    .filter((d) => d.type === "boosted")
                                    .map((decoration, idx) => (
                                      <span
                                        key={idx}
                                        style={{
                                          backgroundColor:
                                            decoration.backgroundColor,
                                          color: decoration.color,
                                          padding: "2px 6px",
                                          borderRadius: 4,
                                          fontSize: "10px",
                                          fontWeight: "bold",
                                          border: `1px solid ${decoration.color}`,
                                        }}
                                        title={`Boosted by rule: ${decoration.label}`}
                                      >
                                        {decoration.label}
                                      </span>
                                    ))}
                                </div>
                              </div>
                              <div
                                style={{
                                  fontSize: 12,
                                  color: "#666",
                                  textAlign: "right",
                                }}
                              ></div>
                            </div>

                            {/* Ranking Reasons */}
                            {item.reasons &&
                              Object.keys(item.reasons).length > 0 && (
                                <div
                                  style={{
                                    marginTop: 8,
                                    padding: 8,
                                    backgroundColor: "#f8f9fa",
                                    borderRadius: 4,
                                    border: "1px solid #e9ecef",
                                  }}
                                >
                                  <div
                                    style={{
                                      fontSize: 11,
                                      fontWeight: "bold",
                                      marginBottom: 4,
                                      color: "#495057",
                                    }}
                                  >
                                    Ranking Reasons:
                                  </div>
                                  <div
                                    style={{
                                      fontSize: 11,
                                      fontFamily: "monospace",
                                    }}
                                  >
                                    {Object.entries(item.reasons).map(
                                      ([key, value]) => {
                                        const reasonStr = String(value);
                                        const decoration = ruleDecorations.find(
                                          (d) => reasonStr.includes(d.type)
                                        );

                                        if (decoration) {
                                          return (
                                            <div
                                              key={key}
                                              style={{
                                                marginBottom: 2,
                                                display: "flex",
                                                alignItems: "center",
                                                gap: 4,
                                              }}
                                            >
                                              <strong>{key}:</strong>
                                              <span
                                                style={{
                                                  backgroundColor:
                                                    decoration.backgroundColor,
                                                  color: decoration.color,
                                                  padding: "2px 6px",
                                                  borderRadius: 4,
                                                  fontSize: "10px",
                                                  fontWeight: "bold",
                                                  border: `1px solid ${decoration.color}`,
                                                  display: "flex",
                                                  alignItems: "center",
                                                  gap: 4,
                                                }}
                                                title={`Rule effect: ${reasonStr}`}
                                              >
                                                <span>{decoration.icon}</span>
                                                <span>{decoration.label}</span>
                                              </span>
                                            </div>
                                          );
                                        }

                                        return (
                                          <div
                                            key={key}
                                            style={{ marginBottom: 2 }}
                                          >
                                            <strong>{key}:</strong> {value}
                                          </div>
                                        );
                                      }
                                    )}
                                  </div>
                                </div>
                              )}
                          </div>
                        );
                      })}
                    </div>
                  ) : (
                    <div
                      style={{
                        padding: 16,
                        textAlign: "center",
                        color: "#666",
                        fontStyle: "italic",
                      }}
                    >
                      No recommendations returned. This could be because:
                      <br />
                      • No items are available in the system
                      <br />
                      • The user has no interaction history
                      <br />• The recommendation algorithm found no suitable
                      items
                    </div>
                  )}
                </div>
              </div>
            </Label>
          </div>
        )}
      </div>

      {/* Decision Trace Drawer */}
      <DecisionTraceDrawer
        isOpen={showDecisionTrace}
        onClose={() => setShowDecisionTrace(false)}
        decisionId={currentDecisionId}
        namespace={namespace}
      />
    </Section>
  );
}
