"use client";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useToast } from "@/components/ToastProvider";

type OrderRow = {
  id: string;
  userId: string;
  total: number;
  currency: string;
  createdAt: string;
};

export default function AdminOrdersPage() {
  const toast = useToast();
  const [items, setItems] = useState<OrderRow[]>([]);
  const [total, setTotal] = useState(0);
  const [limit, setLimit] = useState(20);
  const [offset, setOffset] = useState(0);
  const [selected, setSelected] = useState<Record<string, boolean>>({});
  const selectedIds = useMemo(
    () => Object.keys(selected).filter((k) => selected[k]),
    [selected]
  );

  const load = useCallback(async () => {
    const url = new URL(`/api/orders`, window.location.origin);
    url.searchParams.set("limit", String(limit));
    url.searchParams.set("offset", String(offset));
    const res = await fetch(url);
    const data = await res.json();
    setItems(
      (data.items || []).map(
        (o: {
          id: string;
          userId: string;
          total: number;
          currency: string;
          createdAt: string;
        }) => ({
          id: o.id,
          userId: o.userId,
          total: o.total,
          currency: o.currency,
          createdAt: o.createdAt,
        })
      )
    );
    setTotal(data.total || 0);
  }, [limit, offset]);

  useEffect(() => {
    load();
  }, [limit, offset, load]);

  async function onDeleteSelected() {
    // orders delete API not provided; use nuke for demo
    await fetch("/api/admin/nuke", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ tables: ["order", "orderItem"] }),
    });
    toast("Orders deleted (demo nuke)");
    load();
  }

  return (
    <main className="space-y-4 p-4">
      <h1 className="text-xl font-semibold">Admin · Orders</h1>
      <div className="ml-auto flex gap-2 items-center">
        <button
          className="border rounded px-3 py-2 text-sm"
          onClick={onDeleteSelected}
          disabled={!selectedIds.length}
        >
          Delete selected (demo)
        </button>
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
            <th className="p-2 border">User</th>
            <th className="p-2 border">Total</th>
            <th className="p-2 border">Created</th>
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
              <td className="p-2 border text-xs">{r.userId}</td>
              <td className="p-2 border">
                ${r.total.toFixed(2)} {r.currency}
              </td>
              <td className="p-2 border text-xs">
                {new Date(r.createdAt).toISOString()}
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
          <option value={10}>10</option>
          <option value={20}>20</option>
          <option value={50}>50</option>
        </select>
      </div>
    </main>
  );
}
