/**
 * Query library module exports.
 * Provides a unified interface for TanStack Query with enhanced functionality.
 */

// Core exports
export type {
  QueryConfig,
  PaginationParams,
  PaginatedResponse,
  QueryKeyFactory,
  QueryHooks,
  CancellationManager,
  PaginationHelpers,
  QueryProviderProps,
} from "./types";

// Query client
export {
  createQueryClient,
  defaultQueryClient,
  devQueryClient,
  prodQueryClient,
  getQueryClient,
} from "./QueryClient";

// Query provider
export {
  QueryProvider,
  useQueryClient,
  useQueryKeys,
  useCancellationManager,
} from "./QueryProvider";

// Query key factory
export {
  AppQueryKeyFactory,
  queryKeys,
  userQueryKeys,
  itemQueryKeys,
  recommendationQueryKeys,
  embeddingQueryKeys,
  configQueryKeys,
} from "./QueryKeyFactory";

// Cancellation manager
export {
  RequestCancellationManager,
  cancellationManager,
  userCancellationManager,
  itemCancellationManager,
  recommendationCancellationManager,
  embeddingCancellationManager,
} from "./CancellationManager";

// Pagination helpers
export {
  createPaginationHelpers,
  useInfinitePagination,
  useCursorPagination,
} from "./PaginationHelpers";

// Query hooks
export {
  useAppQuery,
  useAppMutation,
  useOptimisticMutation,
  usePaginatedQuery,
  useInfiniteQuery,
  useCursorQuery,
  usePrefetch,
  useInvalidate,
  useSetQueryData,
  useGetQueryData,
  useRemoveQueries,
  useClearQueries,
} from "./QueryHooks";
