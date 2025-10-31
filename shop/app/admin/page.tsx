"use client";
import { useState } from "react";
import AdminProducts from "@/components/admin/AdminProducts";
import AdminUsers from "@/components/admin/AdminUsers";
import AdminEvents from "@/components/admin/AdminEvents";
import AdminCarts from "@/components/admin/AdminCarts";
import AdminOrders from "@/components/admin/AdminOrders";

export default function AdminHub() {
  const tabs = ["Products", "Users", "Events", "Carts", "Orders"] as const;
  const [tab, setTab] = useState<(typeof tabs)[number]>("Products");

  async function nukeAll() {
    if (!confirm("Delete ALL data (events, orders, carts, products, users)?"))
      return;
    await fetch("/api/admin/nuke", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        tables: ["events", "order", "product", "cart", "user"],
      }),
    });
    alert("All data deleted");
  }

  return (
    <main className="p-4 space-y-4">
      <div className="flex items-center gap-3">
        <h1 className="text-xl font-semibold">Admin</h1>
        <div className="ml-auto flex items-center gap-2">
          <button
            className="border rounded px-3 py-2 text-sm"
            onClick={nukeAll}
          >
            Nuke all data
          </button>
        </div>
      </div>
      <div className="flex gap-2 border-b pb-2">
        {tabs.map((t) => (
          <button
            key={t}
            className={`px-3 py-2 text-sm border rounded ${
              tab === t ? "bg-gray-100 dark:bg-gray-800" : ""
            }`}
            onClick={() => setTab(t)}
          >
            {t}
          </button>
        ))}
      </div>
      {tab === "Products" && <AdminProducts />}
      {tab === "Users" && <AdminUsers />}
      {tab === "Events" && <AdminEvents />}
      {tab === "Carts" && <AdminCarts />}
      {tab === "Orders" && <AdminOrders />}
    </main>
  );
}
