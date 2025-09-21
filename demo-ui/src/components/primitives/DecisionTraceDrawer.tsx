import React, { useState, useEffect } from "react";
import {
  AuditService,
  type types_AuditDecisionDetail,
  type types_AuditTraceConfig,
  type types_AuditTraceBandit,
} from "../../lib/api-client";
import { Button, Code, Th, Td } from "../primitives/UIComponents";
import { WhyItWorks } from "./WhyItWorks";
import { color, spacing, text } from "../../ui/tokens";
import { useToast } from "../../ui/Toast";

interface DecisionTraceDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  decisionId: string | null;
  namespace: string;
}

export function DecisionTraceDrawer({
  isOpen,
  onClose,
  decisionId,
  namespace: _namespace,
}: DecisionTraceDrawerProps) {
  const [trace, setTrace] = useState<types_AuditDecisionDetail | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const toast = useToast();

  useEffect(() => {
    if (isOpen && decisionId) {
      loadDecisionTrace();
    }
  }, [isOpen, decisionId]);

  const loadDecisionTrace = async () => {
    if (!decisionId) return;

    // Check if this is a mock decision ID (starts with "dec-")
    if (decisionId.startsWith("dec-")) {
      setError(
        "Decision trace not available - this is a mock decision ID. In a real implementation, the bandit decision response would include the actual decision_id for audit trail lookup."
      );
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const response = await AuditService.getV1AuditDecisions1(decisionId);
      setTrace(response);
    } catch (err) {
      console.error("Failed to load decision trace:", err);
      const msg =
        err instanceof Error ? err.message : "Failed to load decision trace";
      setError(msg);
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  };

  const exportToJSON = () => {
    if (!trace) return;

    const dataStr = JSON.stringify(trace, null, 2);
    const dataBlob = new Blob([dataStr], { type: "application/json" });
    const url = URL.createObjectURL(dataBlob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `decision-trace-${decisionId}.json`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  const formatTimestamp = (ts: string | undefined) => {
    if (!ts) return "N/A";
    return new Date(ts).toLocaleString();
  };

  const formatConfig = (config: types_AuditTraceConfig | undefined) => {
    if (!config) return {};
    return {
      "Blend Weights": {
        alpha: config.alpha,
        beta: config.beta,
        gamma: config.gamma,
      },
      Personalization: {
        profile_boost: config.profile_boost,
        profile_window_days: config.profile_window_days,
        profile_top_n: config.profile_top_n,
      },
      "Diversity & Caps": {
        mmr_lambda: config.mmr_lambda,
        brand_cap: config.brand_cap,
        category_cap: config.category_cap,
      },
      "Windows & Rules": {
        half_life_days: config.half_life_days,
        co_vis_window_days: config.co_vis_window_days,
        purchased_window_days: config.purchased_window_days,
        rule_exclude_events: config.rule_exclude_events,
        popularity_fanout: config.popularity_fanout,
      },
    };
  };

  const formatBandit = (bandit: types_AuditTraceBandit) => {
    return {
      Policy: bandit.chosen_policy_id,
      Algorithm: bandit.algorithm,
      "Bucket Key": bandit.bucket_key || "N/A",
      Exploration: bandit.explore ? "EXPLORE" : "EXPLOIT",
      "Request ID": bandit.request_id || "N/A",
      Explanation: bandit.explain || {},
    };
  };

  if (!isOpen) return null;

  return (
    <div
      style={{
        position: "fixed",
        top: 0,
        right: 0,
        width: "50%",
        height: "100vh",
        backgroundColor: "white",
        borderLeft: `1px solid ${color.border}`,
        boxShadow: "-2px 0 8px rgba(0,0,0,0.1)",
        zIndex: 1000,
        overflowY: "auto",
        padding: spacing.lg,
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: spacing.lg,
          paddingBottom: spacing.md,
          borderBottom: `1px solid ${color.border}`,
        }}
      >
        <h2 style={{ margin: 0, fontSize: text.lg }}>Decision Trace</h2>
        <div style={{ display: "flex", gap: spacing.md }}>
          {trace && (
            <Button
              onClick={exportToJSON}
              aria-label="Export decision trace as JSON"
              style={{
                backgroundColor: color.success,
                color: color.primaryTextOn,
                border: "none",
                padding: "6px 12px",
                borderRadius: 4,
                fontSize: 12,
              }}
            >
              Export JSON
            </Button>
          )}
          <Button
            onClick={onClose}
            aria-label="Close decision trace drawer"
            style={{
              backgroundColor: color.buttonBg,
              color: color.text,
              border: `1px solid ${color.buttonBorder}`,
              padding: "6px 12px",
              borderRadius: 4,
              fontSize: 12,
            }}
          >
            Close
          </Button>
        </div>
      </div>

      {loading && (
        <div
          role="status"
          aria-live="polite"
          style={{ textAlign: "center", padding: 20 }}
        >
          Loading decision trace...
        </div>
      )}

      {error && (
        <div
          style={{
            backgroundColor: color.dangerBg,
            color: color.text,
            border: `1px solid ${color.panelBorder}`,
            borderRadius: 4,
            padding: 12,
            marginBottom: 16,
          }}
        >
          Error: {error}
        </div>
      )}

      {trace && (
        <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
          <WhyItWorks
            metrics={[
              { label: "Policy", value: trace.bandit?.chosen_policy_id },
              { label: "Algorithm", value: trace.bandit?.algorithm },
              {
                label: "Explore/Exploit",
                value: trace.bandit?.explore ? "EXPLORE" : "EXPLOIT",
              },
              {
                label: "Candidates (pre)",
                value: trace.candidates_pre?.length,
              },
              { label: "Final Items", value: trace.final_items?.length },
            ]}
          />
          {/* Request Meta */}
          <div>
            <h3
              style={{
                margin: "0 0 8px 0",
                fontSize: "14px",
                fontWeight: "bold",
              }}
            >
              Request Meta
            </h3>
            <div
              style={{
                backgroundColor: color.panelSubtle,
                border: `1px solid ${color.panelBorder}`,
                borderRadius: "4px",
                padding: "12px",
                fontSize: "12px",
              }}
            >
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: "8px",
                }}
              >
                <div>
                  <strong>Decision ID:</strong>
                  <br />
                  <code style={{ fontSize: "11px" }}>{trace.decision_id}</code>
                </div>
                <div>
                  <strong>Timestamp:</strong>
                  <br />
                  {formatTimestamp(trace.ts)}
                </div>
                <div>
                  <strong>Namespace:</strong>
                  <br />
                  {trace.namespace}
                </div>
                <div>
                  <strong>Surface:</strong>
                  <br />
                  {trace.surface || "N/A"}
                </div>
                <div>
                  <strong>Request ID:</strong>
                  <br />
                  {trace.request_id || "N/A"}
                </div>
                <div>
                  <strong>User Hash:</strong>
                  <br />
                  {trace.user_hash ? (
                    <code style={{ fontSize: "11px" }}>
                      {trace.user_hash.substring(0, 16)}...
                    </code>
                  ) : (
                    "N/A"
                  )}
                </div>
                <div>
                  <strong>K (Result Size):</strong>
                  <br />
                  {trace.k || "N/A"}
                </div>
              </div>
            </div>
          </div>

          {/* Rules Information */}
          {trace.extras &&
            (trace.extras.rules_evaluated ||
              trace.extras.rules_matched ||
              trace.extras.rule_effects_per_item) && (
              <div>
                <h3
                  style={{
                    margin: "0 0 8px 0",
                    fontSize: "14px",
                    fontWeight: "bold",
                  }}
                >
                  Rule Engine
                </h3>
                <div
                  style={{
                    backgroundColor: color.warningBg,
                    border: `1px solid ${color.warning}`,
                    borderRadius: "4px",
                    padding: "12px",
                  }}
                >
                  {/* Rules Evaluated Counter */}
                  {trace.extras.rules_evaluated && (
                    <div style={{ marginBottom: 12 }}>
                      <div
                        style={{
                          fontSize: "12px",
                          fontWeight: "bold",
                          marginBottom: 4,
                        }}
                      >
                        Rules Evaluated:{" "}
                        {Array.isArray(trace.extras.rules_evaluated)
                          ? trace.extras.rules_evaluated.length
                          : trace.extras.rules_evaluated}
                      </div>
                      {Array.isArray(trace.extras.rules_evaluated) &&
                        trace.extras.rules_evaluated.length > 0 && (
                          <div style={{ fontSize: "11px", color: "#666" }}>
                            Rule IDs: {trace.extras.rules_evaluated.join(", ")}
                          </div>
                        )}
                    </div>
                  )}

                  {/* Rules Matched */}
                  {trace.extras.rules_matched &&
                    Array.isArray(trace.extras.rules_matched) &&
                    trace.extras.rules_matched.length > 0 && (
                      <div style={{ marginBottom: 12 }}>
                        <div
                          style={{
                            fontSize: "12px",
                            fontWeight: "bold",
                            marginBottom: 4,
                          }}
                        >
                          Rules Matched ({trace.extras.rules_matched.length})
                        </div>
                        <div style={{ overflowX: "auto" }}>
                          <table
                            style={{
                              borderCollapse: "collapse",
                              width: "100%",
                              fontSize: "11px",
                            }}
                          >
                            <thead>
                              <tr>
                                <th
                                  style={{
                                    border: `1px solid ${color.border}`,
                                    padding: "4px",
                                    backgroundColor: color.panelSubtle,
                                  }}
                                >
                                  Rule ID
                                </th>
                                <th
                                  style={{
                                    border: `1px solid ${color.border}`,
                                    padding: "4px",
                                    backgroundColor: color.panelSubtle,
                                  }}
                                >
                                  Action
                                </th>
                                <th
                                  style={{
                                    border: `1px solid ${color.border}`,
                                    padding: "4px",
                                    backgroundColor: color.panelSubtle,
                                  }}
                                >
                                  Target
                                </th>
                                <th
                                  style={{
                                    border: `1px solid ${color.border}`,
                                    padding: "4px",
                                    backgroundColor: color.panelSubtle,
                                  }}
                                >
                                  Affected Items
                                </th>
                              </tr>
                            </thead>
                            <tbody>
                              {trace.extras.rules_matched.map(
                                (rule: any, index: number) => (
                                  <tr key={index}>
                                    <td
                                      style={{
                                        border: `1px solid ${color.border}`,
                                        padding: "4px",
                                      }}
                                    >
                                      <code style={{ fontSize: "10px" }}>
                                        {rule.rule_id}
                                      </code>
                                    </td>
                                    <td
                                      style={{
                                        border: `1px solid ${color.border}`,
                                        padding: "4px",
                                      }}
                                    >
                                      <span
                                        style={{
                                          backgroundColor:
                                            rule.action === "BLOCK"
                                              ? color.danger
                                              : rule.action === "PIN"
                                              ? color.warning
                                              : color.success,
                                          color: color.primaryTextOn,
                                          padding: "2px 6px",
                                          borderRadius: 3,
                                          fontSize: "10px",
                                          fontWeight: "bold",
                                        }}
                                      >
                                        {rule.action}
                                      </span>
                                    </td>
                                    <td
                                      style={{
                                        border: `1px solid ${color.border}`,
                                        padding: "4px",
                                      }}
                                    >
                                      <div style={{ fontSize: "10px" }}>
                                        <strong>{rule.target_type}:</strong>{" "}
                                        {rule.target_key ||
                                          rule.item_ids?.join(", ")}
                                      </div>
                                      {rule.boost_value && (
                                        <div
                                          style={{
                                            fontSize: "10px",
                                            color: color.success,
                                          }}
                                        >
                                          +{rule.boost_value}
                                        </div>
                                      )}
                                    </td>
                                    <td
                                      style={{
                                        border: `1px solid ${color.border}`,
                                        padding: "4px",
                                        fontSize: "10px",
                                      }}
                                    >
                                      {rule.affected_item_ids?.join(", ") ||
                                        "None"}
                                    </td>
                                  </tr>
                                )
                              )}
                            </tbody>
                          </table>
                        </div>
                      </div>
                    )}

                  {/* Rule Effects Per Item */}
                  {trace.extras.rule_effects_per_item && (
                    <div>
                      <div
                        style={{
                          fontSize: "12px",
                          fontWeight: "bold",
                          marginBottom: 4,
                        }}
                      >
                        Rule Effects Per Item
                      </div>
                      <div style={{ overflowX: "auto" }}>
                        <table
                          style={{
                            borderCollapse: "collapse",
                            width: "100%",
                            fontSize: "11px",
                          }}
                        >
                          <thead>
                            <tr>
                              <th
                                style={{
                                  border: `1px solid ${color.border}`,
                                  padding: "4px",
                                  backgroundColor: color.panelSubtle,
                                }}
                              >
                                Item ID
                              </th>
                              <th
                                style={{
                                  border: `1px solid ${color.border}`,
                                  padding: "4px",
                                  backgroundColor: color.panelSubtle,
                                }}
                              >
                                Blocked
                              </th>
                              <th
                                style={{
                                  border: `1px solid ${color.border}`,
                                  padding: "4px",
                                  backgroundColor: color.panelSubtle,
                                }}
                              >
                                Pinned
                              </th>
                              <th
                                style={{
                                  border: `1px solid ${color.border}`,
                                  padding: "4px",
                                  backgroundColor: color.panelSubtle,
                                }}
                              >
                                Boost Delta
                              </th>
                            </tr>
                          </thead>
                          <tbody>
                            {Object.entries(
                              trace.extras.rule_effects_per_item
                            ).map(([itemId, effect]: [string, any]) => (
                              <tr key={itemId}>
                                <td
                                  style={{
                                    border: `1px solid ${color.border}`,
                                    padding: "4px",
                                  }}
                                >
                                  <code style={{ fontSize: "10px" }}>
                                    {itemId}
                                  </code>
                                </td>
                                <td
                                  style={{
                                    border: `1px solid ${color.border}`,
                                    padding: "4px",
                                    textAlign: "center",
                                  }}
                                >
                                  <span
                                    style={{
                                      color: effect.blocked
                                        ? color.danger
                                        : color.success,
                                    }}
                                  >
                                    {effect.blocked ? "✓" : "✗"}
                                  </span>
                                </td>
                                <td
                                  style={{
                                    border: `1px solid ${color.border}`,
                                    padding: "4px",
                                    textAlign: "center",
                                  }}
                                >
                                  <span
                                    style={{
                                      color: effect.pinned
                                        ? color.warning
                                        : color.textMuted,
                                    }}
                                  >
                                    {effect.pinned ? "✓" : "✗"}
                                  </span>
                                </td>
                                <td
                                  style={{
                                    border: `1px solid ${color.border}`,
                                    padding: "4px",
                                    textAlign: "center",
                                  }}
                                >
                                  {effect.boost_delta !== 0 ? (
                                    <span
                                      style={{
                                        color:
                                          effect.boost_delta > 0
                                            ? color.success
                                            : color.danger,
                                      }}
                                    >
                                      {effect.boost_delta > 0 ? "+" : ""}
                                      {effect.boost_delta}
                                    </span>
                                  ) : (
                                    "0"
                                  )}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}

          {/* Effective Config */}
          <div>
            <h3
              style={{
                margin: "0 0 8px 0",
                fontSize: "14px",
                fontWeight: "bold",
              }}
            >
              Effective Config
            </h3>
            <div
              style={{
                backgroundColor: "#e3f2fd",
                border: "1px solid #bbdefb",
                borderRadius: "4px",
                padding: "12px",
              }}
            >
              <Code>
                {JSON.stringify(formatConfig(trace.effective_config), null, 2)}
              </Code>
            </div>
          </div>

          {/* Bandit Context */}
          {trace.bandit && (
            <div>
              <h3
                style={{
                  margin: "0 0 8px 0",
                  fontSize: "14px",
                  fontWeight: "bold",
                }}
              >
                Bandit Context
              </h3>
              <div
                style={{
                  backgroundColor: color.warningBg,
                  border: `1px solid ${color.warning}`,
                  borderRadius: "4px",
                  padding: "12px",
                }}
              >
                <Code>
                  {JSON.stringify(formatBandit(trace.bandit), null, 2)}
                </Code>
              </div>
            </div>
          )}

          {/* Candidates (pre-MMR) */}
          <div>
            <h3
              style={{
                margin: "0 0 8px 0",
                fontSize: "14px",
                fontWeight: "bold",
              }}
            >
              Candidates (pre-MMR)
            </h3>
            <div style={{ overflowX: "auto" }}>
              <table
                style={{
                  borderCollapse: "collapse",
                  width: "100%",
                  fontSize: "12px",
                }}
              >
                <thead>
                  <tr>
                    <Th>Item ID</Th>
                    <Th>Score</Th>
                  </tr>
                </thead>
                <tbody>
                  {(trace.candidates_pre || []).map((candidate, index) => (
                    <tr key={index}>
                      <Td mono>{candidate.item_id}</Td>
                      <Td>{candidate.score?.toFixed(6) || "N/A"}</Td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* MMR & Caps Details */}
          {trace.mmr_info && trace.mmr_info.length > 0 && (
            <div>
              <h3
                style={{
                  margin: "0 0 8px 0",
                  fontSize: "14px",
                  fontWeight: "bold",
                }}
              >
                MMR & Caps Details
              </h3>
              <div style={{ overflowX: "auto" }}>
                <table
                  style={{
                    borderCollapse: "collapse",
                    width: "100%",
                    fontSize: "12px",
                  }}
                >
                  <thead>
                    <tr>
                      <Th>Pick</Th>
                      <Th>Item ID</Th>
                      <Th>Max Sim</Th>
                      <Th>Relevance</Th>
                      <Th>Penalty</Th>
                      <Th>Brand Cap</Th>
                      <Th>Category Cap</Th>
                    </tr>
                  </thead>
                  <tbody>
                    {trace.mmr_info.map((mmr, index) => (
                      <tr key={index}>
                        <Td>{mmr.pick_index}</Td>
                        <Td mono>{mmr.item_id}</Td>
                        <Td>{mmr.max_sim?.toFixed(4) || "N/A"}</Td>
                        <Td>{mmr.relevance?.toFixed(4) || "N/A"}</Td>
                        <Td>{mmr.penalty?.toFixed(4) || "N/A"}</Td>
                        <Td>
                          {mmr.brand_cap_hit !== undefined
                            ? mmr.brand_cap_hit
                              ? "HIT"
                              : "OK"
                            : "N/A"}
                        </Td>
                        <Td>
                          {mmr.category_cap_hit !== undefined
                            ? mmr.category_cap_hit
                              ? "HIT"
                              : "OK"
                            : "N/A"}
                        </Td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Final Items */}
          <div>
            <h3
              style={{
                margin: "0 0 8px 0",
                fontSize: "14px",
                fontWeight: "bold",
              }}
            >
              Final Items + Reasons
            </h3>
            <div style={{ overflowX: "auto" }}>
              <table
                style={{
                  borderCollapse: "collapse",
                  width: "100%",
                  fontSize: "12px",
                }}
              >
                <thead>
                  <tr>
                    <Th>Rank</Th>
                    <Th>Item ID</Th>
                    <Th>Score</Th>
                    <Th>Reasons</Th>
                  </tr>
                </thead>
                <tbody>
                  {(trace.final_items || []).map((item, index) => (
                    <tr key={index}>
                      <Td>{index + 1}</Td>
                      <Td mono>{item.item_id}</Td>
                      <Td>{item.score?.toFixed(6) || "N/A"}</Td>
                      <Td>
                        {item.reasons && item.reasons.length > 0
                          ? item.reasons.join(", ")
                          : "N/A"}
                      </Td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Constraints */}
          {trace.constraints && (
            <div>
              <h3
                style={{
                  margin: "0 0 8px 0",
                  fontSize: "14px",
                  fontWeight: "bold",
                }}
              >
                Constraints
              </h3>
              <div
                style={{
                  backgroundColor: color.panelSubtle,
                  border: `1px solid ${color.panelBorder}`,
                  borderRadius: "4px",
                  padding: "12px",
                }}
              >
                <Code>{JSON.stringify(trace.constraints, null, 2)}</Code>
              </div>
            </div>
          )}

          {/* Extras */}
          {trace.extras && Object.keys(trace.extras).length > 0 && (
            <div>
              <h3
                style={{
                  margin: "0 0 8px 0",
                  fontSize: "14px",
                  fontWeight: "bold",
                }}
              >
                Additional Info
              </h3>
              <div
                style={{
                  backgroundColor: color.panelSubtle,
                  border: `1px solid ${color.panelBorder}`,
                  borderRadius: "4px",
                  padding: "12px",
                }}
              >
                <Code>{JSON.stringify(trace.extras, null, 2)}</Code>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
