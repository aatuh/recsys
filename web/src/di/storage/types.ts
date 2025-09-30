/**
 * Storage abstraction types and interfaces.
 */

export interface StorageItem {
  key: string;
  value: string;
  timestamp: number;
  ttl?: number; // Time to live in milliseconds
}

export interface StorageConfig {
  prefix?: string;
  defaultTTL?: number;
  maxSize?: number; // Maximum number of items
}

export interface StorageBackend {
  getItem(key: string): string | null;
  setItem(key: string, value: string): void;
  removeItem(key: string): void;
  clear(): void;
  keys(): string[];
  size(): number;
}

export interface Storage {
  getItem(key: string): string | null;
  setItem(key: string, value: string, ttl?: number): void;
  removeItem(key: string): void;
  clear(): void;
  keys(): string[];
  size(): number;
  hasItem(key: string): boolean;
  getItemWithMetadata(key: string): StorageItem | null;
  setItemWithMetadata(item: StorageItem): void;
}
