/**
 * Preconfigured QueryClient with sensible defaults.
 */

import { QueryClient } from "@tanstack/react-query";
// import type { QueryConfig } from './types';

export interface QueryClientConfig {
  defaultStaleTime?: number;
  defaultGcTime?: number;
  defaultRetry?:
    | boolean
    | number
    | ((failureCount: number, error: Error) => boolean);
  defaultRetryDelay?: number | ((retryAttempt: number, error: Error) => number);
  defaultRefetchOnWindowFocus?: boolean;
  defaultRefetchOnMount?: boolean;
  defaultRefetchOnReconnect?: boolean;
  enableDevtools?: boolean;
}

export function createQueryClient(config: QueryClientConfig = {}): QueryClient {
  const {
    defaultStaleTime = 5 * 60 * 1000, // 5 minutes
    defaultGcTime = 10 * 60 * 1000, // 10 minutes
    defaultRetry = (failureCount: number, error: Error) => {
      // Don't retry on 4xx errors (client errors)
      if (error.message.includes("4")) {
        return false;
      }
      // Retry up to 3 times for other errors
      return failureCount < 3;
    },
    defaultRetryDelay = (retryAttempt: number) =>
      Math.min(1000 * 2 ** retryAttempt, 30000),
    defaultRefetchOnWindowFocus = false,
    defaultRefetchOnMount = true,
    defaultRefetchOnReconnect = true,
  } = config;

  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: defaultStaleTime,
        gcTime: defaultGcTime,
        retry: defaultRetry,
        retryDelay: defaultRetryDelay,
        refetchOnWindowFocus: defaultRefetchOnWindowFocus,
        refetchOnMount: defaultRefetchOnMount,
        refetchOnReconnect: defaultRefetchOnReconnect,
        // Network mode for better offline support
        networkMode: "online",
      },
      mutations: {
        retry: false, // Don't retry mutations by default
        networkMode: "online",
      },
    },
    // Global error handlers are configured via defaultOptions
  });
}

// Default query client instance
export const defaultQueryClient = createQueryClient({
  enableDevtools: import.meta.env.DEV,
});

// Development query client with more aggressive caching
export const devQueryClient = createQueryClient({
  defaultStaleTime: 0, // Always stale in development
  defaultGcTime: 5 * 60 * 1000, // 5 minutes
  defaultRefetchOnWindowFocus: true,
  enableDevtools: true,
});

// Production query client with conservative caching
export const prodQueryClient = createQueryClient({
  defaultStaleTime: 10 * 60 * 1000, // 10 minutes
  defaultGcTime: 30 * 60 * 1000, // 30 minutes
  defaultRefetchOnWindowFocus: false,
  enableDevtools: false,
});

/**
 * Get the appropriate query client based on environment.
 */
export function getQueryClient(): QueryClient {
  const isDevelopment = import.meta.env.DEV;
  return isDevelopment ? devQueryClient : prodQueryClient;
}
