/**
 * Storage module exports for enhanced storage functionality.
 */

export type {
  StorageItem,
  StorageConfig,
  StorageBackend,
  Storage,
} from "./types";

export {
  InMemoryStorageBackend,
  LocalStorageBackend,
  SessionStorageBackend,
  HybridStorageBackend,
} from "./backends";

export { EnhancedStorage, createStorage } from "./enhanced-storage";
