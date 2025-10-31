import { prisma } from "@/server/db/client";

export type UserFilter = {
  q?: string;
  limit?: number;
  offset?: number;
};

export const UserRepository = {
  async list(filter: UserFilter = {}) {
    const where: any = {};
    if (filter.q)
      where.displayName = { contains: filter.q, mode: "insensitive" };
    const limit = filter.limit ?? 20;
    const offset = filter.offset ?? 0;
    const [items, total] = await Promise.all([
      prisma.user.findMany({
        where,
        skip: offset,
        take: limit,
        orderBy: { createdAt: "desc" },
      }),
      prisma.user.count({ where }),
    ]);
    return { items, total, limit, offset };
  },

  async getById(id: string) {
    return prisma.user.findUnique({ where: { id } });
  },

  async create(data: any) {
    return prisma.user.create({ data });
  },

  async update(id: string, data: any) {
    return prisma.user.update({ where: { id }, data });
  },

  async remove(id: string) {
    return prisma.user.delete({ where: { id } });
  },
};
