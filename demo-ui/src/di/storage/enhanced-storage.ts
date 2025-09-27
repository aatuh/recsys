/**
 * Enhanced storage implementation with TTL, metadata, and multiple backends.
 */

import type {
  Storage,
  StorageBackend,
  StorageConfig,
  StorageItem,
} from "./types";
import {
  InMemoryStorageBackend,
  LocalStorageBackend,
  SessionStorageBackend,
  HybridStorageBackend,
} from "./backends";

export class EnhancedStorage implements Storage {
  private backend: StorageBackend;
  private config: StorageConfig;

  constructor(backend: StorageBackend, config: StorageConfig = {}) {
    this.backend = backend;
    this.config = {
      prefix: config.prefix || "recsys_",
      defaultTTL: config.defaultTTL || 24 * 60 * 60 * 1000, // 24 hours
      maxSize: config.maxSize || 1000,
    };
  }

  private getPrefixedKey(key: string): string {
    return `${this.config.prefix}${key}`;
  }

  private getMetadataKey(key: string): string {
    return `${this.config.prefix}__meta_${key}`;
  }

  private isExpired(item: StorageItem): boolean {
    if (!item.ttl) {
      return false;
    }
    return Date.now() - item.timestamp > item.ttl;
  }

  private cleanupExpiredItems(): void {
    const keys = this.backend.keys();
    const prefixedKeys = keys.filter((key) =>
      key.startsWith(this.config.prefix!)
    );

    for (const key of prefixedKeys) {
      if (key.startsWith(`${this.config.prefix}__meta_`)) {
        continue; // Skip metadata keys
      }

      const metadataKey = this.getMetadataKey(
        key.replace(this.config.prefix!, "")
      );
      const metadata = this.backend.getItem(metadataKey);

      if (metadata) {
        try {
          const item: StorageItem = JSON.parse(metadata);
          if (this.isExpired(item)) {
            this.removeItem(key.replace(this.config.prefix!, ""));
          }
        } catch {
          // Invalid metadata, remove the item
          this.removeItem(key.replace(this.config.prefix!, ""));
        }
      }
    }
  }

  private enforceMaxSize(): void {
    const keys = this.backend.keys();
    const prefixedKeys = keys.filter((key) =>
      key.startsWith(this.config.prefix!)
    );

    if (prefixedKeys.length > this.config.maxSize!) {
      // Remove oldest items (simple FIFO)
      const sortedKeys = prefixedKeys.sort((a, b) => {
        const metadataA = this.backend.getItem(
          this.getMetadataKey(a.replace(this.config.prefix!, ""))
        );
        const metadataB = this.backend.getItem(
          this.getMetadataKey(b.replace(this.config.prefix!, ""))
        );

        const timestampA = metadataA ? JSON.parse(metadataA).timestamp : 0;
        const timestampB = metadataB ? JSON.parse(metadataB).timestamp : 0;

        return timestampA - timestampB;
      });

      const toRemove = sortedKeys.slice(
        0,
        sortedKeys.length - this.config.maxSize!
      );
      for (const key of toRemove) {
        this.removeItem(key.replace(this.config.prefix!, ""));
      }
    }
  }

  getItem(key: string): string | null {
    const prefixedKey = this.getPrefixedKey(key);
    const value = this.backend.getItem(prefixedKey);

    if (!value) {
      return null;
    }

    // Check if item has expired
    const metadataKey = this.getMetadataKey(key);
    const metadata = this.backend.getItem(metadataKey);

    if (metadata) {
      try {
        const item: StorageItem = JSON.parse(metadata);
        if (this.isExpired(item)) {
          this.removeItem(key);
          return null;
        }
      } catch {
        // Invalid metadata, return the value anyway
      }
    }

    return value;
  }

  setItem(key: string, value: string, ttl?: number): void {
    const prefixedKey = this.getPrefixedKey(key);
    const metadataKey = this.getMetadataKey(key);

    // Store the value
    this.backend.setItem(prefixedKey, value);

    // Store metadata
    const item: StorageItem = {
      key,
      value,
      timestamp: Date.now(),
      ttl: ttl || this.config.defaultTTL,
    };

    this.backend.setItem(metadataKey, JSON.stringify(item));

    // Cleanup and enforce limits
    this.cleanupExpiredItems();
    this.enforceMaxSize();
  }

  removeItem(key: string): void {
    const prefixedKey = this.getPrefixedKey(key);
    const metadataKey = this.getMetadataKey(key);

    this.backend.removeItem(prefixedKey);
    this.backend.removeItem(metadataKey);
  }

  clear(): void {
    const keys = this.backend.keys();
    const prefixedKeys = keys.filter((key) =>
      key.startsWith(this.config.prefix!)
    );

    for (const key of prefixedKeys) {
      this.backend.removeItem(key);
    }
  }

  keys(): string[] {
    const keys = this.backend.keys();
    return keys
      .filter((key) => key.startsWith(this.config.prefix!))
      .filter((key) => !key.startsWith(`${this.config.prefix}__meta_`))
      .map((key) => key.replace(this.config.prefix!, ""));
  }

  size(): number {
    return this.keys().length;
  }

  hasItem(key: string): boolean {
    return this.getItem(key) !== null;
  }

  getItemWithMetadata(key: string): StorageItem | null {
    const value = this.getItem(key);
    if (!value) {
      return null;
    }

    const metadataKey = this.getMetadataKey(key);
    const metadata = this.backend.getItem(metadataKey);

    if (!metadata) {
      return {
        key,
        value,
        timestamp: Date.now(),
      };
    }

    try {
      return JSON.parse(metadata);
    } catch {
      return {
        key,
        value,
        timestamp: Date.now(),
      };
    }
  }

  setItemWithMetadata(item: StorageItem): void {
    const prefixedKey = this.getPrefixedKey(item.key);
    const metadataKey = this.getMetadataKey(item.key);

    // Store the value
    this.backend.setItem(prefixedKey, item.value);

    // Store metadata
    this.backend.setItem(metadataKey, JSON.stringify(item));

    // Cleanup and enforce limits
    this.cleanupExpiredItems();
    this.enforceMaxSize();
  }
}

/**
 * Factory function to create storage instances with different backends.
 */
export function createStorage(
  type: "memory" | "local" | "session" | "hybrid" = "hybrid",
  config?: StorageConfig
): Storage {
  let backend: StorageBackend;

  switch (type) {
    case "memory":
      backend = new InMemoryStorageBackend();
      break;
    case "local":
      backend = new LocalStorageBackend();
      break;
    case "session":
      backend = new SessionStorageBackend();
      break;
    case "hybrid":
    default:
      backend = new HybridStorageBackend([
        new LocalStorageBackend(),
        new SessionStorageBackend(),
        new InMemoryStorageBackend(),
      ]);
      break;
  }

  return new EnhancedStorage(backend, config);
}
