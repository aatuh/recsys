import React, { useState } from "react";
import { Section, Label, Button } from "../primitives/UIComponents";
import { color } from "../../ui/tokens";
import {
  BanditService,
  type types_BanditRewardRequest,
} from "../../lib/api-client";
import { useViewState } from "../../contexts/ViewStateContext";

interface RewardEntry {
  id: string;
  timestamp: Date;
  requestId: string;
  policyId: string;
  surface: string;
  bucketKey: string;
  algorithm: string;
  reward: boolean;
  success: boolean;
  error?: string;
}

interface RewardFeedbackSectionProps {
  namespace: string;
}

export function RewardFeedbackSection({
  namespace,
}: RewardFeedbackSectionProps) {
  const { banditPlayground, setBanditPlayground } = useViewState();
  const { decisionResult, recommendationResult } = banditPlayground;

  const [reward, setReward] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [rewardHistory, setRewardHistory] = useState<RewardEntry[]>([]);

  // Get the current decision result from either decision section or recommendation section
  const getCurrentDecisionResult = () => {
    // Prefer recommendation result if available (more recent)
    if (recommendationResult && recommendationResult.chosen_policy_id) {
      return {
        policy_id: recommendationResult.chosen_policy_id,
        surface: "homepage", // Recommendation results don't include surface, use default
        bucket_key: recommendationResult.bandit_bucket,
        algorithm: recommendationResult.algorithm,
        request_id: undefined, // Recommendation results don't include request_id
        explore: recommendationResult.explore,
        explain: recommendationResult.bandit_explain,
      };
    }
    // Fall back to decision result
    return decisionResult;
  };

  const currentDecisionResult = getCurrentDecisionResult();
  const canSendReward =
    currentDecisionResult && currentDecisionResult.policy_id;

  const handleSendReward = async (rewardValue: boolean) => {
    if (!currentDecisionResult) {
      setError("No decision result available. Please make a decision first.");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const request: types_BanditRewardRequest = {
        namespace,
        surface: currentDecisionResult.surface || "homepage",
        policy_id: currentDecisionResult.policy_id,
        bucket_key: currentDecisionResult.bucket_key,
        algorithm: currentDecisionResult.algorithm,
        request_id: (currentDecisionResult as any).request_id || undefined,
        reward: rewardValue,
      };

      console.log("Sending reward:", request);
      const response = await BanditService.postV1BanditReward(request);
      console.log("Reward response:", response);

      // Add to reward history
      const rewardEntry: RewardEntry = {
        id: `reward-${Date.now()}`,
        timestamp: new Date(),
        requestId: (currentDecisionResult as any).request_id || "N/A",
        policyId: currentDecisionResult.policy_id || "N/A",
        surface: currentDecisionResult.surface || "N/A",
        bucketKey: currentDecisionResult.bucket_key || "N/A",
        algorithm: currentDecisionResult.algorithm || "N/A",
        reward: rewardValue,
        success: true,
      };

      setRewardHistory((prev) => [rewardEntry, ...prev]);
      // Append to shared reward history for dashboards
      setBanditPlayground((prev) => ({
        ...prev,
        rewardHistory: [
          {
            id: rewardEntry.id,
            timestamp: rewardEntry.timestamp,
            requestId: rewardEntry.requestId,
            policyId: rewardEntry.policyId,
            surface: rewardEntry.surface,
            bucketKey: rewardEntry.bucketKey,
            algorithm: rewardEntry.algorithm,
            reward: rewardValue,
            success: true,
          },
          ...prev.rewardHistory,
        ].slice(0, 1000),
      }));
      setReward(rewardValue);

      // Clear the reward selection after a short delay
      setTimeout(() => {
        setReward(null);
      }, 2000);
    } catch (err) {
      console.error("Failed to send reward:", err);
      const errorMessage =
        err instanceof Error ? err.message : "Failed to send reward";

      // Add failed reward to history
      const rewardEntry: RewardEntry = {
        id: `reward-${Date.now()}`,
        timestamp: new Date(),
        requestId: (currentDecisionResult as any).request_id || "N/A",
        policyId: currentDecisionResult.policy_id || "N/A",
        surface: currentDecisionResult.surface || "N/A",
        bucketKey: currentDecisionResult.bucket_key || "N/A",
        algorithm: currentDecisionResult.algorithm || "N/A",
        reward: rewardValue,
        success: false,
        error: errorMessage,
      };

      setRewardHistory((prev) => [rewardEntry, ...prev]);
      setBanditPlayground((prev) => ({
        ...prev,
        rewardHistory: [
          {
            id: rewardEntry.id,
            timestamp: rewardEntry.timestamp,
            requestId: rewardEntry.requestId,
            policyId: rewardEntry.policyId,
            surface: rewardEntry.surface,
            bucketKey: rewardEntry.bucketKey,
            algorithm: rewardEntry.algorithm,
            reward: rewardValue,
            success: false,
            error: errorMessage,
          },
          ...prev.rewardHistory,
        ].slice(0, 1000),
      }));
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const clearHistory = () => {
    setRewardHistory([]);
  };

  const formatTimestamp = (date: Date) => {
    return date.toLocaleTimeString();
  };

  return (
    <Section title="Reward Feedback">
      <div style={{ marginBottom: 16 }}>
        <p style={{ color: "#666", fontSize: 14, marginBottom: 16 }}>
          Send reward feedback for the last decision to help the bandit
          algorithm learn and adapt. Positive rewards (true) indicate successful
          outcomes, while negative rewards (false) indicate unsuccessful
          outcomes.
        </p>

        {/* Current Decision Info */}
        {currentDecisionResult && (
          <div
            style={{
              backgroundColor: "#f8f9fa",
              border: "1px solid #e9ecef",
              borderRadius: 4,
              padding: 12,
              marginBottom: 16,
            }}
          >
            <div
              style={{
                fontSize: 12,
                fontWeight: "bold",
                marginBottom: 8,
                color: "#495057",
              }}
            >
              Last Decision
            </div>
            <div
              style={{
                fontSize: 12,
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: 8,
              }}
            >
              <div>
                <strong>Policy:</strong> {currentDecisionResult.policy_id}
              </div>
              <div>
                <strong>Surface:</strong> {currentDecisionResult.surface}
              </div>
              <div>
                <strong>Algorithm:</strong> {currentDecisionResult.algorithm}
              </div>
              <div>
                <strong>Request ID:</strong>{" "}
                {(currentDecisionResult as any).request_id || "N/A"}
              </div>
            </div>
          </div>
        )}

        {/* Reward Buttons */}
        <div style={{ marginBottom: 20 }}>
          <Label text="Send Reward Feedback">
            <div style={{ display: "flex", gap: 12, alignItems: "center" }}>
              <Button
                onClick={() => handleSendReward(true)}
                disabled={!canSendReward || loading}
                style={{
                  backgroundColor: color.success,
                  color: color.primaryTextOn,
                  border: "none",
                  padding: "8px 16px",
                  borderRadius: 4,
                  cursor: canSendReward && !loading ? "pointer" : "not-allowed",
                  fontSize: 14,
                  opacity: canSendReward && !loading ? 1 : 0.6,
                  fontWeight: "bold",
                }}
                aria-label="Send positive reward"
              >
                {loading ? "Sending..." : "✓ Positive Reward"}
              </Button>
              <Button
                onClick={() => handleSendReward(false)}
                disabled={!canSendReward || loading}
                style={{
                  backgroundColor: color.danger,
                  color: color.primaryTextOn,
                  border: "none",
                  padding: "8px 16px",
                  borderRadius: 4,
                  cursor: canSendReward && !loading ? "pointer" : "not-allowed",
                  fontSize: 14,
                  opacity: canSendReward && !loading ? 1 : 0.6,
                  fontWeight: "bold",
                }}
                aria-label="Send negative reward"
              >
                {loading ? "Sending..." : "✗ Negative Reward"}
              </Button>
              {!canSendReward && (
                <span
                  style={{
                    fontSize: 12,
                    color: "#6c757d",
                    fontStyle: "italic",
                  }}
                >
                  Make a decision first to send rewards
                </span>
              )}
            </div>
          </Label>
        </div>

        {/* Success/Error Messages */}
        {reward !== null && (
          <div
            style={{
              backgroundColor: reward ? color.successBg : color.dangerBg,
              color: color.text,
              border: `1px solid ${reward ? color.success : color.danger}`,
              borderRadius: 4,
              padding: 12,
              marginBottom: 16,
              fontSize: 14,
            }}
          >
            {reward
              ? "✓ Positive reward sent successfully!"
              : "✗ Negative reward sent successfully!"}
          </div>
        )}

        {error && (
          <div
            style={{
              backgroundColor: color.dangerBg,
              color: color.text,
              border: `1px solid ${color.danger}`,
              borderRadius: 4,
              padding: 12,
              marginBottom: 16,
              fontSize: 14,
            }}
          >
            Error: {error}
          </div>
        )}

        {/* Reward History */}
        <div style={{ marginBottom: 20 }}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: 8,
            }}
          >
            <Label text={`Reward History (${rewardHistory.length})`}>
              <div></div>
            </Label>
            {rewardHistory.length > 0 && (
              <Button
                onClick={clearHistory}
                style={{
                  backgroundColor: "#6c757d",
                  color: "white",
                  border: "none",
                  padding: "4px 8px",
                  borderRadius: 4,
                  cursor: "pointer",
                  fontSize: 12,
                }}
              >
                Clear History
              </Button>
            )}
          </div>

          {rewardHistory.length > 0 ? (
            <div
              style={{
                backgroundColor: color.panelSubtle,
                border: `1px solid ${color.panelBorder}`,
                borderRadius: 4,
                maxHeight: 300,
                overflowY: "auto",
              }}
            >
              {rewardHistory.map((entry) => (
                <div
                  key={entry.id}
                  style={{
                    padding: 12,
                    borderBottom: "1px solid #e9ecef",
                    fontSize: 12,
                  }}
                >
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "flex-start",
                      marginBottom: 4,
                    }}
                  >
                    <div
                      style={{ display: "flex", alignItems: "center", gap: 8 }}
                    >
                      <span
                        style={{
                          backgroundColor: entry.reward
                            ? color.success
                            : color.danger,
                          color: color.primaryTextOn,
                          padding: "2px 6px",
                          borderRadius: 2,
                          fontSize: 10,
                          fontWeight: "bold",
                        }}
                      >
                        {entry.reward ? "✓ POSITIVE" : "✗ NEGATIVE"}
                      </span>
                      <span
                        style={{
                          backgroundColor: entry.success
                            ? color.successBg
                            : color.dangerBg,
                          color: color.text,
                          padding: "2px 6px",
                          borderRadius: 2,
                          fontSize: 10,
                          fontWeight: "bold",
                        }}
                      >
                        {entry.success ? "SUCCESS" : "FAILED"}
                      </span>
                    </div>
                    <div style={{ color: "#6c757d", fontSize: 10 }}>
                      {formatTimestamp(entry.timestamp)}
                    </div>
                  </div>

                  <div
                    style={{
                      display: "grid",
                      gridTemplateColumns: "1fr 1fr",
                      gap: 8,
                      marginBottom: 4,
                    }}
                  >
                    <div>
                      <strong>Policy:</strong> {entry.policyId}
                    </div>
                    <div>
                      <strong>Surface:</strong> {entry.surface}
                    </div>
                    <div>
                      <strong>Algorithm:</strong> {entry.algorithm}
                    </div>
                    <div>
                      <strong>Request ID:</strong> {entry.requestId}
                    </div>
                  </div>

                  <div
                    style={{
                      fontSize: 10,
                      color: "#6c757d",
                      fontFamily: "monospace",
                      wordBreak: "break-all",
                    }}
                  >
                    <strong>Bucket:</strong> {entry.bucketKey}
                  </div>

                  {entry.error && (
                    <div
                      style={{
                        fontSize: 10,
                        color: color.danger,
                        marginTop: 4,
                      }}
                    >
                      <strong>Error:</strong> {entry.error}
                    </div>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <div
              style={{
                backgroundColor: color.panelSubtle,
                border: `1px solid ${color.panelBorder}`,
                borderRadius: 4,
                padding: 16,
                textAlign: "center",
                color: color.textMuted,
                fontSize: 14,
                fontStyle: "italic",
              }}
            >
              No rewards sent yet. Make a decision and send reward feedback to
              see history.
            </div>
          )}
        </div>
      </div>
    </Section>
  );
}
