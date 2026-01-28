/**
 * Generic pagination helpers for paginated endpoints.
 */

import React, { useState, useCallback, useMemo } from "react";
import { useQuery, type QueryKey } from "@tanstack/react-query";
import type {
  PaginationParams,
  PaginatedResponse,
  PaginationHelpers,
} from "./types";
import { cancellationManager } from "./CancellationManager";

export function createPaginationHelpers<T>(): PaginationHelpers<T> {
  return {
    usePaginatedQuery: (
      queryKey: QueryKey,
      queryFn: (params: PaginationParams) => Promise<PaginatedResponse<T>>,
      options: {
        initialPage?: number;
        pageSize?: number;
        enabled?: boolean;
      } = {}
    ) => {
      const { initialPage = 1, pageSize = 20, enabled = true } = options;

      const [currentPage, setCurrentPage] = useState(initialPage);
      const [currentPageSize, setCurrentPageSize] = useState(pageSize);

      const paginationParams: PaginationParams = useMemo(
        () => ({
          page: currentPage,
          limit: currentPageSize,
        }),
        [currentPage, currentPageSize]
      );

      // Create a unique key for this paginated query
      const paginatedQueryKey = [
        ...queryKey,
        "paginated",
        currentPage,
        currentPageSize,
      ];
      const requestKey = cancellationManager.generateKey("paginated-query", {
        key: queryKey.join(":"),
        page: currentPage,
        limit: currentPageSize,
      });

      const {
        data,
        isLoading,
        error,
        refetch: _refetch,
      } = useQuery({
        queryKey: paginatedQueryKey,
        queryFn: async () => {
          const controller =
            cancellationManager.createAbortController(requestKey);

          try {
            const result = await queryFn({
              ...paginationParams,
              signal: controller.signal,
            });

            return result;
          } catch (err) {
            if (err instanceof Error && err.name === "AbortError") {
              throw new Error("Request cancelled");
            }
            throw err;
          }
        },
        enabled,
        staleTime: 5 * 60 * 1000, // 5 minutes
        gcTime: 10 * 60 * 1000, // 10 minutes
      });

      const pagination = data?.pagination;
      const items = data?.data || [];

      const hasNextPage = pagination?.hasNext ?? false;
      const hasPrevPage = pagination?.hasPrev ?? false;

      const fetchNextPage = useCallback(() => {
        if (hasNextPage && pagination) {
          setCurrentPage(pagination.page + 1);
        }
      }, [hasNextPage, pagination]);

      const fetchPrevPage = useCallback(() => {
        if (hasPrevPage && pagination) {
          setCurrentPage(pagination.page - 1);
        }
      }, [hasPrevPage, pagination]);

      const goToPage = useCallback(
        (page: number) => {
          if (pagination && page >= 1 && page <= pagination.totalPages) {
            setCurrentPage(page);
          }
        },
        [pagination]
      );

      const setPageSize = useCallback((size: number) => {
        if (size > 0 && size <= 100) {
          // Reasonable limits
          setCurrentPageSize(size);
          setCurrentPage(1); // Reset to first page when changing page size
        }
      }, []);

      return {
        data: items,
        pagination,
        isLoading,
        error: error as Error | null,
        hasNextPage,
        hasPrevPage,
        fetchNextPage,
        fetchPrevPage,
        goToPage,
        setPageSize,
      };
    },
  };
}

/**
 * Hook for infinite scroll pagination.
 */
