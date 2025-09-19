import React, { useState } from "react";
import { Section, Row, Label, Button } from "./UIComponents";
import {
  BanditService,
  type types_BanditDecideRequest,
  type types_BanditDecideResponse,
  type types_BanditPolicy,
  AuditService,
} from "../lib/api-client";
import {
  useViewState,
  type DecisionContext,
  type BanditDecisionEntry,
} from "../contexts/ViewStateContext";
import { DecisionTraceDrawer } from "./DecisionTraceDrawer";

interface DecisionSectionProps {
  namespace: string;
  availablePolicies: types_BanditPolicy[];
}

export function DecisionSection({
  namespace,
  availablePolicies,
}: DecisionSectionProps) {
  const { banditPlayground, setBanditPlayground } = useViewState();

  // Decision trace drawer state
  const [showDecisionTrace, setShowDecisionTrace] = useState(false);
  const [currentDecisionId, setCurrentDecisionId] = useState<string | null>(
    null
  );

  // Use context state instead of local state
  const {
    surface,
    context,
    candidatePolicyIds,
    algorithm,
    requestId,
    decisionResult,
    loading,
    error,
  } = banditPlayground;

  // Helper functions to update context state
  const setSurface = (value: string) => {
    setBanditPlayground((prev) => ({ ...prev, surface: value }));
  };

  const setContext = (
    value: DecisionContext | ((prev: DecisionContext) => DecisionContext)
  ) => {
    setBanditPlayground((prev) => ({
      ...prev,
      context: typeof value === "function" ? value(prev.context) : value,
    }));
  };

  const setCandidatePolicyIds = (
    value: string[] | ((prev: string[]) => string[])
  ) => {
    setBanditPlayground((prev) => ({
      ...prev,
      candidatePolicyIds:
        typeof value === "function" ? value(prev.candidatePolicyIds) : value,
    }));
  };

  const setAlgorithm = (value: string) => {
    setBanditPlayground((prev) => ({ ...prev, algorithm: value }));
  };

  const setRequestId = (value: string) => {
    setBanditPlayground((prev) => ({ ...prev, requestId: value }));
  };

  const setDecisionResult = (value: types_BanditDecideResponse | null) => {
    setBanditPlayground((prev) => ({ ...prev, decisionResult: value }));
  };

  const setLoading = (value: boolean) => {
    setBanditPlayground((prev) => ({ ...prev, loading: value }));
  };

  const setError = (value: string | null) => {
    setBanditPlayground((prev) => ({ ...prev, error: value }));
  };

  // Note: We don't auto-populate candidate policies anymore
  // Users can choose to select specific policies or let backend use all active policies

  const handleContextChange = (key: string, value: string) => {
    setContext((prev) => ({
      ...prev,
      [key]: value,
    }));
  };

  const handleAddContextField = () => {
    const newKey = prompt("Enter context key:");
    if (newKey && newKey.trim()) {
      setContext((prev) => ({
        ...prev,
        [newKey.trim()]: "",
      }));
    }
  };

  const handleRemoveContextField = (key: string) => {
    if (key === "device" || key === "locale") {
      return; // Don't allow removing required fields
    }
    setContext((prev) => {
      const newContext = { ...prev };
      delete newContext[key];
      return newContext;
    });
  };

  const handlePolicyToggle = (policyId: string) => {
    setCandidatePolicyIds((prev) => {
      if (prev.includes(policyId)) {
        return prev.filter((id) => id !== policyId);
      } else {
        return [...prev, policyId];
      }
    });
  };

  const handleSelectAllPolicies = () => {
    setCandidatePolicyIds(
      availablePolicies
        .map((p) => p.policy_id)
        .filter((id): id is string => id !== undefined)
    );
  };

  const handleClearAllPolicies = () => {
    setCandidatePolicyIds([]);
  };

  const simulateDecision = async () => {
    // Allow empty candidate policy list - backend will use all active policies

    setLoading(true);
    setError(null);
    setDecisionResult(null);

    try {
      const request: types_BanditDecideRequest = {
        namespace,
        surface,
        context,
        candidate_policy_ids: candidatePolicyIds,
        algorithm,
        request_id: requestId || undefined,
      };

      console.log("Making decision request:", request);
      const response = await BanditService.postV1BanditDecide(request);
      console.log("Decision response:", response);
      console.log("Response type:", typeof response);

      // Parse the response if it's a string
      let parsedResponse = response;
      if (typeof response === "string") {
        try {
          parsedResponse = JSON.parse(response);
          console.log("Parsed response:", parsedResponse);
        } catch (e) {
          console.error("Failed to parse response:", e);
          setError("Failed to parse decision response");
          return;
        }
      }

      console.log("Response keys:", Object.keys(parsedResponse || {}));

      setDecisionResult(parsedResponse);

      // Append to decision history for dashboards
      const explain = (parsedResponse as any)?.explain || {};
      const empBest = (
        explain && typeof explain === "object"
          ? (explain as Record<string, string>)["emp_best"]
          : undefined
      ) as string | undefined;
      const entry: BanditDecisionEntry = {
        id: `dec-${Date.now()}`,
        timestamp: new Date(),
        requestId: request.request_id || undefined,
        policyId: ((parsedResponse as any)?.policy_id as string) || "",
        surface: (request.surface as string) || "",
        bucketKey: ((parsedResponse as any)?.bucket_key as string) || "",
        algorithm: String(
          (parsedResponse as any)?.algorithm ?? request.algorithm ?? ""
        ),
        explore: Boolean((parsedResponse as any)?.explore),
        empBestPolicyId: empBest,
        context: ((request as any)?.context ?? {}) as DecisionContext,
      };
      setBanditPlayground((prev) => ({
        ...prev,
        decisionHistory: [entry, ...prev.decisionHistory].slice(0, 500),
      }));
    } catch (err) {
      console.error("Failed to make decision:", err);
      setError(err instanceof Error ? err.message : "Failed to make decision");
    } finally {
      setLoading(false);
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
    // For now, we'll use the first active policy as "empirical best"
    // In a real implementation, this would be based on historical performance
    return availablePolicies.find((p) => p.active) || availablePolicies[0];
  };

  const isExploration = (chosenPolicyId: string) => {
    const empiricalBest = getEmpiricalBestPolicy();
    return empiricalBest && chosenPolicyId !== empiricalBest.policy_id;
  };

  const ensureRequestId = (): string => {
    if (requestId && requestId.trim() !== "") return requestId;
    const rid = `req_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
    setRequestId(rid);
    return rid;
  };

  const handleViewDecisionTrace = async () => {
    try {
      const rid = ensureRequestId();
      // poll the audit list for this request_id (most recent)
      const maxAttempts = 6;
      const delay = (ms: number) => new Promise((r) => setTimeout(r, ms));
      let found: string | null = null;
      for (let i = 0; i < maxAttempts; i++) {
        const list = await AuditService.getV1AuditDecisions(
          namespace,
          undefined,
          undefined,
          undefined,
          rid,
          1
        );
        const dec = list.decisions?.[0];
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
        setError(
          "Decision trace not yet available. Try again in a moment (async write)."
        );
      }
    } catch {
      setError("Failed to load decision trace index");
    }
  };

  const handleCloseDecisionTrace = () => {
    setShowDecisionTrace(false);
    setCurrentDecisionId(null);
  };

  return (
    <Section title="Bandit Decision Simulator">
      <div style={{ marginBottom: 16 }}>
        <p style={{ color: "#666", fontSize: 14, marginBottom: 16 }}>
          Simulate bandit decisions to test policy selection and explore vs
          exploit behavior. The system will choose a policy based on the current
          context and algorithm.
        </p>

        {/* Input Parameters */}
        <div style={{ marginBottom: 20 }}>
          <Label text="Decision Parameters">
            <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
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
                    Surface
                  </label>
                  <input
                    type="text"
                    value={surface}
                    onChange={(e) => setSurface(e.target.value)}
                    style={{
                      width: "100%",
                      padding: "6px 8px",
                      border: "1px solid #ddd",
                      borderRadius: 4,
                      fontSize: 14,
                    }}
                    placeholder="e.g., homepage, search, product"
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
                    Algorithm
                  </label>
                  <select
                    value={algorithm}
                    onChange={(e) => setAlgorithm(e.target.value)}
                    style={{
                      width: "100%",
                      padding: "6px 8px",
                      border: "1px solid #ddd",
                      borderRadius: 4,
                      fontSize: 14,
                    }}
                  >
                    <option value="thompson">Thompson Sampling</option>
                    <option value="ucb1">UCB1</option>
                  </select>
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
                    Request ID (optional)
                  </label>
                  <input
                    type="text"
                    value={requestId}
                    onChange={(e) => setRequestId(e.target.value)}
                    style={{
                      width: "100%",
                      padding: "6px 8px",
                      border: "1px solid #ddd",
                      borderRadius: 4,
                      fontSize: 14,
                    }}
                    placeholder="e.g., req-123"
                  />
                </div>
              </Row>
            </div>
          </Label>
        </div>

        {/* Context */}
        <div style={{ marginBottom: 20 }}>
          <Label text="Context">
            <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
              {Object.entries(context).map(([key, value]) => (
                <Row key={key}>
                  <div style={{ flex: 1 }}>
                    <label
                      style={{
                        display: "block",
                        marginBottom: 4,
                        fontSize: 12,
                        fontWeight: "bold",
                      }}
                    >
                      {key}
                    </label>
                    <div style={{ display: "flex", gap: 4 }}>
                      <input
                        type="text"
                        value={value}
                        onChange={(e) =>
                          handleContextChange(key, e.target.value)
                        }
                        style={{
                          flex: 1,
                          padding: "6px 8px",
                          border: "1px solid #ddd",
                          borderRadius: 4,
                          fontSize: 14,
                        }}
                        placeholder={`Enter ${key} value`}
                      />
                      {key !== "device" && key !== "locale" && (
                        <Button
                          onClick={() => handleRemoveContextField(key)}
                          style={{
                            padding: "6px 8px",
                            backgroundColor: "#dc3545",
                            color: "white",
                            border: "none",
                            borderRadius: 4,
                            fontSize: 12,
                            cursor: "pointer",
                          }}
                          title="Remove context field"
                        >
                          ×
                        </Button>
                      )}
                    </div>
                  </div>
                </Row>
              ))}
              <Button
                onClick={handleAddContextField}
                style={{
                  backgroundColor: "#28a745",
                  color: "white",
                  border: "none",
                  padding: "6px 12px",
                  borderRadius: 4,
                  cursor: "pointer",
                  fontSize: 12,
                  alignSelf: "flex-start",
                }}
              >
                + Add Context Field
              </Button>
            </div>
          </Label>
        </div>

        {/* Candidate Policies */}
        <div style={{ marginBottom: 20 }}>
          <Label
            text={`Candidate Policies (${candidatePolicyIds.length} selected)`}
          >
            <div style={{ marginBottom: 8, fontSize: 12, color: "#666" }}>
              {candidatePolicyIds.length === 0
                ? "No policies selected - backend will use all active policies automatically"
                : "Selected policies will be used for the decision"}
            </div>
            <div style={{ marginBottom: 8 }}>
              <Button
                onClick={handleSelectAllPolicies}
                style={{
                  backgroundColor: "#007acc",
                  color: "white",
                  border: "none",
                  padding: "4px 8px",
                  borderRadius: 4,
                  cursor: "pointer",
                  fontSize: 12,
                  marginRight: 8,
                }}
              >
                Select All
              </Button>
              <Button
                onClick={handleClearAllPolicies}
                style={{
                  backgroundColor:
                    candidatePolicyIds.length === 0 ? "#28a745" : "#6c757d",
                  color: "white",
                  border: "none",
                  padding: "4px 8px",
                  borderRadius: 4,
                  cursor: "pointer",
                  fontSize: 12,
                }}
              >
                {candidatePolicyIds.length === 0
                  ? "Use All Active (Default)"
                  : "Clear All"}
              </Button>
            </div>
            <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
              {availablePolicies.map((policy) => (
                <label
                  key={policy.policy_id}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: 4,
                    padding: "4px 8px",
                    border: "1px solid #ddd",
                    borderRadius: 4,
                    backgroundColor: candidatePolicyIds.includes(
                      policy.policy_id || ""
                    )
                      ? "#e3f2fd"
                      : "#fff",
                    cursor: "pointer",
                    fontSize: 12,
                  }}
                >
                  <input
                    type="checkbox"
                    checked={candidatePolicyIds.includes(
                      policy.policy_id || ""
                    )}
                    onChange={() => handlePolicyToggle(policy.policy_id || "")}
                    style={{ margin: 0 }}
                  />
                  {policy.name}
                  {!policy.active && (
                    <span style={{ color: "#666", fontSize: 10 }}>
                      (inactive)
                    </span>
                  )}
                </label>
              ))}
            </div>
          </Label>
        </div>

        {/* Action Button */}
        <div style={{ marginBottom: 20 }}>
          <Button
            onClick={simulateDecision}
            disabled={loading}
            style={{
              backgroundColor: loading ? "#6c757d" : "#28a745",
              color: "white",
              border: "none",
              padding: "8px 16px",
              borderRadius: 4,
              cursor: loading ? "not-allowed" : "pointer",
              fontSize: 14,
              opacity: loading ? 0.6 : 1,
            }}
          >
            {loading ? "Making Decision..." : "Simulate Decision"}
          </Button>
        </div>

        {/* Error Display */}
        {error && (
          <div
            style={{
              backgroundColor: "#f8d7da",
              color: "#721c24",
              border: "1px solid #f5c6cb",
              borderRadius: 4,
              padding: 12,
              marginBottom: 16,
              fontSize: 14,
            }}
          >
            Error: {error}
          </div>
        )}

        {/* Decision Result */}
        {decisionResult && (
          <div style={{ marginBottom: 20 }}>
            <Label text="Decision Result">
              <div
                style={{ display: "flex", flexDirection: "column", gap: 12 }}
              >
                {/* View Decision Trace Button */}
                <div style={{ marginBottom: 8 }}>
                  <Button
                    onClick={handleViewDecisionTrace}
                    style={{
                      backgroundColor: "#007acc",
                      color: "white",
                      border: "none",
                      padding: "8px 16px",
                      borderRadius: 4,
                      cursor: "pointer",
                      fontSize: 14,
                    }}
                  >
                    View Decision Trace
                  </Button>
                </div>
                <div
                  style={{
                    backgroundColor: "#f8f9fa",
                    border: "1px solid #e9ecef",
                    borderRadius: 4,
                    padding: 12,
                  }}
                >
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
                        {getPolicyDisplayName(decisionResult.policy_id || "")}
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
                        {decisionResult.algorithm}
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
                        {decisionResult.bucket_key}
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
                          color: decisionResult.explore ? "#ff9800" : "#4caf50",
                          fontWeight: "bold",
                        }}
                      >
                        {decisionResult.explore ? "EXPLORE" : "EXPLOIT"}
                      </div>
                    </div>
                  </div>
                </div>

                {/* Exploration vs Exploitation Analysis */}
                <div
                  style={{
                    backgroundColor: "#fff3cd",
                    border: "1px solid #ffeaa7",
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
                  <div style={{ fontSize: 14 }}>
                    <div style={{ marginBottom: 4 }}>
                      <strong>Empirical Best Policy:</strong>{" "}
                      {getEmpiricalBestPolicy()?.name || "None available"}
                    </div>
                    <div style={{ marginBottom: 4 }}>
                      <strong>Chosen Policy:</strong>{" "}
                      {getPolicyDisplayName(decisionResult.policy_id || "")}
                    </div>
                    <div>
                      <strong>Decision Type:</strong>{" "}
                      <span
                        style={{
                          color: isExploration(decisionResult.policy_id || "")
                            ? "#ff9800"
                            : "#4caf50",
                          fontWeight: "bold",
                        }}
                      >
                        {isExploration(decisionResult.policy_id || "")
                          ? "EXPLORATION"
                          : "EXPLOITATION"}
                      </span>
                      {isExploration(decisionResult.policy_id || "") && (
                        <span
                          style={{ fontSize: 12, color: "#666", marginLeft: 8 }}
                        >
                          (Chose different policy than empirical best)
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                {/* Explanation */}
                {decisionResult.explain &&
                  Object.keys(decisionResult.explain).length > 0 && (
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
                        Algorithm Explanation
                      </div>
                      <div style={{ fontSize: 12, fontFamily: "monospace" }}>
                        {Object.entries(decisionResult.explain).map(
                          ([key, value]) => (
                            <div key={key} style={{ marginBottom: 2 }}>
                              <strong>{key}:</strong> {value}
                            </div>
                          )
                        )}
                      </div>
                    </div>
                  )}
              </div>
            </Label>
          </div>
        )}
      </div>

      {/* Decision Trace Drawer */}
      <DecisionTraceDrawer
        isOpen={showDecisionTrace}
        onClose={handleCloseDecisionTrace}
        decisionId={currentDecisionId}
        namespace={namespace}
      />
    </Section>
  );
}
