/**
 * Custom query hooks with enhanced functionality.
 */

import React from "react";
import {
  useQuery,
  useMutation,
  useQueryClient,
  type QueryKey,
  type UseQueryOptions,
  type UseMutationOptions,
} from "@tanstack/react-query";
import type { QueryConfig, PaginationParams } from "./types";
import { cancellationManager } from "./CancellationManager";
import { createPaginationHelpers } from "./PaginationHelpers";

/**
 * Enhanced useQuery hook with automatic cancellation.
 */
export function useAppQuery<TData = unknown, TError = Error>(
  queryKey: QueryKey,
  queryFn: () => Promise<TData>,
  options: Partial<UseQueryOptions<TData, TError, TData, QueryKey>> &
    QueryConfig = {}
) {
  const requestKey = cancellationManager.generateKey("query", {
    key: queryKey.join(":"),
  });

  return useQuery({
    ...options,
    queryKey,
    queryFn: async () => {
      const _controller = cancellationManager.createAbortController(requestKey);

      try {
        const result = await queryFn();
        return result;
      } catch (error) {
        if (error instanceof Error && error.name === "AbortError") {
          throw new Error("Request cancelled");
        }
        throw error;
      }
    },
  });
}

/**
 * Enhanced useMutation hook with automatic cancellation.
 */
export function useAppMutation<
  TData = unknown,
  TError = Error,
  TVariables = unknown,
  TContext = unknown
>(
  mutationFn: (variables: TVariables) => Promise<TData>,
  options: UseMutationOptions<TData, TError, TVariables, TContext> = {}
) {
  const _queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (variables: TVariables) => {
      const requestKey = cancellationManager.generateKey("mutation", {
        variables: JSON.stringify(variables),
      });

      const _controller = cancellationManager.createAbortController(requestKey);

      try {
        const result = await mutationFn(variables);
        return result;
      } catch (error) {
        if (error instanceof Error && error.name === "AbortError") {
          throw new Error("Request cancelled");
        }
        throw error;
      }
    },
    ...options,
  });
}

/**
 * Hook for optimistic updates.
 */
export function useOptimisticMutation<
  TData = unknown,
  TError = Error,
  TVariables = unknown,
  TContext = { previousData: unknown }
>(
  mutationFn: (variables: TVariables) => Promise<TData>,
  options: {
    queryKey: QueryKey;
    optimisticUpdate: (variables: TVariables) => TData;
    rollbackUpdate: (variables: TVariables) => TData;
  } & UseMutationOptions<TData, TError, TVariables, TContext>
) {
  const _queryClient = useQueryClient();

  return useMutation({
    mutationFn,
    onMutate: async (variables) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: options.queryKey });

      // Snapshot the previous value
      const previousData = queryClient.getQueryData(options.queryKey);

      // Optimistically update to the new value
      queryClient.setQueryData(
        options.queryKey,
        options.optimisticUpdate(variables)
      );

      // Return a context object with the snapshotted value
      return { previousData } as TContext;
    },
    onError: (error, variables, context) => {
      // If the mutation fails, use the context returned from onMutate to roll back
      if (context && typeof context === "object" && "previousData" in context) {
        queryClient.setQueryData(
          options.queryKey,
          (context as any).previousData
        );
      }
    },
    onSettled: () => {
      // Always refetch after error or success
      queryClient.invalidateQueries({ queryKey: options.queryKey });
    },
    ...options,
  });
}

/**
 * Hook for paginated queries.
 */
export function usePaginatedQuery<T>(
  queryKey: QueryKey,
  queryFn: (
    params: PaginationParams
  ) => Promise<{ data: T[]; pagination: any }>,
  options: {
    initialPage?: number;
    pageSize?: number;
    enabled?: boolean;
  } = {}
) {
  const paginationHelpers = createPaginationHelpers<T>();
  return paginationHelpers.usePaginatedQuery(queryKey, queryFn, options);
}

/**
 * Hook for infinite scroll pagination.
 */
export function useInfiniteQuery<T>(
  queryKey: QueryKey,
  queryFn: (
    params: PaginationParams
  ) => Promise<{ data: T[]; pagination: any }>,
  options: {
    pageSize?: number;
    enabled?: boolean;
  } = {}
) {
  const { useInfinitePagination } = require("./PaginationHelpers");
  return useInfinitePagination(queryKey, queryFn, options);
}

/**
 * Hook for cursor-based pagination.
 */
export function useCursorQuery<T>(
  queryKey: QueryKey,
  queryFn: (
    params: PaginationParams
  ) => Promise<{ data: T[]; pagination: any }>,
  options: {
    pageSize?: number;
    enabled?: boolean;
  } = {}
) {
  const { useCursorPagination } = require("./PaginationHelpers");
  return useCursorPagination(queryKey, queryFn, options);
}

/**
 * Hook for prefetching data.
 */
export function usePrefetch() {
  const _queryClient = useQueryClient();

  return React.useCallback(
    async (queryKey: QueryKey, queryFn: () => Promise<any>) => {
      await queryClient.prefetchQuery({
        queryKey,
        queryFn,
        staleTime: 5 * 60 * 1000, // 5 minutes
      });
    },
    [queryClient]
  );
}

/**
 * Hook for invalidating queries.
 */
export function useInvalidate() {
  const _queryClient = useQueryClient();

  return React.useCallback(
    async (queryKey: QueryKey) => {
      await queryClient.invalidateQueries({ queryKey });
    },
    [queryClient]
  );
}

/**
 * Hook for setting query data.
 */
export function useSetQueryData() {
  const _queryClient = useQueryClient();

  return React.useCallback(
    (queryKey: QueryKey, data: any) => {
      queryClient.setQueryData(queryKey, data);
    },
    [queryClient]
  );
}

/**
 * Hook for getting query data.
 */
export function useGetQueryData() {
  const _queryClient = useQueryClient();

  return React.useCallback(
    (queryKey: QueryKey): any => {
      return queryClient.getQueryData(queryKey);
    },
    [queryClient]
  );
}

/**
 * Hook for removing queries.
 */
export function useRemoveQueries() {
  const _queryClient = useQueryClient();

  return React.useCallback(
    (queryKey: QueryKey) => {
      queryClient.removeQueries({ queryKey });
    },
    [queryClient]
  );
}

/**
 * Hook for clearing all queries.
 */
export function useClearQueries() {
  const _queryClient = useQueryClient();

  return React.useCallback(() => {
    queryClient.clear();
  }, [queryClient]);
}
