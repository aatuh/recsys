"use client";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useToast } from "@/components/ToastProvider";

type EventRow = {
  id: string;
  type: string;
  userId: string;
  productId?: string;
  ts: string;
  recsysStatus: string;
};

export default function AdminEvents() {
  const toast = useToast();
  const [items, setItems] = useState<EventRow[]>([]);
  const [total, setTotal] = useState(0);
  const [limit, setLimit] = useState(50);
  const [offset, setOffset] = useState(0);
  const [type, setType] = useState("");
  const [selected, setSelected] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(false);

  const selectedIds = useMemo(
    () => Object.keys(selected).filter((k) => selected[k]),
    [selected]
  );

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const url = new URL(`/api/events`, window.location.origin);
      url.searchParams.set("limit", String(limit));
      url.searchParams.set("offset", String(offset));
      if (type) url.searchParams.set("type", type);
      const res = await fetch(url);
      const data = await res.json();
      setItems(
        (data.items || []).map(
          (e: {
            id: string;
            type: string;
            userId: string;
            productId: string | null;
            ts: string;
            recsysStatus: string;
          }) => ({
            id: e.id,
            type: e.type,
            userId: e.userId,
            productId: e.productId,
            ts: e.ts,
            recsysStatus: e.recsysStatus,
          })
        )
      );
      setTotal(data.total || 0);
      setSelected({});
    } finally {
      setLoading(false);
    }
  }, [limit, offset, type]);

  useEffect(() => {
    load();
  }, [limit, offset, load]);

  async function onDeleteSelected() {
    if (!selectedIds.length) return;
    await fetch("/api/events/batch", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ action: "delete", ids: selectedIds }),
    });
    toast(`Deleted ${selectedIds.length} events`);
    load();
  }
  async function onRetryFailed() {
    await fetch("/api/events/retry-failed", { method: "POST" });
    toast("Retry triggered");
    load();
  }
  async function onFlushPending() {
    await fetch("/api/events/flush", { method: "POST" });
    toast("Flush triggered");
    load();
  }
  async function onDeletePending() {
    if (!confirm("Delete ALL pending events?")) return;
    await fetch("/api/events/batch", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ action: "delete-pending" }),
    });
    toast("All pending events deleted");
    load();
  }

  async function onDeleteFromRecsys() {
    if (!selectedIds.length) {
      alert("Please select events to delete from Recsys");
      return;
    }
    if (!confirm(`Delete ${selectedIds.length} events from Recsys?`)) return;

    try {
      const response = await fetch("/api/events/batch", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({
          action: "delete-from-recsys",
          ids: selectedIds,
        }),
      });

      if (!response.ok) {
        throw new Error(
          `Failed to delete events from Recsys: ${response.status}`
        );
      }

      const result = await response.json();
      toast(`Deleted ${result.deleted || 0} events from Recsys`);
      load();
    } catch (error) {
      console.error("Failed to delete events from Recsys:", error);
      toast("Failed to delete events from Recsys");
    }
  }

  async function onNuke() {
    if (!confirm("Delete ALL events?")) return;
    await fetch("/api/admin/nuke", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ tables: ["events"] }),
    });
    toast("All events deleted");
    load();
  }

  return (
    <section className="space-y-4">
      <div className="flex gap-2 items-center">
        <select
          className="border p-2 text-sm"
          value={type}
          onChange={(e) => setType(e.target.value)}
        >
          <option value="">All types</option>
          <option value="view">view</option>
          <option value="click">click</option>
          <option value="add">add</option>
          <option value="purchase">purchase</option>
          <option value="custom">custom</option>
        </select>
        <button
          className="border rounded px-3 py-2 text-sm"
          onClick={() => {
            setOffset(0);
            load();
          }}
          disabled={loading}
        >
          Filter
        </button>
        <div className="ml-auto flex gap-2 items-center">
          <button
            className="border rounded px-3 py-2 text-sm"
            onClick={onFlushPending}
          >
            Flush pending
          </button>
          <button
            className="border rounded px-3 py-2 text-sm"
            onClick={onRetryFailed}
          >
            Retry failed
          </button>
          <button
            className="border rounded px-3 py-2 text-sm bg-red-50 text-red-700 hover:bg-red-100"
            onClick={onDeletePending}
          >
            Delete pending
          </button>
          <button
            className="border rounded px-3 py-2 text-sm"
            onClick={onDeleteSelected}
            disabled={!selectedIds.length}
          >
            Delete selected
          </button>
          <button
            className="border rounded px-3 py-2 text-sm bg-orange-50 text-orange-700 hover:bg-orange-100"
            onClick={onDeleteFromRecsys}
            disabled={!selectedIds.length}
          >
            Delete from Recsys
          </button>
          <button
            className="border rounded px-3 py-2 text-sm bg-red-50 text-red-700 hover:bg-red-100"
            onClick={onNuke}
          >
            Nuke events
          </button>
        </div>
      </div>

      <table className="w-full text-sm border">
        <thead>
          <tr className="bg-gray-50">
            <th className="p-2 border">
              <input
                type="checkbox"
                checked={
                  items.length > 0 && selectedIds.length === items.length
                }
                onChange={(e) => {
                  const next: Record<string, boolean> = {};
                  if (e.target.checked)
                    items.forEach((r) => (next[r.id] = true));
                  setSelected(next);
                }}
              />
            </th>
            <th className="p-2 border">Type</th>
            <th className="p-2 border">User</th>
            <th className="p-2 border">Product</th>
            <th className="p-2 border">Status</th>
            <th className="p-2 border">Timestamp</th>
          </tr>
        </thead>
        <tbody>
          {items.map((r) => (
            <tr key={r.id} className="border-b">
              <td className="p-2 border">
                <input
                  type="checkbox"
                  checked={!!selected[r.id]}
                  onChange={(e) =>
                    setSelected((s) => ({ ...s, [r.id]: e.target.checked }))
                  }
                />
              </td>
              <td className="p-2 border">{r.type}</td>
              <td className="p-2 border text-xs">{r.userId}</td>
              <td className="p-2 border text-xs">{r.productId || ""}</td>
              <td className="p-2 border">{r.recsysStatus}</td>
              <td className="p-2 border text-xs">
                {new Date(r.ts).toISOString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      <div className="flex items-center gap-2">
        <button
          className="border rounded px-2 py-1 text-sm"
          onClick={() => setOffset(Math.max(0, offset - limit))}
          disabled={offset === 0}
        >
          Prev
        </button>
        <div className="text-xs text-gray-600">
          {offset + 1} - {Math.min(offset + limit, total)} of {total}
        </div>
        <button
          className="border rounded px-2 py-1 text-sm"
          onClick={() => setOffset(offset + limit)}
          disabled={offset + limit >= total}
        >
          Next
        </button>
        <select
          className="border p-1 text-sm ml-2"
          value={limit}
          onChange={(e) => setLimit(parseInt(e.target.value, 10))}
        >
          <option value={20}>20</option>
          <option value={50}>50</option>
          <option value={100}>100</option>
        </select>
      </div>
    </section>
  );
}
