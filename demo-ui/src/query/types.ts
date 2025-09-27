/**
 * Query library types and interfaces.
 */

import type {
  QueryKey,
  QueryClient,
  UseQueryOptions,
  UseMutationOptions,
} from "@tanstack/react-query";

export interface QueryConfig {
  staleTime?: number;
  gcTime?: number;
  retry?: boolean | number | ((failureCount: number, error: Error) => boolean);
  retryDelay?: number | ((retryAttempt: number, error: Error) => number);
  refetchOnWindowFocus?: boolean;
  refetchOnMount?: boolean;
  refetchOnReconnect?: boolean;
  enabled?: boolean;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
  offset?: number;
  cursor?: string;
  signal?: AbortSignal;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasNext: boolean;
    hasPrev: boolean;
    nextCursor?: string;
    prevCursor?: string;
  };
}

export interface QueryKeyFactory {
  // Base keys
  readonly base: readonly string[];

  // Scoped keys (for future auth integration)
  scoped(scope: string): readonly string[];

  // Specific query keys
  all(): readonly string[];
  lists(): readonly string[];
  list(filters: Record<string, unknown>): readonly string[];
  details(): readonly string[];
  detail(id: string): readonly string[];

  // Pagination keys
  paginatedList(
    filters: Record<string, unknown>,
    pagination: PaginationParams
  ): readonly string[];

  // Cache invalidation helpers
  invalidateQueries(
    queryClient: QueryClient,
    key: readonly string[]
  ): Promise<void>;
  invalidateAll(queryClient: QueryClient): Promise<void>;
}

export interface QueryHooks<TData = unknown, TError = Error> {
  useQuery: <TQueryData = TData>(
    queryKey: QueryKey,
    queryFn: () => Promise<TQueryData>,
    options?: UseQueryOptions<TQueryData, TError, TQueryData, QueryKey>
  ) => any;

  useMutation: <TVariables = unknown, TContext = unknown>(
    mutationFn: (variables: TVariables) => Promise<TData>,
    options?: UseMutationOptions<TData, TError, TVariables, TContext>
  ) => any;
}

export interface CancellationManager {
  cancel(key: string): void;
  cancelAll(): void;
  isCancelled(key: string): boolean;
  createAbortController(key: string): AbortController;
}

export interface PaginationHelpers<T> {
  usePaginatedQuery: (
    queryKey: QueryKey,
    queryFn: (params: PaginationParams) => Promise<PaginatedResponse<T>>,
    options?: {
      initialPage?: number;
      pageSize?: number;
      enabled?: boolean;
    }
  ) => {
    data: T[] | undefined;
    pagination: PaginatedResponse<T>["pagination"] | undefined;
    isLoading: boolean;
    error: Error | null;
    hasNextPage: boolean;
    hasPrevPage: boolean;
    fetchNextPage: () => void;
    fetchPrevPage: () => void;
    goToPage: (page: number) => void;
    setPageSize: (size: number) => void;
  };
}

export interface QueryProviderProps {
  children: React.ReactNode;
  client?: QueryClient;
  devtools?: boolean;
}
