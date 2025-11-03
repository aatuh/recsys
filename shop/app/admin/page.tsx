"use client";
import { useSearchParams, useRouter } from "next/navigation";
import { useMemo, useEffect } from "react";
import Link from "next/link";
import AdminProducts from "@/components/admin/AdminProducts";
import AdminUsers from "@/components/admin/AdminUsers";
import AdminEvents from "@/components/admin/AdminEvents";
import AdminCarts from "@/components/admin/AdminCarts";
import AdminOrders from "@/components/admin/AdminOrders";
import { AnalyticsDashboard } from "@/components/AnalyticsDashboard";
import { AdminTools } from "@/components/AdminTools";
import AdminRecommendationSettings from "@/components/admin/AdminRecommendationSettings";
import { AdminBanditPolicies } from "@/components/admin/AdminBanditPolicies";

const tabs = [
  "Products",
  "Users",
  "Events",
  "Carts",
  "Orders",
  "Recommendations",
  "Bandit",
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

const ADMIN_TAB_STORAGE_KEY = "admin_last_tab";

export default function AdminHub() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const tab = useMemo(
    () => getTabFromSlug(searchParams.get("tab")),
    [searchParams]
  );

  // On mount, if no tab in URL, redirect to last accessed tab
  useEffect(() => {
    if (!searchParams.get("tab")) {
      try {
        const lastTab = localStorage.getItem(ADMIN_TAB_STORAGE_KEY);
        if (lastTab) {
          const lastTabName = getTabFromSlug(lastTab);
          router.replace(`/admin?tab=${getTabSlug(lastTabName)}`);
        }
      } catch (err) {
        // Ignore localStorage errors
      }
    }
  }, [searchParams, router]);

  // Save current tab to localStorage when it changes
  useEffect(() => {
    try {
      localStorage.setItem(ADMIN_TAB_STORAGE_KEY, getTabSlug(tab));
    } catch (err) {
      // Ignore localStorage errors
    }
  }, [tab]);

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
      {tab === "Bandit" && <AdminBanditPolicies />}
      {tab === "Analytics" && <AnalyticsDashboard />}
      {tab === "Tools" && <AdminTools onAction={handleAdminAction} />}
    </main>
  );
}
