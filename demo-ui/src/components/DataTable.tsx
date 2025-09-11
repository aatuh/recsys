import React, { useState, useMemo } from "react";
import { Button } from "./UIComponents";

export interface Column<T> {
  key: keyof T | string;
  title: string;
  width?: string;
  sortable?: boolean;
  filterable?: boolean;
  render?: (value: any, row: T) => React.ReactNode;
  align?: "left" | "center" | "right";
}

export interface DataTableProps<T> {
  data: T[];
  columns: Column<T>[];
  loading?: boolean;
  selectable?: boolean;
  selectedRows?: Set<string>;
  onSelectionChange?: (selectedRows: Set<string>) => void;
  onRowClick?: (row: T) => void;
  sortBy?: string;
  sortDirection?: "asc" | "desc";
  onSort?: (column: string, direction: "asc" | "desc") => void;
  filters?: Record<string, any>;
  onFilterChange?: (filters: Record<string, any>) => void;
  pagination?: {
    page: number;
    pageSize: number;
    total: number;
    onPageChange: (page: number) => void;
    onPageSizeChange: (pageSize: number) => void;
  };
  emptyMessage?: string;
  rowKey?: (row: T) => string;
}

export function DataTable<T extends Record<string, any>>({
  data,
  columns,
  loading = false,
  selectable = false,
  selectedRows = new Set(),
  onSelectionChange,
  onRowClick,
  sortBy,
  sortDirection,
  onSort,
  filters = {},
  onFilterChange,
  pagination,
  emptyMessage = "No data available",
  rowKey = (row) => row.id || row.user_id || row.item_id || JSON.stringify(row),
}: DataTableProps<T>) {
  const [localFilters, setLocalFilters] =
    useState<Record<string, any>>(filters);

  const handleSort = (column: string) => {
    if (!onSort) return;

    const newDirection =
      sortBy === column && sortDirection === "asc" ? "desc" : "asc";
    onSort(column, newDirection);
  };

  const handleSelectAll = () => {
    if (!onSelectionChange) return;

    if (selectedRows.size === data.length) {
      onSelectionChange(new Set());
    } else {
      onSelectionChange(new Set(data.map(rowKey)));
    }
  };

  const handleSelectRow = (row: T) => {
    if (!onSelectionChange) return;

    const key = rowKey(row);
    const newSelection = new Set(selectedRows);

    if (newSelection.has(key)) {
      newSelection.delete(key);
    } else {
      newSelection.add(key);
    }

    onSelectionChange(newSelection);
  };

  const handleFilterChange = (column: string, value: any) => {
    const newFilters = { ...localFilters, [column]: value };
    setLocalFilters(newFilters);
    onFilterChange?.(newFilters);
  };

  const renderCell = (column: Column<T>, row: T) => {
    const value = String(column.key).includes(".")
      ? String(column.key)
          .split(".")
          .reduce((obj: any, key: string) => obj?.[key], row)
      : row[column.key as keyof T];

    if (column.render) {
      return column.render(value, row);
    }

    // Default rendering
    if (value === null || value === undefined) {
      return <span style={{ color: "#999" }}>-</span>;
    }

    if (typeof value === "boolean") {
      return (
        <span style={{ color: value ? "#4caf50" : "#f44336" }}>
          {value ? "✓" : "✗"}
        </span>
      );
    }

    if (typeof value === "object") {
      const jsonStr = JSON.stringify(value) || "";
      return (
        <code
          style={{
            fontSize: "11px",
            backgroundColor: "#f5f5f5",
            padding: "2px 4px",
            borderRadius: "3px",
          }}
        >
          {jsonStr.substring(0, 50)}
          {jsonStr.length > 50 ? "..." : ""}
        </code>
      );
    }

    return String(value);
  };

  const formatDate = (dateString: string) => {
    try {
      return new Date(dateString).toLocaleString();
    } catch {
      return dateString;
    }
  };

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: "40px", color: "#666" }}>
        <div>Loading...</div>
      </div>
    );
  }

  return (
    <div
      style={{
        border: "1px solid #e0e0e0",
        borderRadius: "8px",
        overflow: "hidden",
      }}
    >
      {/* Table */}
      <div style={{ overflowX: "auto" }}>
        <table
          style={{
            width: "100%",
            borderCollapse: "collapse",
            fontSize: "13px",
          }}
        >
          <thead>
            <tr
              style={{
                backgroundColor: "#f8f9fa",
                borderBottom: "2px solid #e0e0e0",
              }}
            >
              {selectable && (
                <th
                  style={{
                    padding: "12px 8px",
                    textAlign: "center",
                    width: "40px",
                  }}
                >
                  <input
                    type="checkbox"
                    checked={
                      data.length > 0 && selectedRows.size === data.length
                    }
                    onChange={handleSelectAll}
                    style={{ cursor: "pointer" }}
                  />
                </th>
              )}
              {columns.map((column) => (
                <th
                  key={String(column.key)}
                  style={{
                    padding: "12px 8px",
                    textAlign: column.align || "left",
                    fontWeight: "600",
                    width: column.width,
                    cursor: column.sortable ? "pointer" : "default",
                    userSelect: "none",
                  }}
                  onClick={() =>
                    column.sortable && handleSort(String(column.key))
                  }
                >
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "4px",
                    }}
                  >
                    {column.title}
                    {column.sortable && (
                      <span style={{ fontSize: "10px", color: "#666" }}>
                        {sortBy === column.key
                          ? sortDirection === "asc"
                            ? "↑"
                            : "↓"
                          : "↕"}
                      </span>
                    )}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.length === 0 ? (
              <tr>
                <td
                  colSpan={columns.length + (selectable ? 1 : 0)}
                  style={{
                    padding: "40px",
                    textAlign: "center",
                    color: "#666",
                    fontStyle: "italic",
                  }}
                >
                  {emptyMessage}
                </td>
              </tr>
            ) : (
              data.map((row, index) => {
                const key = rowKey(row);
                const isSelected = selectedRows.has(key);

                return (
                  <tr
                    key={key}
                    style={{
                      backgroundColor: isSelected
                        ? "#e3f2fd"
                        : index % 2 === 0
                        ? "#fff"
                        : "#fafafa",
                      borderBottom: "1px solid #f0f0f0",
                      cursor: onRowClick ? "pointer" : "default",
                    }}
                    onClick={() => onRowClick?.(row)}
                  >
                    {selectable && (
                      <td style={{ padding: "8px", textAlign: "center" }}>
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => handleSelectRow(row)}
                          onClick={(e) => e.stopPropagation()}
                          style={{ cursor: "pointer" }}
                        />
                      </td>
                    )}
                    {columns.map((column) => (
                      <td
                        key={String(column.key)}
                        style={{
                          padding: "8px",
                          textAlign: column.align || "left",
                          maxWidth: column.width || "200px",
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                          whiteSpace: "nowrap",
                        }}
                      >
                        {renderCell(column, row)}
                      </td>
                    ))}
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {pagination && (
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            padding: "12px 16px",
            backgroundColor: "#f8f9fa",
            borderTop: "1px solid #e0e0e0",
            fontSize: "12px",
          }}
        >
          <div style={{ color: "#666" }}>
            Showing{" "}
            {Math.min(
              (pagination.page - 1) * pagination.pageSize + 1,
              pagination.total
            )}{" "}
            to{" "}
            {Math.min(pagination.page * pagination.pageSize, pagination.total)}{" "}
            of {pagination.total} entries
          </div>

          <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
            <span>Rows per page:</span>
            <select
              value={pagination.pageSize}
              onChange={(e) =>
                pagination.onPageSizeChange(Number(e.target.value))
              }
              style={{
                padding: "2px 4px",
                border: "1px solid #ccc",
                borderRadius: "3px",
                fontSize: "12px",
              }}
            >
              <option value={10}>10</option>
              <option value={25}>25</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
            </select>

            <div style={{ display: "flex", gap: "4px" }}>
              <Button
                onClick={() => pagination.onPageChange(pagination.page - 1)}
                disabled={pagination.page <= 1}
                style={{
                  padding: "4px 8px",
                  fontSize: "12px",
                  minWidth: "auto",
                }}
              >
                Previous
              </Button>

              <span style={{ padding: "4px 8px", color: "#666" }}>
                Page {pagination.page} of{" "}
                {Math.ceil(pagination.total / pagination.pageSize)}
              </span>

              <Button
                onClick={() => pagination.onPageChange(pagination.page + 1)}
                disabled={
                  pagination.page >=
                  Math.ceil(pagination.total / pagination.pageSize)
                }
                style={{
                  padding: "4px 8px",
                  fontSize: "12px",
                  minWidth: "auto",
                }}
              >
                Next
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
