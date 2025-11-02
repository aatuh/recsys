import { prisma } from "@/server/db/client";

const SEEN_EVENT_TYPES = ["view", "click", "add", "purchase"] as const;
const SEEN_LOOKBACK_DAYS = 30;
const SEEN_EVENT_LIMIT = 500;

export type SeenItemsConfig = {
  userId: string;
  lookbackDays?: number;
  limit?: number;
};

export async function getSeenItemIds({
  userId,
  lookbackDays = SEEN_LOOKBACK_DAYS,
  limit = SEEN_EVENT_LIMIT,
}: SeenItemsConfig): Promise<Set<string>> {
  const since = new Date(Date.now() - lookbackDays * 24 * 60 * 60 * 1000);

  const events = await prisma.event.findMany({
    where: {
      userId,
      productId: { not: null },
      type: { in: [...SEEN_EVENT_TYPES] },
      ts: { gte: since },
    },
    select: { productId: true },
    orderBy: { ts: "desc" },
    take: limit,
  });

  return new Set(
    events
      .map((event) => event.productId)
      .filter((productId): productId is string => Boolean(productId))
  );
}

export function summarizeExclusion(
  seen: Set<string>,
  excluded: number
): string {
  return `excluded ${excluded} seen items (cache_size=${seen.size})`;
}
