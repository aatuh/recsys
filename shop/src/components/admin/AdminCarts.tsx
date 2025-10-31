"use client";
import { useEffect, useMemo, useState } from "react";
import { useToast } from "@/components/ToastProvider";

type CartRow = { id: string; userId: string; items: number; updatedAt: string };

export default function AdminCarts() {
  const toast = useToast();
  const [items, setItems] = useState<CartRow[]>([]);
  const [total, setTotal] = useState(0);
  const [limit, setLimit] = useState(20);
  const [offset, setOffset] = useState(0);
  const [selected, setSelected] = useState<Record<string, boolean>>({});
  const selectedIds = useMemo(
    () => Object.keys(selected).filter((k) => selected[k]),
    [selected]
  );

  async function load() {
    const url = new URL(`/api/admin/carts`, window.location.origin);
    url.searchParams.set("limit", String(limit));
    url.searchParams.set("offset", String(offset));
    const res = await fetch(url.toString());
    const data = await res.json();
    setItems(data.items || []);
    setTotal(data.total || 0);
  }

  useEffect(() => {
    load();
  }, [limit, offset]);

  async function onDeleteSelected() {
    for (const uid of selectedIds) {
      await fetch(`/api/cart?userId=${encodeURIComponent(uid)}`, {
        method: "DELETE",
      });
    }
    toast(`Cleared carts for ${selectedIds.length} users`);
    load();
  }

  return (
    <section className="space-y-4">
      <div className="ml-auto flex gap-2 items-center">
        <button
          className="border rounded px-3 py-2 text-sm"
          onClick={onDeleteSelected}
          disabled={!selectedIds.length}
        >
          Clear selected carts
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
            <th className="p-2 border">Items</th>
            <th className="p-2 border">Updated</th>
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
              <td className="p-2 border">{r.items}</td>
              <td className="p-2 border text-xs">
                {new Date(r.updatedAt).toISOString()}
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
    </section>
  );
}
