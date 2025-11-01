"use client";
import { useEffect, useState } from "react";

interface AnalyticsData {
  eventCounts: Record<string, number>;
  funnelMetrics: {
    views: number;
    clicks: number;
    adds: number;
    purchases: number;
    ctr: number;
    atcRate: number;
    conversionRate: number;
  };
  recommendationEffectiveness: {
    totalClicks: number;
    recommendedClicks: number;
    recommendationCtr: number;
  };
  topProducts: Array<{
    productId: string;
    views: number;
    clicks: number;
    purchases: number;
  }>;
}

export function AnalyticsDashboard() {
  const [data, setData] = useState<AnalyticsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchAnalytics() {
      try {
        setLoading(true);
        setError(null);
        
        const response = await fetch("/api/admin/analytics");
        if (!response.ok) {
          throw new Error(`Failed to fetch analytics: ${response.status}`);
        }
        
        const analyticsData = await response.json();
        setData(analyticsData);
      } catch (err) {
        console.error("Failed to fetch analytics:", err);
        setError(err instanceof Error ? err.message : "Unknown error");
      } finally {
        setLoading(false);
      }
    }

    fetchAnalytics();
  }, []);

  if (loading) {
    return (
      <div className="space-y-4">
        <h2 className="text-xl font-semibold">Analytics Dashboard</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="border rounded p-4 animate-pulse">
              <div className="h-4 bg-gray-200 rounded mb-2"></div>
              <div className="h-8 bg-gray-200 rounded"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-4">
        <h2 className="text-xl font-semibold">Analytics Dashboard</h2>
        <div className="text-red-600">Error: {error}</div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="space-y-4">
        <h2 className="text-xl font-semibold">Analytics Dashboard</h2>
        <div className="text-gray-500">No analytics data available</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">Analytics Dashboard</h2>
      
      {/* Event Counts */}
      <div>
        <h3 className="text-lg font-medium mb-3">Event Volume</h3>
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
          {Object.entries(data.eventCounts).map(([type, count]) => (
            <div key={type} className="border rounded p-3 text-center">
              <div className="text-sm text-gray-600 capitalize">{type}</div>
              <div className="text-2xl font-bold">{count.toLocaleString()}</div>
            </div>
          ))}
        </div>
      </div>

      {/* Funnel Metrics */}
      <div>
        <h3 className="text-lg font-medium mb-3">Funnel Metrics</h3>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="border rounded p-3 text-center">
            <div className="text-sm text-gray-600">CTR</div>
            <div className="text-2xl font-bold">{(data.funnelMetrics.ctr * 100).toFixed(2)}%</div>
          </div>
          <div className="border rounded p-3 text-center">
            <div className="text-sm text-gray-600">Add-to-Cart Rate</div>
            <div className="text-2xl font-bold">{(data.funnelMetrics.atcRate * 100).toFixed(2)}%</div>
          </div>
          <div className="border rounded p-3 text-center">
            <div className="text-sm text-gray-600">Conversion Rate</div>
            <div className="text-2xl font-bold">{(data.funnelMetrics.conversionRate * 100).toFixed(2)}%</div>
          </div>
          <div className="border rounded p-3 text-center">
            <div className="text-sm text-gray-600">Total Revenue</div>
            <div className="text-2xl font-bold">${(data.funnelMetrics.purchases * 50).toFixed(0)}</div>
          </div>
        </div>
      </div>

      {/* Recommendation Effectiveness */}
      <div>
        <h3 className="text-lg font-medium mb-3">Recommendation Effectiveness</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="border rounded p-3 text-center">
            <div className="text-sm text-gray-600">Recommendation CTR</div>
            <div className="text-2xl font-bold">{(data.recommendationEffectiveness.recommendationCtr * 100).toFixed(2)}%</div>
          </div>
          <div className="border rounded p-3 text-center">
            <div className="text-sm text-gray-600">Recommended Clicks</div>
            <div className="text-2xl font-bold">{data.recommendationEffectiveness.recommendedClicks}</div>
          </div>
          <div className="border rounded p-3 text-center">
            <div className="text-sm text-gray-600">Total Clicks</div>
            <div className="text-2xl font-bold">{data.recommendationEffectiveness.totalClicks}</div>
          </div>
        </div>
      </div>

      {/* Top Products */}
      <div>
        <h3 className="text-lg font-medium mb-3">Top Performing Products</h3>
        <div className="border rounded overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-2 text-left text-sm font-medium">Product ID</th>
                <th className="px-4 py-2 text-left text-sm font-medium">Views</th>
                <th className="px-4 py-2 text-left text-sm font-medium">Clicks</th>
                <th className="px-4 py-2 text-left text-sm font-medium">Purchases</th>
                <th className="px-4 py-2 text-left text-sm font-medium">CTR</th>
              </tr>
            </thead>
            <tbody>
              {data.topProducts.map((product) => (
                <tr key={product.productId} className="border-t">
                  <td className="px-4 py-2 text-sm">{product.productId}</td>
                  <td className="px-4 py-2 text-sm">{product.views}</td>
                  <td className="px-4 py-2 text-sm">{product.clicks}</td>
                  <td className="px-4 py-2 text-sm">{product.purchases}</td>
                  <td className="px-4 py-2 text-sm">
                    {product.views > 0 ? ((product.clicks / product.views) * 100).toFixed(2) : 0}%
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
