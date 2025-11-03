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
  imageUrl: string | null;
  stockCount: number;
  tagsCsv: string;
  description: string | null;
  attributesJson: string | null;
};

function formatAttributes(raw: string | null | undefined): string {
  if (!raw) return "";
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

function parseCsvList(value: string): string[] {
  return value
    .split(",")
    .map((entry) => entry.trim())
    .filter(Boolean);
}

function parseCategoryLines(value: string): string[] {
  return value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);
}

function parseAttributesConfig(
  value: string
): Record<string, string[]> | undefined {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  try {
    const parsed = JSON.parse(trimmed) as Record<string, unknown>;
    const result: Record<string, string[]> = {};
    Object.entries(parsed).forEach(([key, entry]) => {
      if (Array.isArray(entry)) {
        result[key] = entry.map((item) => String(item));
      } else if (entry !== null && entry !== undefined) {
        result[key] = [String(entry)];
      }
    });
    return Object.keys(result).length > 0 ? result : undefined;
  } catch {
    return undefined;
  }
}

function prepareAttributesPayload(
  raw: string
): Record<string, unknown> | string | undefined {
  const trimmed = raw.trim();
  if (!trimmed) return undefined;
  try {
    return JSON.parse(trimmed);
  } catch {
    return trimmed;
  }
}

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
    categoryPath: "",
    imageUrl: "",
    stockCount: "",
    tagsCsv: "",
    description: "",
    attributesJson: "",
  });
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const [showSeedOptions, setShowSeedOptions] = useState(false);
  const [seedBrands, setSeedBrands] = useState("");
  const [seedCategories, setSeedCategories] = useState(
    ["Electronics > Audio > Headphones", "Fitness > Wearables > Smartwatch", "Wellness > Yoga > Accessories"].join(
      "\n"
    )
  );
  const [seedTags, setSeedTags] = useState("featured,seasonal,recommended");
  const [seedMinPrice, setSeedMinPrice] = useState("12");
  const [seedMaxPrice, setSeedMaxPrice] = useState("320");
  const [seedAttributes, setSeedAttributes] = useState(
    JSON.stringify(
      {
        audience: ["beginner", "enthusiast", "professional"],
        sustainability: ["recycled", "low_impact"],
      },
      null,
      2
    )
  );

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
      setItems(
        (data.items || []).map((item: Product) => ({
          ...item,
          imageUrl: item.imageUrl ?? "",
          tagsCsv: item.tagsCsv ?? "",
          description: item.description ?? "",
          attributesJson: item.attributesJson ?? "",
        }))
      );
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
          categoryPath: p.category || "",
          imageUrl: p.imageUrl || "",
          stockCount: String(p.stockCount ?? ""),
          tagsCsv: p.tagsCsv || "",
          description: p.description || "",
          attributesJson: formatAttributes(p.attributesJson),
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
        categoryPath: allEq((p) => p.category) || "",
        imageUrl: allEq((p) => p.imageUrl ?? "") || "",
        stockCount: allEq((p) => String(p.stockCount)) || "",
        tagsCsv: allEq((p) => p.tagsCsv ?? "") || "",
        description: allEq((p) => p.description ?? "") || "",
        attributesJson: "",
      });
    } else {
      setForm({
        name: "",
        sku: "",
        price: "",
        currency: "USD",
        brand: "",
        category: "",
        categoryPath: "",
        imageUrl: "",
        stockCount: "",
        tagsCsv: "",
        description: "",
        attributesJson: "",
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
    const payload: Record<string, unknown> = { count };
    const brands = parseCsvList(seedBrands);
    if (brands.length > 0) {
      payload.brands = brands;
    }
    const categoryLines = parseCategoryLines(seedCategories);
    if (categoryLines.length > 0) {
      payload.categories = categoryLines;
    }
    const tagList = parseCsvList(seedTags);
    if (tagList.length > 0) {
      payload.tags = tagList;
    }
    const priceRange: { min?: number; max?: number } = {};
    if (seedMinPrice.trim()) {
      const parsed = Number(seedMinPrice);
      if (Number.isNaN(parsed)) {
        toast("Invalid minimum price");
        return;
      }
      priceRange.min = parsed;
    }
    if (seedMaxPrice.trim()) {
      const parsed = Number(seedMaxPrice);
      if (Number.isNaN(parsed)) {
        toast("Invalid maximum price");
        return;
      }
      priceRange.max = parsed;
    }
    if (
      priceRange.min !== undefined &&
      priceRange.max !== undefined &&
      priceRange.min > priceRange.max
    ) {
      toast("Minimum price must be less than maximum price");
      return;
    }
    if (Object.keys(priceRange).length > 0) {
      payload.priceRange = priceRange;
    }
    if (seedAttributes.trim()) {
      const parsedAttributes = parseAttributesConfig(seedAttributes);
      if (!parsedAttributes) {
        toast("Seed attributes must be valid JSON (object of arrays).");
        return;
      }
      payload.attributes = parsedAttributes;
    }

    await fetch("/api/products/seed", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(payload),
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
      const payload: Record<string, unknown> = {
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
      if (form.categoryPath.trim()) {
        payload.categoryPath = form.categoryPath;
      }
      const attributesPayload = prepareAttributesPayload(form.attributesJson);
      if (attributesPayload !== undefined) {
        payload.attributes = attributesPayload;
        payload.attributesJson = form.attributesJson;
      } else if (form.attributesJson.trim() === "") {
        payload.attributesJson = "";
      }
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
        categoryPath: "",
        imageUrl: "",
        stockCount: "",
        tagsCsv: "",
        description: "",
        attributesJson: "",
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
    if (touched.categoryPath) data.categoryPath = form.categoryPath;
    if (touched.imageUrl) data.imageUrl = form.imageUrl;
    if (touched.stockCount)
      data.stockCount = form.stockCount ? parseInt(form.stockCount, 10) : 0;
    if (touched.tagsCsv) data.tagsCsv = form.tagsCsv;
    if (touched.description) data.description = form.description;
    if (touched.attributesJson) {
      const trimmed = form.attributesJson.trim();
      if (!trimmed) {
        data.attributesJson = "";
      } else {
        const attributesPayload = prepareAttributesPayload(form.attributesJson);
        if (attributesPayload !== undefined) {
          data.attributes = attributesPayload;
        }
        data.attributesJson = form.attributesJson;
      }
    }
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
            type="button"
            className="border rounded px-3 py-2 text-xs"
            onClick={() => setShowSeedOptions((prev) => !prev)}
          >
            {showSeedOptions ? "Hide seed options" : "Seed options"}
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

      {showSeedOptions ? (
        <div className="border rounded p-4 bg-gray-50 space-y-3">
          <h3 className="text-sm font-semibold text-gray-700">
            Seed generator options
          </h3>
          <div className="grid gap-3 md:grid-cols-2">
            <label className="text-xs text-gray-700">
              Brands (comma separated)
              <input
                className="mt-1 w-full border rounded p-2 text-sm"
                placeholder="Acme, Globex, Umbrella"
                value={seedBrands}
                onChange={(e) => setSeedBrands(e.target.value)}
              />
            </label>
            <label className="text-xs text-gray-700">
              Tags (comma separated)
              <input
                className="mt-1 w-full border rounded p-2 text-sm"
                placeholder="featured, seasonal, recommended"
                value={seedTags}
                onChange={(e) => setSeedTags(e.target.value)}
              />
            </label>
            <label className="text-xs text-gray-700 md:col-span-2">
              Category paths (one per line)
              <textarea
                className="mt-1 w-full border rounded p-2 text-sm h-24"
                placeholder="Electronics > Audio > Headphones"
                value={seedCategories}
                onChange={(e) => setSeedCategories(e.target.value)}
              />
            </label>
            <label className="text-xs text-gray-700">
              Minimum price
              <input
                className="mt-1 w-full border rounded p-2 text-sm"
                type="number"
                min={1}
                value={seedMinPrice}
                onChange={(e) => setSeedMinPrice(e.target.value)}
              />
            </label>
            <label className="text-xs text-gray-700">
              Maximum price
              <input
                className="mt-1 w-full border rounded p-2 text-sm"
                type="number"
                min={1}
                value={seedMaxPrice}
                onChange={(e) => setSeedMaxPrice(e.target.value)}
              />
            </label>
            <label className="text-xs text-gray-700 md:col-span-2">
              Attribute pools (JSON object of arrays)
              <textarea
                className="mt-1 w-full border rounded p-2 text-sm h-28 font-mono"
                value={seedAttributes}
                onChange={(e) => setSeedAttributes(e.target.value)}
              />
            </label>
          </div>
        </div>
      ) : null}

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
            className="border p-2 md:col-span-2"
            name="categoryPath"
            placeholder="Category path (e.g. Electronics > Audio > Headphones)"
            value={form.categoryPath}
            onChange={(e) => setField("categoryPath", e.target.value)}
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
            placeholder="Tags (comma separated)"
            value={form.tagsCsv}
            onChange={(e) => setField("tagsCsv", e.target.value)}
          />
          <textarea
            className="border p-2 md:col-span-3 font-mono text-xs"
            name="attributesJson"
            placeholder='{"color":"black","usage":"commute"}'
            value={form.attributesJson}
            onChange={(e) => setField("attributesJson", e.target.value)}
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
