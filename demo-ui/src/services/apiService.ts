import {
  ConfigService,
  IngestionService,
  RankingService,
} from "../lib/api-client";
import type {
  types_EventTypeConfigUpsertRequest,
  types_UsersUpsertRequest,
  types_ItemsUpsertRequest,
  types_EventsBatchRequest,
  types_RecommendRequest,
  types_RecommendResponse,
  types_Overrides,
} from "../lib/api-client";

/**
 * API service functions for the RecSys demo UI application.
 */

export async function upsertEventTypes(
  namespace: string,
  eventTypes: Array<{
    id: string;
    title: string;
    index: number;
    weight: number;
    halfLifeDays: number;
  }>,
  append: (s: string) => void
) {
  const payload: types_EventTypeConfigUpsertRequest = {
    namespace,
    types: eventTypes.map((et) => ({
      type: et.index,
      name: et.id,
      weight: et.weight,
      half_life_days: et.halfLifeDays,
      is_active: true,
    })),
  };
  await ConfigService.upsertEventTypes(payload);
  append("✔ event-types:upsert");
}

export async function upsertUsers(
  namespace: string,
  users: any[],
  append: (s: string) => void
) {
  const payload: types_UsersUpsertRequest = { namespace, users };
  await IngestionService.upsertUsers(payload);
  append(`✔ users:upsert (${users.length})`);
}

export async function upsertItems(
  namespace: string,
  items: any[],
  append: (s: string) => void
) {
  const payload: types_ItemsUpsertRequest = { namespace, items };
  await IngestionService.upsertItems(payload);
  append(`✔ items:upsert (${items.length})`);
}

export async function batchEvents(
  namespace: string,
  events: any[],
  append: (s: string) => void
) {
  const payload: types_EventsBatchRequest = { namespace, events };
  await IngestionService.batchEvents(payload);
  append(`✔ events:batch (${events.length})`);
}

export async function recommend(
  userId: string,
  namespace: string,
  kVal: number,
  blendVal: { pop: number; cooc: number; als: number },
  overrides?: types_Overrides | null
): Promise<types_RecommendResponse> {
  const payload: types_RecommendRequest = {
    user_id: userId,
    namespace,
    k: kVal,
    include_reasons: true,
    constraints: {},
    blend: blendVal,
    overrides: overrides || undefined,
  };
  return RankingService.postV1Recommendations(payload);
}

export async function similar(itemId: string, namespace: string, kVal: number) {
  return RankingService.getV1ItemsSimilar(itemId, namespace, kVal);
}

// Data Management API methods

export interface ListParams {
  namespace: string;
  limit?: number;
  offset?: number;
  user_id?: string;
  item_id?: string;
  event_type?: number;
  created_after?: string;
  created_before?: string;
}

export interface ListResponse {
  items: any[];
  total: number;
  limit: number;
  offset: number;
  has_more: boolean;
  next_offset?: number;
}

export interface DeleteParams {
  namespace: string;
  user_id?: string;
  item_id?: string;
  event_type?: number;
  created_after?: string;
  created_before?: string;
}

export interface DeleteResponse {
  deleted_count: number;
  message: string;
}

const API_BASE =
  (import.meta as any).env?.VITE_API_BASE_URL?.toString() || "/api";

async function apiRequest<T>(
  url: string,
  options: RequestInit = {}
): Promise<T> {
  const response = await fetch(`${API_BASE}${url}`, {
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    ...options,
  });

  if (!response.ok) {
    let errorMessage = `HTTP ${response.status}`;
    try {
      const contentType = response.headers.get("Content-Type");
      if (contentType && contentType.includes("application/json")) {
        const error = await response.json();
        errorMessage = error.message || errorMessage;
      } else {
        const text = await response.text();
        errorMessage = text || errorMessage;
      }
    } catch (e) {
      // If we can't parse the error response, use the status
      errorMessage = `HTTP ${response.status}: ${response.statusText}`;
    }
    throw new Error(errorMessage);
  }

  return response.json();
}

export async function listUsers(params: ListParams): Promise<ListResponse> {
  const searchParams = new URLSearchParams();
  searchParams.set("namespace", params.namespace);
  if (params.limit) searchParams.set("limit", params.limit.toString());
  if (params.offset) searchParams.set("offset", params.offset.toString());
  if (params.user_id) searchParams.set("user_id", params.user_id);
  if (params.created_after)
    searchParams.set("created_after", params.created_after);
  if (params.created_before)
    searchParams.set("created_before", params.created_before);

  return apiRequest<ListResponse>(`/v1/users?${searchParams.toString()}`);
}

