/**
 * Embeddings module exports.
 * Provides a unified interface for text embeddings with local and remote providers.
 */

export type {
  EmbeddingsProvider,
  EmbeddingsConfig,
  EmbeddingsResult,
  EmbeddingsError,
  ChunkingOptions,
  BatchingOptions,
} from "./types";

export { LocalEmbeddingsProvider } from "./LocalEmbeddingsProvider";
export { RemoteEmbeddingsProvider } from "./RemoteEmbeddingsProvider";
export { WorkerEmbeddingsProvider } from "./WorkerEmbeddingsProvider";
export { EmbeddingsFactory } from "./EmbeddingsFactory";
export { ChunkingUtils } from "./ChunkingUtils";

// Re-export worker types for external use
export type {
  WorkerMessage,
  WorkerResponse,
  EmbedRequest,
  EmbedResponse,
} from "./EmbeddingsWorker";
