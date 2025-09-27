/**
 * Request cancellation manager to prevent race conditions in fast UIs.
 */

import type { CancellationManager } from "./types";

export class RequestCancellationManager implements CancellationManager {
  private controllers = new Map<string, AbortController>();
  private cancelledKeys = new Set<string>();

  cancel(key: string): void {
    const controller = this.controllers.get(key);
    if (controller) {
      controller.abort();
      this.controllers.delete(key);
    }
    this.cancelledKeys.add(key);
  }

  cancelAll(): void {
    for (const [key, controller] of this.controllers) {
      controller.abort();
      this.cancelledKeys.add(key);
    }
    this.controllers.clear();
  }

  isCancelled(key: string): boolean {
    return this.cancelledKeys.has(key);
  }

  createAbortController(key: string): AbortController {
    // Cancel any existing request with the same key
    this.cancel(key);

    // Remove from cancelled set
    this.cancelledKeys.delete(key);

    // Create new controller
    const controller = new AbortController();
    this.controllers.set(key, controller);

    // Clean up when request completes
    controller.signal.addEventListener("abort", () => {
      this.controllers.delete(key);
    });

    return controller;
  }

  /**
   * Generate a unique key for a request based on parameters.
   */
  generateKey(prefix: string, params: Record<string, unknown> = {}): string {
    const sortedParams = Object.entries(params)
      .filter(([, value]) => value !== undefined && value !== null)
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([key, value]) => `${key}:${String(value)}`)
      .join("|");

    return sortedParams ? `${prefix}:${sortedParams}` : prefix;
  }

  /**
   * Create a scoped cancellation manager for a specific domain.
   */
  static createScoped(scope: string): RequestCancellationManager {
    const manager = new RequestCancellationManager();
    const originalGenerateKey = manager.generateKey.bind(manager);

    // Override generateKey to include scope
    manager.generateKey = (
      prefix: string,
      params: Record<string, unknown> = {}
    ) => {
      return originalGenerateKey(`${scope}:${prefix}`, params);
    };

    return manager;
  }
}

// Global cancellation manager
export const cancellationManager = new RequestCancellationManager();

// Scoped managers for different domains
export const userCancellationManager =
  RequestCancellationManager.createScoped("users");
export const itemCancellationManager =
  RequestCancellationManager.createScoped("items");
export const recommendationCancellationManager =
  RequestCancellationManager.createScoped("recommendations");
export const embeddingCancellationManager =
  RequestCancellationManager.createScoped("embeddings");
