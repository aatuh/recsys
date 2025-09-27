/**
 * Query provider component with preconfigured QueryClient.
 */

import React from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import type { QueryProviderProps } from "./types";
import { getQueryClient } from "./QueryClient";

export function QueryProvider({
  children,
  client,
  devtools = process.env.NODE_ENV === "development",
}: QueryProviderProps) {
  const queryClient = client || getQueryClient();

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      {devtools && <ReactQueryDevtools initialIsOpen={false} />}
    </QueryClientProvider>
  );
}

/**
 * Hook to access the query client from context.
 */
export function useQueryClient() {
  const {
    useQueryClient: useTanStackQueryClient,
  } = require("@tanstack/react-query");
  return useTanStackQueryClient();
}

/**
 * Hook to access query keys factory.
 */
export function useQueryKeys() {
  return {
    users: require("./QueryKeyFactory").userQueryKeys,
    items: require("./QueryKeyFactory").itemQueryKeys,
    recommendations: require("./QueryKeyFactory").recommendationQueryKeys,
    embeddings: require("./QueryKeyFactory").embeddingQueryKeys,
    config: require("./QueryKeyFactory").configQueryKeys,
  };
}

/**
 * Hook to access cancellation manager.
 */
export function useCancellationManager() {
  return require("./CancellationManager").cancellationManager;
}
