import { z } from "zod";

// Event contract for Recsys API
export const EventContract = z.object({
  user_id: z.string(),
  item_id: z.string().optional(),
  type: z.number().min(0).max(4), // 0=view, 1=click, 2=add, 3=purchase, 4=custom
  value: z.number().default(1),
  ts: z.string(), // ISO-8601 UTC
  meta: z.record(z.any()).optional(),
  source_event_id: z.string().optional(),
});

export type EventContract = z.infer<typeof EventContract>;

// Event type mapping
export const EVENT_TYPES = {
  VIEW: 0,
  CLICK: 1,
  ADD: 2,
  PURCHASE: 3,
  CUSTOM: 4,
} as const;

export const EVENT_TYPE_NAMES = {
  [EVENT_TYPES.VIEW]: "view",
  [EVENT_TYPES.CLICK]: "click",
  [EVENT_TYPES.ADD]: "add",
  [EVENT_TYPES.PURCHASE]: "purchase",
  [EVENT_TYPES.CUSTOM]: "custom",
} as const;

// Recommendation context metadata
export const RecommendationMeta = z.object({
  surface: z.enum(["home", "pdp", "cart", "checkout", "products"]).optional(),
  widget: z.string().optional(),
  recommended: z.boolean().optional(),
  request_id: z.string().optional(),
  rank: z.number().optional(),
  experiment: z.string().optional(),
  ab_bucket: z.string().optional(),
  session_id: z.string().optional(),
  referrer: z.string().optional(),
  href: z.string().optional(),
  text: z.string().optional(),
  cart_id: z.string().optional(),
  unit_price: z.number().optional(),
  currency: z.string().optional(),
  order_id: z.string().optional(),
  line_item_id: z.string().optional(),
  kind: z.string().optional(), // for custom events
  total: z.number().optional(),
  items: z
    .array(
      z.object({
        item_id: z.string(),
        qty: z.number(),
        unit_price: z.number(),
      })
    )
    .optional(),
});

export type RecommendationMeta = z.infer<typeof RecommendationMeta>;

export function buildEventContract(event: {
  userId: string;
  productId?: string | null;
  type: "view" | "click" | "add" | "purchase" | "custom";
  value?: number;
  ts?: string;
  meta?: Record<string, unknown>;
  sourceEventId?: string;
}): EventContract {
  const typeMap: Record<string, number> = {
    view: EVENT_TYPES.VIEW,
    click: EVENT_TYPES.CLICK,
    add: EVENT_TYPES.ADD,
    purchase: EVENT_TYPES.PURCHASE,
    custom: EVENT_TYPES.CUSTOM,
  };

  return {
    user_id: event.userId,
    item_id: event.productId || undefined,
    type: typeMap[event.type] ?? EVENT_TYPES.CUSTOM,
    value: event.value ?? 1,
    ts: event.ts || new Date().toISOString(),
    meta: event.meta,
    source_event_id: event.sourceEventId,
  };
}

// Event-type weights and decay configuration
export const EventTypeConfig = z.object({
  namespace: z.string(),
  types: z.array(
    z.object({
      type: z.number(),
      name: z.string(),
      weight: z.number(),
      half_life_days: z.number(),
      is_active: z.boolean(),
    })
  ),
});

export type EventTypeConfig = z.infer<typeof EventTypeConfig>;

export const DEFAULT_EVENT_TYPE_CONFIG: EventTypeConfig = {
  namespace: "default",
  types: [
    {
      type: EVENT_TYPES.VIEW,
      name: "view",
      weight: 0.05,
      half_life_days: 3,
      is_active: true,
    },
    {
      type: EVENT_TYPES.CLICK,
      name: "click",
      weight: 0.2,
      half_life_days: 7,
      is_active: true,
    },
    {
      type: EVENT_TYPES.ADD,
      name: "add",
      weight: 0.7,
      half_life_days: 21,
      is_active: true,
    },
    {
      type: EVENT_TYPES.PURCHASE,
      name: "purchase",
      weight: 1.0,
      half_life_days: 60,
      is_active: true,
    },
    {
      type: EVENT_TYPES.CUSTOM,
      name: "custom",
      weight: 0.1,
      half_life_days: 3,
      is_active: true,
    },
  ],
};
