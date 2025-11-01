"use client";
import { useState } from "react";
import AdminProducts from "@/components/admin/AdminProducts";
import AdminUsers from "@/components/admin/AdminUsers";
import AdminEvents from "@/components/admin/AdminEvents";
import AdminCarts from "@/components/admin/AdminCarts";
import AdminOrders from "@/components/admin/AdminOrders";
import { AnalyticsDashboard } from "@/components/AnalyticsDashboard";
import { AdminTools } from "@/components/AdminTools";

export default function AdminHub() {
  const tabs = [
    "Products",
    "Users",
    "Events",
    "Carts",
    "Orders",
    "Analytics",
    "Tools",
  ] as const;
  const [tab, setTab] = useState<(typeof tabs)[number]>("Products");

  const handleAdminAction = async (
    action: string,
    params?: Record<string, unknown>
  ) => {
    try {
      const response = await fetch("/api/admin/tools", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ action, params }),
      });

      if (!response.ok) {
        throw new Error(`Action failed: ${response.status}`);
      }

      const result = await response.json();
      alert(`Action completed: ${JSON.stringify(result)}`);
    } catch (error) {
      console.error("Admin action failed:", error);
      alert(
        `Action failed: ${
          error instanceof Error ? error.message : "Unknown error"
        }`
      );
    }
  };

  return (
    <main className="p-4 space-y-4">
      <div className="flex items-center gap-3">
        <h1 className="text-xl font-semibold">Admin Hub</h1>
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
      {tab === "Analytics" && <AnalyticsDashboard />}
      {tab === "Tools" && <AdminTools onAction={handleAdminAction} />}
    </main>
  );
}
