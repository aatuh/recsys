import { prisma } from "@/server/db/client";

const DEFAULT_MAX_AGE_DAYS = Number(
  process.env.SHOP_FRESH_ITEM_MAX_AGE_DAYS ?? "7"
);
const DEFAULT_LIMIT = Number(process.env.SHOP_FRESH_ITEM_LIMIT ?? "20");

export type FreshItemOptions = {
  maxAgeDays?: number;
  limit?: number;
  excludeIds?: Set<string> | string[];
};

export async function getFreshItemIds(
  options: FreshItemOptions = {}
): Promise<string[]> {
  const maxAgeDays = options.maxAgeDays ?? DEFAULT_MAX_AGE_DAYS;
  const limit = options.limit ?? DEFAULT_LIMIT;
  const exclude =
    options.excludeIds instanceof Set
      ? options.excludeIds
      : new Set(options.excludeIds ?? []);

  if (limit <= 0) {
    return [];
  }

  const since = new Date(
    Date.now() - maxAgeDays * 24 * 60 * 60 * 1000
  );

  const freshest = await prisma.product.findMany({
    where: {
      createdAt: { gte: since },
      stockCount: { gt: 0 },
      id: { notIn: Array.from(exclude) },
    },
    orderBy: { createdAt: "desc" },
    select: { id: true },
    take: limit,
  });

  return freshest.map((product) => product.id);
}