export async function listItems(params: ListParams): Promise<ListResponse> {
  const searchParams = new URLSearchParams();
  searchParams.set("namespace", params.namespace);
  if (params.limit) searchParams.set("limit", params.limit.toString());
  if (params.offset) searchParams.set("offset", params.offset.toString());
  if (params.item_id) searchParams.set("item_id", params.item_id);
  if (params.created_after)
    searchParams.set("created_after", params.created_after);
  if (params.created_before)
    searchParams.set("created_before", params.created_before);

  return apiRequest<ListResponse>(`/v1/items?${searchParams.toString()}`);
}

export async function listEvents(params: ListParams): Promise<ListResponse> {
  const searchParams = new URLSearchParams();
  searchParams.set("namespace", params.namespace);
  if (params.limit) searchParams.set("limit", params.limit.toString());
  if (params.offset) searchParams.set("offset", params.offset.toString());
  if (params.user_id) searchParams.set("user_id", params.user_id);
  if (params.item_id) searchParams.set("item_id", params.item_id);
  if (params.event_type !== undefined)
    searchParams.set("event_type", params.event_type.toString());
  if (params.created_after)
    searchParams.set("created_after", params.created_after);
  if (params.created_before)
    searchParams.set("created_before", params.created_before);

  return apiRequest<ListResponse>(`/v1/events?${searchParams.toString()}`);
}

export async function deleteUsers(
  params: DeleteParams
): Promise<DeleteResponse> {
  return apiRequest<DeleteResponse>("/v1/users:delete", {
    method: "POST",
    body: JSON.stringify(params),
  });
}

export async function deleteItems(
  params: DeleteParams
): Promise<DeleteResponse> {
  return apiRequest<DeleteResponse>("/v1/items:delete", {
    method: "POST",
    body: JSON.stringify(params),
  });
}

export async function deleteEvents(
  params: DeleteParams
): Promise<DeleteResponse> {
  return apiRequest<DeleteResponse>("/v1/events:delete", {
    method: "POST",
    body: JSON.stringify(params),
  });
}

// Export API methods

/**
 * Fetches all data for a specific table type with pagination
 */
export async function fetchAllData(
  dataType: "users" | "items" | "events",
  params: Omit<ListParams, "limit" | "offset">
): Promise<any[]> {
  const allData: any[] = [];
  let offset = 0;
  const limit = 1000; // Large batch size for efficiency

  while (true) {
    const batchParams: ListParams = {
      ...params,
      limit,
      offset,
    };

    let response: ListResponse;
    switch (dataType) {
      case "users":
        response = await listUsers(batchParams);
        break;
      case "items":
        response = await listItems(batchParams);
        break;
      case "events":
        response = await listEvents(batchParams);
        break;
      default:
        throw new Error("Invalid data type");
    }

    allData.push(...response.items);

    // If we got fewer items than requested, we've reached the end
    if (response.items.length < limit || !response.has_more) {
      break;
    }

    offset += limit;
  }

  return allData;
}

/**
 * Fetches all data for multiple table types
 */
export async function fetchAllDataForTables(
  namespace: string,
  selectedTables: string[],
  filters: {
    user_id: string;
    item_id: string;
    event_type: string;
    created_after: string;
    created_before: string;
  }
): Promise<{
  users?: any[];
  items?: any[];
  events?: any[];
}> {
  const baseParams = {
    namespace,
    user_id: filters.user_id || undefined,
    item_id: filters.item_id || undefined,
    event_type: filters.event_type ? parseInt(filters.event_type) : undefined,
    created_after: filters.created_after || undefined,
    created_before: filters.created_before || undefined,
  };

  const result: {
    users?: any[];
    items?: any[];
    events?: any[];
  } = {};

  // Fetch data for each selected table in parallel
  const promises = selectedTables.map(async (table) => {
    switch (table) {
      case "users":
        result.users = await fetchAllData("users", baseParams);
        break;
      case "items":
        result.items = await fetchAllData("items", baseParams);
        break;
      case "events":
        result.events = await fetchAllData("events", baseParams);
        break;
    }
  });

  await Promise.all(promises);
  return result;
}
