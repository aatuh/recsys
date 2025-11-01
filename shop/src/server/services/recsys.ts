import { loadConfig } from "./config";
import { OpenAPI } from "@/lib/api-client/core/OpenAPI";
import { IngestionService } from "@/lib/api-client/services/IngestionService";
import { RankingService } from "@/lib/api-client/services/RankingService";
import { DataManagementService } from "@/lib/api-client/services/DataManagementService";
import { ItemContract } from "@/lib/contracts/item";
import { UserContract } from "@/lib/contracts/user";
import { EventContract } from "@/lib/contracts/event";
import { prisma } from "@/server/db/client";

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

export async function forwardEventsBatch(events: EventContract[]) {
  const cfg = loadConfig();
  initRecsysClient();
  return IngestionService.batchEvents({
    namespace: cfg.recsysNamespace,
    events,
  });
}

export async function upsertItems(items: ItemContract[]) {
  const cfg = loadConfig();
  initRecsysClient();

  try {
    console.log(`Syncing ${items.length} items to recsys...`);
    const result = await IngestionService.upsertItems({
      namespace: cfg.recsysNamespace,
      items,
    });
    console.log(`Successfully synced ${items.length} items to recsys`);
    return result;
  } catch (error) {
    console.error("Failed to upsert items to recsys:", error);
    throw error;
  }
}

export async function upsertUsers(users: UserContract[]) {
  const cfg = loadConfig();
  initRecsysClient();

  try {
    console.log(`Syncing ${users.length} users to recsys...`);
    const result = await IngestionService.upsertUsers({
      namespace: cfg.recsysNamespace,
      users,
    });
    console.log(`Successfully synced ${users.length} users to recsys`);
    return result;
  } catch (error) {
    console.error("Failed to upsert users to recsys:", error);
    throw error;
  }
}

export async function upsertEventTypeConfig() {
  initRecsysClient();
  // Note: This endpoint may not exist in the generated client yet
  // For now, we'll skip this functionality
  console.log("Event type config upsert not implemented yet");
  return { status: "skipped" };
}

export async function getRecommendations(params: {
  userId: string;
  k?: number;
  includeReasons?: boolean;
  constraints?: {
    price_between?: [number, number];
    include_tags_any?: string[];
    exclude_tags_any?: string[];
    brand_cap?: number;
    category_cap?: number;
  };
}) {
  const cfg = loadConfig();
  initRecsysClient();

  // Build the request payload
  const requestPayload: Record<string, unknown> = {
    user_id: params.userId,
    namespace: cfg.recsysNamespace,
    k: params.k ?? 12,
    include_reasons: params.includeReasons ?? false,
  };

  // Add constraints if provided
  if (params.constraints) {
    const constraints: Record<string, unknown> = {};

    if (params.constraints.price_between) {
      constraints.price_between = params.constraints.price_between;
    }

    if (params.constraints.include_tags_any) {
      constraints.include_tags_any = params.constraints.include_tags_any;
    }

    if (params.constraints.exclude_tags_any) {
      constraints.exclude_tags_any = params.constraints.exclude_tags_any;
    }

    if (params.constraints.brand_cap) {
      requestPayload.brand_cap = params.constraints.brand_cap;
    }

    if (params.constraints.category_cap) {
      requestPayload.category_cap = params.constraints.category_cap;
    }

    if (Object.keys(constraints).length > 0) {
      requestPayload.constraints = constraints;
    }
  }

  return RankingService.postV1Recommendations(requestPayload);
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

  try {
    console.log(`Deleting ${itemIds.length} items from recsys...`);
    for (const id of itemIds) {
      await DataManagementService.deleteItems({
        namespace: cfg.recsysNamespace,
        item_id: id,
      });
    }
    console.log(`Successfully deleted ${itemIds.length} items from recsys`);
  } catch (error) {
    console.error("Failed to delete items from recsys:", error);
    throw error;
  }
}

export async function deleteAllItemsInNamespace() {
  const cfg = loadConfig();
  initRecsysClient();

  try {
    console.log("Deleting ALL items in recsys namespace...");
    const result = await DataManagementService.deleteItems({
      namespace: cfg.recsysNamespace,
    });
    console.log(
      `Successfully deleted ${result.deleted_count ?? 0} items from recsys`
    );
    return result;
  } catch (error) {
    console.error("Failed to delete all items from recsys:", error);
    throw error;
  }
}

export async function deleteUsers(userIds: string[]) {
  const cfg = loadConfig();
  initRecsysClient();

  try {
    console.log(`Deleting ${userIds.length} users from recsys...`);
    for (const id of userIds) {
      // Delete events tied to the user
      await DataManagementService.deleteEvents({
        namespace: cfg.recsysNamespace,
        user_id: id,
      });
      // Delete the user entity/profile itself
      await DataManagementService.deleteUsers({
        namespace: cfg.recsysNamespace,
        user_id: id,
      });
    }
    console.log(
      `Successfully deleted events for ${userIds.length} users from recsys`
    );
  } catch (error) {
    console.error("Failed to delete user events from recsys:", error);
    throw error;
  }
}

export async function deleteAllUsersInNamespace() {
  const cfg = loadConfig();
  initRecsysClient();

  try {
    console.log("Deleting ALL users in recsys namespace...");
    const result = await DataManagementService.deleteUsers({
      namespace: cfg.recsysNamespace,
    });
    console.log(
      `Successfully deleted ${result.deleted_count ?? 0} users from recsys`
    );
    return result;
  } catch (error) {
    console.error("Failed to delete all users from recsys:", error);
    throw error;
  }
}

export async function deleteAllEventsInNamespace() {
  const cfg = loadConfig();
  initRecsysClient();

  try {
    console.log("Deleting ALL events in recsys namespace...");
    const result = await DataManagementService.deleteEvents({
      namespace: cfg.recsysNamespace,
    });
    console.log(
      `Successfully deleted ${result.deleted_count ?? 0} events from recsys`
    );
    return result;
  } catch (error) {
    console.error("Failed to delete all events from recsys:", error);
    throw error;
  }
}

export async function deleteEvents(eventIds: string[]) {
  const cfg = loadConfig();
  initRecsysClient();

  try {
    console.log(`Deleting ${eventIds.length} events from recsys...`);
    // Fetch details needed to target deletions via filters
    const events = await prisma.event.findMany({
      where: { id: { in: eventIds } },
      select: { id: true, userId: true, productId: true, type: true, ts: true },
    });

    for (const e of events) {
      // Build a narrow filter per event. We match by user, optional item,
      // event type code, and an exact timestamp window (same start/end).
      const eventTypeCode = mapEventTypeToCode(
        e.type as unknown as "view" | "click" | "add" | "purchase" | "custom"
      );
      const tsDate = e.ts as unknown as Date;
      const start = new Date(tsDate.getTime() - 1000).toISOString();
      const end = new Date(tsDate.getTime() + 1000).toISOString();
      await DataManagementService.deleteEvents({
        namespace: cfg.recsysNamespace,
        user_id: e.userId || undefined,
        item_id: e.productId || undefined,
        event_type: eventTypeCode,
        created_after: start,
        created_before: end,
      });
    }
    console.log(`Successfully requested deletion for ${events.length} events`);
  } catch (error) {
    console.error("Failed to delete events from recsys:", error);
    throw error;
  }
}
