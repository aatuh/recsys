import { getHttpClient, getLogger, type HttpClient, type Logger } from "../di";
import {
  ConfigService,
  IngestionService,
  RankingService,
  types_RecommendRequest,
  type types_EventTypeConfigUpsertRequest,
  type types_UsersUpsertRequest,
  type types_ItemsUpsertRequest,
  type types_EventsBatchRequest,
  type types_RecommendResponse,
  type types_Overrides,
} from "../lib/api-client";

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
    users: any[],
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
    items: any[],
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
    events: any[],
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
}

// Export a default instance for convenience
export const enhancedApiService = new EnhancedApiService();
