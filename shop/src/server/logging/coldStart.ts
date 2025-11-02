import { RecommendationMeta } from "@/lib/contracts/event";

export type ColdStartMetric = {
  type: "impression" | "click" | "add" | "purchase";
  userId: string;
  itemId: string;
  surface?: string;
  widget?: string;
  rank?: number;
  requestId?: string;
  sessionId?: string;
};

export function logColdStart(metric: ColdStartMetric) {
  const payload = {
    ...metric,
    cold_start: true,
  };
  console.info(`[metrics] cold_start_${metric.type} ${JSON.stringify(payload)}`);
}

export function maybeLogColdStart(
  type: ColdStartMetric["type"],
  userId: string,
  itemId: string | null | undefined,
  meta?: RecommendationMeta
) {
  if (!meta?.cold_start || !itemId) return;
  logColdStart({
    type,
    userId,
    itemId,
    surface: meta.surface,
    widget: meta.widget,
    rank: meta.rank,
    requestId: meta.request_id,
    sessionId: meta.session_id,
  });
}
