import React, { useState } from "react";
import { Button } from "../primitives/UIComponents";
import { color, radius, spacing, text } from "../../ui/tokens";
import type { PaginationState } from "../../types/ui";

export interface Column<T> {
  key: keyof T | string;
  title: string;
  width?: string;
  sortable?: boolean;
  filterable?: boolean;
  render?: (
    value: any,
    row: T,
    toggleExpansion?: () => void
  ) => React.ReactNode;
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
  pagination?: PaginationState & {
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
  pagination,
  emptyMessage = "No data available",
  rowKey = (row) => row.id || row.user_id || row.item_id || JSON.stringify(row),
}: DataTableProps<T>) {
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());
  // Simple virtualization: render only a window around visible page
  const windowedData = React.useMemo(() => {
    if (!pagination) return data;
    const start = (pagination.page - 1) * pagination.pageSize;
    const end = Math.min(start + pagination.pageSize, data.length);
    return data.slice(start, end);
  }, [data, pagination?.page, pagination?.pageSize]);

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

  const toggleRowExpansion = (rowKey: string) => {
    const newExpandedRows = new Set(expandedRows);
    if (newExpandedRows.has(rowKey)) {
      newExpandedRows.delete(rowKey);
    } else {
      newExpandedRows.add(rowKey);
    }
    setExpandedRows(newExpandedRows);
  };

  const renderCell = (column: Column<T>, row: T) => {
    const value = String(column.key).includes(".")
      ? String(column.key)
          .split(".")
          .reduce((obj: any, key: string) => obj?.[key], row)
      : row[column.key as keyof T];

    if (column.render) {
      return column.render(value, row, () => toggleRowExpansion(rowKey(row)));
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
      const isTruncated = jsonStr.length > 50;

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
            if (isTruncated) {
              e.stopPropagation();
              toggleRowExpansion(rowKey(row));
            }
          }}
          title={isTruncated ? "Click to expand row" : ""}
        >
          {jsonStr.substring(0, 50)}
          {isTruncated ? "..." : ""}
        </code>
      );
    }

    // For long strings, also make them clickable
    const stringValue = String(value);
    if (stringValue.length > 50) {
      return (
        <span
          style={{
            cursor: "pointer",
            borderBottom: "1px dotted #666",
          }}
          onClick={(e) => {
            e.stopPropagation();
            toggleRowExpansion(rowKey(row));
          }}
          title="Click to expand row"
        >
          {stringValue.substring(0, 50)}...
        </span>
      );
    }

    return stringValue;
  };

  if (loading) {
    return (
      <div
        role="status"
        aria-live="polite"
        style={{ textAlign: "center", padding: 40, color: color.textMuted }}
      >
        <div>Loading...</div>
      </div>
    );
  }

  return (
    <>
      <div
        style={{
          border: `1px solid ${color.panelBorder}`,
          borderRadius: radius.lg,
          overflow: "hidden",
        }}
      >
        {/* Table */}
        <div style={{ overflowX: "auto" }}>
          <table
            style={{
              width: "100%",
              borderCollapse: "collapse",
              fontSize: text.sm,
              minWidth: "600px", // Ensure table doesn't get too cramped on mobile
            }}
          >
            <thead>
              <tr
                style={{
                  backgroundColor: color.panelSubtle,
                  borderBottom: `2px solid ${color.border}`,
                }}
              >
                {selectable && (
                  <th
                    style={{
                      padding: spacing.sm,
                      textAlign: "center",
                      width: "40px",
                      minWidth: "40px",
                    }}
                    scope="col"
                  >
                    <input
                      type="checkbox"
                      checked={
                        data.length > 0 && selectedRows.size === data.length
                      }
                      onChange={handleSelectAll}
                      style={{ cursor: "pointer" }}
                      aria-label="Select all rows"
                    />
                  </th>
                )}
                {columns.map((column) => (
                  <th
                    key={String(column.key)}
                    style={{
                      padding: spacing.sm,
                      textAlign: column.align || "left",
                      fontWeight: 600,
                      width: column.width,
                      cursor: column.sortable ? "pointer" : "default",
                      userSelect: "none",
                      minWidth: "100px", // Prevent columns from getting too narrow
                    }}
                    scope="col"
                    onClick={() =>
                      column.sortable && handleSort(String(column.key))
                    }
                  >
                    <div
                      style={{
                        display: "flex",
                        alignItems: "center",
                        gap: spacing.xs,
                      }}
                    >
                      <span style={{ fontSize: text.xs }}>{column.title}</span>
                      {column.sortable && (
                        <span
                          style={{ fontSize: text.xs, color: color.textMuted }}
                        >
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
                      padding: 40,
                      textAlign: "center",
                      color: color.textMuted,
                      fontStyle: "italic",
                    }}
                  >
                    {emptyMessage}
                  </td>
                </tr>
              ) : (
                windowedData.map((row, index) => {
                  const key = rowKey(row);
                  const isSelected = selectedRows.has(key);
                  const isExpanded = expandedRows.has(key);

                  return (
                    <React.Fragment key={key}>
                      <tr
                        style={{
                          backgroundColor: isSelected
                            ? "#e3f2fd"
                            : index % 2 === 0
                            ? "#fff"
                            : color.panelSubtle,
                          borderBottom: `1px solid ${color.panelBorder}`,
                          cursor: onRowClick ? "pointer" : "default",
                        }}
                        onClick={() => onRowClick?.(row)}
                      >
                        {selectable && (
                          <td
                            style={{ padding: spacing.md, textAlign: "center" }}
                          >
                            <input
                              type="checkbox"
                              checked={isSelected}
                              onChange={() => handleSelectRow(row)}
                              onClick={(e) => e.stopPropagation()}
                              style={{ cursor: "pointer" }}
                              aria-label={`Select row ${key}`}
                            />
                          </td>
                        )}
                        {columns.map((column) => (
                          <td
                            key={String(column.key)}
                            style={{
                              padding: spacing.md,
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
                      {isExpanded && (
                        <tr>
                          <td
                            colSpan={columns.length + (selectable ? 1 : 0)}
                            style={{
                              padding: spacing.lg,
                              backgroundColor: color.panelSubtle,
                              borderBottom: `1px solid ${color.panelBorder}`,
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
                              <h4
                                style={{
                                  margin: 0,
                                  fontSize: text.md,
                                  fontWeight: 600,
                                  color: color.text,
                                }}
                              >
                                Full Data - {key}
                              </h4>
                              <Button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  toggleRowExpansion(key);
                                }}
                                style={{
                                  padding: "4px 8px",
                                  fontSize: 12,
                                  backgroundColor: color.panelSubtle,
                                  border: `1px solid ${color.border}`,
                                }}
                              >
                                ✕ Close
                              </Button>
                            </div>
                            <div
                              style={{
                                backgroundColor: color.panelSubtle,
                                border: `1px solid ${color.panelBorder}`,
                                borderRadius: radius.md,
                                padding: spacing.lg,
                                fontFamily: "monospace",
                                fontSize: 12,
                                lineHeight: "1.4",
                                maxHeight: "300px",
                                overflow: "auto",
                              }}
                            >
                              <pre
                                style={{
                                  margin: 0,
                                  whiteSpace: "pre-wrap",
                                  wordBreak: "break-word",
                                  color: color.text,
                                }}
                              >
                                {JSON.stringify(row, null, 2)}
                              </pre>
                            </div>
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination - Mobile-friendly */}
        {pagination && (
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: spacing.sm,
              padding: spacing.md,
              backgroundColor: color.panelSubtle,
              borderTop: `1px solid ${color.border}`,
            }}
          >
            <div style={{ textAlign: "center" }}>
              <span style={{ fontSize: text.sm, color: color.textMuted }}>
                Showing{" "}
                {Math.min(
                  (pagination.page - 1) * pagination.pageSize + 1,
                  pagination.total
                )}{" "}
                to{" "}
                {Math.min(
                  pagination.page * pagination.pageSize,
                  pagination.total
                )}{" "}
                of {pagination.total} entries
              </span>
            </div>

            <div
              style={{
                display: "flex",
                justifyContent: "center",
                alignItems: "center",
                gap: spacing.sm,
                flexWrap: "wrap",
              }}
            >
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: spacing.xs,
                }}
              >
                <span style={{ fontSize: text.xs }}>Rows:</span>
                <select
                  value={pagination.pageSize}
                  onChange={(e) =>
                    pagination.onPageSizeChange(Number(e.target.value))
                  }
                  style={{
                    padding: `${spacing.xs}px`,
                    border: `1px solid ${color.border}`,
                    borderRadius: radius.sm,
                    fontSize: text.xs,
                  }}
                >
                  <option value={10}>10</option>
                  <option value={25}>25</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
              </div>

              <div
                style={{
                  display: "flex",
                  gap: spacing.xs,
                  alignItems: "center",
                }}
              >
                <Button
                  onClick={() => pagination.onPageChange(pagination.page - 1)}
                  disabled={pagination.page <= 1}
                  style={{
                    padding: `${spacing.xs}px ${spacing.sm}px`,
                    fontSize: text.xs,
                    minWidth: "60px",
                  }}
                >
                  Prev
                </Button>

                <span
                  style={{
                    padding: `${spacing.xs}px ${spacing.sm}px`,
                    color: color.textMuted,
                    fontSize: text.xs,
                    minWidth: "80px",
                    textAlign: "center",
                  }}
                >
                  {pagination.page} /{" "}
                  {Math.ceil(pagination.total / pagination.pageSize)}
                </span>

                <Button
                  onClick={() => pagination.onPageChange(pagination.page + 1)}
                  disabled={
                    pagination.page >=
                    Math.ceil(pagination.total / pagination.pageSize)
                  }
                  style={{
                    padding: `${spacing.xs}px ${spacing.sm}px`,
                    fontSize: text.xs,
                    minWidth: "60px",
                  }}
                >
                  Next
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  );
}
