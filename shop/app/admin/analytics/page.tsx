"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { AnalyticsDashboard } from "@/components/AnalyticsDashboard";

export default function AdminAnalyticsPage() {
  const router = useRouter();

  useEffect(() => {
    router.replace("/admin?tab=analytics");
  }, [router]);

  return (
    <main className="p-4 space-y-4">
      <h1 className="text-xl font-semibold">Admin · Analytics</h1>
      <AnalyticsDashboard />
    </main>
  );
}
