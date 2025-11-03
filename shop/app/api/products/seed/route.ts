import { NextRequest, NextResponse } from "next/server";
import { randomUUID } from "crypto";
import { prisma } from "@/server/db/client";
import { upsertItems } from "@/server/services/recsys";
import { buildItemContract } from "@/lib/contracts/item";

type CategoryTemplate = {
  path: string[];
  tags: string[];
  attributes: Record<string, string[]>;
};

type SeedPayload = {
  count?: number;
  brands?: string[];
  categories?: Array<
    | string
    | {
        path?: string | string[];
        tags?: string[];
        attributes?: Record<string, string[]>;
      }
  >;
  priceRange?: { min?: number; max?: number };
  attributes?: Record<string, string[]>;
  tags?: string[];
};

const DEFAULT_BRANDS = [
  "Acme",
  "Globex",
  "Umbrella",
  "Stark",
  "Wayne",
  "InGen",
  "Tyrell",
  "Cyberdyne",
];

const DEFAULT_CATEGORY_TEMPLATES: CategoryTemplate[] = [
  {
    path: ["Electronics", "Audio", "Headphones"],
    tags: ["electronics", "audio", "headphones", "wireless"],
    attributes: {
      color: ["Black", "Silver", "White", "Space Gray"],
      form_factor: ["over_ear", "in_ear", "open_back"],
      connectivity: ["bluetooth", "wired", "bluetooth 5.3"],
      battery_life_hours: ["18", "24", "30"],
      noise_control: ["anc", "passive"],
    },
  },
  {
    path: ["Fitness", "Wearables", "Smartwatch"],
    tags: ["fitness", "wearable", "health", "smartwatch"],
    attributes: {
      strap_material: ["silicone", "leather", "nylon"],
      water_resistance: ["5ATM", "10ATM"],
      sensors: ["heart_rate", "spo2", "gps"],
      battery_life_days: ["3", "5", "7"],
      lifestyle: ["outdoor", "wellness", "performance"],
    },
  },
  {
    path: ["Home", "Kitchen", "Appliances"],
    tags: ["home", "kitchen", "appliance", "smart_home"],
    attributes: {
      power: ["800W", "1000W", "1200W"],
      material: ["stainless_steel", "ceramic", "glass"],
      warranty_years: ["1", "2"],
      energy_rating: ["A+", "A++"],
      automation: ["app_controlled", "voice_controlled"],
    },
  },
  {
    path: ["Outdoors", "Adventure", "Backpacks"],
    tags: ["outdoors", "adventure", "backpack", "gear"],
    attributes: {
      capacity_liters: ["18", "22", "28"],
      waterproof: ["true", "false"],
      frame: ["internal", "frameless"],
      season: ["all_season", "summer"],
      gender: ["unisex", "women"],
    },
  },
  {
    path: ["Wellness", "Yoga", "Accessories"],
    tags: ["wellness", "yoga", "accessory", "mindfulness"],
    attributes: {
      material: ["natural_rubber", "cork", "microfiber"],
      thickness_mm: ["3", "5", "8"],
      eco_friendly: ["true", "false"],
      grip_rating: ["premium", "standard"],
      weight_kg: ["1.1", "1.5", "1.8"],
    },
  },
];

const DEFAULT_ATTRIBUTE_POOL: Record<string, string[]> = {
  audience: ["beginner", "enthusiast", "professional"],
  release_year: ["2024", "2025"],
  sustainability: ["recycled", "carbon_neutral", "low_impact"],
  subscription_eligible: ["true", "false"],
};

const ADJECTIVES = [
  "Advanced",
  "Bold",
  "Chic",
  "Compact",
  "Deluxe",
  "Eco",
  "Elegant",
  "Hyper",
  "Innovative",
  "Minimalist",
  "Modular",
  "NextGen",
  "Premium",
  "Refined",
  "Robust",
  "Smart",
  "Sophisticated",
  "Ultralight",
  "Versatile",
  "Wireless",
];

const NOUNS = [
  "Adapter",
  "Camera",
  "Console",
  "Controller",
  "Drone",
  "Earbuds",
  "Fitness Tracker",
  "Headphones",
  "Monitor",
  "Notebook",
  "Speaker",
  "Smartwatch",
  "Thermostat",
  "Tripod",
  "VR Headset",
  "Yoga Mat",
];