export function useInfinitePagination<T>(
  queryKey: QueryKey,
  queryFn: (params: PaginationParams) => Promise<PaginatedResponse<T>>,
  options: {
    pageSize?: number;
    enabled?: boolean;
  } = {}
) {
  const { pageSize = 20, enabled = true } = options;

  const [allData, setAllData] = useState<T[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [hasNextPage, setHasNextPage] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);

  const requestKey = cancellationManager.generateKey("infinite-pagination", {
    key: queryKey.join(":"),
    page: currentPage,
    limit: pageSize,
  });

  const { data, isLoading, error } = useQuery({
    queryKey: [...queryKey, "infinite", currentPage, pageSize],
    queryFn: async () => {
      const controller = cancellationManager.createAbortController(requestKey);

      try {
        const result = await queryFn({
          page: currentPage,
          limit: pageSize,
          signal: controller.signal,
        });

        return result;
      } catch (err) {
        if (err instanceof Error && err.name === "AbortError") {
          throw new Error("Request cancelled");
        }
        throw err;
      }
    },
    enabled: enabled && hasNextPage,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });

  // Update accumulated data when new data arrives
  React.useEffect(() => {
    if (data) {
      if (currentPage === 1) {
        // First page - replace all data
        setAllData(data.data);
      } else {
        // Subsequent pages - append data
        setAllData((prev) => [...prev, ...data.data]);
      }

      setHasNextPage(data.pagination.hasNext);
    }
  }, [data, currentPage]);

  const loadMore = useCallback(() => {
    if (hasNextPage && !isLoadingMore) {
      setIsLoadingMore(true);
      setCurrentPage((prev) => prev + 1);
    }
  }, [hasNextPage, isLoadingMore]);

  const reset = useCallback(() => {
    setAllData([]);
    setCurrentPage(1);
    setHasNextPage(true);
    setIsLoadingMore(false);
  }, []);

  // Reset loading state when data changes
  React.useEffect(() => {
    if (data) {
      setIsLoadingMore(false);
    }
  }, [data]);

  return {
    data: allData,
    isLoading: isLoading && currentPage === 1,
    isLoadingMore,
    error: error as Error | null,
    hasNextPage,
    loadMore,
    reset,
    totalPages: data?.pagination.totalPages ?? 0,
    currentPage,
  };
}

/**
 * Hook for cursor-based pagination.
 */
export function useCursorPagination<T>(
  queryKey: QueryKey,
  queryFn: (params: PaginationParams) => Promise<PaginatedResponse<T>>,
  options: {
    pageSize?: number;
    enabled?: boolean;
  } = {}
) {
  const { pageSize = 20, enabled = true } = options;

  const [allData, setAllData] = useState<T[]>([]);
  const [nextCursor, setNextCursor] = useState<string | undefined>();
  const [prevCursor, setPrevCursor] = useState<string | undefined>();
  const [isLoadingMore, setIsLoadingMore] = useState(false);

  const requestKey = cancellationManager.generateKey("cursor-pagination", {
    key: queryKey.join(":"),
    cursor: nextCursor,
    limit: pageSize,
  });

  const { data, isLoading, error } = useQuery({
    queryKey: [...queryKey, "cursor", nextCursor, pageSize],
    queryFn: async () => {
      const controller = cancellationManager.createAbortController(requestKey);

      try {
        const result = await queryFn({
          cursor: nextCursor,
          limit: pageSize,
          signal: controller.signal,
        });

        return result;
      } catch (err) {
        if (err instanceof Error && err.name === "AbortError") {
          throw new Error("Request cancelled");
        }
        throw err;
      }
    },
    enabled: enabled && (nextCursor !== undefined || allData.length === 0),
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });

  // Update data when new data arrives
  React.useEffect(() => {
    if (data) {
      setAllData((prev) => [...prev, ...data.data]);
      setNextCursor(data.pagination.nextCursor);
      setPrevCursor(data.pagination.prevCursor);
    }
  }, [data]);

  const loadMore = useCallback(() => {
    if (nextCursor && !isLoadingMore) {
      setIsLoadingMore(true);
    }
  }, [nextCursor, isLoadingMore]);

  const reset = useCallback(() => {
    setAllData([]);
    setNextCursor(undefined);
    setPrevCursor(undefined);
    setIsLoadingMore(false);
  }, []);

  // Reset loading state when data changes
  React.useEffect(() => {
    if (data) {
      setIsLoadingMore(false);
    }
  }, [data]);

  return {
    data: allData,
    isLoading: isLoading && allData.length === 0,
    isLoadingMore,
    error: error as Error | null,
    hasNextPage: !!nextCursor,
    loadMore,
    reset,
    nextCursor,
    prevCursor,
  };
}
