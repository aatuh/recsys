"use client";
import { useState } from "react";

export default function NewProductPage() {
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    const form = e.currentTarget as HTMLFormElement & {
      name: { value: string };
      sku: { value: string };
      price: { value: string };
      currency: { value: string };
      brand: { value: string };
      category: { value: string };
      imageUrl: { value: string };
      stockCount: { value: string };
      tagsCsv: { value: string };
      description: { value: string };
    };
    const data = {
      name: form.name.value,
      sku: form.sku.value,
      price: parseFloat(form.price.value || "0"),
      currency: form.currency.value || "USD",
      brand: form.brand.value,
      category: form.category.value,
      imageUrl: form.imageUrl.value,
      stockCount: parseInt(form.stockCount.value || "0", 10),
      tagsCsv: form.tagsCsv.value,
      description: form.description.value,
    };
    try {
      const res = await fetch("/api/products", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(data),
      });
      if (!res.ok) throw new Error("Failed to create");
      const created = await res.json();
      window.location.href = `/products/${encodeURIComponent(created.id)}`;
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="space-y-4">
      <h1 className="text-xl font-semibold">New product</h1>
      {error && <div className="text-red-600 text-sm">{error}</div>}
      <form
        onSubmit={onSubmit}
        className="grid grid-cols-1 md:grid-cols-2 gap-3"
      >
        <input className="border p-2" name="name" placeholder="Name" required />
        <input className="border p-2" name="sku" placeholder="SKU" required />
        <input
          className="border p-2"
          name="price"
          placeholder="Price"
          type="number"
          step="0.01"
        />
        <input
          className="border p-2"
          name="currency"
          placeholder="Currency"
          defaultValue="USD"
        />
        <input className="border p-2" name="brand" placeholder="Brand" />
        <input className="border p-2" name="category" placeholder="Category" />
        <input className="border p-2" name="imageUrl" placeholder="Image URL" />
        <input
          className="border p-2"
          name="stockCount"
          placeholder="Stock"
          type="number"
        />
        <input
          className="border p-2 md:col-span-2"
          name="tagsCsv"
          placeholder="Tags CSV (e.g. electronics,new)"
        />
        <textarea
          className="border p-2 md:col-span-2"
          name="description"
          placeholder="Description"
        />
        <div className="md:col-span-2">
          <button
            disabled={submitting}
            className="border rounded px-3 py-2 text-sm"
          >
            {submitting ? "Creating..." : "Create"}
          </button>
        </div>
      </form>
    </main>
  );
}
