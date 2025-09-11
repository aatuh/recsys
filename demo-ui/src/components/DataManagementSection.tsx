import React, { useState, useEffect, useCallback } from "react";
import { Section, Button, Code } from "./UIComponents";
import { DataTable, Column } from "./DataTable";
import {
  listUsers,
  listItems,
  listEvents,
  deleteUsers,
  deleteItems,
  deleteEvents,
  type ListParams,
  type DeleteParams,
  type ListResponse,
} from "../services/apiService";
import { updateItemEmbedding } from "../actions/updateItemEmbedding";

interface DataManagementSectionProps {
  namespace: string;
}

type DataType = "users" | "items" | "events";

interface FilterState {
  user_id: string;
  item_id: string;
  event_type: string;
  created_after: string;
  created_before: string;
}

const initialFilters: FilterState = {
  user_id: "",
  item_id: "",
  event_type: "",
  created_after: "",
  created_before: "",
};

export function DataManagementSection({
  namespace,
}: DataManagementSectionProps) {
  const [dataType, setDataType] = useState<DataType>("users");
  const [data, setData] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedRows, setSelectedRows] = useState<Set<string>>(new Set());
  const [filters, setFilters] = useState<FilterState>(initialFilters);
  const [sortBy, setSortBy] = useState<string>("");
  const [sortDirection, setSortDirection] = useState<"asc" | "desc">("desc");
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 25,
    total: 0,
  });
  const [embeddingsLoading, setEmbeddingsLoading] = useState(false);
  const [embeddingsProgress, setEmbeddingsProgress] = useState({
    current: 0,
    total: 0,
    message: "",
  });

  // Define columns for each data type
  const getColumns = (): Column<any>[] => {
    switch (dataType) {
      case "users":
        return [
          { key: "user_id", title: "User ID", width: "150px", sortable: true },
          {
            key: "traits",
            title: "Traits",
            width: "200px",
            render: (value) => (
              <code
                style={{
                  fontSize: "11px",
                  backgroundColor: "#f5f5f5",
                  padding: "2px 4px",
                  borderRadius: "3px",
                }}
              >
                {typeof value === "string"
                  ? value.substring(0, 50)
                  : (JSON.stringify(value) || "").substring(0, 50)}
                {(typeof value === "string"
                  ? value.length
                  : (JSON.stringify(value) || "").length) > 50
                  ? "..."
                  : ""}
              </code>
            ),
          },
          {
            key: "created_at",
            title: "Created",
            width: "150px",
            sortable: true,
          },
          {
            key: "updated_at",
            title: "Updated",
            width: "150px",
            sortable: true,
          },
        ];
      case "items":
        return [
          { key: "item_id", title: "Item ID", width: "150px", sortable: true },
          {
            key: "available",
            title: "Available",
            width: "80px",
            align: "center",
          },
          { key: "price", title: "Price", width: "80px", align: "right" },
          {
            key: "tags",
            title: "Tags",
            width: "150px",
            render: (value) => (
              <div style={{ display: "flex", flexWrap: "wrap", gap: "2px" }}>
                {Array.isArray(value)
                  ? value.slice(0, 3).map((tag, i) => (
                      <span
                        key={i}
                        style={{
                          backgroundColor: "#e3f2fd",
                          color: "#1565c0",
                          padding: "1px 4px",
                          borderRadius: "3px",
                          fontSize: "10px",
                        }}
                      >
                        {tag}
                      </span>
                    ))
                  : "-"}
                {Array.isArray(value) && value.length > 3 && (
                  <span style={{ fontSize: "10px", color: "#666" }}>
                    +{value.length - 3}
                  </span>
                )}
              </div>
            ),
          },
          {
            key: "props",
            title: "Properties",
            width: "150px",
            render: (value) => (
              <code
                style={{
                  fontSize: "11px",
                  backgroundColor: "#f5f5f5",
                  padding: "2px 4px",
                  borderRadius: "3px",
                }}
              >
                {typeof value === "string"
                  ? value.substring(0, 30)
                  : (JSON.stringify(value) || "").substring(0, 30)}
                {(typeof value === "string"
                  ? value.length
                  : (JSON.stringify(value) || "").length) > 30
                  ? "..."
                  : ""}
              </code>
            ),
          },
          {
            key: "created_at",
            title: "Created",
            width: "150px",
            sortable: true,
          },
        ];
      case "events":
        return [
          { key: "user_id", title: "User ID", width: "120px", sortable: true },
          { key: "item_id", title: "Item ID", width: "120px", sortable: true },
          {
            key: "type",
            title: "Type",
            width: "80px",
            align: "center",
            render: (value) => {
              const typeNames = {
                0: "View",
                1: "Click",
                2: "Add",
                3: "Purchase",
              };
              return (
                <span
                  style={{
                    backgroundColor:
                      value === 3
                        ? "#4caf50"
                        : value === 2
                        ? "#ff9800"
                        : value === 1
                        ? "#2196f3"
                        : "#9e9e9e",
                    color: "white",
                    padding: "2px 6px",
                    borderRadius: "3px",
                    fontSize: "10px",
                    fontWeight: "bold",
                  }}
                >
                  {typeNames[value as keyof typeof typeNames] || value}
                </span>
              );
            },
          },
          { key: "value", title: "Value", width: "60px", align: "right" },
          { key: "ts", title: "Timestamp", width: "150px", sortable: true },
          {
            key: "meta",
            title: "Metadata",
            width: "100px",
            render: (value) => (
              <code
                style={{
                  fontSize: "11px",
                  backgroundColor: "#f5f5f5",
                  padding: "2px 4px",
                  borderRadius: "3px",
                }}
              >
                {typeof value === "string"
                  ? value.substring(0, 20)
                  : (JSON.stringify(value) || "").substring(0, 20)}
                {(typeof value === "string"
                  ? value.length
                  : (JSON.stringify(value) || "").length) > 20
                  ? "..."
                  : ""}
              </code>
            ),
          },
        ];
      default:
        return [];
    }
  };

  const loadData = useCallback(async () => {
    if (!namespace) return;

    setLoading(true);
    setError(null);

    try {
      const params: ListParams = {
        namespace,
        limit: pagination.pageSize,
        offset: (pagination.page - 1) * pagination.pageSize,
      };

      // Add filters
      if (filters.user_id) params.user_id = filters.user_id;
      if (filters.item_id) params.item_id = filters.item_id;
      if (filters.event_type) params.event_type = parseInt(filters.event_type);
      if (filters.created_after) params.created_after = filters.created_after;
      if (filters.created_before)
        params.created_before = filters.created_before;

      let response: ListResponse;
      switch (dataType) {
        case "users":
          response = await listUsers(params);
          break;
        case "items":
          response = await listItems(params);
          break;
        case "events":
          response = await listEvents(params);
          break;
        default:
          throw new Error("Invalid data type");
      }

      setData(response.items);
      setPagination((prev) => ({ ...prev, total: response.total }));
    } catch (err: any) {
      setError(err.message || "Failed to load data");
    } finally {
      setLoading(false);
    }
  }, [namespace, dataType, pagination.page, pagination.pageSize, filters]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleDelete = async (selectedOnly = false) => {
    if (!namespace) return;

    const confirmMessage = selectedOnly
      ? `Are you sure you want to delete ${selectedRows.size} selected ${dataType}?`
      : `Are you sure you want to delete ALL ${dataType} in this namespace? This action cannot be undone.`;

    if (!window.confirm(confirmMessage)) return;

    setLoading(true);
    setError(null);

    try {
      const params: DeleteParams = { namespace };

      // If deleting selected only, we need to get the IDs and delete them individually
      // For now, we'll delete all and let the user filter first
      if (selectedOnly && selectedRows.size > 0) {
        // TODO: Implement individual deletion by ID
        alert(
          "Individual deletion not yet implemented. Please use filters to narrow down your selection."
        );
        setLoading(false);
        return;
      }

      // Add filters for targeted deletion
      if (filters.user_id) params.user_id = filters.user_id;
      if (filters.item_id) params.item_id = filters.item_id;
      if (filters.event_type) params.event_type = parseInt(filters.event_type);
      if (filters.created_after) params.created_after = filters.created_after;
      if (filters.created_before)
        params.created_before = filters.created_before;

      let response;
      switch (dataType) {
        case "users":
          response = await deleteUsers(params);
          break;
        case "items":
          response = await deleteItems(params);
          break;
        case "events":
          response = await deleteEvents(params);
          break;
        default:
          throw new Error("Invalid data type");
      }

      alert(`Successfully deleted ${response.deleted_count} ${dataType}`);
      setSelectedRows(new Set());
      loadData();
    } catch (err: any) {
      setError(err.message || "Failed to delete data");
    } finally {
      setLoading(false);
    }
  };

  const handleSort = (column: string, direction: "asc" | "desc") => {
    setSortBy(column);
    setSortDirection(direction);
    // Note: In a real implementation, you'd pass sort parameters to the API
  };

  const handleFilterChange = (newFilters: Record<string, any>) => {
    setFilters((prev) => ({ ...prev, ...newFilters }));
    setPagination((prev) => ({ ...prev, page: 1 })); // Reset to first page
  };

  const handlePageChange = (page: number) => {
    setPagination((prev) => ({ ...prev, page }));
  };

  const handlePageSizeChange = (pageSize: number) => {
    setPagination((prev) => ({ ...prev, pageSize, page: 1 }));
  };

  const clearFilters = () => {
    setFilters(initialFilters);
    setPagination((prev) => ({ ...prev, page: 1 }));
  };

  const handleUpdateEmbeddings = async () => {
    if (dataType !== "items" || !namespace) return;

    setEmbeddingsLoading(true);
    setEmbeddingsProgress({ current: 0, total: 0, message: "Starting..." });
    setError(null);

    try {
      // Get all items to process (not just current page)
      const allItemsParams: ListParams = {
        namespace,
        limit: 1000, // Get a large batch
        offset: 0,
      };

      // Add current filters
      if (filters.item_id) allItemsParams.item_id = filters.item_id;
      if (filters.created_after)
        allItemsParams.created_after = filters.created_after;
      if (filters.created_before)
        allItemsParams.created_before = filters.created_before;

      const response = await listItems(allItemsParams);
      const itemsToProcess = response.items;

      if (itemsToProcess.length === 0) {
        setEmbeddingsProgress({
          current: 0,
          total: 0,
          message: "No items to process",
        });
        return;
      }

      setEmbeddingsProgress({
        current: 0,
        total: itemsToProcess.length,
        message: `Processing ${itemsToProcess.length} items...`,
      });

      // Process items in small batches to avoid overwhelming the browser
      const batchSize = 5;
      let processed = 0;

      for (let i = 0; i < itemsToProcess.length; i += batchSize) {
        const batch = itemsToProcess.slice(i, i + batchSize);

        // Process batch in parallel
        await Promise.all(
          batch.map(async (item) => {
            try {
              await updateItemEmbedding(
                namespace,
                {
                  item_id: item.item_id,
                  tags: item.tags,
                  price: item.price,
                  props: item.props,
                },
                (message) => {
                  setEmbeddingsProgress((prev) => ({
                    ...prev,
                    message: `${item.item_id}: ${message}`,
                  }));
                }
              );
              processed++;
              setEmbeddingsProgress((prev) => ({
                ...prev,
                current: processed,
                message: `Processed ${processed}/${itemsToProcess.length} items`,
              }));
            } catch (err: any) {
              console.error(`Failed to process item ${item.item_id}:`, err);
              setEmbeddingsProgress((prev) => ({
                ...prev,
                message: `Error processing ${item.item_id}: ${err.message}`,
              }));
            }
          })
        );

        // Small delay between batches to keep UI responsive
        await new Promise((resolve) => setTimeout(resolve, 100));
      }

      setEmbeddingsProgress({
        current: processed,
        total: itemsToProcess.length,
        message: `Completed! Processed ${processed} items`,
      });

      // Refresh the data to show updated items
      loadData();
    } catch (err: any) {
      setError(err.message || "Failed to update embeddings");
      setEmbeddingsProgress({ current: 0, total: 0, message: "Failed" });
    } finally {
      setEmbeddingsLoading(false);
    }
  };

  return (
    <Section title="Data Management">
      <div style={{ marginBottom: "16px" }}>
        <p style={{ color: "#666", fontSize: "14px", marginBottom: "16px" }}>
          Manage your data with advanced filtering, sorting, and bulk
          operations. Use filters to narrow down your selection before deleting.
        </p>

        {/* Data Type Selector */}
        <div style={{ marginBottom: "16px" }}>
          <label
            style={{ display: "block", marginBottom: "8px", fontWeight: "600" }}
          >
            Data Type:
          </label>
          <div style={{ display: "flex", gap: "8px" }}>
            {(["users", "items", "events"] as DataType[]).map((type) => (
              <Button
                key={type}
                onClick={() => {
                  setDataType(type);
                  setSelectedRows(new Set());
                  setPagination((prev) => ({ ...prev, page: 1 }));
                }}
                style={{
                  backgroundColor: dataType === type ? "#1976d2" : "#f5f5f5",
                  color: dataType === type ? "white" : "#333",
                  textTransform: "capitalize",
                }}
              >
                {type}
              </Button>
            ))}
          </div>
        </div>

        {/* Filters */}
        <div
          style={{
            border: "1px solid #e0e0e0",
            borderRadius: "6px",
            padding: "16px",
            marginBottom: "16px",
            backgroundColor: "#fafafa",
          }}
        >
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              marginBottom: "12px",
            }}
          >
            <h4 style={{ margin: 0, fontSize: "14px", color: "#333" }}>
              Filters
            </h4>
            <Button
              onClick={clearFilters}
              style={{
                padding: "4px 8px",
                fontSize: "12px",
                backgroundColor: "#f5f5f5",
              }}
            >
              Clear All
            </Button>
          </div>

          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
              gap: "12px",
            }}
          >
            {dataType === "users" && (
              <div>
                <label
                  style={{
                    display: "block",
                    fontSize: "12px",
                    marginBottom: "4px",
                  }}
                >
                  User ID:
                </label>
                <input
                  type="text"
                  value={filters.user_id}
                  onChange={(e) =>
                    handleFilterChange({ user_id: e.target.value })
                  }
                  placeholder="Filter by user ID"
                  style={{
                    width: "100%",
                    padding: "6px 8px",
                    border: "1px solid #ccc",
                    borderRadius: "4px",
                    fontSize: "12px",
                  }}
                />
              </div>
            )}

            {dataType === "items" && (
              <div>
                <label
                  style={{
                    display: "block",
                    fontSize: "12px",
                    marginBottom: "4px",
                  }}
                >
                  Item ID:
                </label>
                <input
                  type="text"
                  value={filters.item_id}
                  onChange={(e) =>
                    handleFilterChange({ item_id: e.target.value })
                  }
                  placeholder="Filter by item ID"
                  style={{
                    width: "100%",
                    padding: "6px 8px",
                    border: "1px solid #ccc",
                    borderRadius: "4px",
                    fontSize: "12px",
                  }}
                />
              </div>
            )}

            {dataType === "events" && (
              <>
                <div>
                  <label
                    style={{
                      display: "block",
                      fontSize: "12px",
                      marginBottom: "4px",
                    }}
                  >
                    User ID:
                  </label>
                  <input
                    type="text"
                    value={filters.user_id}
                    onChange={(e) =>
                      handleFilterChange({ user_id: e.target.value })
                    }
                    placeholder="Filter by user ID"
                    style={{
                      width: "100%",
                      padding: "6px 8px",
                      border: "1px solid #ccc",
                      borderRadius: "4px",
                      fontSize: "12px",
                    }}
                  />
                </div>
                <div>
                  <label
                    style={{
                      display: "block",
                      fontSize: "12px",
                      marginBottom: "4px",
                    }}
                  >
                    Item ID:
                  </label>
                  <input
                    type="text"
                    value={filters.item_id}
                    onChange={(e) =>
                      handleFilterChange({ item_id: e.target.value })
                    }
                    placeholder="Filter by item ID"
                    style={{
                      width: "100%",
                      padding: "6px 8px",
                      border: "1px solid #ccc",
                      borderRadius: "4px",
                      fontSize: "12px",
                    }}
                  />
                </div>
                <div>
                  <label
                    style={{
                      display: "block",
                      fontSize: "12px",
                      marginBottom: "4px",
                    }}
                  >
                    Event Type:
                  </label>
                  <select
                    value={filters.event_type}
                    onChange={(e) =>
                      handleFilterChange({ event_type: e.target.value })
                    }
                    style={{
                      width: "100%",
                      padding: "6px 8px",
                      border: "1px solid #ccc",
                      borderRadius: "4px",
                      fontSize: "12px",
                    }}
                  >
                    <option value="">All Types</option>
                    <option value="0">View</option>
                    <option value="1">Click</option>
                    <option value="2">Add to Cart</option>
                    <option value="3">Purchase</option>
                  </select>
                </div>
              </>
            )}

            <div>
              <label
                style={{
                  display: "block",
                  fontSize: "12px",
                  marginBottom: "4px",
                }}
              >
                Created After:
              </label>
              <input
                type="datetime-local"
                value={filters.created_after}
                onChange={(e) =>
                  handleFilterChange({ created_after: e.target.value })
                }
                style={{
                  width: "100%",
                  padding: "6px 8px",
                  border: "1px solid #ccc",
                  borderRadius: "4px",
                  fontSize: "12px",
                }}
              />
            </div>

            <div>
              <label
                style={{
                  display: "block",
                  fontSize: "12px",
                  marginBottom: "4px",
                }}
              >
                Created Before:
              </label>
              <input
                type="datetime-local"
                value={filters.created_before}
                onChange={(e) =>
                  handleFilterChange({ created_before: e.target.value })
                }
                style={{
                  width: "100%",
                  padding: "6px 8px",
                  border: "1px solid #ccc",
                  borderRadius: "4px",
                  fontSize: "12px",
                }}
              />
            </div>
          </div>
        </div>

        {/* Actions */}
        <div
          style={{
            display: "flex",
            gap: "8px",
            marginBottom: "16px",
            alignItems: "center",
            flexWrap: "wrap",
          }}
        >
          <Button
            onClick={() => loadData()}
            disabled={loading}
            style={{ backgroundColor: "#4caf50" }}
          >
            {loading ? "Loading..." : "Refresh"}
          </Button>

          {dataType === "items" && (
            <Button
              onClick={handleUpdateEmbeddings}
              disabled={loading || embeddingsLoading}
              style={{ backgroundColor: "#9c27b0" }}
            >
              {embeddingsLoading ? "Updating..." : "Update Embeddings"}
            </Button>
          )}

          {selectedRows.size > 0 && (
            <Button
              onClick={() => handleDelete(true)}
              disabled={loading}
              style={{ backgroundColor: "#f44336" }}
            >
              Delete Selected ({selectedRows.size})
            </Button>
          )}

          <Button
            onClick={() => handleDelete(false)}
            disabled={loading}
            style={{ backgroundColor: "#ff9800" }}
          >
            Delete All (Filtered)
          </Button>

          <span style={{ fontSize: "12px", color: "#666", marginLeft: "16px" }}>
            {pagination.total} total {dataType}
          </span>
        </div>

        {/* Embeddings Progress */}
        {embeddingsLoading && (
          <div
            style={{
              backgroundColor: "#e8f5e8",
              border: "1px solid #4caf50",
              borderRadius: "4px",
              padding: "12px",
              marginBottom: "16px",
            }}
          >
            <div
              style={{
                display: "flex",
                alignItems: "center",
                marginBottom: "8px",
              }}
            >
              <div
                style={{
                  fontSize: "14px",
                  fontWeight: "600",
                  color: "#2e7d32",
                }}
              >
                Updating Embeddings
              </div>
              <div
                style={{ marginLeft: "auto", fontSize: "12px", color: "#666" }}
              >
                {embeddingsProgress.current}/{embeddingsProgress.total}
              </div>
            </div>

            {embeddingsProgress.total > 0 && (
              <div
                style={{
                  width: "100%",
                  backgroundColor: "#e0e0e0",
                  borderRadius: "4px",
                  height: "8px",
                  marginBottom: "8px",
                }}
              >
                <div
                  style={{
                    width: `${
                      (embeddingsProgress.current / embeddingsProgress.total) *
                      100
                    }%`,
                    backgroundColor: "#4caf50",
                    height: "100%",
                    borderRadius: "4px",
                    transition: "width 0.3s ease",
                  }}
                />
              </div>
            )}

            <div style={{ fontSize: "12px", color: "#666" }}>
              {embeddingsProgress.message}
            </div>
          </div>
        )}

        {/* Error Display */}
        {error && (
          <div
            style={{
              backgroundColor: "#ffebee",
              color: "#c62828",
              padding: "12px",
              borderRadius: "4px",
              marginBottom: "16px",
              fontSize: "14px",
            }}
          >
            Error: {error}
          </div>
        )}

        {/* Data Table */}
        <DataTable
          data={data}
          columns={getColumns()}
          loading={loading}
          selectable={true}
          selectedRows={selectedRows}
          onSelectionChange={setSelectedRows}
          sortBy={sortBy}
          sortDirection={sortDirection}
          onSort={handleSort}
          pagination={{
            page: pagination.page,
            pageSize: pagination.pageSize,
            total: pagination.total,
            onPageChange: handlePageChange,
            onPageSizeChange: handlePageSizeChange,
          }}
          emptyMessage={`No ${dataType} found. Try adjusting your filters.`}
        />
      </div>
    </Section>
  );
}
