/**
 * React hook for embeddings functionality.
 * Provides easy access to embeddings providers with feature flag integration.
 */

import { useState, useEffect, useCallback, useRef } from "react";
import { useFeatureFlags } from "../contexts/FeatureFlagsContext";
import {
  EmbeddingsFactory,
  type EmbeddingsProvider,
  type EmbeddingsConfig,
} from "../embeddings";
import { getLogger } from "../di";

export interface UseEmbeddingsOptions {
  config?: EmbeddingsConfig;
  autoInitialize?: boolean;
}

export interface UseEmbeddingsReturn {
  provider: EmbeddingsProvider | null;
  isReady: boolean;
  isLoading: boolean;
  error: Error | null;
  embed: (texts: string[]) => Promise<number[][]>;
  initialize: () => Promise<void>;
  dispose: () => Promise<void>;
  isProviderReady: () => Promise<boolean>;
}

export function useEmbeddings(
  options: UseEmbeddingsOptions = {}
): UseEmbeddingsReturn {
  const { isEnabled } = useFeatureFlags();
  const [provider, setProvider] = useState<EmbeddingsProvider | null>(null);
  const [isReady, setIsReady] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const logger = getLogger().child({ hook: "useEmbeddings" });
  const factoryRef = useRef<EmbeddingsFactory | null>(null);
  const initializationRef = useRef<Promise<void> | null>(null);

  const { config = {}, autoInitialize = true } = options;

  // Initialize provider when feature flags change
  useEffect(() => {
    if (autoInitialize) {
      initialize();
    }
  }, [isEnabled("useRemoteEmbeddings")]);

  const initialize = useCallback(async (): Promise<void> => {
    if (isLoading || isReady) {
      return;
    }

    // Prevent multiple simultaneous initializations
    if (initializationRef.current) {
      return initializationRef.current;
    }

    setIsLoading(true);
    setError(null);

    initializationRef.current = (async () => {
      try {
        logger.debug("Initializing embeddings provider", {
          useRemoteEmbeddings: isEnabled("useRemoteEmbeddings"),
          config,
        });

        // Create factory if not exists
        if (!factoryRef.current) {
          factoryRef.current = EmbeddingsFactory.getInstance();
        }

        // Create provider based on feature flags
        const newProvider = await factoryRef.current.createProvider(
          isEnabled("useRemoteEmbeddings"),
          config
        );

        // Check if provider is ready
        const ready = await newProvider.isReady();
        if (!ready) {
          throw new Error("Provider initialization failed");
        }

        setProvider(newProvider);
        setIsReady(true);

        logger.info("Embeddings provider initialized", {
          model: newProvider.getModelName(),
          dimension: newProvider.getDimension(),
          provider: newProvider.constructor.name,
        });
      } catch (err) {
        const error = err instanceof Error ? err : new Error(String(err));
        setError(error);
        setIsReady(false);
        logger.error("Failed to initialize embeddings provider", {
          error: error.message,
        });
        throw error;
      } finally {
        setIsLoading(false);
        initializationRef.current = null;
      }
    })();

    return initializationRef.current;
  }, [isEnabled, config, isLoading, isReady, logger]);

  const embed = useCallback(
    async (texts: string[]): Promise<number[][]> => {
      if (!provider) {
        throw new Error("Embeddings provider not initialized");
      }

      if (!isReady) {
        throw new Error("Embeddings provider not ready");
      }

      try {
        logger.debug("Generating embeddings", { textCount: texts.length });
        const startTime = Date.now();

        const embeddings = await provider.embed(texts);

        const processingTime = Date.now() - startTime;
        logger.debug("Embeddings generated successfully", {
          textCount: texts.length,
          dimension: embeddings[0]?.length || 0,
          processingTime,
        });

        return embeddings;
      } catch (err) {
        const error = err instanceof Error ? err : new Error(String(err));
        logger.error("Embedding generation failed", {
          textCount: texts.length,
          error: error.message,
        });
        throw error;
      }
    },
    [provider, isReady, logger]
  );

  const dispose = useCallback(async (): Promise<void> => {
    if (provider) {
      try {
        await provider.dispose();
        logger.debug("Embeddings provider disposed");
      } catch (err) {
        logger.warn("Error disposing embeddings provider", { error: err });
      }
    }

    setProvider(null);
    setIsReady(false);
    setIsLoading(false);
    setError(null);
  }, [provider, logger]);

  const isProviderReady = useCallback(async (): Promise<boolean> => {
    if (!provider) {
      return false;
    }

    try {
      return await provider.isReady();
    } catch (err) {
      logger.warn("Error checking provider readiness", { error: err });
      return false;
    }
  }, [provider, logger]);

  return {
    provider,
    isReady,
    isLoading,
    error,
    embed,
    initialize,
    dispose,
    isProviderReady,
  };
}

// Convenience hook for simple embedding generation
export function useEmbedText() {
  const { embed, isReady, error } = useEmbeddings();

  return {
    embedText: embed,
    isReady,
    error,
  };
}

// Hook for batch embedding with chunking
export function useBatchEmbeddings(
  options: UseEmbeddingsOptions & {
    maxBatchSize?: number;
    chunking?: boolean;
  } = {}
) {
  const { embed, isReady, error } = useEmbeddings(options);
  const { maxBatchSize = 32 } = options;

  const embedBatch = useCallback(
    async (texts: string[]): Promise<number[][]> => {
      if (!isReady) {
        throw new Error("Embeddings provider not ready");
      }

      if (texts.length <= maxBatchSize) {
        return embed(texts);
      }

      // Process in batches
      const results: number[][] = [];
      for (let i = 0; i < texts.length; i += maxBatchSize) {
        const batch = texts.slice(i, i + maxBatchSize);
        const batchEmbeddings = await embed(batch);
        results.push(...batchEmbeddings);
      }

      return results;
    },
    [embed, isReady, maxBatchSize]
  );

  return {
    embedBatch,
    isReady,
    error,
  };
}
