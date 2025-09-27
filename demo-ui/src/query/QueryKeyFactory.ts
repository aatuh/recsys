/**
 * Centralized query key factory for consistent cache key management.
 * Supports future auth scoping without changing callers.
 */

import type { QueryKey, QueryClient } from "@tanstack/react-query";
import type { QueryKeyFactory, PaginationParams } from "./types";

export class AppQueryKeyFactory implements QueryKeyFactory {
  private readonly baseKeys: readonly string[];
  private readonly scope?: string;

  constructor(baseKeys: readonly string[] = ["app"], scope?: string) {
    this.baseKeys = baseKeys;
    this.scope = scope;
  }

  get base(): readonly string[] {
    return this.scope ? [...this.baseKeys, "scope", this.scope] : this.baseKeys;
  }

  scoped(scope: string): readonly string[] {
    return [...this.baseKeys, "scope", scope];
  }

  all(): readonly string[] {
    return this.base;
  }

  lists(): readonly string[] {
    return [...this.base, "lists"];
  }

  list(filters: Record<string, unknown>): readonly string[] {
    const filterKeys = Object.entries(filters)
      .filter(([, value]) => value !== undefined && value !== null)
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([key, value]) => `${key}:${String(value)}`);

    return [...this.base, "lists", ...filterKeys];
  }

  details(): readonly string[] {
    return [...this.base, "details"];
  }

  detail(id: string): readonly string[] {
    return [...this.base, "details", id];
  }

  paginatedList(
    filters: Record<string, unknown>,
    pagination: PaginationParams
  ): readonly string[] {
    const listKey = this.list(filters);
    const paginationKey = this.buildPaginationKey(pagination);
    return [...listKey, "paginated", ...paginationKey];
  }

  private buildPaginationKey(pagination: PaginationParams): readonly string[] {
    const keys: string[] = [];

    if (pagination.page !== undefined) {
      keys.push(`page:${pagination.page}`);
    }

    if (pagination.limit !== undefined) {
      keys.push(`limit:${pagination.limit}`);
    }

    if (pagination.offset !== undefined) {
      keys.push(`offset:${pagination.offset}`);
    }

    if (pagination.cursor !== undefined) {
      keys.push(`cursor:${pagination.cursor}`);
    }

    return keys;
  }

  async invalidateQueries(
    queryClient: QueryClient,
    key: QueryKey
  ): Promise<void> {
    await queryClient.invalidateQueries({ queryKey: key });
  }

  async invalidateAll(queryClient: QueryClient): Promise<void> {
    await queryClient.invalidateQueries({ queryKey: this.base });
  }
}

// Default factory instance
export const queryKeys = new AppQueryKeyFactory();

// Specific domain factories
export const userQueryKeys = new AppQueryKeyFactory(["app", "users"]);
export const itemQueryKeys = new AppQueryKeyFactory(["app", "items"]);
export const recommendationQueryKeys = new AppQueryKeyFactory([
  "app",
  "recommendations",
]);
export const embeddingQueryKeys = new AppQueryKeyFactory(["app", "embeddings"]);
export const configQueryKeys = new AppQueryKeyFactory(["app", "config"]);
