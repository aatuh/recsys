/**
 * Feature flags configuration panel for development and testing.
 */

import React from "react";
import { useFeatureFlags } from "../contexts/FeatureFlagsContext";
import { Button } from "./primitives/UIComponents";

export function FeatureFlagsPanel() {
  const { flags, updateFlag, resetFlags } = useFeatureFlags();

  const handleToggle = (flag: keyof typeof flags) => {
    updateFlag(flag, !flags[flag]);
  };

  const handleReset = () => {
    resetFlags();
  };

  return (
    <div style={{ padding: "20px", border: "1px solid #ccc", margin: "10px" }}>
      <h3>Feature Flags</h3>
      <p style={{ fontSize: "14px", color: "#666", marginBottom: "20px" }}>
        Toggle feature flags to control application behavior. Changes are
        persisted to storage.
      </p>

      <div style={{ display: "grid", gap: "15px" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          <input
            type="checkbox"
            id="authEnabled"
            checked={flags.authEnabled}
            onChange={() => handleToggle("authEnabled")}
          />
          <label htmlFor="authEnabled">
            <strong>Authentication Enabled</strong>
            <br />
            <small style={{ color: "#666" }}>
              Enable authentication features and user management
            </small>
          </label>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          <input
            type="checkbox"
            id="useRemoteEmbeddings"
            checked={flags.useRemoteEmbeddings}
            onChange={() => handleToggle("useRemoteEmbeddings")}
          />
          <label htmlFor="useRemoteEmbeddings">
            <strong>Remote Embeddings</strong>
            <br />
            <small style={{ color: "#666" }}>
              Use remote embedding service instead of local computation
            </small>
          </label>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          <input
            type="checkbox"
            id="circuitBreakerEnabled"
            checked={flags.circuitBreakerEnabled}
            onChange={() => handleToggle("circuitBreakerEnabled")}
          />
          <label htmlFor="circuitBreakerEnabled">
            <strong>Circuit Breaker</strong>
            <br />
            <small style={{ color: "#666" }}>
              Enable circuit breaker for HTTP resilience
            </small>
          </label>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          <input
            type="checkbox"
            id="retryEnabled"
            checked={flags.retryEnabled}
            onChange={() => handleToggle("retryEnabled")}
          />
          <label htmlFor="retryEnabled">
            <strong>Retry Logic</strong>
            <br />
            <small style={{ color: "#666" }}>
              Enable automatic retry for failed requests
            </small>
          </label>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          <input
            type="checkbox"
            id="analyticsEnabled"
            checked={flags.analyticsEnabled}
            onChange={() => handleToggle("analyticsEnabled")}
          />
          <label htmlFor="analyticsEnabled">
            <strong>Analytics</strong>
            <br />
            <small style={{ color: "#666" }}>
              Enable analytics tracking and logging
            </small>
          </label>
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          <input
            type="checkbox"
            id="debugMode"
            checked={flags.debugMode}
            onChange={() => handleToggle("debugMode")}
          />
          <label htmlFor="debugMode">
            <strong>Debug Mode</strong>
            <br />
            <small style={{ color: "#666" }}>
              Enable debug logging and development features
            </small>
          </label>
        </div>
      </div>

      <div style={{ marginTop: "20px", display: "flex", gap: "10px" }}>
        <Button
          onClick={handleReset}
          style={{ backgroundColor: "#6c757d", color: "white" }}
        >
          Reset to Defaults
        </Button>
      </div>

      <div style={{ marginTop: "20px", fontSize: "12px", color: "#666" }}>
        <strong>Current Flags:</strong>
        <pre
          style={{
            backgroundColor: "#f5f5f5",
            padding: "10px",
            borderRadius: "4px",
            marginTop: "5px",
            overflow: "auto",
          }}
        >
          {JSON.stringify(flags, null, 2)}
        </pre>
      </div>
    </div>
  );
}
