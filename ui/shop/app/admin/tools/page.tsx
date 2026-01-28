"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { AdminTools } from "@/components/AdminTools";

export default function AdminToolsPage() {
  const router = useRouter();

  useEffect(() => {
    router.replace("/admin?tab=tools");
  }, [router]);

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
      <h1 className="text-xl font-semibold">Admin · Tools</h1>
      <AdminTools onAction={handleAdminAction} />
    </main>
  );
}
