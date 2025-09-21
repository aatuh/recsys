import React, { useState, useEffect, useCallback, useRef } from "react";
import { Section, Button } from "../primitives/UIComponents";
import { DataTable, Column } from "../primitives/DataTable";
import { useViewState } from "../../contexts/ViewStateContext";
import {
  listUsers,
  listItems,
  listEvents,
  listSegments,
  deleteUsers,
  deleteItems,
  deleteEvents,
  deleteSegments,
  fetchAllDataForTables,
  type ListParams,
  type DeleteParams,
  type ListResponse,
} from "../../services/apiService";
import { ConfigService } from "../../lib/api-client";
import { updateItemEmbedding } from "../../actions/updateItemEmbedding";
import {
  downloadJsonFile,
  generateExportFilename,
  formatExportData,
} from "../../utils/exportUtils";
import { useToast } from "../../ui/Toast";
import { logger } from "../../utils/logger";

interface DataManagementSectionProps {
  namespace: string;
}

type DataType = "users" | "items" | "events" | "segments";

interface FilterState {
  user_id: string;
  item_id: string;
  event_type: string;
  created_after: string;
  created_before: string;
  segment_id: string;
  profile_id: string;
}

const initialFilters: FilterState = {
  user_id: "",
  item_id: "",
  event_type: "",
  created_after: "",
  created_before: "",
  segment_id: "",
  profile_id: "",
};

