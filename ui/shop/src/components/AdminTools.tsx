"use client";
import { useState } from "react";

interface AdminToolsProps {
  onAction: (action: string, params?: Record<string, unknown>) => Promise<void>;
}

export function AdminTools({ onAction }: AdminToolsProps) {
  const [loading, setLoading] = useState<string | null>(null);

  const handleAction = async (
    action: string,
    params?: Record<string, unknown>
  ) => {
    setLoading(action);
    try {
      await onAction(action, params);
    } catch (error) {
      console.error(`Failed to execute ${action}:`, error);
    } finally {
      setLoading(null);
    }
  };

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">Admin Tools</h2>

      {/* Data Sync Tools */}
      <div className="border rounded p-4">
        <h3 className="text-lg font-medium mb-3">Data Synchronization</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <button
            onClick={() => handleAction("sync-all-items")}
            disabled={loading === "sync-all-items"}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {loading === "sync-all-items"
              ? "Syncing..."
              : "Sync All Items to Recsys"}
          </button>

          <button
            onClick={() => handleAction("sync-all-users")}
            disabled={loading === "sync-all-users"}
            className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
          >
            {loading === "sync-all-users"
              ? "Syncing..."
              : "Sync All Users to Recsys"}
          </button>

          <button
            onClick={() => handleAction("flush-events")}
            disabled={loading === "flush-events"}
            className="px-4 py-2 bg-purple-600 text-white rounded hover:bg-purple-700 disabled:opacity-50"
          >
            {loading === "flush-events"
              ? "Flushing..."
              : "Flush Pending Events"}
          </button>

          <button
            onClick={() => handleAction("retry-failed-events")}
            disabled={loading === "retry-failed-events"}
            className="px-4 py-2 bg-orange-600 text-white rounded hover:bg-orange-700 disabled:opacity-50"
          >
            {loading === "retry-failed-events"
              ? "Retrying..."
              : "Retry Failed Events"}
          </button>
        </div>
      </div>

      {/* Data Management Tools */}
      <div className="border rounded p-4">
        <h3 className="text-lg font-medium mb-3">Data Management</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <button
            onClick={() => handleAction("seed-products", { count: 50 })}
            disabled={loading === "seed-products"}
            className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
          >
            {loading === "seed-products"
              ? "Seeding..."
              : "Seed 50 Random Products"}
          </button>

          <button
            onClick={() => handleAction("seed-users", { count: 20 })}
            disabled={loading === "seed-users"}
            className="px-4 py-2 bg-teal-600 text-white rounded hover:bg-teal-700 disabled:opacity-50"
          >
            {loading === "seed-users" ? "Seeding..." : "Seed 20 Random Users"}
          </button>

          <button
            onClick={() => handleAction("init-event-types")}
            disabled={loading === "init-event-types"}
            className="px-4 py-2 bg-cyan-600 text-white rounded hover:bg-cyan-700 disabled:opacity-50"
          >
            {loading === "init-event-types"
              ? "Initializing..."
              : "Initialize Event Type Config"}
          </button>

          <button
            onClick={() => {
              if (confirm("This will delete ALL data. Are you sure?")) {
                handleAction("nuke-all");
              }
            }}
            disabled={loading === "nuke-all"}
            className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 disabled:opacity-50"
          >
            {loading === "nuke-all" ? "Nuking..." : "Nuke All Data"}
          </button>
        </div>
      </div>

      {/* Health Checks */}
      <div className="border rounded p-4">
        <h3 className="text-lg font-medium mb-3">Health Checks</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <button
            onClick={() => handleAction("check-recsys-health")}
            disabled={loading === "check-recsys-health"}
            className="px-4 py-2 bg-emerald-600 text-white rounded hover:bg-emerald-700 disabled:opacity-50"
          >
            {loading === "check-recsys-health"
              ? "Checking..."
              : "Check Recsys Health"}
          </button>

          <button
            onClick={() => handleAction("validate-data-integrity")}
            disabled={loading === "validate-data-integrity"}
            className="px-4 py-2 bg-amber-600 text-white rounded hover:bg-amber-700 disabled:opacity-50"
          >
            {loading === "validate-data-integrity"
              ? "Validating..."
              : "Validate Data Integrity"}
          </button>
      </div>
    </div>

  </div>
);
}
