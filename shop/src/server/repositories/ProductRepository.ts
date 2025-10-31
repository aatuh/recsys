import { prisma } from "@/server/db/client";

export type ProductFilter = {
  q?: string;
  brand?: string;
  category?: string;
  minPrice?: number;
  maxPrice?: number;
  available?: boolean;
  limit?: number;
  offset?: number;
};

export const ProductRepository = {
  async list(filter: ProductFilter = {}) {
    const where: any = {};
    if (filter.q) {
      where.OR = [
        { name: { contains: filter.q, mode: "insensitive" } },
        { description: { contains: filter.q, mode: "insensitive" } },
      ];
    }
    if (filter.brand) where.brand = filter.brand;
    if (filter.category) where.category = filter.category;
    if (filter.available !== undefined) {
      where.stockCount = filter.available ? { gt: 0 } : 0;
    }
    if (filter.minPrice !== undefined || filter.maxPrice !== undefined) {
      where.price = {};
      if (filter.minPrice !== undefined) where.price.gte = filter.minPrice;
      if (filter.maxPrice !== undefined) where.price.lte = filter.maxPrice;
    }
    const limit = filter.limit ?? 20;
    const offset = filter.offset ?? 0;
    const [items, total] = await Promise.all([
      prisma.product.findMany({
        where,
        skip: offset,
        take: limit,
        orderBy: { createdAt: "desc" },
      }),
      prisma.product.count({ where }),
    ]);
    return { items, total, limit, offset };
  },

  async getById(id: string) {
    return prisma.product.findUnique({ where: { id } });
  },

  async create(data: any) {
    return prisma.product.create({ data });
  },

  async update(id: string, data: any) {
    return prisma.product.update({ where: { id }, data });
  },

  async remove(id: string) {
    return prisma.product.delete({ where: { id } });
  },
};
