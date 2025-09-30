/**
 * Storage backend implementations for different storage mechanisms.
 */

import type { StorageBackend } from "./types";

/**
 * In-memory storage backend using Map.
 * Data is lost when the page is refreshed.
 */
export class InMemoryStorageBackend implements StorageBackend {
  private storage = new Map<string, string>();

  getItem(key: string): string | null {
    return this.storage.get(key) || null;
  }

  setItem(key: string, value: string): void {
    this.storage.set(key, value);
  }

  removeItem(key: string): void {
    this.storage.delete(key);
  }

  clear(): void {
    this.storage.clear();
  }

  keys(): string[] {
    return Array.from(this.storage.keys());
  }

  size(): number {
    return this.storage.size;
  }
}

/**
 * LocalStorage backend implementation.
 * Data persists across browser sessions.
 */
export class LocalStorageBackend implements StorageBackend {
  getItem(key: string): string | null {
    try {
      return localStorage.getItem(key);
    } catch {
      return null;
    }
  }

  setItem(key: string, value: string): void {
    try {
      localStorage.setItem(key, value);
    } catch {
      // Silently fail if storage is not available
    }
  }

  removeItem(key: string): void {
    try {
      localStorage.removeItem(key);
    } catch {
      // Silently fail if storage is not available
    }
  }

  clear(): void {
    try {
      localStorage.clear();
    } catch {
      // Silently fail if storage is not available
    }
  }

  keys(): string[] {
    try {
      return Object.keys(localStorage);
    } catch {
      return [];
    }
  }

  size(): number {
    try {
      return localStorage.length;
    } catch {
      return 0;
    }
  }
}

/**
 * SessionStorage backend implementation.
 * Data persists only for the current browser session.
 */
export class SessionStorageBackend implements StorageBackend {
  getItem(key: string): string | null {
    try {
      return sessionStorage.getItem(key);
    } catch {
      return null;
    }
  }

  setItem(key: string, value: string): void {
    try {
      sessionStorage.setItem(key, value);
    } catch {
      // Silently fail if storage is not available
    }
  }

  removeItem(key: string): void {
    try {
      sessionStorage.removeItem(key);
    } catch {
      // Silently fail if storage is not available
    }
  }

  clear(): void {
    try {
      sessionStorage.clear();
    } catch {
      // Silently fail if storage is not available
    }
  }

  keys(): string[] {
    try {
      return Object.keys(sessionStorage);
    } catch {
      return [];
    }
  }

  size(): number {
    try {
      return sessionStorage.length;
    } catch {
      return 0;
    }
  }
}

/**
 * Hybrid storage backend that tries multiple backends in order.
 * Falls back to in-memory if others fail.
 */
export class HybridStorageBackend implements StorageBackend {
  private backends: StorageBackend[];

  constructor(backends: StorageBackend[]) {
    this.backends = backends;
  }

  private getWorkingBackend(): StorageBackend {
    // Try to find a working backend
    for (const backend of this.backends) {
      try {
        // Test if backend is working
        const testKey = "__storage_test__";
        backend.setItem(testKey, "test");
        const result = backend.getItem(testKey);
        backend.removeItem(testKey);

        if (result === "test") {
          return backend;
        }
      } catch {
        // Continue to next backend
      }
    }

    // Fall back to in-memory storage (should always be the last one)
    return (
      this.backends[this.backends.length - 1] || new InMemoryStorageBackend()
    );
  }

  getItem(key: string): string | null {
    return this.getWorkingBackend().getItem(key);
  }

  setItem(key: string, value: string): void {
    this.getWorkingBackend().setItem(key, value);
  }

  removeItem(key: string): void {
    this.getWorkingBackend().removeItem(key);
  }

  clear(): void {
    this.getWorkingBackend().clear();
  }

  keys(): string[] {
    return this.getWorkingBackend().keys();
  }

  size(): number {
    return this.getWorkingBackend().size();
  }
}
