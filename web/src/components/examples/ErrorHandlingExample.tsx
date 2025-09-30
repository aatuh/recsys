/**
 * Example component demonstrating error handling, route guards, toasts, and telemetry.
 */

import React, { useState } from "react";
import { GuardedRoute } from "../GuardedRoute";
import { useToast } from "../../contexts/ToastContext";
import { useTelemetry, TELEMETRY_EVENTS } from "../../hooks/useTelemetry";
import { Button } from "../primitives/UIComponents";

export function ErrorHandlingExample() {
  const [logs, setLogs] = useState<string[]>([]);
  const [hasAccess, setHasAccess] = useState(true);
  const [throwError, setThrowError] = useState(false);

  const toast = useToast();
  const telemetry = useTelemetry();

  const addLog = (message: string) => {
    setLogs((prev) => [...prev, `${new Date().toISOString()}: ${message}`]);
  };

  // Simulate an error for testing error boundary
  if (throwError) {
    throw new Error("This is a test error for the error boundary!");
  }

  const handleShowSuccessToast = () => {
    toast.showSuccess("Success!", "Operation completed successfully");
    addLog("✅ Success toast shown");
    telemetry.track(TELEMETRY_EVENTS.BUTTON_CLICK, {
      action: "show_success_toast",
    });
  };

  const handleShowErrorToast = () => {
    toast.showError("Error!", "Something went wrong", {
      action: {
        label: "Retry",
        onClick: () => {
          addLog("🔄 Retry action clicked");
        },
      },
    });
    addLog("❌ Error toast shown");
    telemetry.track(TELEMETRY_EVENTS.BUTTON_CLICK, {
      action: "show_error_toast",
    });
  };

  const handleShowWarningToast = () => {
    toast.showWarning("Warning!", "Please check your input");
    addLog("⚠️ Warning toast shown");
    telemetry.track(TELEMETRY_EVENTS.BUTTON_CLICK, {
      action: "show_warning_toast",
    });
  };

  const handleShowInfoToast = () => {
    toast.showInfo("Info", "Here's some useful information");
    addLog("ℹ️ Info toast shown");
    telemetry.track(TELEMETRY_EVENTS.BUTTON_CLICK, {
      action: "show_info_toast",
    });
  };

  const handleToggleAccess = () => {
    setHasAccess(!hasAccess);
    addLog(`🔒 Access ${hasAccess ? "denied" : "granted"}`);
    telemetry.track(TELEMETRY_EVENTS.FEATURE_USED, {
      feature: "access_toggle",
      hasAccess: !hasAccess,
    });
  };

  const handleThrowError = () => {
    setThrowError(true);
    addLog("💥 Error thrown for testing");
    telemetry.track(TELEMETRY_EVENTS.ERROR_OCCURRED, {
      type: "test_error",
      message: "User triggered test error",
    });
  };

  const handleTrackCustomEvent = () => {
    telemetry.track("custom_event", {
      component: "ErrorHandlingExample",
      action: "custom_tracking",
      timestamp: Date.now(),
    });
    addLog("📊 Custom telemetry event tracked");
  };

  const handleIdentifyUser = () => {
    telemetry.identify("demo_user_123", {
      name: "Demo User",
      role: "developer",
      environment: "demo",
    });
    addLog("👤 User identified in telemetry");
  };

  const handlePageView = () => {
    telemetry.page("Error Handling Example", {
      section: "demo",
      features: ["toasts", "guards", "telemetry"],
    });
    addLog("📄 Page view tracked");
  };

  const handleClearLogs = () => {
    setLogs([]);
    addLog("🧹 Logs cleared");
  };

  return (
    <div style={{ padding: "20px", border: "1px solid #ccc", margin: "10px" }}>
      <h3>Error Handling & UX Features Example</h3>
      <p style={{ fontSize: "14px", color: "#666", marginBottom: "20px" }}>
        Demonstrates error boundaries, route guards, toast notifications, and
        telemetry tracking.
      </p>

      <div style={{ marginBottom: "20px" }}>
        <h4>Toast Notifications</h4>
        <div
          style={{
            display: "flex",
            gap: "10px",
            flexWrap: "wrap",
            marginBottom: "10px",
          }}
        >
          <Button onClick={handleShowSuccessToast}>Show Success Toast</Button>
          <Button onClick={handleShowErrorToast}>Show Error Toast</Button>
          <Button onClick={handleShowWarningToast}>Show Warning Toast</Button>
          <Button onClick={handleShowInfoToast}>Show Info Toast</Button>
        </div>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Route Guards</h4>
        <div
          style={{
            display: "flex",
            gap: "10px",
            alignItems: "center",
            marginBottom: "10px",
          }}
        >
          <Button onClick={handleToggleAccess}>
            {hasAccess ? "Deny Access" : "Grant Access"}
          </Button>
          <span style={{ fontSize: "14px", color: "#666" }}>
            Current access: {hasAccess ? "Granted" : "Denied"}
          </span>
        </div>

        <GuardedRoute
          canActivate={() => hasAccess}
          fallback={
            <div
              style={{
                padding: "20px",
                backgroundColor: "#f8f9fa",
                borderRadius: "4px",
                textAlign: "center",
                color: "#6c757d",
              }}
            >
              🔒 This content is protected by route guard
            </div>
          }
        >
          <div
            style={{
              padding: "20px",
              backgroundColor: "#d4edda",
              borderRadius: "4px",
              textAlign: "center",
              color: "#155724",
            }}
          >
            ✅ Protected content is accessible
          </div>
        </GuardedRoute>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Telemetry Tracking</h4>
        <div
          style={{
            display: "flex",
            gap: "10px",
            flexWrap: "wrap",
            marginBottom: "10px",
          }}
        >
          <Button onClick={handleTrackCustomEvent}>Track Custom Event</Button>
          <Button onClick={handleIdentifyUser}>Identify User</Button>
          <Button onClick={handlePageView}>Track Page View</Button>
        </div>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Error Boundary Testing</h4>
        <div style={{ display: "flex", gap: "10px", marginBottom: "10px" }}>
          <Button
            onClick={handleThrowError}
            style={{ backgroundColor: "#dc3545", color: "white" }}
          >
            Throw Test Error
          </Button>
        </div>
        <p style={{ fontSize: "12px", color: "#666" }}>
          ⚠️ This will trigger the error boundary. The page will show a fallback
          UI.
        </p>
      </div>

      <div>
        <h4>Activity Logs</h4>
        <div style={{ display: "flex", gap: "10px", marginBottom: "10px" }}>
          <Button
            onClick={handleClearLogs}
            style={{ backgroundColor: "#6c757d", color: "white" }}
          >
            Clear Logs
          </Button>
        </div>
        <div
          style={{
            height: "200px",
            overflow: "auto",
            border: "1px solid #ddd",
            padding: "10px",
            backgroundColor: "#f9f9f9",
            fontFamily: "monospace",
            fontSize: "12px",
          }}
        >
          {logs.map((log, index) => (
            <div key={index} style={{ marginBottom: "2px" }}>
              {log}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
