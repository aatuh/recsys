import { NextResponse } from "next/server";
import { prisma } from "@/server/db/client";

export async function GET() {
  try {
    // Get event counts by type
    const eventCounts = await prisma.event.groupBy({
      by: ["type"],
      _count: {
        type: true,
      },
    });

    const eventCountsMap: Record<string, number> = {};
    eventCounts.forEach(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      ({ type, _count }: any) => {
        const typeName =
          ["view", "click", "add", "purchase", "custom"][type] || "unknown";
        eventCountsMap[typeName] = _count.type;
      }
    );

    // Calculate funnel metrics
    const views = eventCountsMap.view || 0;
    const clicks = eventCountsMap.click || 0;
    const adds = eventCountsMap.add || 0;
    const purchases = eventCountsMap.purchase || 0;

    const ctr = views > 0 ? clicks / views : 0;
    const atcRate = clicks > 0 ? adds / clicks : 0;
    const conversionRate = adds > 0 ? purchases / adds : 0;

    // Get recommendation effectiveness
    const recommendedEvents = await prisma.event.findMany({
      where: {
        metaText: {
          contains: '"recommended":true',
        },
      },
    });

    const recommendedClicks = recommendedEvents.filter(
      (e: { type: string }) => e.type === "click"
    ).length;
    const totalClicks = clicks;
    const recommendationCtr =
      totalClicks > 0 ? recommendedClicks / totalClicks : 0;

    // Get top products by performance
    const productStats = await prisma.event.groupBy({
      by: ["productId"],
      where: {
        productId: {
          not: null,
        },
      },
      _count: {
        productId: true,
      },
    });

    const topProducts = await Promise.all(
      productStats
        .slice(0, 10)
        .map(async (stat: { productId: string | null }) => {
          const productId = stat.productId!;

          const [views, clicks, purchases] = await Promise.all([
            prisma.event.count({
              where: { productId, type: "view" },
            }),
            prisma.event.count({
              where: { productId, type: "click" },
            }),
            prisma.event.count({
              where: { productId, type: "purchase" },
            }),
          ]);

          return {
            productId,
            views,
            clicks,
            purchases,
          };
        })
    );

    const analyticsData = {
      eventCounts: eventCountsMap,
      funnelMetrics: {
        views,
        clicks,
        adds,
        purchases,
        ctr,
        atcRate,
        conversionRate,
      },
      recommendationEffectiveness: {
        totalClicks,
        recommendedClicks,
        recommendationCtr,
      },
      topProducts,
    };

    return NextResponse.json(analyticsData);
  } catch (error) {
    console.error("Analytics API error:", error);
    return NextResponse.json(
      { error: "Failed to fetch analytics data" },
      { status: 500 }
    );
  }
}
