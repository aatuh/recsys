/**
 * Thin app facade for API calls.
 * Re-exports app-level methods that call the generated services.
 * Provides light logging and standard error mapping to UI toasts.
 */

import {
  AuditService,
  BanditService,
  ConfigService,
  DataManagementService,
  ExplainService,
  IngestionService,
  RankingService,
  ApiError,
  type types_RecommendRequest,
  type types_RecommendResponse,
  type types_RecommendWithBanditRequest,
  type types_RecommendWithBanditResponse,
  type types_EventTypeConfigUpsertRequest,
  type types_EventTypeConfigUpsertResponse,
  type types_UsersUpsertRequest,
  type types_ItemsUpsertRequest,
  type types_EventsBatchRequest,
  type types_DeleteRequest,
  type types_SegmentProfilesUpsertRequest,
  type types_SegmentsUpsertRequest,
  type types_IDListRequest,
  type types_SegmentDryRunRequest,
  type types_SegmentDryRunResponse,
  type types_ExplainLLMRequest,
  type types_ExplainLLMResponse,
  type types_BanditDecideRequest,
  type types_BanditDecideResponse,
  type types_BanditRewardRequest,
  type types_BanditPoliciesUpsertRequest,
  type specs_types_ScoredItem,
} from "../lib/api-client";

// Re-export types for compatibility
export type ListParams = {
  namespace: string;
  limit?: number;
  offset?: number;
  user_id?: string;
  item_id?: string;
  event_type?: number;
  created_after?: string;
  created_before?: string;
};

export type DeleteParams = {
  namespace: string;
  ids?: string[];
  user_id?: string;
  item_id?: string;
  event_type?: number;
  created_after?: string;
  created_before?: string;
};

export type ListResponse = {
  items: any[];
  total: number;
  limit: number;
  offset: number;
  has_more: boolean;
  next_offset?: number;
};

/**
 * Map API errors to user-friendly messages.
 */
function mapApiError(error: unknown): string {
  if (error instanceof ApiError) {
    switch (error.status) {
      case 400:
        return "Invalid request. Please check your input and try again.";
      case 401:
        return "Authentication required. Please log in and try again.";
      case 403:
        return "Access denied. You don't have permission to perform this action.";
      case 404:
        return "Resource not found. The requested item may have been deleted.";
      case 409:
        return "Conflict. The resource already exists or is in use.";
      case 422:
        return "Validation error. Please check your input and try again.";
      case 429:
        return "Too many requests. Please wait a moment and try again.";
      case 500:
        return "Server error. Please try again later.";
      case 502:
      case 503:
      case 504:
        return "Service temporarily unavailable. Please try again later.";
      default:
        return `Request failed (${error.status}). Please try again.`;
    }
  }

  if (error instanceof Error) {
    // Network or other errors
    if (error.message.includes("fetch")) {
      return "Network error. Please check your connection and try again.";
    }
    return error.message;
  }

  return "An unexpected error occurred. Please try again.";
}

/**
 * Log API calls for debugging.
 */
function logApiCall(operation: string, params?: any): void {
  if (import.meta.env.DEV) {
    console.log(`[API] ${operation}`, params);
  }
}

/**
 * Handle API errors with logging and user-friendly messages.
 */
function handleApiError(operation: string, error: unknown): never {
  const userMessage = mapApiError(error);

  // Log detailed error for debugging
  console.error(`[API] ${operation} failed:`, error);

  // Throw error with user-friendly message
  throw new Error(userMessage);
}

// ============================================================================
// RECOMMENDATION SERVICES
// ============================================================================

/**
 * Get recommendations for a user.
 */
export async function recommend(
  payload: types_RecommendRequest
): Promise<types_RecommendResponse> {
  logApiCall("recommend", payload);

  try {
    return await RankingService.postV1Recommendations(payload);
  } catch (error) {
    handleApiError("recommend", error);
  }
}

/**
 * Get recommendations with bandit-selected policy.
 */
export async function recommendWithBandit(
  payload: types_RecommendWithBanditRequest
): Promise<types_RecommendWithBanditResponse> {
  logApiCall("recommendWithBandit", payload);

  try {
    return await RankingService.postV1BanditRecommendations(payload);
  } catch (error) {
    handleApiError("recommendWithBandit", error);
  }
}

/**
 * Get similar items for a given item.
 */
export async function getSimilarItems(
  itemId: string,
  namespace: string = "default",
  k: number = 20
): Promise<specs_types_ScoredItem[]> {
  logApiCall("getSimilarItems", { itemId, namespace, k });

  try {
    return await RankingService.getV1ItemsSimilar(itemId, namespace, k);
  } catch (error) {
    handleApiError("getSimilarItems", error);
  }
}

// ============================================================================
// DATA MANAGEMENT SERVICES
// ============================================================================

/**
 * List users with pagination and filtering.
 */