export async function POST(req: NextRequest) {
  const payload = (await req.json().catch(() => ({}))) as SeedPayload;
  const {
    count = 50,
    brands,
    categories,
    priceRange,
    attributes,
    tags,
  } = payload;

  const brandPool =
    Array.isArray(brands) && brands.length > 0 ? brands : DEFAULT_BRANDS;
  const categoryTemplates = resolveCategoryTemplates(categories);
  const attributePool = { ...DEFAULT_ATTRIBUTE_POOL, ...(attributes ?? {}) };
  const globalTags =
    Array.isArray(tags) && tags.length > 0
      ? tags.map((tag) => tag.toLowerCase())
      : ["featured", "seasonal", "recommended"];
  const priceMin =
    typeof priceRange?.min === "number" ? Math.max(priceRange.min, 5) : 12;
  const priceMax =
    typeof priceRange?.max === "number"
      ? Math.max(priceRange.max, priceMin + 5)
      : 320;

  const now = Date.now();
  const productData = Array.from({ length: count }).map((_, index) => {
    const template =
      categoryTemplates[index % categoryTemplates.length] ??
      DEFAULT_CATEGORY_TEMPLATES[0];
    const brand = brandPool[index % brandPool.length];
    const adjective = ADJECTIVES[index % ADJECTIVES.length];
    const noun = NOUNS[index % NOUNS.length];
    const categoryPath = template.path.join(" > ");
    const localAttributes = buildAttributes(template.attributes, attributePool);
    const mergedAttributes = {
      brand,
      category: template.path[template.path.length - 1],
      ...localAttributes,
    };
    const attributeTags = Object.entries(mergedAttributes)
      .filter(([, value]) => value !== undefined && value !== null)
      .map(([key, value]) => `${key}:${String(value).toLowerCase()}`);
    const tagsCsv = [
      ...new Set([
        ...template.tags,
        ...globalTags,
        `brand:${brand.toLowerCase()}`,
        ...attributeTags,
      ]),
    ].join(",");

    const price =
      Math.round(
        (priceMin + Math.random() * (priceMax - priceMin)) * 100
      ) / 100;
    const stockCount = 5 + Math.floor(Math.random() * 60);

    return {
      id: randomUUID(),
      sku: `SKU-${now}-${index}`,
      name: `${adjective} ${noun}`,
      description: `${brand} ${noun.toLowerCase()} for ${template.path
        .slice(-1)
        .join("")
        .toLowerCase()} use`,
      price,
      currency: "USD",
      brand,
      category: categoryPath,
      imageUrl: "",
      stockCount,
      tagsCsv,
      attributesJson: JSON.stringify(mergedAttributes),
    };
  });

  await prisma.product.createMany({
    data: productData,
    skipDuplicates: true,
  });

  const createdProducts = await prisma.product.findMany({
    where: {
      id: {
        in: productData.map((product) => product.id),
      },
    },
  });

  void upsertItems(createdProducts.map((product) => buildItemContract(product)))
    .catch((error) => {
      console.error("Failed to sync products to recsys:", error);
    });

  return NextResponse.json({ inserted: createdProducts.length });
}

function resolveCategoryTemplates(
  categories: SeedPayload["categories"]
): CategoryTemplate[] {
  if (!Array.isArray(categories) || categories.length === 0) {
    return DEFAULT_CATEGORY_TEMPLATES;
  }
  const templates: CategoryTemplate[] = [];
  categories.forEach((entry) => {
    if (typeof entry === "string") {
      const path = entry
        .split(">")
        .map((part) => part.trim())
        .filter(Boolean);
      if (path.length > 0) {
        templates.push({
          path,
          tags: path.map((part) => part.toLowerCase()),
          attributes: {},
        });
      }
      return;
    }
    const pathSource = entry?.path;
    const pathArray = Array.isArray(pathSource)
      ? pathSource
      : typeof pathSource === "string"
        ? pathSource.split(">").map((part) => part.trim())
        : [];
    if (pathArray.length === 0) {
      return;
    }
    templates.push({
      path: pathArray,
      tags: Array.isArray(entry?.tags)
        ? entry.tags.map((tag) => tag.toLowerCase())
        : pathArray.map((part) => part.toLowerCase()),
      attributes: sanitizeAttributePool(entry?.attributes ?? {}),
    });
  });
  return templates.length > 0 ? templates : DEFAULT_CATEGORY_TEMPLATES;
}

function sanitizeAttributePool(
  attributes: Record<string, string[]> | undefined
): Record<string, string[]> {
  if (!attributes) return {};
  return Object.fromEntries(
    Object.entries(attributes).map(([key, values]) => [
      key,
      Array.isArray(values) ? values.map((value) => String(value)) : [],
    ])
  );
}

function buildAttributes(
  base: Record<string, string[]>,
  supplement: Record<string, string[]>
): Record<string, string> {
  const result: Record<string, string> = {};
  const merged = { ...supplement, ...base };
  Object.entries(merged).forEach(([key, values]) => {
    if (!Array.isArray(values) || values.length === 0) {
      return;
    }
    const choice = values[Math.floor(Math.random() * values.length)];
    result[key] = choice;
  });
  return result;
}
