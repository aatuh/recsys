import { prisma } from "@/server/db/client";

export type EventFilter = {
  type?: string;
  userId?: string;
  productId?: string;
  limit?: number;
  offset?: number;
};

export const EventRepository = {
  async list(filter: EventFilter = {}) {
    const where: Record<string, unknown> = {};
    if (filter.type) where.type = filter.type;
    if (filter.userId) where.userId = filter.userId;
    if (filter.productId) where.productId = filter.productId;
    const limit = filter.limit ?? 50;
    const offset = filter.offset ?? 0;
    const [items, total] = await Promise.all([
      prisma.event.findMany({
        where,
        orderBy: { ts: "desc" },
        skip: offset,
        take: limit,
      }),
      prisma.event.count({ where }),
    ]);
    return { items, total, limit, offset };
  },

  async createBatch(
    events: Array<{
      type: string;
      userId: string;
      productId?: string | null;
      value: number;
      ts: Date;
      metaText?: string | null;
    }>
  ) {
    return prisma.$transaction(
      events.map((e) => prisma.event.create({ data: e }))
    );
  },

  async mark(ids: string[], status: "pending" | "sent" | "failed") {
    await prisma.event.updateMany({
      where: { id: { in: ids } },
      data: {
        recsysStatus: status,
        sentAt: status === "sent" ? new Date() : null,
      },
    });
  },
};
