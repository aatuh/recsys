/**
 * Example component demonstrating query library functionality.
 * Shows pagination, cancellation, and optimistic updates.
 */

import React, { useState } from "react";
import {
  useAppQuery,
  useAppMutation,
  useOptimisticMutation,
  usePaginatedQuery,
  useInfiniteQuery,
  usePrefetch,
  useInvalidate,
  userQueryKeys,
  cancellationManager,
} from "../../query";
import { Button } from "../primitives/UIComponents";

// Mock API functions
const mockUsers = Array.from({ length: 100 }, (_, i) => ({
  id: `user-${i + 1}`,
  name: `User ${i + 1}`,
  email: `user${i + 1}@example.com`,
  status: i % 3 === 0 ? "active" : i % 3 === 1 ? "inactive" : "pending",
}));

const fetchUsers = async (
  params: { page?: number; limit?: number; search?: string } = {}
) => {
  const { page = 1, limit = 10, search = "" } = params;

  // Simulate network delay
  await new Promise((resolve) => setTimeout(resolve, 1000));

  let filteredUsers = mockUsers;
  if (search) {
    filteredUsers = mockUsers.filter(
      (user) =>
        user.name.toLowerCase().includes(search.toLowerCase()) ||
        user.email.toLowerCase().includes(search.toLowerCase())
    );
  }

  const startIndex = (page - 1) * limit;
  const endIndex = startIndex + limit;
  const paginatedUsers = filteredUsers.slice(startIndex, endIndex);

  return {
    data: paginatedUsers,
    pagination: {
      page,
      limit,
      total: filteredUsers.length,
      totalPages: Math.ceil(filteredUsers.length / limit),
      hasNext: endIndex < filteredUsers.length,
      hasPrev: page > 1,
    },
  };
};

const fetchUser = async (id: string) => {
  await new Promise((resolve) => setTimeout(resolve, 500));
  return mockUsers.find((user) => user.id === id) || null;
};

const updateUser = async (
  id: string,
  updates: { name?: string; status?: string }
) => {
  await new Promise((resolve) => setTimeout(resolve, 1000));
  const user = mockUsers.find((u) => u.id === id);
  if (user) {
    Object.assign(user, updates);
    return user;
  }
  throw new Error("User not found");
};

