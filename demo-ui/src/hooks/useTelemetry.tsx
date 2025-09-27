/**
 * Telemetry hooks for tracking events and analytics.
 * No-op by default; future integrations can plug in analytics services.
 */

import React, { useCallback, useRef } from "react";
import { getLogger } from "../di";
import { useFeatureFlags } from "../contexts/FeatureFlagsContext";

export interface TelemetryEvent {
  event: string;
  properties?: Record<string, unknown>;
  timestamp?: number;
  userId?: string;
  sessionId?: string;
}

export interface TelemetryContext {
  track: (event: string, properties?: Record<string, unknown>) => void;
  identify: (userId: string, traits?: Record<string, unknown>) => void;
  page: (name: string, properties?: Record<string, unknown>) => void;
  alias: (userId: string, previousId?: string) => void;
  group: (groupId: string, traits?: Record<string, unknown>) => void;
  reset: () => void;
}

export interface TelemetryProviderProps {
  children: React.ReactNode;
  enabled?: boolean;
  debug?: boolean;
}

// Default no-op telemetry implementation
class NoOpTelemetry implements TelemetryContext {
  track(_event: string, _properties?: Record<string, unknown>): void {
    // No-op
  }

  identify(_userId: string, _traits?: Record<string, unknown>): void {
    // No-op
  }

  page(_name: string, _properties?: Record<string, unknown>): void {
    // No-op
  }

  alias(_userId: string, _previousId?: string): void {
    // No-op
  }

  group(_groupId: string, _traits?: Record<string, unknown>): void {
    // No-op
  }

  reset(): void {
    // No-op
  }
}

// Enhanced telemetry with logging and feature flag integration
class EnhancedTelemetry implements TelemetryContext {
  private logger = getLogger().child({ component: "Telemetry" });
  private userId: string | null = null;
  private sessionId: string | null = null;
  private groupId: string | null = null;
  private enabled: boolean;
  private debug: boolean;

  constructor(enabled: boolean = true, debug: boolean = false) {
    this.enabled = enabled;
    this.debug = debug;
  }

  private shouldTrack(): boolean {
    return this.enabled;
  }

  private logEvent(event: string, properties?: Record<string, unknown>): void {
    if (this.debug) {
      this.logger.debug("Telemetry event", {
        event,
        properties,
        userId: this.userId,
        sessionId: this.sessionId,
        groupId: this.groupId,
      });
    }
  }

  track(event: string, properties?: Record<string, unknown> | undefined): void {
    if (!this.shouldTrack()) {
      return;
    }

    const _telemetryEvent: TelemetryEvent = {
      event,
      properties,
      timestamp: Date.now(),
      userId: this.userId || undefined,
      sessionId: this.sessionId || undefined,
    };

    this.logEvent(event, properties);

    // Future: Send to analytics service
    // Example: analytics.track(event, properties);
  }

  identify(userId: string, traits?: Record<string, unknown>): void {
    if (!this.shouldTrack()) {
      return;
    }

    this.userId = userId;
    this.logEvent("identify", { userId, traits });

    // Future: Send to analytics service
    // Example: analytics.identify(userId, traits);
  }

  page(name: string, properties?: Record<string, unknown>): void {
    if (!this.shouldTrack()) {
      return;
    }

    this.logEvent("page", { name, properties });

    // Future: Send to analytics service
    // Example: analytics.page(name, properties);
  }

  alias(userId: string, previousId?: string): void {
    if (!this.shouldTrack()) {
      return;
    }

    this.logEvent("alias", { userId, previousId });

    // Future: Send to analytics service
    // Example: analytics.alias(userId, previousId);
  }

  group(groupId: string, traits?: Record<string, unknown>): void {
    if (!this.shouldTrack()) {
      return;
    }

    this.groupId = groupId;
    this.logEvent("group", { groupId, traits });

    // Future: Send to analytics service
    // Example: analytics.group(groupId, traits);
  }

  reset(): void {
    this.userId = null;
    this.sessionId = null;
    this.groupId = null;
    this.logEvent("reset");
  }
}

// Hook for using telemetry
export function useTelemetry(): TelemetryContext {
  const { isEnabled } = useFeatureFlags();
  const telemetryRef = useRef<TelemetryContext | null>(null);

  if (!telemetryRef.current) {
    if (isEnabled("analyticsEnabled")) {
      telemetryRef.current = new EnhancedTelemetry(
        true,
        (globalThis as any).process?.env?.NODE_ENV === "development"
      );
    } else {
      telemetryRef.current = new NoOpTelemetry();
    }
  }

  return telemetryRef.current;
}

// Convenience hooks for specific telemetry actions
export function useTrack() {
  const telemetry = useTelemetry();

  return useCallback(
    (event: string, properties?: Record<string, unknown>) => {
      telemetry.track(event, properties);
    },
    [telemetry]
  );
}

export function useIdentify() {
  const telemetry = useTelemetry();

  return useCallback(
    (userId: string, traits?: Record<string, unknown>) => {
      telemetry.identify(userId, traits);
    },
    [telemetry]
  );
}

export function usePage() {
  const telemetry = useTelemetry();

  return useCallback(
    (name: string, properties?: Record<string, unknown>) => {
      telemetry.page(name, properties);
    },
    [telemetry]
  );
}

// Higher-order component for automatic page tracking
export function withTelemetry<P extends object>(
  Component: React.ComponentType<P>,
  pageName?: string
) {
  return function TelemetryComponent(props: P) {
    const _track = useTrack();
    const page = usePage();

    React.useEffect(() => {
      if (pageName) {
        page(pageName);
      }
    }, [page, pageName]);

    return <Component {...props} />;
  };
}

// Telemetry event constants for consistency
export const TELEMETRY_EVENTS = {
  // User events
  USER_LOGIN: "user_login",
  USER_LOGOUT: "user_logout",
  USER_REGISTER: "user_register",

  // Navigation events
  PAGE_VIEW: "page_view",
  NAVIGATION: "navigation",

  // Feature usage
  FEATURE_USED: "feature_used",
  FEATURE_ENABLED: "feature_enabled",
  FEATURE_DISABLED: "feature_disabled",

  // API events
  API_REQUEST: "api_request",
  API_SUCCESS: "api_success",
  API_ERROR: "api_error",

  // Error events
  ERROR_OCCURRED: "error_occurred",
  ERROR_BOUNDARY_TRIGGERED: "error_boundary_triggered",

  // Performance events
  PERFORMANCE_MARK: "performance_mark",
  PERFORMANCE_MEASURE: "performance_measure",

  // UI events
  BUTTON_CLICK: "button_click",
  FORM_SUBMIT: "form_submit",
  MODAL_OPEN: "modal_open",
  MODAL_CLOSE: "modal_close",

  // Business events
  RECOMMENDATION_REQUESTED: "recommendation_requested",
  RECOMMENDATION_RECEIVED: "recommendation_received",
  ITEM_VIEWED: "item_viewed",
  ITEM_CLICKED: "item_clicked",
} as const;

export type TelemetryEventName =
  (typeof TELEMETRY_EVENTS)[keyof typeof TELEMETRY_EVENTS];
