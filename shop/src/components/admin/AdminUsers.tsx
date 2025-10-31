"use client";
import { useEffect, useMemo, useState } from "react";
import { useToast } from "@/components/ToastProvider";

type User = { id: string; displayName: string };

export default function AdminUsers() {
  const toast = useToast();
  const [items, setItems] = useState<User[]>([]);
  const [total, setTotal] = useState(0);
  const [limit, setLimit] = useState(20);
  const [offset, setOffset] = useState(0);
  const [q, setQ] = useState("");
  const [selected, setSelected] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(false);
  const [batchCount, setBatchCount] = useState(20);
  const [form, setForm] = useState({ displayName: "" });
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const selectedIds = useMemo(
    () => Object.keys(selected).filter((k) => selected[k]),
    [selected]
  );

  function setField(val: string) {
    setForm({ displayName: val });
    setTouched({ displayName: true });
  }

  async function load() {
    setLoading(true);
    try {
      const url = new URL(`/api/users`, window.location.origin);
      url.searchParams.set("limit", String(limit));
      url.searchParams.set("offset", String(offset));
      if (q) url.searchParams.set("q", q);
      const res = await fetch(url);
      const data = await res.json();
      setItems(data.items || []);
      setTotal(data.total || 0);
      setSelected({});
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, [limit, offset]);

  useEffect(() => {
    setTouched({});
    if (selectedIds.length === 1) {
      const u = items.find((x) => x.id === selectedIds[0]);
      setForm({ displayName: u?.displayName || "" });
    } else if (selectedIds.length > 1) {
      const sel = items.filter((x) => selected[x.id]);
      const allEq =
        sel.length > 0 &&
        sel.every((s) => s.displayName === sel[0].displayName);
      setForm({ displayName: allEq ? sel[0].displayName : "" });
    } else {
      setForm({ displayName: "" });
    }
  }, [selectedIds.join("|"), items]);

  async function onDeleteSelected() {
    if (!selectedIds.length) return;
    await fetch("/api/users/batch", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ action: "delete", ids: selectedIds }),
    });
    toast(`Deleted ${selectedIds.length} users`);
    load();
  }

  async function onUpdateSelected() {
    if (!selectedIds.length) return;
    const data: any = {};
    if (touched.displayName) data.displayName = form.displayName;
    if (!Object.keys(data).length) return toast("No changes to apply");
    await fetch("/api/users/batch", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ action: "update", ids: selectedIds, data }),
    });
    toast(`Updated ${selectedIds.length} users`);
    setTouched({});
    load();
  }

  async function onSeed(count: number) {
    await fetch("/api/users/seed", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ count }),
    });
    toast(`Inserted ${count} users`);
    load();
  }

  async function onNuke() {
    if (!confirm("Delete ALL users?")) return;
    await fetch("/api/admin/nuke", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ tables: ["user"] }),
    });
    toast("All users deleted");
    load();
  }

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!selectedIds.length) {
      await fetch("/api/users", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ displayName: form.displayName }),
      });
      toast("User created");
      setForm({ displayName: "" });
      load();
      return;
    }
    await onUpdateSelected();
  }

  return (
    <section className="space-y-4">
      <div className="flex gap-2 items-center">
        <input
          className="border p-2 text-sm"
          placeholder="Search…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
        />
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
          <label className="text-xs text-gray-600">Batch count</label>
          <input
            className="border p-1 w-20 text-sm"
            type="number"
            min={1}
            max={1000}
            value={batchCount}
            onChange={(e) => setBatchCount(parseInt(e.target.value || "0", 10))}
          />
          <button
            className="border rounded px-3 py-2 text-sm"
            onClick={() => onSeed(Math.max(1, Math.min(1000, batchCount || 0)))}
          >
            Create batch
          </button>
          <button
            className="border rounded px-3 py-2 text-sm"
            onClick={onDeleteSelected}
            disabled={!selectedIds.length}
          >
            Delete selected
          </button>
          <button className="border rounded px-3 py-2 text-sm" onClick={onNuke}>
            Nuke users
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
                    items.forEach((u) => (next[u.id] = true));
                  setSelected(next);
                }}
              />
            </th>
            <th className="p-2 border">Display name</th>
            <th className="p-2 border">ID</th>
          </tr>
        </thead>
        <tbody>
          {items.map((u) => (
            <tr key={u.id} className="border-b">
              <td className="p-2 border">
                <input
                  type="checkbox"
                  checked={!!selected[u.id]}
                  onChange={(e) =>
                    setSelected((s) => ({ ...s, [u.id]: e.target.checked }))
                  }
                />
              </td>
              <td className="p-2 border">{u.displayName}</td>
              <td className="p-2 border text-xs text-gray-600">{u.id}</td>
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

      <section className="space-y-2 border-t pt-4">
        <h2 className="font-medium">Create or update selected</h2>
        <form
          className="grid grid-cols-1 md:grid-cols-3 gap-2"
          onSubmit={onSubmit}
        >
          <input
            className="border p-2"
            name="displayName"
            placeholder={
              selectedIds.length > 1 && !form.displayName
                ? "— multiple —"
                : "Display name"
            }
            value={form.displayName}
            onChange={(e) => setField(e.target.value)}
          />
          <div className="md:col-span-3">
            <button className="border rounded px-3 py-2 text-sm">
              {selectedIds.length ? "Apply to selected" : "Create"}
            </button>
          </div>
        </form>
      </section>
    </section>
  );
}
