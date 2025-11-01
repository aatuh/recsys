"use client";
import { useEffect, useMemo, useState } from "react";
import { useToast } from "@/components/ToastProvider";

type Product = {
  id: string;
  name: string;
  sku: string;
  price: number;
  currency: string;
  brand: string;
  category: string;
  stockCount: number;
};

export default function AdminProducts() {
  const toast = useToast();
  const [items, setItems] = useState<Product[]>([]);
  const [total, setTotal] = useState(0);
  const [limit, setLimit] = useState(20);
  const [offset, setOffset] = useState(0);
  const [q, setQ] = useState("");
  const [selected, setSelected] = useState<Record<string, boolean>>({});
  const [loading, setLoading] = useState(false);
  const [batchCount, setBatchCount] = useState(50);
  const [form, setForm] = useState({
    name: "",
    sku: "",
    price: "",
    currency: "USD",
    brand: "",
    category: "",
    imageUrl: "",
    stockCount: "",
    tagsCsv: "",
    description: "",
  });
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  function setField<K extends keyof typeof form>(key: K, val: string) {
    setForm((f) => ({ ...f, [key]: val }));
    setTouched((t) => ({ ...t, [key]: true }));
  }

  const selectedIds = useMemo(
    () => Object.keys(selected).filter((k) => selected[k]),
    [selected]
  );

  async function load() {
    setLoading(true);
    try {
      const url = new URL(`/api/products`, window.location.origin);
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [limit, offset]);

  // Populate form when selection changes
  useEffect(() => {
    setTouched({});
    const ids = selectedIds;
    if (ids.length === 1) {
      const p = items.find((x) => x.id === ids[0]);
      if (p) {
        setForm({
          name: p.name || "",
          sku: p.sku || "",
          price: String(p.price ?? ""),
          currency: p.currency || "USD",
          brand: p.brand || "",
          category: p.category || "",
          imageUrl: "",
          stockCount: String(p.stockCount ?? ""),
          tagsCsv: "",
          description: "",
        });
      }
    } else if (ids.length > 1) {
      const sel = items.filter((x) => selected[x.id]);
      const allEq = (getter: (p: Product) => string | number) => {
        if (sel.length === 0) return "";
        const v = getter(sel[0]);
        for (const s of sel) if (getter(s) !== v) return "";
        return String(v ?? "");
      };
      setForm({
        name: allEq((p) => p.name) || "",
        sku: allEq((p) => p.sku) || "",
        price: allEq((p) => String(p.price)) || "",
        currency: allEq((p) => p.currency) || "USD",
        brand: allEq((p) => p.brand) || "",
        category: allEq((p) => p.category) || "",
        imageUrl: "",
        stockCount: allEq((p) => String(p.stockCount)) || "",
        tagsCsv: "",
        description: "",
      });
    } else {
      setForm({
        name: "",
        sku: "",
        price: "",
        currency: "USD",
        brand: "",
        category: "",
        imageUrl: "",
        stockCount: "",
        tagsCsv: "",
        description: "",
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedIds.join("|"), items]);

  async function onDeleteSelected() {
    if (selectedIds.length === 0) return;
    await fetch("/api/products/batch", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ action: "delete", ids: selectedIds }),
    });
    toast(`Deleted ${selectedIds.length} products`);
    load();
  }

  async function onUpdateSelected(data: Partial<Product>) {
    if (selectedIds.length === 0) return;
    await fetch("/api/products/batch", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ action: "update", ids: selectedIds, data }),
    });
    toast(`Updated ${selectedIds.length} products`);
    load();
  }

  async function onSeed(count = 50) {
    await fetch("/api/products/seed", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ count }),
    });
    toast(`Inserted ${count} products`);
    load();
  }

  async function onNuke() {
    if (!confirm("Delete ALL products?")) return;
    await fetch("/api/admin/nuke", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ tables: ["product"] }),
    });
    toast("All products deleted");
    load();
  }

  async function onSyncAll() {
    try {
      const response = await fetch("/api/admin/sync-items", {
        method: "POST",
        headers: { "content-type": "application/json" },
      });
      const result = await response.json();
      if (response.ok) {
        toast(
          `Successfully synced ${result.synced || "all"} products to recsys`
        );
      } else {
        toast(`Sync failed: ${result.message || "Unknown error"}`);
      }
    } catch (error) {
      toast(
        `Sync failed: ${
          error instanceof Error ? error.message : "Unknown error"
        }`
      );
    }
  }

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const ids = selectedIds;
    if (!ids.length) {
      const payload = {
        name: form.name,
        sku: form.sku,
        price: form.price ? parseFloat(form.price) : 0,
        currency: form.currency || "USD",
        brand: form.brand,
        category: form.category,
        imageUrl: form.imageUrl,
        stockCount: form.stockCount ? parseInt(form.stockCount, 10) : 0,
        tagsCsv: form.tagsCsv,
        description: form.description,
      };
      await fetch("/api/products", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(payload),
      });
      toast("Product created");
      setForm({
        name: "",
        sku: "",
        price: "",
        currency: "USD",
        brand: "",
        category: "",
        imageUrl: "",
        stockCount: "",
        tagsCsv: "",
        description: "",
      });
      load();
      return;
    }
    const data: Record<string, unknown> = {};
    if (touched.name) data.name = form.name;
    if (touched.sku) data.sku = form.sku;
    if (touched.price) data.price = form.price ? parseFloat(form.price) : 0;
    if (touched.currency) data.currency = form.currency || "USD";
    if (touched.brand) data.brand = form.brand;
    if (touched.category) data.category = form.category;
    if (touched.imageUrl) data.imageUrl = form.imageUrl;
    if (touched.stockCount)
      data.stockCount = form.stockCount ? parseInt(form.stockCount, 10) : 0;
    if (touched.tagsCsv) data.tagsCsv = form.tagsCsv;
    if (touched.description) data.description = form.description;
    if (!Object.keys(data).length) return toast("No changes to apply");
    await fetch("/api/products/batch", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ action: "update", ids, data }),
    });
    toast(`Updated ${ids.length} products`);
    setTouched({});
    load();
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
          <button
            className="border rounded px-3 py-2 text-sm"
            onClick={() => onUpdateSelected({ brand: "Acme" })}
            disabled={!selectedIds.length}
          >
            Set brand=Acme
          </button>
          <button
            className="border rounded px-3 py-2 text-sm bg-blue-50 hover:bg-blue-100"
            onClick={onSyncAll}
            disabled={loading}
          >
            Sync to Recsys
          </button>
          <button className="border rounded px-3 py-2 text-sm" onClick={onNuke}>
            Nuke products
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
                    items.forEach((p) => (next[p.id] = true));
                  setSelected(next);
                }}
              />
            </th>
            <th className="p-2 border">ID</th>
            <th className="p-2 border">Name</th>
            <th className="p-2 border">Brand</th>
            <th className="p-2 border">Category</th>
            <th className="p-2 border">Price</th>
            <th className="p-2 border">Stock</th>
          </tr>
        </thead>
        <tbody>
          {items.map((p) => (
            <tr key={p.id} className="border-b">
              <td className="p-2 border">
                <input
                  type="checkbox"
                  checked={!!selected[p.id]}
                  onChange={(e) =>
                    setSelected((s) => ({ ...s, [p.id]: e.target.checked }))
                  }
                />
              </td>
              <td className="p-2 border text-xs font-mono text-gray-600">
                {p.id}
              </td>
              <td className="p-2 border">{p.name}</td>
              <td className="p-2 border">{p.brand}</td>
              <td className="p-2 border">{p.category}</td>
              <td className="p-2 border">
                ${p.price.toFixed(2)} {p.currency}
              </td>
              <td className="p-2 border">{p.stockCount}</td>
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
            name="name"
            placeholder="Name"
            value={form.name}
            onChange={(e) => setField("name", e.target.value)}
          />
          <input
            className="border p-2"
            name="sku"
            placeholder="SKU"
            value={form.sku}
            onChange={(e) => setField("sku", e.target.value)}
          />
          <input
            className="border p-2"
            name="price"
            placeholder="Price"
            type="number"
            step="0.01"
            value={form.price}
            onChange={(e) => setField("price", e.target.value)}
          />
          <input
            className="border p-2"
            name="currency"
            placeholder="Currency"
            value={form.currency}
            onChange={(e) => setField("currency", e.target.value)}
          />
          <input
            className="border p-2"
            name="brand"
            placeholder={
              selectedIds.length > 1 && !form.brand ? "— multiple —" : "Brand"
            }
            value={form.brand}
            onChange={(e) => setField("brand", e.target.value)}
          />
          <input
            className="border p-2"
            name="category"
            placeholder={
              selectedIds.length > 1 && !form.category
                ? "— multiple —"
                : "Category"
            }
            value={form.category}
            onChange={(e) => setField("category", e.target.value)}
          />
          <input
            className="border p-2"
            name="imageUrl"
            placeholder="Image URL"
            value={form.imageUrl}
            onChange={(e) => setField("imageUrl", e.target.value)}
          />
          <input
            className="border p-2"
            name="stockCount"
            placeholder="Stock"
            type="number"
            value={form.stockCount}
            onChange={(e) => setField("stockCount", e.target.value)}
          />
          <input
            className="border p-2"
            name="tagsCsv"
            placeholder="Tags CSV"
            value={form.tagsCsv}
            onChange={(e) => setField("tagsCsv", e.target.value)}
          />
          <textarea
            className="border p-2 md:col-span-3"
            name="description"
            placeholder="Description"
            value={form.description}
            onChange={(e) => setField("description", e.target.value)}
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