export async function listUsers(params: ListParams): Promise<ListResponse> {
  logApiCall("listUsers", params);

  try {
    const result = await DataManagementService.listUsers(
      params.namespace,
      params.limit,
      params.offset,
      params.user_id,
      params.created_after,
      params.created_before
    );

    return {
      items: result.items || [],
      total: result.total || 0,
      limit: result.limit || params.limit || 0,
      offset: result.offset || params.offset || 0,
      has_more: Boolean(result.has_more),
      next_offset: result.next_offset,
    };
  } catch (error) {
    handleApiError("listUsers", error);
  }
}

/**
 * List items with pagination and filtering.
 */
export async function listItems(params: ListParams): Promise<ListResponse> {
  logApiCall("listItems", params);

  try {
    const result = await DataManagementService.listItems(
      params.namespace,
      params.limit,
      params.offset,
      params.item_id,
      params.created_after,
      params.created_before
    );

    return {
      items: result.items || [],
      total: result.total || 0,
      limit: result.limit || params.limit || 0,
      offset: result.offset || params.offset || 0,
      has_more: Boolean(result.has_more),
      next_offset: result.next_offset,
    };
  } catch (error) {
    handleApiError("listItems", error);
  }
}

/**
 * List events with pagination and filtering.
 */
export async function listEvents(params: ListParams): Promise<ListResponse> {
  logApiCall("listEvents", params);

  try {
    const result = await DataManagementService.listEvents(
      params.namespace,
      params.limit,
      params.offset,
      params.user_id,
      params.item_id,
      params.event_type,
      params.created_after,
      params.created_before
    );

    return {
      items: result.items || [],
      total: result.total || 0,
      limit: result.limit || params.limit || 0,
      offset: result.offset || params.offset || 0,
      has_more: Boolean(result.has_more),
      next_offset: result.next_offset,
    };
  } catch (error) {
    handleApiError("listEvents", error);
  }
}

/**
 * Delete users with optional filtering.
 */
export async function deleteUsers(
  params: DeleteParams
): Promise<{ deleted_count: number }> {
  logApiCall("deleteUsers", params);

  try {
    const payload: types_DeleteRequest = {
      namespace: params.namespace,
      user_id: params.user_id,
      event_type: params.event_type,
      created_after: params.created_after,
      created_before: params.created_before,
    };

    const result = await DataManagementService.deleteUsers(payload);
    return { deleted_count: result.deleted_count || 0 };
  } catch (error) {
    handleApiError("deleteUsers", error);
  }
}

/**
 * Delete items with optional filtering.
 */
export async function deleteItems(
  params: DeleteParams
): Promise<{ deleted_count: number }> {
  logApiCall("deleteItems", params);

  try {
    const payload: types_DeleteRequest = {
      namespace: params.namespace,
      item_id: params.item_id,
      event_type: params.event_type,
      created_after: params.created_after,
      created_before: params.created_before,
    };

    const result = await DataManagementService.deleteItems(payload);
    return { deleted_count: result.deleted_count || 0 };
  } catch (error) {
    handleApiError("deleteItems", error);
  }
}

/**
 * Delete events with optional filtering.
 */
export async function deleteEvents(
  params: DeleteParams
): Promise<{ deleted_count: number }> {
  logApiCall("deleteEvents", params);

  try {
    const payload: types_DeleteRequest = {
      namespace: params.namespace,
      user_id: params.user_id,
      item_id: params.item_id,
      event_type: params.event_type,
      created_after: params.created_after,
      created_before: params.created_before,
    };

    const result = await DataManagementService.deleteEvents(payload);
    return { deleted_count: result.deleted_count || 0 };
  } catch (error) {
    handleApiError("deleteEvents", error);
  }
}

// ============================================================================
// INGESTION SERVICES
// ============================================================================

/**
 * Upsert users.
 */
export async function upsertUsers(
  payload: types_UsersUpsertRequest,
  append: (value: string) => void
): Promise<void> {
  logApiCall("upsertUsers", payload);

  try {
    await IngestionService.upsertUsers(payload);
    append(`✅ Upserted ${payload.users?.length || 0} users`);
  } catch (error) {
    handleApiError("upsertUsers", error);
  }
}

/**
 * Upsert items.
 */
export async function upsertItems(
  payload: types_ItemsUpsertRequest,
  append: (value: string) => void
): Promise<void> {
  logApiCall("upsertItems", payload);

  try {
    await IngestionService.upsertItems(payload);
    append(`✅ Upserted ${payload.items?.length || 0} items`);
  } catch (error) {
    handleApiError("upsertItems", error);
  }
}

/**
 * Batch events.
 */
export async function batchEvents(
  payload: types_EventsBatchRequest,
  append: (value: string) => void
): Promise<void> {
  logApiCall("batchEvents", payload);

  try {
    await IngestionService.batchEvents(payload);
    append(`✅ Batched ${payload.events?.length || 0} events`);
  } catch (error) {
    handleApiError("batchEvents", error);
  }
}

/**
 * Upsert event types.
 */
export async function upsertEventTypes(
  payload: types_EventTypeConfigUpsertRequest,
  append: (value: string) => void
): Promise<void> {
  logApiCall("upsertEventTypes", payload);

  try {
    await ConfigService.upsertEventTypes(payload);
    append(`✅ Upserted ${payload.types?.length || 0} event types`);
  } catch (error) {
    handleApiError("upsertEventTypes", error);
  }
}

