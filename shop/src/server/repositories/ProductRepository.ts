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
    const where: Record<string, unknown> = {};
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
      where.price = {} as Record<string, unknown>;
      if (filter.minPrice !== undefined)
        (where.price as Record<string, unknown>).gte = filter.minPrice;
      if (filter.maxPrice !== undefined)
        (where.price as Record<string, unknown>).lte = filter.maxPrice;
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

  async create(data: {
    sku: string;
    name: string;
    description: string;
    price: number;
    currency: string;
    stockCount: number;
    brand?: string | null;
    category?: string | null;
    imageUrl?: string | null;
    tagsCsv?: string | null;
  }) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return prisma.product.create({ data: data as any });
  },

  async update(
    id: string,
    data: Partial<{
      sku: string;
      name: string;
      description: string;
      price: number;
      currency: string;
      stockCount: number;
      brand: string | null;
      category: string | null;
      imageUrl: string | null;
      tagsCsv: string | null;
    }>
  ) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return prisma.product.update({ where: { id }, data: data as any });
  },

  async remove(id: string) {
    return prisma.product.delete({ where: { id } });
  },
};
