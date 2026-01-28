/**
 * Embeddings provider types and interfaces.
 */

export interface EmbeddingsProvider {
  /**
   * Generate embeddings for a batch of texts.
   * @param texts Array of text strings to embed
   * @returns Promise resolving to array of embedding vectors (number[][])
   */
  embed(texts: string[]): Promise<number[][]>;

  /**
   * Get the dimension of embeddings produced by this provider.
   * @returns The embedding dimension
   */
  getDimension(): number;

  /**
   * Get the model name/identifier used by this provider.
   * @returns The model identifier
   */
  getModelName(): string;

  /**
   * Check if the provider is ready to process requests.
   * @returns Promise resolving to true when ready
   */
  isReady(): Promise<boolean>;

  /**
   * Clean up resources when the provider is no longer needed.
   */
  dispose(): Promise<void>;
}

export interface EmbeddingsConfig {
  model?: string;
  dimension?: number;
  maxBatchSize?: number;
  maxTextLength?: number;
  timeout?: number;
  useWorker?: boolean;
  workerUrl?: string;
}

export interface EmbeddingsResult {
  embeddings: number[][];
  model: string;
  dimension: number;
  processingTime: number;
  batchSize: number;
}

export interface EmbeddingsError extends Error {
  code:
    | "MODEL_LOAD_ERROR"
    | "EMBEDDING_ERROR"
    | "TIMEOUT_ERROR"
    | "NETWORK_ERROR"
    | "VALIDATION_ERROR";
  provider: string;
  model?: string;
  retryable: boolean;
}

export interface ChunkingOptions {
  maxChunkSize: number;
  overlap: number;
  strategy: "sentence" | "word" | "character";
  preserveWhitespace: boolean;
}

export interface BatchingOptions {
  maxBatchSize: number;
  maxConcurrentBatches: number;
  timeout: number;
  retryAttempts: number;
}
