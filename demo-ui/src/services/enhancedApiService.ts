import { getHttpClient, getLogger, type HttpClient, type Logger } from "../di";
import {
  ConfigService,
  IngestionService,
  RankingService,
  DataManagementService,
  types_RecommendRequest,
  type types_EventTypeConfigUpsertRequest,
  type types_UsersUpsertRequest,
  type types_ItemsUpsertRequest,
  type types_EventsBatchRequest,
  type types_RecommendResponse,
  type types_Overrides,
  type types_User,
  type types_Item,
  type types_Event,
  type types_ListResponse,
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
 * Enhanced API service with dependency injection and structured logging.
 * This service demonstrates the new DI patterns and can be used alongside
 * the existing apiService.ts for gradual migration.
 */

export interface ApiServiceOptions {
  httpClient?: HttpClient;
  logger?: Logger;
}

export class EnhancedApiService {
  private httpClient: HttpClient;
  private logger: Logger;

  constructor(options: ApiServiceOptions = {}) {
    this.httpClient = options.httpClient || getHttpClient();
    this.logger = options.logger || getLogger();
  }

  async upsertEventTypes(
    namespace: string,
    eventTypes: Array<{
      id: string;
      title: string;
      index: number;
      weight: number;
      halfLifeDays: number;
    }>,
    append: (value: string) => void
  ) {
    const startTime = Date.now();
    this.logger.info("Starting event types upsert", {
      namespace,
      count: eventTypes.length,
    });

    try {
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

      const duration = Date.now() - startTime;
      this.logger.info("Event types upsert completed", {
        namespace,
        count: eventTypes.length,
        duration,
      });

      append("✔ event-types:upsert");
    } catch (error) {
      this.logger.error("Event types upsert failed", {
        namespace,
        count: eventTypes.length,
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  async upsertUsers(
    namespace: string,
    users: types_User[],
    append: (value: string) => void
  ) {
    const startTime = Date.now();
    this.logger.info("Starting users upsert", {
      namespace,
      count: users.length,
    });

    try {
      const payload: types_UsersUpsertRequest = { namespace, users };
      await IngestionService.upsertUsers(payload);

      const duration = Date.now() - startTime;
      this.logger.info("Users upsert completed", {
        namespace,
        count: users.length,
        duration,
      });

      append(`✔ users:upsert (${users.length})`);
    } catch (error) {
      this.logger.error("Users upsert failed", {
        namespace,
        count: users.length,
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  async upsertItems(
    namespace: string,
    items: types_Item[],
    append: (value: string) => void
  ) {
    const startTime = Date.now();
    this.logger.info("Starting items upsert", {
      namespace,
      count: items.length,
    });

    try {
      const payload: types_ItemsUpsertRequest = { namespace, items };
      await IngestionService.upsertItems(payload);

      const duration = Date.now() - startTime;
      this.logger.info("Items upsert completed", {
        namespace,
        count: items.length,
        duration,
      });

      append(`✔ items:upsert (${items.length})`);
    } catch (error) {
      this.logger.error("Items upsert failed", {
        namespace,
        count: items.length,
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  async batchEvents(
    namespace: string,
    events: types_Event[],
    append: (value: string) => void
  ) {
    const startTime = Date.now();
    this.logger.info("Starting events batch", {
      namespace,
      count: events.length,
    });

    try {
      const payload: types_EventsBatchRequest = { namespace, events };
      await IngestionService.batchEvents(payload);

      const duration = Date.now() - startTime;
      this.logger.info("Events batch completed", {
        namespace,
        count: events.length,
        duration,
      });

      append(`✔ events:batch (${events.length})`);
    } catch (error) {
      this.logger.error("Events batch failed", {
        namespace,
        count: events.length,
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  async recommend(
    userId: string,
    namespace: string,
    kVal: number,
    blendVal: { pop: number; cooc: number; als: number },
    overrides?: types_Overrides | null
  ): Promise<types_RecommendResponse> {
    const startTime = Date.now();
    this.logger.info("Starting recommendation request", {
      userId,
      namespace,
      k: kVal,
      blend: blendVal,
    });

    try {
      const payload: types_RecommendRequest = {
        user_id: userId,
        namespace,
        k: kVal,
        include_reasons: true,
        explain_level: types_RecommendRequest.explain_level.NUMERIC,
        constraints: {},
        blend: blendVal,
        overrides: overrides ?? undefined,
      };

      const result = await RankingService.postV1Recommendations(payload);

      const duration = Date.now() - startTime;
      this.logger.info("Recommendation request completed", {
        userId,
        namespace,
        k: kVal,
        duration,
        resultCount: result.items?.length || 0,
      });

      return result;
    } catch (error) {
      this.logger.error("Recommendation request failed", {
        userId,
        namespace,
        k: kVal,
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  async similar(itemId: string, namespace: string, kVal: number) {
    const startTime = Date.now();
    this.logger.info("Starting similar items request", {
      itemId,
      namespace,
      k: kVal,
    });

    try {
      const result = await RankingService.getV1ItemsSimilar(
        itemId,
        namespace,
        kVal
      );

      const duration = Date.now() - startTime;
      this.logger.info("Similar items request completed", {
        itemId,
        namespace,
        k: kVal,
        duration,
        resultCount: result.length || 0,
      });

      return result;
    } catch (error) {
      this.logger.error("Similar items request failed", {
        itemId,
        namespace,
        k: kVal,
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  // Data management methods
  async listUsers(params: {
    namespace: string;
    limit?: number;
    offset?: number;
    user_id?: string;
    created_after?: string;
    created_before?: string;
  }) {
    const startTime = Date.now();
    this.logger.info("Starting list users request", { params });

    try {
      const res: types_ListResponse = await DataManagementService.listUsers(
        params.namespace,
        params.limit,
        params.offset,
        params.user_id,
        params.created_after,
        params.created_before
      );

      const duration = Date.now() - startTime;
      this.logger.info("List users request completed", {
        params,
        duration,
        resultCount: res.items?.length || 0,
      });

      return {
        items: res.items || [],
        total: res.total || 0,
        limit: res.limit || params.limit || 0,
        offset: res.offset || params.offset || 0,
        has_more: Boolean(res.has_more),
        next_offset: res.next_offset,
      };
    } catch (error) {
      const duration = Date.now() - startTime;
      this.logger.error("List users request failed", {
        params,
        duration,
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  async listItems(params: {
    namespace: string;
    limit?: number;
    offset?: number;
    item_id?: string;
    created_after?: string;
    created_before?: string;
  }) {
    const startTime = Date.now();
    this.logger.info("Starting list items request", { params });

    try {
      const res: types_ListResponse = await DataManagementService.listItems(
        params.namespace,
        params.limit,
        params.offset,
        params.item_id,
        params.created_after,
        params.created_before
      );

      const duration = Date.now() - startTime;
      this.logger.info("List items request completed", {
        params,
        duration,
        resultCount: res.items?.length || 0,
      });

      return {
        items: res.items || [],
        total: res.total || 0,
        limit: res.limit || params.limit || 0,
        offset: res.offset || params.offset || 0,
        has_more: Boolean(res.has_more),
        next_offset: res.next_offset,
      };
    } catch (error) {
      const duration = Date.now() - startTime;
      this.logger.error("List items request failed", {
        params,
        duration,
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  async listEvents(params: {
    namespace: string;
    limit?: number;
    offset?: number;
    user_id?: string;
    item_id?: string;
    created_after?: string;
    created_before?: string;
  }) {
    const startTime = Date.now();
    this.logger.info("Starting list events request", { params });

    try {
      const res: types_ListResponse = await DataManagementService.listEvents(
        params.namespace,
        params.limit,
        params.offset,
        params.user_id,
        params.item_id,
        undefined, // eventType
        params.created_after,
        params.created_before
      );

      const duration = Date.now() - startTime;
      this.logger.info("List events request completed", {
        params,
        duration,
        resultCount: res.items?.length || 0,
      });

      return {
        items: res.items || [],
        total: res.total || 0,
        limit: res.limit || params.limit || 0,
        offset: res.offset || params.offset || 0,
        has_more: Boolean(res.has_more),
        next_offset: res.next_offset,
      };
    } catch (error) {
      const duration = Date.now() - startTime;
      this.logger.error("List events request failed", {
        params,
        duration,
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  async deleteUsers(namespace: string, userIds?: string[]) {
    const startTime = Date.now();
    this.logger.info("Starting delete users request", { namespace, userIds });

    try {
      // Delete users - if userIds provided, filter by them, otherwise delete all
      const deleteRequest: any = { namespace };
      if (userIds && userIds.length > 0) {
        // Note: The API doesn't support deleting by specific IDs, so we delete all
        // This is a limitation of the current API design
        this.logger.warn(
          "Delete by specific IDs not supported, deleting all users",
          { userIds }
        );
      }

      await DataManagementService.deleteUsers(deleteRequest);

      const duration = Date.now() - startTime;
      this.logger.info("Delete users request completed", {
        namespace,
        userIds,
        duration,
      });

      return { deleted_count: userIds?.length || 0 };
    } catch (error) {
      const duration = Date.now() - startTime;
      this.logger.error("Delete users request failed", {
        namespace,
        userIds,
        duration,
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  async deleteItems(namespace: string, itemIds?: string[]) {
    const startTime = Date.now();
    this.logger.info("Starting delete items request", { namespace, itemIds });

    try {
      // Delete items - if itemIds provided, filter by them, otherwise delete all
      const deleteRequest: any = { namespace };
      if (itemIds && itemIds.length > 0) {
        // Note: The API doesn't support deleting by specific IDs, so we delete all
        // This is a limitation of the current API design
        this.logger.warn(
          "Delete by specific IDs not supported, deleting all items",
          { itemIds }
        );
      }

      await DataManagementService.deleteItems(deleteRequest);

      const duration = Date.now() - startTime;
      this.logger.info("Delete items request completed", {
        namespace,
        itemIds,
        duration,
      });

      return { deleted_count: itemIds?.length || 0 };
    } catch (error) {
      const duration = Date.now() - startTime;
      this.logger.error("Delete items request failed", {
        namespace,
        itemIds,
        duration,
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  async deleteEvents(namespace: string, eventIds?: string[]) {
    const startTime = Date.now();
    this.logger.info("Starting delete events request", { namespace, eventIds });

    try {
      // Delete events - if eventIds provided, filter by them, otherwise delete all
      const deleteRequest: any = { namespace };
      if (eventIds && eventIds.length > 0) {
        // Note: The API doesn't support deleting by specific IDs, so we delete all
        // This is a limitation of the current API design
        this.logger.warn(
          "Delete by specific IDs not supported, deleting all events",
          { eventIds }
        );
      }

      await DataManagementService.deleteEvents(deleteRequest);

      const duration = Date.now() - startTime;
      this.logger.info("Delete events request completed", {
        namespace,
        eventIds,
        duration,
      });

      return { deleted_count: eventIds?.length || 0 };
    } catch (error) {
      const duration = Date.now() - startTime;
      this.logger.error("Delete events request failed", {
        namespace,
        eventIds,
        duration,
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  // Stub methods for compatibility (these would need to be implemented)
  async listSegments(_params: ListParams): Promise<ListResponse> {
    // This would need to be implemented based on the actual API
    throw new Error("listSegments not implemented in enhanced service");
  }

  async deleteSegments(_namespace: string, _segmentIds: string[]) {
    // This would need to be implemented based on the actual API
    throw new Error("deleteSegments not implemented in enhanced service");
  }

  async fetchAllDataForTables(_namespace: string) {
    // This would need to be implemented based on the actual API
    throw new Error(
      "fetchAllDataForTables not implemented in enhanced service"
    );
  }
}

// Export a default instance for convenience
export const enhancedApiService = new EnhancedApiService();