export function QueryExample() {
  const [logs, setLogs] = useState<string[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);

  const addLog = (message: string) => {
    setLogs((prev) => [...prev, `${new Date().toISOString()}: ${message}`]);
  };

  // Basic query example
  const {
    data: userData,
    isLoading: userLoading,
    error: userError,
  } = useAppQuery(
    userQueryKeys.detail(selectedUserId || ""),
    () => fetchUser(selectedUserId!),
    {
      enabled: !!selectedUserId,
      staleTime: 5 * 60 * 1000, // 5 minutes
    }
  );

  // Paginated query example
  const {
    data: paginatedUsers,
    pagination,
    isLoading: paginatedLoading,
    hasNextPage,
    hasPrevPage,
    fetchNextPage,
    fetchPrevPage,
    goToPage: _goToPage,
    setPageSize,
  } = usePaginatedQuery(
    userQueryKeys.list({ search: searchTerm }),
    (params) => fetchUsers({ ...params, search: searchTerm }),
    {
      initialPage: 1,
      pageSize: 5,
    }
  );

  // Infinite query example
  const {
    data: infiniteUsers,
    isLoading: infiniteLoading,
    isLoadingMore,
    hasNextPage: hasMoreInfinite,
    loadMore: loadMoreInfinite,
    reset: resetInfinite,
  } = useInfiniteQuery(userQueryKeys.lists(), (params) => fetchUsers(params), {
    pageSize: 8,
  });

  // Mutation example
  const updateUserMutation = useAppMutation(
    ({
      id,
      updates,
    }: {
      id: string;
      updates: { name?: string; status?: string };
    }) => updateUser(id, updates),
    {
      onSuccess: (data) => {
        addLog(`✅ User updated: ${data?.name || "Unknown"}`);
      },
      onError: (error) => {
        addLog(`❌ Update failed: ${error.message}`);
      },
    }
  );

  // Optimistic mutation example
  const optimisticUpdateMutation = useOptimisticMutation(
    ({
      id,
      updates,
    }: {
      id: string;
      updates: { name?: string; status?: string };
    }) => updateUser(id, updates),
    {
      queryKey: userQueryKeys.detail(selectedUserId || ""),
      optimisticUpdate: ({ updates }) => {
        if (userData) {
          return { ...userData, ...updates };
        }
        return userData;
      },
      rollbackUpdate: ({ updates }) => {
        if (userData) {
          return { ...userData, ...updates };
        }
        return userData;
      },
      onSuccess: (data) => {
        addLog(`✅ Optimistic update successful: ${data?.name || "Unknown"}`);
      },
      onError: (error) => {
        addLog(`❌ Optimistic update failed: ${error.message}`);
      },
    }
  );

  // Prefetch example
  const prefetch = usePrefetch();
  const invalidate = useInvalidate();

  const handlePrefetchUser = async (userId: string) => {
    try {
      await prefetch(userQueryKeys.detail(userId), () => fetchUser(userId));
      addLog(`🔄 Prefetched user: ${userId}`);
    } catch (error) {
      addLog(
        `❌ Prefetch failed: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  };

  const handleInvalidateUser = async (userId: string) => {
    try {
      await invalidate(userQueryKeys.detail(userId));
      addLog(`🔄 Invalidated user: ${userId}`);
    } catch (error) {
      addLog(
        `❌ Invalidation failed: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  };

  const handleCancelRequests = () => {
    cancellationManager.cancelAll();
    addLog("🛑 Cancelled all requests");
  };

  const handleUpdateUser = () => {
    if (selectedUserId) {
      updateUserMutation.mutate({
        id: selectedUserId,
        updates: {
          name: `Updated User ${Date.now()}`,
          status: "active",
        },
      });
    }
  };

  const handleOptimisticUpdate = () => {
    if (selectedUserId) {
      optimisticUpdateMutation.mutate({
        id: selectedUserId,
        updates: {
          name: `Optimistic User ${Date.now()}`,
          status: "pending",
        },
      });
    }
  };

  const handleClearLogs = () => {
    setLogs([]);
  };

  return (
    <div style={{ padding: "20px", border: "1px solid #ccc", margin: "10px" }}>
      <h3>Query Library Example</h3>
      <p style={{ fontSize: "14px", color: "#666", marginBottom: "20px" }}>
        Demonstrates TanStack Query integration with pagination, cancellation,
        and optimistic updates.
      </p>

      <div style={{ marginBottom: "20px" }}>
        <h4>User Selection</h4>
        <div style={{ display: "flex", gap: "10px", marginBottom: "10px" }}>
          <select
            value={selectedUserId || ""}
            onChange={(e) => setSelectedUserId(e.target.value || null)}
            style={{ padding: "5px", minWidth: "200px" }}
          >
            <option value="">Select a user...</option>
            {mockUsers.slice(0, 10).map((user) => (
              <option key={user.id} value={user.id}>
                {user.name} ({user.status})
              </option>
            ))}
          </select>
          <Button
            onClick={() => selectedUserId && handlePrefetchUser(selectedUserId)}
            disabled={!selectedUserId}
          >
            Prefetch User
          </Button>
          <Button
            onClick={() =>
              selectedUserId && handleInvalidateUser(selectedUserId)
            }
            disabled={!selectedUserId}
          >
            Invalidate User
          </Button>
        </div>

        {userData && (
          <div
            style={{
              padding: "10px",
              backgroundColor: "#f5f5f5",
              borderRadius: "4px",
            }}
          >
            <strong>Selected User:</strong> {userData.name} ({userData.status})
            <br />
            <strong>Email:</strong> {userData.email}
            {userLoading && (
              <span style={{ color: "#666" }}> - Loading...</span>
            )}
            {userError && (
              <span style={{ color: "#dc3545" }}>
                {" "}
                - Error: {userError.message}
              </span>
            )}
          </div>
        )}
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>User Mutations</h4>
        <div style={{ display: "flex", gap: "10px", marginBottom: "10px" }}>
          <Button
            onClick={handleUpdateUser}
            disabled={!selectedUserId || updateUserMutation.isPending}
          >
            {updateUserMutation.isPending ? "Updating..." : "Update User"}
          </Button>
          <Button
            onClick={handleOptimisticUpdate}
            disabled={!selectedUserId || optimisticUpdateMutation.isPending}
          >
            {optimisticUpdateMutation.isPending
              ? "Updating..."
              : "Optimistic Update"}
          </Button>
        </div>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Search & Pagination</h4>
        <div style={{ marginBottom: "10px" }}>
          <input
            type="text"
            placeholder="Search users..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            style={{ padding: "5px", marginRight: "10px", minWidth: "200px" }}
          />
          <Button onClick={handleCancelRequests}>Cancel All Requests</Button>
        </div>

        <div style={{ display: "flex", gap: "10px", marginBottom: "10px" }}>
          <Button
            onClick={fetchPrevPage}
            disabled={!hasPrevPage || paginatedLoading}
          >
            Previous Page
          </Button>
          <Button
            onClick={fetchNextPage}
            disabled={!hasNextPage || paginatedLoading}
          >
            Next Page
          </Button>
          <select
            value={pagination?.limit || 5}
            onChange={(e) => setPageSize(Number(e.target.value))}
            style={{ padding: "5px" }}
          >
            <option value={5}>5 per page</option>
            <option value={10}>10 per page</option>
            <option value={20}>20 per page</option>
          </select>
        </div>

        {pagination && (
          <div
            style={{ fontSize: "12px", color: "#666", marginBottom: "10px" }}
          >
            Page {pagination.page} of {pagination.totalPages} (
            {pagination.total} total users)
          </div>
        )}

        {paginatedLoading ? (
          <div>Loading paginated users...</div>
        ) : (
          <div
            style={{
              maxHeight: "200px",
              overflowY: "auto",
              border: "1px solid #ddd",
              padding: "10px",
            }}
          >
            {paginatedUsers?.map((user) => (
              <div
                key={user.id}
                style={{
                  marginBottom: "5px",
                  padding: "5px",
                  backgroundColor: "#f9f9f9",
                }}
              >
                {user.name} ({user.status}) - {user.email}
              </div>
            ))}
          </div>
        )}
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Infinite Scroll</h4>
        <div style={{ display: "flex", gap: "10px", marginBottom: "10px" }}>
          <Button
            onClick={loadMoreInfinite}
            disabled={!hasMoreInfinite || infiniteLoading || isLoadingMore}
          >
            {isLoadingMore ? "Loading More..." : "Load More"}
          </Button>
          <Button onClick={resetInfinite}>Reset</Button>
        </div>

        {infiniteLoading ? (
          <div>Loading infinite users...</div>
        ) : (
          <div
            style={{
              maxHeight: "200px",
              overflowY: "auto",
              border: "1px solid #ddd",
              padding: "10px",
            }}
          >
            {infiniteUsers?.map((user: any) => (
              <div
                key={user.id}
                style={{
                  marginBottom: "5px",
                  padding: "5px",
                  backgroundColor: "#f9f9f9",
                }}
              >
                {user.name} ({user.status}) - {user.email}
              </div>
            ))}
          </div>
        )}
      </div>

      <div>
        <h4>Activity Logs</h4>
        <div style={{ display: "flex", gap: "10px", marginBottom: "10px" }}>
          <Button
            onClick={handleClearLogs}
            style={{ backgroundColor: "#6c757d", color: "white" }}
          >
            Clear Logs
          </Button>
        </div>

        <div
          style={{
            height: "200px",
            overflow: "auto",
            border: "1px solid #ddd",
            padding: "10px",
            backgroundColor: "#f9f9f9",
            fontFamily: "monospace",
            fontSize: "12px",
          }}
        >
          {logs.map((log, index) => (
            <div key={index} style={{ marginBottom: "2px" }}>
              {log}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
