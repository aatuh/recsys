/**
 * Session management component for testing and demonstrating session state machine.
 */

import React, { useState, useEffect } from "react";
import {
  useSession,
  useIsAuthenticated,
  useIsExpired,
} from "../contexts/SessionStateMachine";
import { Button } from "./primitives/UIComponents";

export function SessionManager() {
  const {
    state,
    userId,
    sessionId,
    login,
    logout,
    refresh,
    getSessionDuration,
    getInactivityDuration,
  } = useSession();

  const isAuthenticated = useIsAuthenticated();
  const isExpired = useIsExpired();

  const [testUserId, setTestUserId] = useState("user_123");
  const [sessionDuration, setSessionDuration] = useState(0);
  const [inactivityDuration, setInactivityDuration] = useState(0);

  // Update session duration display
  useEffect(() => {
    if (isAuthenticated) {
      const interval = setInterval(() => {
        setSessionDuration(getSessionDuration());
        setInactivityDuration(getInactivityDuration());
      }, 1000);

      return () => clearInterval(interval);
    } else {
      setSessionDuration(0);
      setInactivityDuration(0);
    }
  }, [isAuthenticated, getSessionDuration, getInactivityDuration]);

  const handleLogin = () => {
    if (testUserId.trim()) {
      login(testUserId.trim());
    }
  };

  const handleLogout = () => {
    logout();
  };

  const handleRefresh = () => {
    refresh();
  };

  const formatDuration = (ms: number): string => {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) {
      return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${seconds % 60}s`;
    } else {
      return `${seconds}s`;
    }
  };

  const getStateColor = (state: string): string => {
    switch (state) {
      case "anonymous":
        return "#666";
      case "authenticated":
        return "#28a745";
      case "expired":
        return "#dc3545";
      default:
        return "#666";
    }
  };

  return (
    <div style={{ padding: "20px", border: "1px solid #ccc", margin: "10px" }}>
      <h3>Session State Machine</h3>
      <p style={{ fontSize: "14px", color: "#666", marginBottom: "20px" }}>
        Manage user session state: anonymous → authenticated → expired
      </p>

      <div style={{ marginBottom: "20px" }}>
        <h4>Current State</h4>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "10px",
            marginBottom: "10px",
          }}
        >
          <div
            style={{
              width: "12px",
              height: "12px",
              borderRadius: "50%",
              backgroundColor: getStateColor(state),
            }}
          />
          <span
            style={{
              fontWeight: "bold",
              color: getStateColor(state),
              textTransform: "uppercase",
            }}
          >
            {state}
          </span>
        </div>

        {isAuthenticated && (
          <div style={{ fontSize: "14px", color: "#666" }}>
            <div>
              User ID: <strong>{userId}</strong>
            </div>
            <div>
              Session ID: <code>{sessionId}</code>
            </div>
            <div>
              Session Duration:{" "}
              <strong>{formatDuration(sessionDuration)}</strong>
            </div>
            <div>
              Inactivity: <strong>{formatDuration(inactivityDuration)}</strong>
            </div>
          </div>
        )}

        {isExpired && (
          <div style={{ color: "#dc3545", fontWeight: "bold" }}>
            Session has expired. Please log in again.
          </div>
        )}
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Actions</h4>
        <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
          {!isAuthenticated && (
            <>
              <input
                type="text"
                value={testUserId}
                onChange={(e) => setTestUserId(e.target.value)}
                placeholder="Enter user ID"
                style={{
                  padding: "8px",
                  border: "1px solid #ddd",
                  borderRadius: "4px",
                }}
              />
              <Button onClick={handleLogin}>Login</Button>
            </>
          )}

          {isAuthenticated && (
            <>
              <Button
                onClick={handleRefresh}
                style={{ backgroundColor: "#6c757d", color: "white" }}
              >
                Refresh Session
              </Button>
              <Button
                onClick={handleLogout}
                style={{ backgroundColor: "#dc3545", color: "white" }}
              >
                Logout
              </Button>
            </>
          )}

          {isExpired && <Button onClick={handleLogin}>Login Again</Button>}
        </div>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>State Transitions</h4>
        <div style={{ fontSize: "14px", color: "#666" }}>
          <div>
            • <strong>anonymous</strong> → <strong>authenticated</strong>: User
            logs in
          </div>
          <div>
            • <strong>authenticated</strong> → <strong>expired</strong>: Session
            timeout or inactivity
          </div>
          <div>
            • <strong>expired</strong> → <strong>anonymous</strong>: User logs
            out
          </div>
          <div>
            • <strong>authenticated</strong> → <strong>anonymous</strong>: User
            logs out
          </div>
        </div>
      </div>

      <div>
        <h4>Session Data</h4>
        <pre
          style={{
            backgroundColor: "#f5f5f5",
            padding: "10px",
            borderRadius: "4px",
            fontSize: "12px",
            overflow: "auto",
          }}
        >
          {JSON.stringify(
            {
              state,
              userId,
              sessionId,
              isAuthenticated,
              isExpired,
              sessionDuration: formatDuration(sessionDuration),
              inactivityDuration: formatDuration(inactivityDuration),
            },
            null,
            2
          )}
        </pre>
      </div>
    </div>
  );
}
