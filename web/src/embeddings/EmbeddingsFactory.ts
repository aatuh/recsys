/**
 * Factory for creating embeddings providers based on configuration and feature flags.
 */

import type { EmbeddingsProvider, EmbeddingsConfig } from "./types";
import { LocalEmbeddingsProvider } from "./LocalEmbeddingsProvider";
import { RemoteEmbeddingsProvider } from "./RemoteEmbeddingsProvider";
import { getLogger } from "../di";

export type ProviderType = "local" | "remote" | "auto";

export interface EmbeddingsFactoryConfig {
  provider?: ProviderType;
  localConfig?: EmbeddingsConfig;
  remoteConfig?: EmbeddingsConfig;
  fallbackToLocal?: boolean;
}

export class EmbeddingsFactory {
  private static instance: EmbeddingsFactory | null = null;
  private logger = getLogger().child({ component: "EmbeddingsFactory" });
  private config: EmbeddingsFactoryConfig;

  constructor(config: EmbeddingsFactoryConfig = {}) {
    this.config = {
      provider: "auto",
      fallbackToLocal: true,
      ...config,
    };
  }

  static getInstance(config?: EmbeddingsFactoryConfig): EmbeddingsFactory {
    if (!EmbeddingsFactory.instance) {
      EmbeddingsFactory.instance = new EmbeddingsFactory(config);
    }
    return EmbeddingsFactory.instance;
  }

  /**
   * Create an embeddings provider based on configuration and feature flags.
   */
  async createProvider(
    useRemoteEmbeddings: boolean = false,
    customConfig?: EmbeddingsConfig
  ): Promise<EmbeddingsProvider> {
    const providerType = this.determineProviderType(useRemoteEmbeddings);

    this.logger.info("Creating embeddings provider", {
      providerType,
      useRemoteEmbeddings,
      hasCustomConfig: !!customConfig,
    });

    try {
      switch (providerType) {
        case "local":
          return this.createLocalProvider(customConfig);

        case "remote": {
          return this.createRemoteProvider(customConfig);
        }

        case "auto":
        default:
          return this.createAutoProvider(useRemoteEmbeddings, customConfig);
      }
    } catch (error) {
      this.logger.error("Failed to create embeddings provider", {
        providerType,
        error: error instanceof Error ? error.message : String(error),
      });

      // Fallback to local provider if configured
      if (this.config.fallbackToLocal && providerType !== "local") {
        this.logger.warn("Falling back to local embeddings provider");
        return this.createLocalProvider(customConfig);
      }

      throw error;
    }
  }

  private determineProviderType(useRemoteEmbeddings: boolean): ProviderType {
    if (this.config.provider === "local") {
      return "local";
    }

    if (this.config.provider === "remote") {
      return "remote";
    }

    // Auto mode: use remote if flag is enabled, otherwise local
    return useRemoteEmbeddings ? "remote" : "local";
  }

  private createLocalProvider(
    customConfig?: EmbeddingsConfig
  ): LocalEmbeddingsProvider {
    const config = {
      ...this.config.localConfig,
      ...customConfig,
    };

    this.logger.debug("Creating local embeddings provider", { config });
    return new LocalEmbeddingsProvider(config);
  }

  private createRemoteProvider(
    customConfig?: EmbeddingsConfig
  ): RemoteEmbeddingsProvider {
    const config = {
      ...this.config.remoteConfig,
      ...customConfig,
    };

    this.logger.debug("Creating remote embeddings provider", { config });
    return new RemoteEmbeddingsProvider(config);
  }

  private async createAutoProvider(
    useRemoteEmbeddings: boolean,
    customConfig?: EmbeddingsConfig
  ): Promise<EmbeddingsProvider> {
    if (useRemoteEmbeddings) {
      try {
        return this.createRemoteProvider(customConfig);
      } catch (error) {
        this.logger.warn("Remote provider failed, falling back to local", {
          error,
        });
        if (this.config.fallbackToLocal) {
          return this.createLocalProvider(customConfig);
        }
        throw error;
      }
    } else {
      return this.createLocalProvider(customConfig);
    }
  }

  /**
   * Get the recommended provider configuration for the current environment.
   */
  getRecommendedConfig(): EmbeddingsConfig {
    const isDevelopment =
      (globalThis as any).process?.env?.NODE_ENV === "development";

    return {
      model: "Xenova/all-MiniLM-L6-v2",
      dimension: 384,
      maxBatchSize: isDevelopment ? 16 : 32,
      maxTextLength: 512,
      timeout: 30000,
      useWorker: !isDevelopment, // Use worker in production
    };
  }

  /**
   * Test if a provider type is available in the current environment.
   */
  async isProviderAvailable(type: ProviderType): Promise<boolean> {
    try {
      switch (type) {
        case "local":
          // Local provider is always available
          return true;

        case "remote": {
          // Test if remote endpoint is reachable
          const remoteProvider = new RemoteEmbeddingsProvider();
          return await remoteProvider.isReady();
        }

        default:
          return false;
      }
    } catch (error) {
      this.logger.debug("Provider availability check failed", { type, error });
      return false;
    }
  }

  /**
   * Get information about available providers.
   */
  async getProviderInfo(): Promise<{
    available: ProviderType[];
    recommended: ProviderType;
    local: { model: string; dimension: number };
    remote: { endpoint?: string; available: boolean };
  }> {
    const available: ProviderType[] = [];
    const localAvailable = await this.isProviderAvailable("local");
    const remoteAvailable = await this.isProviderAvailable("remote");

    if (localAvailable) available.push("local");
    if (remoteAvailable) available.push("remote");

    return {
      available,
      recommended: available.includes("local") ? "local" : "remote",
      local: {
        model: this.config.localConfig?.model || "Xenova/all-MiniLM-L6-v2",
        dimension: this.config.localConfig?.dimension || 384,
      },
      remote: {
        endpoint: this.config.remoteConfig?.model, // Using model field for endpoint
        available: remoteAvailable,
      },
    };
  }
}
