import { z } from "zod";

type RawTags = string | string[] | undefined | null;
type RawCategoryPath = string | string[] | undefined | null;
type RawAttributes =
  | Record<string, unknown>
  | Array<unknown>
  | string
  | undefined
  | null;

const ProductCreateSchema = z.object({
  name: z.string().min(1),
  sku: z.string().min(1),
  price: z.coerce.number().min(0),
  currency: z.string().default("USD"),
  brand: z.string().optional().nullable(),
  category: z.string().optional().nullable(),
  categoryPath: z.union([z.string(), z.array(z.string())]).optional().nullable(),
  description: z.string().optional().nullable(),
  imageUrl: z.string().optional().nullable(),
  stockCount: z.coerce.number().int().min(0).default(0),
  tags: z.union([z.string(), z.array(z.string())]).optional().nullable(),
  tagsCsv: z.string().optional().nullable(),
  attributes: z
    .union([z.record(z.any()), z.array(z.any()), z.string()])
    .optional()
    .nullable(),
  attributesJson: z.string().optional().nullable(),
});

const ProductUpdateSchema = ProductCreateSchema.partial();

function toCategory(category: string | null | undefined): string | undefined {
  if (category === null || category === undefined) {
    return undefined;
  }
  const trimmed = category.trim();
  return trimmed.length > 0 ? trimmed : "";
}

function toCategoryFromPath(path: RawCategoryPath): string | undefined {
  if (path === null || path === undefined) {
    return undefined;
  }
  if (Array.isArray(path)) {
    return path.map((part) => part.trim()).filter(Boolean).join(" > ");
  }
  return path
    .split(">")
    .map((part) => part.trim())
    .filter(Boolean)
    .join(" > ");
}

function toTagsCsv(tags: RawTags, tagsCsv?: string | null | undefined): string {
  if (typeof tagsCsv === "string" && tagsCsv.trim().length > 0) {
    return tagsCsv;
  }
  if (tags === null || tags === undefined) {
    return "";
  }
  if (Array.isArray(tags)) {
    return tags
      .map((tag) => tag.trim())
      .filter((tag) => tag.length > 0)
      .join(",");
  }
  return tags
    .split(",")
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0)
    .join(",");
}

function toAttributesJson(
  attributes: RawAttributes,
  fallback?: string | null | undefined
): string {
  if (typeof fallback === "string" && fallback.trim().length > 0) {
    try {
      JSON.parse(fallback);
      return fallback;
    } catch {
      // fall through
    }
  }
  if (attributes === null || attributes === undefined) {
    return "{}";
  }
  if (typeof attributes === "string") {
    const trimmed = attributes.trim();
    if (!trimmed) {
      return "{}";
    }
    try {
      const parsed = JSON.parse(trimmed);
      return JSON.stringify(parsed);
    } catch {
      return JSON.stringify({ notes: trimmed });
    }
  }
  if (Array.isArray(attributes)) {
    return JSON.stringify({ list: attributes });
  }
  return JSON.stringify(attributes);
}

function sanitizeString(value: string | null | undefined): string {
  if (value === null || value === undefined) {
    return "";
  }
  return value.trim();
}

export type NormalizedProductInput = {
  name: string;
  sku: string;
  description: string;
  price: number;
  currency: string;
  brand: string;
  category: string;
  imageUrl: string;
  stockCount: number;
  tagsCsv: string;
  attributesJson: string;
};

export function normalizeProductPayload(payload: unknown): NormalizedProductInput {
  const parsed = ProductCreateSchema.parse(payload ?? {});
  const categoryFromPath = toCategoryFromPath(parsed.categoryPath);
  const resolvedCategory =
    categoryFromPath ??
    toCategory(parsed.category)?.trim() ??
    (parsed.category === undefined ? undefined : sanitizeString(parsed.category ?? ""));
  return {
    name: sanitizeString(parsed.name),
    sku: sanitizeString(parsed.sku),
    description: sanitizeString(parsed.description ?? ""),
    price: parsed.price,
    currency: sanitizeString(parsed.currency ?? "USD") || "USD",
    brand: sanitizeString(parsed.brand ?? ""),
    category: resolvedCategory && resolvedCategory.length > 0 ? resolvedCategory : "General",
    imageUrl: sanitizeString(parsed.imageUrl ?? ""),
    stockCount: Number.isFinite(parsed.stockCount) ? parsed.stockCount : 0,
    tagsCsv: toTagsCsv(parsed.tags ?? undefined, parsed.tagsCsv),
    attributesJson: toAttributesJson(parsed.attributes, parsed.attributesJson),
  };
}

export function normalizeProductPatch(
  payload: unknown
): Partial<NormalizedProductInput> {
  const parsed = ProductUpdateSchema.parse(payload ?? {});
  const update: Partial<NormalizedProductInput> = {};
  if (parsed.name !== undefined) update.name = sanitizeString(parsed.name);
  if (parsed.sku !== undefined) update.sku = sanitizeString(parsed.sku);
  if (parsed.description !== undefined)
    update.description = sanitizeString(parsed.description);
  if (parsed.price !== undefined) update.price = parsed.price;
  if (parsed.currency !== undefined)
    update.currency = sanitizeString(parsed.currency) || "USD";
  if (parsed.brand !== undefined) update.brand = sanitizeString(parsed.brand);

  if (parsed.categoryPath !== undefined) {
    const category = toCategoryFromPath(parsed.categoryPath);
    update.category = category && category.length > 0 ? category : "General";
  } else if (parsed.category !== undefined) {
    const category = toCategory(parsed.category);
    update.category = category && category.length > 0 ? category : "General";
  }

  if (parsed.imageUrl !== undefined) {
    update.imageUrl = sanitizeString(parsed.imageUrl);
  }
  if (parsed.stockCount !== undefined) {
    update.stockCount = Number.isFinite(parsed.stockCount)
      ? parsed.stockCount
      : 0;
  }
  if (parsed.tags !== undefined || parsed.tagsCsv !== undefined) {
    update.tagsCsv = toTagsCsv(parsed.tags, parsed.tagsCsv);
  }
  if (parsed.attributes !== undefined || parsed.attributesJson !== undefined) {
    update.attributesJson = toAttributesJson(
      parsed.attributes,
      parsed.attributesJson
    );
  }
  return update;
}
