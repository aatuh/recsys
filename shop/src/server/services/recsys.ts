import { randomUUID } from "crypto";
import { loadConfig } from "./config";
import { OpenAPI } from "@/lib/api-client/core/OpenAPI";
import { IngestionService } from "@/lib/api-client/services/IngestionService";
import { RankingService } from "@/lib/api-client/services/RankingService";
import { DataManagementService } from "@/lib/api-client/services/DataManagementService";
import { BanditService } from "@/lib/api-client/services/BanditService";
import { ItemContract } from "@/lib/contracts/item";
import { UserContract } from "@/lib/contracts/user";
import { EventContract } from "@/lib/contracts/event";
import { prisma } from "@/server/db/client";
import type { specs_types_ScoredItem } from "@/lib/api-client/models/specs_types_ScoredItem";

export function initRecsysClient() {
  const cfg = loadConfig();
  OpenAPI.BASE = cfg.recsysBaseUrl;
  OpenAPI.HEADERS = async () => ({
    "X-Org-ID": cfg.recsysOrgId,
  });
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
  const result = await IngestionService.batchEvents({
    namespace: cfg.recsysNamespace,
    events,
  });
  if (BANDIT_ENABLED) {
    const rewardCandidates = events.filter((event) =>
      shouldSendBanditReward(event)
    );
    for (const event of rewardCandidates) {
      const meta = (event.meta ?? {}) as Record<string, unknown>;
      const policyId = meta["bandit_policy_id"] as string | undefined;
      const requestId = meta["bandit_request_id"] as string | undefined;
      if (!policyId) {
        continue;
      }
      await BanditService.postV1BanditReward({
        namespace: cfg.recsysNamespace,
        policy_id: policyId,
        request_id: requestId,
        bucket_key: meta["bandit_bucket"] as string | undefined,
        algorithm: meta["bandit_algorithm"] as string | undefined,
        surface: meta["surface"] as string | undefined,
        reward: true,
        experiment: meta["bandit_experiment"] as string | undefined,
        variant: meta["bandit_variant"] as string | undefined,
      }).catch((err) => {
        console.warn("bandit reward send failed", err);
      });
    }
  }
  return result;
}

function shouldSendBanditReward(event: EventContract): boolean {
  if (!BANDIT_ENABLED) {
    return false;
  }
  const meta = event.meta as Record<string, unknown> | undefined;
  if (!meta || !meta["bandit_policy_id"]) {
    return false;
  }
  switch (event.type) {
    case 1:
      return BANDIT_REWARD_ON_CLICK;
    case 2:
      return BANDIT_REWARD_ON_ADD;
    case 3:
      return BANDIT_REWARD_ON_PURCHASE;
    default:
      return false;
  }
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

const BANDIT_ENABLED = process.env.SHOP_BANDIT_ENABLED === "true";
const BANDIT_POLICY_IDS =
  process.env.SHOP_BANDIT_POLICY_IDS?.split(",")
    .map((id) => id.trim())
    .filter((id) => id.length > 0) ?? [];
const BANDIT_REWARD_ON_CLICK =
  process.env.SHOP_BANDIT_REWARD_ON_CLICK !== "false";
const BANDIT_REWARD_ON_ADD = process.env.SHOP_BANDIT_REWARD_ON_ADD !== "false";
const BANDIT_REWARD_ON_PURCHASE =
  process.env.SHOP_BANDIT_REWARD_ON_PURCHASE !== "false";

type RecommendationConstraintsInput = {
  price_between?: [number, number];
  include_tags_any?: string[];
  exclude_tags_any?: string[];
  brand_cap?: number;
  category_cap?: number;
};

export type RecommendationResponse = {
  items?: Array<{ item_id: string; score: number; reasons?: string[] }>;
  model_version?: string;
  segment_id?: string;
  profile_id?: string;
  bandit?: {
    chosen_policy_id?: string;
    algorithm?: string;
    bucket?: string;
    explore?: boolean;
    explain?: Record<string, string>;
    request_id?: string;
    experiment?: string;
    variant?: string;
  };
};

export async function getRecommendations(params: {
  userId: string;
  k?: number;
  includeReasons?: boolean;
  constraints?: RecommendationConstraintsInput;
  surface?: string;
  widget?: string;
  context?: Record<string, string>;
}): Promise<RecommendationResponse> {
  const cfg = loadConfig();
  initRecsysClient();

  // Build the request payload
  const basePayload: Record<string, unknown> = {
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
      basePayload.brand_cap = params.constraints.brand_cap;
    }

    if (params.constraints.category_cap) {
      basePayload.category_cap = params.constraints.category_cap;
    }

    if (Object.keys(constraints).length > 0) {
      basePayload.constraints = constraints;
    }
  }

  const surface = params.surface ?? "home";
  const context: Record<string, string> = {
    surface,
    ...(params.widget ? { widget: params.widget } : {}),
    ...(params.context ?? {}),
  };

  if (BANDIT_ENABLED) {
    const requestId = randomUUID();
    const banditPayload: Record<string, unknown> = {
      ...basePayload,
      surface,
      context,
      request_id: requestId,
    };
    if (BANDIT_POLICY_IDS.length > 0) {
      banditPayload.candidate_policy_ids = BANDIT_POLICY_IDS;
    }

    try {
      const banditResponse = await RankingService.postV1BanditRecommendations(
        banditPayload
      );

      return {
        items: banditResponse.items
          ?.filter(
            (item): item is specs_types_ScoredItem & { item_id: string } =>
              Boolean(item.item_id)
          )
          .map((item) => ({
            item_id: item.item_id,
            score: item.score ?? 0,
            reasons: item.reasons,
          })),
        model_version: banditResponse.model_version,
        segment_id: banditResponse.segment_id,
        profile_id: banditResponse.profile_id,
        bandit: {
          chosen_policy_id: banditResponse.chosen_policy_id,
          algorithm: banditResponse.algorithm ?? undefined,
          bucket: banditResponse.bandit_bucket ?? undefined,
          explore: banditResponse.explore ?? undefined,
          explain: banditResponse.bandit_explain ?? undefined,
          request_id: banditResponse.request_id ?? requestId,
          experiment: banditResponse.bandit_experiment ?? undefined,
          variant: banditResponse.bandit_variant ?? undefined,
        },
      };
    } catch (error) {
      console.warn(
        "Bandit recommendations failed, falling back to standard ranking",
        error
      );
    }
  }

  const recommendationResponse = await RankingService.postV1Recommendations(
    basePayload
  );
  return {
    items: recommendationResponse.items
      ?.filter((item): item is specs_types_ScoredItem & { item_id: string } =>
        Boolean(item.item_id)
      )
      .map((item) => ({
        item_id: item.item_id,
        score: item.score ?? 0,
        reasons: item.reasons,
      })),
    model_version: recommendationResponse.model_version,
    segment_id: recommendationResponse.segment_id,
    profile_id: recommendationResponse.profile_id,
  };
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