export function DataManagementSection({
  namespace,
}: DataManagementSectionProps) {
  const { dataManagement, setDataManagement } = useViewState();
  const toast = useToast();

  // Local state that doesn't need to be preserved across navigation
  const [data, setData] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Helper function to create clickable JSON renderer
  const createJsonRenderer = (fieldName: string, maxLength: number = 50) => {
    return (value: any, row: any, toggleExpansion?: () => void) => {
      const jsonStr =
        typeof value === "string" ? value : JSON.stringify(value) || "";
      const isTruncated = jsonStr.length > maxLength;

      return (
        <code
          style={{
            fontSize: "11px",
            backgroundColor: "#f5f5f5",
            padding: "2px 4px",
            borderRadius: "3px",
            cursor: isTruncated ? "pointer" : "default",
            border: isTruncated ? "1px solid #e0e0e0" : "none",
            display: "inline-block",
          }}
          onClick={(e) => {
            if (isTruncated && toggleExpansion) {
              e.stopPropagation();
              toggleExpansion();
            }
          }}
          title={isTruncated ? "Click to expand row" : ""}
        >
          {jsonStr.substring(0, maxLength)}
          {isTruncated ? "..." : ""}
        </code>
      );
    };
  };

  // Define columns for each data type
  const getColumns = (): Column<any>[] => {
    switch (dataManagement.dataType) {
      case "users":
        return [
          { key: "user_id", title: "User ID", width: "150px", sortable: true },
          {
            key: "traits",
            title: "Traits",
            width: "200px",
            render: createJsonRenderer("User Traits"),
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
            render: createJsonRenderer("Item Properties", 30),
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
            render: createJsonRenderer("Event Metadata", 20),
          },
        ];
      case "segments":
        return [
          {
            key: "segment_id",
            title: "Segment ID",
            width: "150px",
            sortable: true,
          },
          { key: "name", title: "Name", width: "200px", sortable: true },
          { key: "description", title: "Description", width: "250px" },
          {
            key: "priority",
            title: "Priority",
            width: "80px",
            align: "center",
          },
          { key: "active", title: "Active", width: "80px", align: "center" },
          { key: "profile_id", title: "Profile ID", width: "150px" },
          {
            key: "rules",
            title: "Rules",
            width: "200px",
            render: createJsonRenderer("Rules"),
          },
          {
            key: "created_at",
            title: "Created",
            width: "150px",
            sortable: true,
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
        limit: dataManagement.pagination.pageSize,
        offset:
          (dataManagement.pagination.page - 1) *
          dataManagement.pagination.pageSize,
      };

      // Add filters
      if (dataManagement.filters.user_id)
        params.user_id = dataManagement.filters.user_id;
      if (dataManagement.filters.item_id)
        params.item_id = dataManagement.filters.item_id;
      if (dataManagement.filters.event_type)
        params.event_type = parseInt(dataManagement.filters.event_type);
      if (dataManagement.filters.created_after)
        params.created_after = dataManagement.filters.created_after;
      if (dataManagement.filters.created_before)
        params.created_before = dataManagement.filters.created_before;

      let response: ListResponse;
      switch (dataManagement.dataType) {
        case "users":
          response = await listUsers(params);
          break;
        case "items":
          response = await listItems(params);
          break;
        case "events":
          response = await listEvents(params);
          break;
        case "segments":
          response = await listSegments(params);
          break;
        default:
          throw new Error("Invalid data type");
      }

      setData(response.items);
      setDataManagement((prev) => ({
        ...prev,
        pagination: { ...prev.pagination, total: response.total },
      }));
    } catch (err: any) {
      const msg = err?.message || "Failed to load data";
      setError(msg);
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  }, [
    namespace,
    dataManagement.dataType,
    dataManagement.pagination.page,
    dataManagement.pagination.pageSize,
    dataManagement.filters,
    setDataManagement,
    toast,
  ]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleDelete = async (selectedOnly = false) => {
    if (!namespace) return;

    const confirmMessage = selectedOnly
      ? `Are you sure you want to delete ${dataManagement.selectedRows.size} selected ${dataManagement.dataType}?`
      : `Are you sure you want to delete ALL ${dataManagement.dataType} in this namespace? This action cannot be undone.`;

    if (!window.confirm(confirmMessage)) return;

    setLoading(true);
    setError(null);

    try {
      logger.info("data.delete.start", {
        namespace,
        dataType: dataManagement.dataType,
        selectedOnly,
        selectedCount: dataManagement.selectedRows.size,
      });
      const params: DeleteParams = { namespace };

      // If deleting selected only, we need to get the IDs and delete them individually
      if (selectedOnly && dataManagement.selectedRows.size > 0) {
        if (dataManagement.dataType === "segments") {
          // For segments, we need to delete by specific IDs
          const selectedSegmentIds = Array.from(dataManagement.selectedRows);
          await ConfigService.segmentsDelete({
            namespace: params.namespace,
            ids: selectedSegmentIds,
          });

          toast.success(
            `Deleted ${selectedSegmentIds.length} selected segments`
          );
          setDataManagement((prev) => ({ ...prev, selectedRows: new Set() }));
          loadData();
          logger.info("data.delete.success", {
            namespace,
            dataType: "segments",
            deleted: selectedSegmentIds.length,
          });
          setLoading(false);
          return;
        } else {
          // TODO: Implement individual deletion by ID for other data types
          toast.info(
            "Individual deletion not yet implemented for this data type. Please use filters to narrow down your selection."
          );
          setLoading(false);
          return;
        }
      }

      // Add filters for targeted deletion
      if (dataManagement.filters.user_id)
        params.user_id = dataManagement.filters.user_id;
      if (dataManagement.filters.item_id)
        params.item_id = dataManagement.filters.item_id;
      if (dataManagement.filters.event_type)
        params.event_type = parseInt(dataManagement.filters.event_type);
      if (dataManagement.filters.created_after)
        params.created_after = dataManagement.filters.created_after;
      if (dataManagement.filters.created_before)
        params.created_before = dataManagement.filters.created_before;

      let response;
      switch (dataManagement.dataType) {
        case "users":
          response = await deleteUsers(params);
          break;
        case "items":
          response = await deleteItems(params);
          break;
        case "events":
          response = await deleteEvents(params);
          break;
        case "segments":
          response = await deleteSegments(params);
          break;
        default:
          throw new Error("Invalid data type");
      }

      toast.success(
        `Deleted ${response.deleted_count} ${dataManagement.dataType}`
      );
      setDataManagement((prev) => ({ ...prev, selectedRows: new Set() }));
      loadData();
      logger.info("data.delete.success", {
        namespace,
        dataType: dataManagement.dataType,
        deleted: response.deleted_count,
        filtered: !selectedOnly,
      });
    } catch (err: any) {
      const msg = err?.message || "Failed to delete data";
      setError(msg);
      toast.error(msg);
      logger.error("data.delete.error", {
        namespace,
        dataType: dataManagement.dataType,
        error: msg,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleSort = (column: string, direction: "asc" | "desc") => {
    setDataManagement((prev) => ({
      ...prev,
      sortBy: column,
      sortDirection: direction,
    }));
    // Note: In a real implementation, you'd pass sort parameters to the API
  };

  // Debounce filter changes to reduce API calls
  const debounceTimerRef = useRef<number | null>(null);
  const handleFilterChange = (newFilters: Record<string, any>) => {
    if (debounceTimerRef.current) {
      window.clearTimeout(debounceTimerRef.current);
    }
    debounceTimerRef.current = window.setTimeout(() => {
      setDataManagement((prev) => ({
        ...prev,
        filters: { ...prev.filters, ...newFilters },
        pagination: { ...prev.pagination, page: 1 },
      }));
    }, 300);
  };

  const handlePageChange = (page: number) => {
    setDataManagement((prev) => ({
      ...prev,
      pagination: { ...prev.pagination, page },
    }));
  };

  const handlePageSizeChange = (pageSize: number) => {
    setDataManagement((prev) => ({
      ...prev,
      pagination: { ...prev.pagination, pageSize, page: 1 },
    }));
  };

  const clearFilters = () => {
    setDataManagement((prev) => ({
      ...prev,
      filters: initialFilters,
      pagination: { ...prev.pagination, page: 1 },
    }));
  };

  const handleUpdateEmbeddings = async () => {
    if (dataManagement.dataType !== "items" || !namespace) return;

    setDataManagement((prev) => ({
      ...prev,
      embeddingsLoading: true,
      embeddingsProgress: { current: 0, total: 0, message: "Starting..." },
    }));
    setError(null);

    try {
      // Get all items to process (not just current page)
      const allItemsParams: ListParams = {
        namespace,
        limit: 1000, // Get a large batch
        offset: 0,
      };

      // Add current filters
      if (dataManagement.filters.item_id)
        allItemsParams.item_id = dataManagement.filters.item_id;
      if (dataManagement.filters.created_after)
        allItemsParams.created_after = dataManagement.filters.created_after;
      if (dataManagement.filters.created_before)
        allItemsParams.created_before = dataManagement.filters.created_before;

      const response = await listItems(allItemsParams);
      const itemsToProcess = response.items;

      if (itemsToProcess.length === 0) {
        setDataManagement((prev) => ({
          ...prev,
          embeddingsProgress: {
            current: 0,
            total: 0,
            message: "No items to process",
          },
        }));
        return;
      }

      setDataManagement((prev) => ({
        ...prev,
        embeddingsProgress: {
          current: 0,
          total: itemsToProcess.length,
          message: `Processing ${itemsToProcess.length} items...`,
        },
      }));

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
                  setDataManagement((prev) => ({
                    ...prev,
                    embeddingsProgress: {
                      ...prev.embeddingsProgress,
                      message: `${item.item_id}: ${message}`,
                    },
                  }));
                }
              );
              processed++;
              setDataManagement((prev) => ({
                ...prev,
                embeddingsProgress: {
                  ...prev.embeddingsProgress,
                  current: processed,
                  message: `Processed ${processed}/${itemsToProcess.length} items`,
                },
              }));
            } catch (err: any) {
              console.error(`Failed to process item ${item.item_id}:`, err);
              setDataManagement((prev) => ({
                ...prev,
                embeddingsProgress: {
                  ...prev.embeddingsProgress,
                  message: `Error processing ${item.item_id}: ${err.message}`,
                },
              }));
            }
          })
        );

        // Small delay between batches to keep UI responsive
        await new Promise((resolve) => setTimeout(resolve, 100));
      }

      setDataManagement((prev) => ({
        ...prev,
        embeddingsProgress: {
          current: processed,
          total: itemsToProcess.length,
          message: `Completed! Processed ${processed} items`,
        },
      }));

      // Refresh the data to show updated items
      loadData();
    } catch (err: any) {
      setError(err.message || "Failed to update embeddings");
      toast.error("Failed to update embeddings");
    } finally {
      setDataManagement((prev) => ({ ...prev, embeddingsLoading: false }));
    }
  };

  const handleDestroyAllData = async () => {
    if (!namespace) return;

    // Show a very clear warning dialog
    const confirmMessage = `⚠️ DANGER: This will PERMANENTLY DELETE ALL DATA in namespace "${namespace}"!

This includes:
• ALL users
• ALL items  
• ALL events

This action CANNOT be undone!

Are you sure you want to continue?`;

    const confirmed = window.confirm(confirmMessage);
    if (!confirmed) {
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const params: DeleteParams = { namespace };

      // Delete all data types in sequence
      const results = {
        users: 0,
        items: 0,
        events: 0,
        segments: 0,
      };

      // Delete users
      try {
        const userResult = await deleteUsers(params);
        results.users = userResult.deleted_count;
      } catch (err: any) {
        console.warn("Failed to delete users:", err);
      }

      // Delete items
      try {
        const itemResult = await deleteItems(params);
        results.items = itemResult.deleted_count;
      } catch (err: any) {
        console.warn("Failed to delete items:", err);
      }

      // Delete events
      try {
        const eventResult = await deleteEvents(params);
        results.events = eventResult.deleted_count;
      } catch (err: any) {
        console.warn("Failed to delete events:", err);
      }

      // Delete segments
      try {
        const segmentResult = await deleteSegments(params);
        results.segments = segmentResult.deleted_count;
      } catch (err: any) {
        console.warn("Failed to delete segments:", err);
      }

      const totalDeleted =
        results.users +
        results.items +
        results.events +
        (results.segments || 0);

      toast.success(
        `Destroyed data in "${namespace}". Deleted ${totalDeleted} records (users: ${
          results.users
        }, items: ${results.items}, events: ${results.events}, segments: ${
          results.segments || 0
        }).`
      );
      logger.info("data.destroy_all.success", {
        namespace,
        totals: results,
        totalDeleted,
      });

      // Reset state and refresh
      setDataManagement((prev) => ({
        ...prev,
        selectedRows: new Set(),
        filters: initialFilters,
        pagination: { ...prev.pagination, page: 1, total: 0 },
      }));
      setData([]);
    } catch (err: any) {
      const msg = err?.message || "Failed to destroy all data";
      setError(msg);
      toast.error(msg);
      logger.error("data.destroy_all.error", { namespace, error: msg });
    } finally {
      setLoading(false);
    }
  };

  const handleExportData = async () => {
    if (!namespace || dataManagement.selectedExportTables.length === 0) return;

    setDataManagement((prev) => ({
      ...prev,
      exportLoading: true,
      exportProgress: { current: 0, total: 0, message: "Starting export..." },
    }));
    setError(null);

    try {
      // Fetch all data for selected tables
      const data = await fetchAllDataForTables(
        namespace,
        dataManagement.selectedExportTables,
        dataManagement.filters
      );

      // Format the export data
      const exportData = formatExportData(
        namespace,
        dataManagement.selectedExportTables,
        data
      );

      // Generate filename and download
      const filename = generateExportFilename(
        namespace,
        dataManagement.selectedExportTables
      );
      downloadJsonFile(exportData, filename);

      setDataManagement((prev) => ({
        ...prev,
        exportProgress: {
          current: 1,
          total: 1,
          message: `Export completed! Downloaded ${filename}`,
        },
      }));
      toast.success(
        `Exported ${dataManagement.selectedExportTables.join(
          ", "
        )} to ${filename}`
      );
      logger.info("data.export.success", {
        namespace,
        tables: dataManagement.selectedExportTables,
        filename,
      });

      // Clear progress after a delay
      setTimeout(() => {
        setDataManagement((prev) => ({
          ...prev,
          exportLoading: false,
          exportProgress: { current: 0, total: 0, message: "" },
        }));
      }, 3000);
    } catch (err: any) {
      setError(err.message || "Failed to export data");
      toast.error("Failed to export data");
      logger.error("data.export.error", {
        namespace,
        tables: dataManagement.selectedExportTables,
        error: err?.message,
      });
    } finally {
      setDataManagement((prev) => ({
        ...prev,
        exportLoading: false,
        exportProgress: { current: 0, total: 0, message: "Export failed" },
      }));
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
            {(["users", "items", "events", "segments"] as DataType[]).map(
              (type) => (
                <Button
                  key={type}
                  onClick={() => {
                    setDataManagement((prev) => ({
                      ...prev,
                      dataType: type,
                      selectedRows: new Set(),
                      pagination: { ...prev.pagination, page: 1 },
                    }));
                  }}
                  style={{
                    backgroundColor:
                      dataManagement.dataType === type ? "#e3f2fd" : "#f5f5f5",
                    color: dataManagement.dataType === type ? "1565c0" : "#666",
                    textTransform: "capitalize",
                  }}
                >
                  {type}
                </Button>
              )
            )}
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
            {dataManagement.dataType === "users" && (
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
                  value={dataManagement.filters.user_id}
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

            {dataManagement.dataType === "items" && (
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
                  value={dataManagement.filters.item_id}
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

            {dataManagement.dataType === "events" && (
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
                    value={dataManagement.filters.user_id}
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
                    value={dataManagement.filters.item_id}
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
                    value={dataManagement.filters.event_type}
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
                value={dataManagement.filters.created_after}
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
                value={dataManagement.filters.created_before}
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

        {/* Export Section */}
        <div
          style={{
            border: "1px solid #e0e0e0",
            borderRadius: "6px",
            padding: "16px",
            marginBottom: "16px",
            backgroundColor: "#f8f9fa",
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
              Export Data
            </h4>
          </div>

          <div style={{ marginBottom: "12px" }}>
            <label
              style={{
                display: "block",
                fontSize: "12px",
                marginBottom: "8px",
                fontWeight: "600",
              }}
            >
              Select tables to export:
            </label>
            <div style={{ display: "flex", gap: "12px", flexWrap: "wrap" }}>
              {(["users", "items", "events", "segments"] as const).map(
                (table) => (
                  <label
                    key={table}
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "6px",
                      cursor: "pointer",
                      fontSize: "12px",
                    }}
                  >
                    <input
                      type="checkbox"
                      checked={dataManagement.selectedExportTables.includes(
                        table
                      )}
                      onChange={(e) => {
                        const newTables = e.target.checked
                          ? [...dataManagement.selectedExportTables, table]
                          : dataManagement.selectedExportTables.filter(
                              (t) => t !== table
                            );
                        setDataManagement((prev) => ({
                          ...prev,
                          selectedExportTables: newTables,
                        }));
                      }}
                      style={{ cursor: "pointer" }}
                    />
                    <span style={{ textTransform: "capitalize" }}>{table}</span>
                  </label>
                )
              )}
            </div>
          </div>

          <div style={{ display: "flex", gap: "8px", alignItems: "center" }}>
            <Button
              onClick={handleExportData}
              disabled={
                loading ||
                dataManagement.exportLoading ||
                dataManagement.selectedExportTables.length === 0
              }
              style={{
                backgroundColor: "#4caf50",
                color: "white",
                border: "none",
              }}
            >
              {dataManagement.exportLoading ? "Exporting..." : "Export JSON"}
            </Button>

            <span style={{ fontSize: "11px", color: "#666" }}>
              {dataManagement.selectedExportTables.length === 0
                ? "Select at least one table to export"
                : `Will export: ${dataManagement.selectedExportTables.join(
                    ", "
                  )}`}
            </span>
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
            aria-label="Refresh data"
          >
            {loading ? "Loading..." : "Refresh"}
          </Button>

          {dataManagement.dataType === "items" && (
            <Button
              onClick={handleUpdateEmbeddings}
              disabled={loading || dataManagement.embeddingsLoading}
              aria-label="Update item embeddings"
            >
              {dataManagement.embeddingsLoading
                ? "Updating..."
                : "Update Embeddings"}
            </Button>
          )}

          {dataManagement.selectedRows.size > 0 && (
            <Button
              onClick={() => handleDelete(true)}
              disabled={loading}
              aria-label="Delete selected rows"
            >
              Delete Selected ({dataManagement.selectedRows.size})
            </Button>
          )}

          <Button
            onClick={() => handleDelete(false)}
            disabled={loading}
            aria-label="Delete all filtered rows"
          >
            Delete All (Filtered)
          </Button>

          <Button
            onClick={handleDestroyAllData}
            disabled={loading}
            style={{
              backgroundColor: "#f5f5f5",
              color: "#d32f2f",
              border: "1px solid #e0e0e0",
              fontSize: "12px",
            }}
            title="Permanently delete ALL users, items, events, and segments in this namespace"
            aria-label="Clear all data in namespace"
          >
            {loading ? "Destroying..." : "Clear All Data"}
          </Button>

          <span style={{ fontSize: "12px", color: "#666", marginLeft: "16px" }}>
            {dataManagement.pagination.total} total {dataManagement.dataType}
          </span>
        </div>

        {/* Embeddings Progress */}
        {dataManagement.embeddingsLoading && (
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
                {dataManagement.embeddingsProgress.current}/
                {dataManagement.embeddingsProgress.total}
              </div>
            </div>

            {dataManagement.embeddingsProgress.total > 0 && (
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
                      (dataManagement.embeddingsProgress.current /
                        dataManagement.embeddingsProgress.total) *
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
              {dataManagement.embeddingsProgress.message}
            </div>
          </div>
        )}

        {/* Export Progress */}
        {dataManagement.exportLoading && (
          <div
            style={{
              backgroundColor: "#e3f2fd",
              border: "1px solid #2196f3",
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
                  color: "#1565c0",
                }}
              >
                Exporting Data
              </div>
              <div
                style={{ marginLeft: "auto", fontSize: "12px", color: "#666" }}
              >
                {dataManagement.exportProgress.current}/
                {dataManagement.exportProgress.total}
              </div>
            </div>

            {dataManagement.exportProgress.total > 0 && (
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
                      (dataManagement.exportProgress.current /
                        dataManagement.exportProgress.total) *
                      100
                    }%`,
                    backgroundColor: "#2196f3",
                    height: "100%",
                    borderRadius: "4px",
                    transition: "width 0.3s ease",
                  }}
                />
              </div>
            )}

            <div style={{ fontSize: "12px", color: "#666" }}>
              {dataManagement.exportProgress.message}
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
          selectedRows={dataManagement.selectedRows}
          onSelectionChange={(rows) =>
            setDataManagement((prev) => ({ ...prev, selectedRows: rows }))
          }
          sortBy={dataManagement.sortBy}
          sortDirection={dataManagement.sortDirection}
          onSort={handleSort}
          pagination={{
            page: dataManagement.pagination.page,
            pageSize: dataManagement.pagination.pageSize,
            total: dataManagement.pagination.total,
            onPageChange: handlePageChange,
            onPageSizeChange: handlePageSizeChange,
          }}
          emptyMessage={`No ${dataManagement.dataType} found. Try adjusting your filters.`}
        />
      </div>
    </Section>
  );
}
