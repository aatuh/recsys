/**
 * Dependency injection module exports.
 * Provides a clean API for accessing services throughout the application.
 */

export type {
  HttpClient,
  Storage,
  Logger,
  Embeddings,
  Container,
} from "./interfaces";

export { BrowserEmbeddings } from "./implementations";

export { StructuredLogger, NoOpLogger } from "./logger";

export {
  DefaultContainer,
  getContainer,
  setContainer,
  resetContainer,
  getHttpClient,
  getStorage,
  getLogger,
  getEmbeddings,
} from "./container";

// Export HTTP module
export * from "./http";

// Export Storage module
export * from "./storage";
