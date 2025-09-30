/**
 * Dependency injection interfaces for the web UI application.
 * These interfaces define contracts that can be easily swapped for testing
 * or different environments.
 */

export interface HttpClient {
  get<T>(
    url: string,
    options?: Partial<import("./http/types").HttpRequest>
  ): Promise<import("./http/types").HttpResponse<T>>;
  post<T>(
    url: string,
    data?: unknown,
    options?: Partial<import("./http/types").HttpRequest>
  ): Promise<import("./http/types").HttpResponse<T>>;
  put<T>(
    url: string,
    data?: unknown,
    options?: Partial<import("./http/types").HttpRequest>
  ): Promise<import("./http/types").HttpResponse<T>>;
  delete<T>(
    url: string,
    options?: Partial<import("./http/types").HttpRequest>
  ): Promise<import("./http/types").HttpResponse<T>>;

  addRequestInterceptor(
    interceptor: import("./http/types").RequestInterceptor
  ): void;
  addResponseInterceptor(
    interceptor: import("./http/types").ResponseInterceptor
  ): void;
  removeRequestInterceptor(
    interceptor: import("./http/types").RequestInterceptor
  ): void;
  removeResponseInterceptor(
    interceptor: import("./http/types").ResponseInterceptor
  ): void;

  createCancellationToken(): import("./http/types").CancellationToken;
}

export interface Storage {
  getItem(key: string): string | null;
  setItem(key: string, value: string, ttl?: number): void;
  removeItem(key: string): void;
  clear(): void;
  keys(): string[];
  size(): number;
  hasItem(key: string): boolean;
  getItemWithMetadata(
    key: string
  ): import("./storage/types").StorageItem | null;
  setItemWithMetadata(item: import("./storage/types").StorageItem): void;
}

export interface Logger {
  debug(message: string, fields?: Record<string, unknown>): void;
  info(message: string, fields?: Record<string, unknown>): void;
  warn(message: string, fields?: Record<string, unknown>): void;
  error(message: string, fields?: Record<string, unknown>): void;
  child(fields: Record<string, unknown>): Logger;
}

export interface Embeddings {
  embedText(text: string): Promise<number[]>;
  itemToText(item: {
    item_id: string;
    tags?: string[];
    price?: number;
    props?: Record<string, any>;
  }): string;
}

/**
 * Container interface for dependency injection.
 */
export interface Container {
  getHttpClient(): HttpClient;
  getStorage(): Storage;
  getLogger(): Logger;
  getEmbeddings(): Embeddings;
}
