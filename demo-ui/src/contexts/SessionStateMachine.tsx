/**
 * Session state machine for managing user authentication states.
 * Implements a simple state machine: anonymous -> authenticated -> expired
 */

import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react";
import { getStorage, getLogger } from "../di";

export type SessionState = "anonymous" | "authenticated" | "expired";

export interface SessionContext {
  state: SessionState;
  userId?: string;
  sessionId?: string;
  loginTime?: number;
  lastActivity?: number;
  transitionTo: (newState: SessionState, data?: SessionData) => void;
  login: (userId: string) => void;
  logout: () => void;
  refresh: () => void;
  isAuthenticated: boolean;
  isExpired: boolean;
  getSessionDuration: () => number;
  getInactivityDuration: () => number;
}

export interface SessionData {
  userId?: string;
  sessionId?: string;
  loginTime?: number;
  lastActivity?: number;
}

const SessionContextProvider = createContext<SessionContext | undefined>(
  undefined
);

export interface SessionProviderProps {
  children: ReactNode;
  sessionTimeout?: number; // in milliseconds
  inactivityTimeout?: number; // in milliseconds
  autoRefresh?: boolean;
}

export function SessionProvider({
  children,
  sessionTimeout = 24 * 60 * 60 * 1000, // 24 hours
  inactivityTimeout = 30 * 60 * 1000, // 30 minutes
  autoRefresh = true,
}: SessionProviderProps) {
  const [state, setState] = useState<SessionState>("anonymous");
  const [sessionData, setSessionData] = useState<SessionData>({});

  const storage = getStorage();
  const logger = getLogger();
  const sessionLogger = logger.child({ component: "SessionStateMachine" });

  // Load session from storage on mount
  useEffect(() => {
    try {
      const storedSession = storage.getItem("session_data");
      if (storedSession) {
        const data: SessionData = JSON.parse(storedSession);
        const now = Date.now();

        // Check if session has expired
        if (data.loginTime && now - data.loginTime > sessionTimeout) {
          sessionLogger.info("Session expired due to timeout", {
            loginTime: data.loginTime,
            duration: now - data.loginTime,
          });
          setState("expired");
          return;
        }

        // Check if session is inactive
        if (data.lastActivity && now - data.lastActivity > inactivityTimeout) {
          sessionLogger.info("Session expired due to inactivity", {
            lastActivity: data.lastActivity,
            inactivity: now - data.lastActivity,
          });
          setState("expired");
          return;
        }

        // Session is still valid
        setState("authenticated");
        setSessionData(data);
        sessionLogger.info("Session restored from storage", {
          userId: data.userId,
        });
      }
    } catch (error) {
      sessionLogger.error("Failed to load session from storage", { error });
    }
  }, [sessionTimeout, inactivityTimeout, storage, sessionLogger]);

  // Save session to storage when it changes
  useEffect(() => {
    if (state === "authenticated" && sessionData.userId) {
      try {
        storage.setItem("session_data", JSON.stringify(sessionData));
        sessionLogger.debug("Session saved to storage", {
          userId: sessionData.userId,
        });
      } catch (error) {
        sessionLogger.error("Failed to save session to storage", { error });
      }
    } else if (state === "anonymous" || state === "expired") {
      try {
        storage.removeItem("session_data");
        sessionLogger.debug("Session removed from storage");
      } catch (error) {
        sessionLogger.error("Failed to remove session from storage", { error });
      }
    }
  }, [state, sessionData, storage, sessionLogger]);

  // Auto-refresh session activity
  useEffect(() => {
    if (!autoRefresh || state !== "authenticated") {
      return;
    }

    const interval = setInterval(() => {
      const now = Date.now();
      setSessionData((prev) => ({
        ...prev,
        lastActivity: now,
      }));
    }, 60000); // Update every minute

    return () => clearInterval(interval);
  }, [autoRefresh, state]);

  const transitionTo = (newState: SessionState, data?: SessionData) => {
    const oldState = state;
    setState(newState);

    if (data) {
      setSessionData(data);
    }

    sessionLogger.info("Session state transition", {
      from: oldState,
      to: newState,
      userId: data?.userId,
    });
  };

  const login = (userId: string) => {
    const now = Date.now();
    const sessionId = `session_${now}_${Math.random()
      .toString(36)
      .substr(2, 9)}`;

    const data: SessionData = {
      userId,
      sessionId,
      loginTime: now,
      lastActivity: now,
    };

    transitionTo("authenticated", data);
    sessionLogger.info("User logged in", { userId, sessionId });
  };

  const logout = () => {
    const userId = sessionData.userId;
    transitionTo("anonymous", {});
    sessionLogger.info("User logged out", { userId });
  };

  const refresh = () => {
    if (state === "authenticated") {
      const now = Date.now();
      setSessionData((prev) => ({
        ...prev,
        lastActivity: now,
      }));
      sessionLogger.debug("Session refreshed", { userId: sessionData.userId });
    }
  };

  const isAuthenticated = state === "authenticated";
  const isExpired = state === "expired";

  const getSessionDuration = (): number => {
    if (!sessionData.loginTime) {
      return 0;
    }
    return Date.now() - sessionData.loginTime;
  };

  const getInactivityDuration = (): number => {
    if (!sessionData.lastActivity) {
      return 0;
    }
    return Date.now() - sessionData.lastActivity;
  };

  const contextValue: SessionContext = {
    state,
    userId: sessionData.userId,
    sessionId: sessionData.sessionId,
    loginTime: sessionData.loginTime,
    lastActivity: sessionData.lastActivity,
    transitionTo,
    login,
    logout,
    refresh,
    isAuthenticated,
    isExpired,
    getSessionDuration,
    getInactivityDuration,
  };

  return (
    <SessionContextProvider.Provider value={contextValue}>
      {children}
    </SessionContextProvider.Provider>
  );
}

export function useSession(): SessionContext {
  const context = useContext(SessionContextProvider);
  if (context === undefined) {
    throw new Error("useSession must be used within a SessionProvider");
  }
  return context;
}

// Convenience hooks for specific session states
export function useIsAuthenticated(): boolean {
  const { isAuthenticated } = useSession();
  return isAuthenticated;
}

export function useIsExpired(): boolean {
  const { isExpired } = useSession();
  return isExpired;
}

export function useSessionData(): SessionData {
  const { userId, sessionId, loginTime, lastActivity } = useSession();
  return { userId, sessionId, loginTime, lastActivity };
}
