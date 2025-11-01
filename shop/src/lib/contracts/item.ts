import { z } from "zod";

// Item contract for Recsys API
export const ItemContract = z.object({
  item_id: z.string(),
  available: z.boolean(),
  price: z.number().optional(),
  tags: z.array(z.string()).optional(),
  props: z.object({
    name: z.string(),
    sku: z.string().optional(),
    url: z.string().optional(),
    image_url: z.string().optional(),
    brand: z.string().optional(),
    category: z.string().optional(),
    currency: z.string().optional(),
    description: z.string().optional(),
    attributes: z.record(z.string()).optional(),
  }).optional(),
  embedding: z.array(z.number()).optional(),
});

export type ItemContract = z.infer<typeof ItemContract>;

// Tag conventions
export const TAG_CONVENTIONS = {
  BRAND_PREFIX: "brand:",
  CATEGORY_PREFIX: "category:",
  CAT_PREFIX: "cat:",
} as const;

export function buildItemTags(product: {
  brand?: string | null;
  category?: string | null;
  tagsCsv?: string | null;
}): string[] {
  const tags: string[] = [];
  
  if (product.brand) {
    tags.push(`${TAG_CONVENTIONS.BRAND_PREFIX}${product.brand}`);
  }
  
  if (product.category) {
    tags.push(`${TAG_CONVENTIONS.CATEGORY_PREFIX}${product.category}`);
    // Add lowercase category alias
    tags.push(`${TAG_CONVENTIONS.CAT_PREFIX}${product.category.toLowerCase()}`);
  }
  
  // Add CSV tags as facets
  if (product.tagsCsv) {
    const csvTags = product.tagsCsv.split(",").map(s => s.trim()).filter(Boolean);
    csvTags.forEach(tag => {
      if (!tag.includes(":")) {
        // Simple facet tags
        tags.push(tag.toLowerCase());
      } else {
        tags.push(tag);
      }
    });
  }
  
  return tags;
}

export function buildItemContract(product: {
  id: string;
  name: string;
  sku: string;
  price: number;
  currency: string;
  brand?: string | null;
  category?: string | null;
  description?: string | null;
  imageUrl?: string | null;
  stockCount: number;
  tagsCsv?: string | null;
}): ItemContract {
  return {
    item_id: product.id,
    available: product.stockCount > 0,
    price: product.price,
    tags: buildItemTags(product),
    props: {
      name: product.name,
      sku: product.sku,
      url: `/products/${product.id}`,
      image_url: product.imageUrl || undefined,
      brand: product.brand || undefined,
      category: product.category || undefined,
      currency: product.currency,
      description: product.description || undefined,
    },
  };
}
