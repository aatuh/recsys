"use client";
import { useSearchParams } from "next/navigation";
import { useMemo } from "react";
import Link from "next/link";
import AdminProducts from "@/components/admin/AdminProducts";
import AdminUsers from "@/components/admin/AdminUsers";
import AdminEvents from "@/components/admin/AdminEvents";
import AdminCarts from "@/components/admin/AdminCarts";
import AdminOrders from "@/components/admin/AdminOrders";
import { AnalyticsDashboard } from "@/components/AnalyticsDashboard";
import { AdminTools } from "@/components/AdminTools";
import AdminRecommendationSettings from "@/components/admin/AdminRecommendationSettings";

const tabs = [
  "Products",
  "Users",
  "Events",
  "Carts",
  "Orders",
  "Recommendations",
  "Analytics",
  "Tools",
] as const;

type TabName = (typeof tabs)[number];

function getTabSlug(tab: TabName): string {
  return tab.toLowerCase();
}

function getTabFromSlug(slug: string | null): TabName {
  const normalized = slug?.toLowerCase();
  const found = tabs.find((t) => getTabSlug(t) === normalized);
  return found ?? "Products";
}

export default function AdminHub() {
  const searchParams = useSearchParams();
  const tab = useMemo(
    () => getTabFromSlug(searchParams.get("tab")),
    [searchParams]
  );

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
          <Link
            key={t}
            href={`/admin?tab=${getTabSlug(t)}`}
            className={`px-3 py-2 text-sm border rounded ${
              tab === t ? "bg-gray-100 dark:bg-gray-800" : ""
            }`}
          >
            {t}
          </Link>
        ))}
      </div>

      {tab === "Products" && <AdminProducts />}
      {tab === "Users" && <AdminUsers />}
      {tab === "Events" && <AdminEvents />}
      {tab === "Carts" && <AdminCarts />}
      {tab === "Orders" && <AdminOrders />}
      {tab === "Recommendations" && <AdminRecommendationSettings />}
      {tab === "Analytics" && <AnalyticsDashboard />}
      {tab === "Tools" && <AdminTools onAction={handleAdminAction} />}
    </main>
  );
}
