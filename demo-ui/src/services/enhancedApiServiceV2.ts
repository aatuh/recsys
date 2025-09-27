/**
 * Enhanced API service using the new HTTP client with interceptors, retry, and circuit breaker.
 * This service demonstrates the advanced HTTP patterns and can replace the existing API service.
 */

import { getHttpClient, getLogger, type HttpClient, type Logger } from "../di";
import {
  types_RecommendRequest,
  type types_EventTypeConfigUpsertRequest,
  type types_UsersUpsertRequest,
  type types_ItemsUpsertRequest,
  type types_EventsBatchRequest,
  type types_RecommendResponse,
  type types_Overrides,
} from "../lib/api-client";
import { ApiError } from "../di/http";

export interface EnhancedApiServiceOptions {
  httpClient?: HttpClient;
  logger?: Logger;
}

export class EnhancedApiServiceV2 {
  private httpClient: HttpClient;
  private logger: Logger;

  constructor(options: EnhancedApiServiceOptions = {}) {
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
    const requestId = (globalThis as any).crypto.randomUUID();

    this.logger.info("Starting event types upsert", {
      namespace,
      count: eventTypes.length,
      requestId,
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

      // Use the enhanced HTTP client for the request
      await this.httpClient.post("/config/event-types", payload, {
        headers: {
          "Content-Type": "application/json",
        },
      });

      const duration = Date.now() - startTime;
      this.logger.info("Event types upsert completed", {
        namespace,
        count: eventTypes.length,
        duration,
        requestId,
      });

      append("✔ event-types:upsert");
    } catch (error) {
      this.logger.error("Event types upsert failed", {
        namespace,
        count: eventTypes.length,
        requestId,
        error: error instanceof Error ? error.message : String(error),
        isRetryable: error instanceof ApiError ? error.retryable : false,
      });
      throw error;
    }
  }

  async upsertUsers(
    namespace: string,
    users: any[],
    append: (value: string) => void
  ) {
    const startTime = Date.now();
    const requestId = (globalThis as any).crypto.randomUUID();

    this.logger.info("Starting users upsert", {
      namespace,
      count: users.length,
      requestId,
    });

    try {
      const payload: types_UsersUpsertRequest = { namespace, users };

      await this.httpClient.post("/ingestion/users", payload, {
        headers: {
          "Content-Type": "application/json",
        },
      });

      const duration = Date.now() - startTime;
      this.logger.info("Users upsert completed", {
        namespace,
        count: users.length,
        duration,
        requestId,
      });

      append(`✔ users:upsert (${users.length})`);
    } catch (error) {
      this.logger.error("Users upsert failed", {
        namespace,
        count: users.length,
        requestId,
        error: error instanceof Error ? error.message : String(error),
        isRetryable: error instanceof ApiError ? error.retryable : false,
      });
    }
  }

  async upsertItems(
    namespace: string,
    items: any[],
    append: (value: string) => void
  ) {
    const startTime = Date.now();
    const requestId = (globalThis as any).crypto.randomUUID();

    this.logger.info("Starting items upsert", {
      namespace,
      count: items.length,
      requestId,
    });

    try {
      const payload: types_ItemsUpsertRequest = { namespace, items };

      await this.httpClient.post("/ingestion/items", payload, {
        headers: {
          "Content-Type": "application/json",
        },
      });

      const duration = Date.now() - startTime;
      this.logger.info("Items upsert completed", {
        namespace,
        count: items.length,
        duration,
        requestId,
      });

      append(`✔ items:upsert (${items.length})`);
    } catch (error) {
      this.logger.error("Items upsert failed", {
        namespace,
        count: items.length,
        requestId,
        error: error instanceof Error ? error.message : String(error),
        isRetryable: error instanceof ApiError ? error.retryable : false,
      });
      throw error;
    }
  }

  async batchEvents(
    namespace: string,
    events: any[],
    append: (value: string) => void
  ) {
    const startTime = Date.now();
    const requestId = (globalThis as any).crypto.randomUUID();

    this.logger.info("Starting events batch", {
      namespace,
      count: events.length,
      requestId,
    });

    try {
      const payload: types_EventsBatchRequest = { namespace, events };

      await this.httpClient.post("/ingestion/events", payload, {
        headers: {
          "Content-Type": "application/json",
        },
      });

      const duration = Date.now() - startTime;
      this.logger.info("Events batch completed", {
        namespace,
        count: events.length,
        duration,
        requestId,
      });

      append(`✔ events:batch (${events.length})`);
    } catch (error) {
      this.logger.error("Events batch failed", {
        namespace,
        count: events.length,
        requestId,
        error: error instanceof Error ? error.message : String(error),
        isRetryable: error instanceof ApiError ? error.retryable : false,
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
    const requestId = (globalThis as any).crypto.randomUUID();

    this.logger.info("Starting recommendation request", {
      userId,
      namespace,
      k: kVal,
      blend: blendVal,
      requestId,
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

      const result = await this.httpClient.post<types_RecommendResponse>(
        "/ranking/recommendations",
        payload,
        {
          headers: {
            "Content-Type": "application/json",
          },
        }
      );

      const duration = Date.now() - startTime;
      this.logger.info("Recommendation request completed", {
        userId,
        namespace,
        k: kVal,
        duration,
        requestId,
        resultCount: result.data.items?.length || 0,
      });

      return result.data as types_RecommendResponse;
    } catch (error) {
      this.logger.error("Recommendation request failed", {
        userId,
        namespace,
        k: kVal,
        requestId,
        error: error instanceof Error ? error.message : String(error),
        isRetryable: error instanceof ApiError ? error.retryable : false,
      });
      throw error;
    }
  }

  async similar(itemId: string, namespace: string, kVal: number) {
    const startTime = Date.now();
    const requestId = (globalThis as any).crypto.randomUUID();

    this.logger.info("Starting similar items request", {
      itemId,
      namespace,
      k: kVal,
      requestId,
    });

    try {
      const result = await this.httpClient.get(
        `/ranking/items/${itemId}/similar?namespace=${namespace}&k=${kVal}`,
        {
          headers: {
            "Content-Type": "application/json",
          },
        }
      );

      const duration = Date.now() - startTime;
      this.logger.info("Similar items request completed", {
        itemId,
        namespace,
        k: kVal,
        duration,
        requestId,
        resultCount: (result.data as any[])?.length || 0,
      });

      return result.data as any[];
    } catch (error) {
      this.logger.error("Similar items request failed", {
        itemId,
        namespace,
        k: kVal,
        requestId,
        error: error instanceof Error ? error.message : String(error),
        isRetryable: error instanceof ApiError ? error.retryable : false,
      });
      throw error;
    }
  }
}

// Export a default instance for convenience
export const enhancedApiServiceV2 = new EnhancedApiServiceV2();
