import { prisma } from "@/server/db/client";

export type RankedItem = {
  item_id: string;
  score: number;
  reasons?: string[];
};

export type DiversityOptions = {
  brandCap?: number;
  categoryCap?: number;
};

const DEFAULT_BRAND_CAP = 2;
const DEFAULT_CATEGORY_CAP = 3;

export async function applyDiversityCaps(
  items: RankedItem[],
  options: DiversityOptions = {}
): Promise<{ kept: RankedItem[]; excluded: Record<string, number> }> {
  const brandCap = options.brandCap ?? DEFAULT_BRAND_CAP;
  const categoryCap = options.categoryCap ?? DEFAULT_CATEGORY_CAP;

  if (items.length === 0) {
    return { kept: items, excluded: {} };
  }

  const productMetadata = await prisma.product.findMany({
    where: { id: { in: items.map((item) => item.item_id) } },
    select: { id: true, brand: true, category: true },
  });

  const productById = new Map(
    productMetadata.map((product) => [product.id, product])
  );

  const brandCounts = new Map<string, number>();
  const categoryCounts = new Map<string, number>();
  const excluded: Record<string, number> = {};

  const kept: RankedItem[] = [];

  for (const item of items) {
    const meta = productById.get(item.item_id);
    if (!meta) {
      kept.push(item);
      continue;
    }

    const brand = meta.brand ?? null;
    const category = meta.category ?? null;

    const brandCount: number | undefined = brand
      ? brandCounts.get(brand)
      : undefined;
    const categoryCount: number | undefined = category
      ? categoryCounts.get(category)
      : undefined;

    const wouldExceedBrand =
      brandCap > 0 && brand && (brandCount ?? 0) >= brandCap;
    const wouldExceedCategory =
      categoryCap > 0 && category && (categoryCount ?? 0) >= categoryCap;

    if (wouldExceedBrand) {
      excluded.brand = (excluded.brand ?? 0) + 1;
      continue;
    }

    if (wouldExceedCategory) {
      excluded.category = (excluded.category ?? 0) + 1;
      continue;
    }

    if (brand) {
      brandCounts.set(brand, (brandCount ?? 0) + 1);
    }
    if (category) {
      categoryCounts.set(category, (categoryCount ?? 0) + 1);
    }
    kept.push(item);
  }

  return { kept, excluded };
}
