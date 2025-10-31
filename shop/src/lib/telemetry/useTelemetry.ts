"use client";
import { useCallback } from "react";
import { telemetryEventSchema } from "./schema";

type TelemetryEvent = {
  userId: string;
  productId?: string;
  type: "view" | "click" | "add" | "purchase" | "custom";
  value?: number;
  ts?: string;
  meta?: unknown;
  sourceEventId?: string;
};

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
    const payload: TelemetryEvent = {
      userId,
      ts: new Date().toISOString(),
      ...ev,
    };
    const parsed = telemetryEventSchema.safeParse(payload);
    if (!parsed.success) return;
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