// ============================================================================
// CONFIG SERVICES
// ============================================================================

/**
 * Get event types configuration.
 */
export async function getEventTypes(
  namespace: string
): Promise<types_EventTypeConfigUpsertResponse[]> {
  logApiCall("getEventTypes", { namespace });

  try {
    return await ConfigService.getV1EventTypes(namespace);
  } catch (error) {
    handleApiError("getEventTypes", error);
  }
}

/**
 * Get segment profiles.
 */
export async function getSegmentProfiles(namespace: string = "default") {
  logApiCall("getSegmentProfiles", { namespace });

  try {
    return await ConfigService.getV1SegmentProfiles(namespace);
  } catch (error) {
    handleApiError("getSegmentProfiles", error);
  }
}

/**
 * Upsert segment profiles.
 */
export async function upsertSegmentProfiles(
  payload: types_SegmentProfilesUpsertRequest
) {
  logApiCall("upsertSegmentProfiles", payload);

  try {
    return await ConfigService.segmentProfilesUpsert(payload);
  } catch (error) {
    handleApiError("upsertSegmentProfiles", error);
  }
}

/**
 * Delete segment profiles.
 */
export async function deleteSegmentProfiles(payload: types_IDListRequest) {
  logApiCall("deleteSegmentProfiles", payload);

  try {
    return await ConfigService.segmentProfilesDelete(payload);
  } catch (error) {
    handleApiError("deleteSegmentProfiles", error);
  }
}

/**
 * Get segments.
 */
export async function getSegments(namespace: string = "default") {
  logApiCall("getSegments", { namespace });

  try {
    return await ConfigService.getV1Segments(namespace);
  } catch (error) {
    handleApiError("getSegments", error);
  }
}

/**
 * Upsert segments.
 */
export async function upsertSegments(payload: types_SegmentsUpsertRequest) {
  logApiCall("upsertSegments", payload);

  try {
    return await ConfigService.segmentsUpsert(payload);
  } catch (error) {
    handleApiError("upsertSegments", error);
  }
}

/**
 * Delete segments.
 */
export async function deleteSegments(payload: types_IDListRequest) {
  logApiCall("deleteSegments", payload);

  try {
    return await ConfigService.segmentsDelete(payload);
  } catch (error) {
    handleApiError("deleteSegments", error);
  }
}

/**
 * Dry run segment selection.
 */
export async function dryRunSegments(
  payload: types_SegmentDryRunRequest
): Promise<types_SegmentDryRunResponse> {
  logApiCall("dryRunSegments", payload);

  try {
    return await ConfigService.segmentsDryRun(payload);
  } catch (error) {
    handleApiError("dryRunSegments", error);
  }
}

// ============================================================================
// BANDIT SERVICES
// ============================================================================

/**
 * Decide bandit policy.
 */
export async function decideBandit(
  payload: types_BanditDecideRequest
): Promise<types_BanditDecideResponse> {
  logApiCall("decideBandit", payload);

  try {
    return await BanditService.postV1BanditDecide(payload);
  } catch (error) {
    handleApiError("decideBandit", error);
  }
}

/**
 * Reward bandit.
 */
export async function rewardBandit(payload: types_BanditRewardRequest) {
  logApiCall("rewardBandit", payload);

  try {
    return await BanditService.postV1BanditReward(payload);
  } catch (error) {
    handleApiError("rewardBandit", error);
  }
}

/**
 * Upsert bandit policies.
 */
export async function upsertBanditPolicies(
  payload: types_BanditPoliciesUpsertRequest
) {
  logApiCall("upsertBanditPolicies", payload);

  try {
    return await BanditService.upsertBanditPolicies(payload);
  } catch (error) {
    handleApiError("upsertBanditPolicies", error);
  }
}

// ============================================================================
// EXPLAIN SERVICES
// ============================================================================

/**
 * Explain with LLM.
 */
export async function explainWithLLM(
  payload: types_ExplainLLMRequest
): Promise<types_ExplainLLMResponse> {
  logApiCall("explainWithLLM", payload);

  try {
    return await ExplainService.postV1ExplainLlm(payload);
  } catch (error) {
    handleApiError("explainWithLLM", error);
  }
}

// ============================================================================
// AUDIT SERVICES
// ============================================================================

/**
 * Get audit decisions.
 */
export async function getAuditDecisions(params: any) {
  logApiCall("getAuditDecisions", params);

  try {
    return await AuditService.getV1AuditDecisions(params);
  } catch (error) {
    handleApiError("getAuditDecisions", error);
  }
}

/**
 * Get audit decision detail.
 */
export async function getAuditDecisionDetail(decisionId: string) {
  logApiCall("getAuditDecisionDetail", { decisionId });

  try {
    return await AuditService.getV1AuditDecisions1(decisionId);
  } catch (error) {
    handleApiError("getAuditDecisionDetail", error);
  }
}
