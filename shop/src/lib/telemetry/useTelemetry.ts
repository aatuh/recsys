"use client";
import { useCallback } from "react";
import { telemetryEventSchema } from "./schema";
import { RecommendationMeta } from "@/lib/contracts/event";

type TelemetryEvent = {
  userId: string;
  productId?: string;
  type: "view" | "click" | "add" | "purchase" | "custom";
  value?: number;
  ts?: string;
  meta?: RecommendationMeta;
  sourceEventId?: string;
};

// Generate session ID for tracking
function getSessionId(): string {
  let sessionId = window.sessionStorage.getItem("shop_session_id");
  if (!sessionId) {
    sessionId = `session_${Date.now()}_${Math.random()
      .toString(36)
      .substr(2, 9)}`;
    window.sessionStorage.setItem("shop_session_id", sessionId);
  }
  return sessionId;
}

// Generate request ID for recommendation tracking
function generateRequestId(): string {
  return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

async function postEventFetch(ev: TelemetryEvent) {
  await fetch("/api/events", {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(ev),
  });
}

export function useTelemetry() {
  const emit = useCallback(async (ev: Omit<TelemetryEvent, "userId">) => {
    const userId = window.localStorage.getItem("shop_user_id");
    if (!userId) return;

    const sessionId = ev.meta?.session_id ?? getSessionId();
    const requestId =
      ev.meta?.request_id ?? ev.meta?.bandit_request_id ?? generateRequestId();

    const baseMeta: RecommendationMeta = {
      ...(ev.meta ?? {}),
      session_id: sessionId,
      request_id: requestId,
      referrer:
        ev.meta?.referrer ?? (document.referrer || window.location.pathname),
    };

    const payload: TelemetryEvent = {
      userId,
      ts: new Date().toISOString(),
      meta: baseMeta,
      ...ev,
    };

    const parsed = telemetryEventSchema.safeParse(payload);
    if (!parsed.success) {
      console.warn("Invalid telemetry event:", parsed.error, payload);
      return;
    }

    if (navigator.sendBeacon) {
      try {
        const blob = new Blob([JSON.stringify(payload)], {
          type: "application/json",
        });
        navigator.sendBeacon("/api/events", blob);
        return;
      } catch {}
    }
    // Fallback to fetch
    try {
      await postEventFetch(payload);
    } catch {
      // minimal retry once after 1s
      setTimeout(() => postEventFetch(payload).catch(() => {}), 1000);
    }
  }, []);

  return { emit };
}
