import { loadConfig } from "./config";
import { OpenAPI } from "@/lib/api-client/core/OpenAPI";
import { IngestionService } from "@/lib/api-client/services/IngestionService";
import { RankingService } from "@/lib/api-client/services/RankingService";
import { DataManagementService } from "@/lib/api-client/services/DataManagementService";

export function initRecsysClient() {
  const cfg = loadConfig();
  OpenAPI.BASE = cfg.recsysBaseUrl;
}

export function mapEventTypeToCode(
  type: "view" | "click" | "add" | "purchase" | "custom"
): number {
  switch (type) {
    case "view":
      return 0;
    case "click":
      return 1;
    case "add":
      return 2;
    case "purchase":
      return 3;
    case "custom":
      return 4;
  }
}

export async function forwardEventsBatch(events: any[]) {
  const cfg = loadConfig();
  initRecsysClient();
  return IngestionService.batchEvents({
    namespace: cfg.recsysNamespace,
    events,
  });
}

export async function upsertItems(items: any[]) {
  const cfg = loadConfig();
  initRecsysClient();
  return IngestionService.upsertItems({
    namespace: cfg.recsysNamespace,
    items,
  });
}

export async function upsertUsers(users: any[]) {
  const cfg = loadConfig();
  initRecsysClient();
  return IngestionService.upsertUsers({
    namespace: cfg.recsysNamespace,
    users,
  });
}

export async function getRecommendations(params: {
  userId: string;
  k?: number;
  includeReasons?: boolean;
}) {
  const cfg = loadConfig();
  initRecsysClient();
  return RankingService.postV1Recommendations({
    user_id: params.userId,
    namespace: cfg.recsysNamespace,
    k: params.k ?? 12,
    include_reasons: params.includeReasons ?? false,
  } as any);
}

export async function getSimilar(params: { itemId: string; k?: number }) {
  const cfg = loadConfig();
  initRecsysClient();
  return RankingService.getV1ItemsSimilar(
    params.itemId,
    cfg.recsysNamespace,
    params.k ?? 10
  );
}

export async function deleteItems(itemIds: string[]) {
  const cfg = loadConfig();
  initRecsysClient();
  for (const id of itemIds) {
    await DataManagementService.deleteItems({
      namespace: cfg.recsysNamespace,
      item_id: id,
    } as any);
  }
}
