"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import AdminRecommendationSettings from "@/components/admin/AdminRecommendationSettings";

export default function AdminRecommendationsPage() {
  const router = useRouter();

  useEffect(() => {
    router.replace("/admin?tab=recommendations");
  }, [router]);

  return (
    <main className="p-4 space-y-4">
      <h1 className="text-xl font-semibold">Admin · Recommendations</h1>
      <AdminRecommendationSettings />
    </main>
  );
}
