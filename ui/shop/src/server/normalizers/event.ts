import { z } from "zod";

const EventSchema = z.object({
  userId: z.string().min(1),
  productId: z.string().optional().nullable(),
  type: z
    .union([
      z.literal("view"),
      z.literal("click"),
      z.literal("add"),
      z.literal("purchase"),
      z.literal("custom"),
    ])
    .default("view"),
  value: z.coerce.number().optional(),
  ts: z.union([z.string(), z.date()]).optional(),
  meta: z.union([z.record(z.any()), z.string()]).optional().nullable(),
  metaJson: z.string().optional().nullable(),
  sourceEventId: z.string().optional().nullable(),
});

type ParsedEvent = z.infer<typeof EventSchema>;

function parseTimestamp(value: ParsedEvent["ts"]): Date | undefined {
  if (!value) return undefined;
  if (value instanceof Date) return value;
  const ms = Date.parse(value);
  if (Number.isNaN(ms)) return undefined;
  return new Date(ms);
}

function parseMeta(
  meta: ParsedEvent["meta"],
  metaJson: ParsedEvent["metaJson"]
): Record<string, unknown> | undefined {
  if (meta && typeof meta === "object" && !Array.isArray(meta)) {
    return meta as Record<string, unknown>;
  }
  const raw = typeof meta === "string" ? meta : metaJson ?? undefined;
  if (!raw) return undefined;
  try {
    const parsed = JSON.parse(raw);
    if (parsed && typeof parsed === "object") {
      return parsed as Record<string, unknown>;
    }
  } catch {
    return { notes: raw };
  }
  return undefined;
}

export type NormalizedEventInput = {
  data: {
    userId: string;
    productId?: string | null;
    type: string;
    value: number;
    ts?: Date;
    metaText?: string | null;
    sourceEventId?: string | null;
    recsysStatus: "pending";
  };
  meta?: Record<string, unknown>;
};

export function normalizeEventPayload(payload: unknown): NormalizedEventInput {
  const parsed = EventSchema.parse(payload ?? {});
  const meta = parseMeta(parsed.meta, parsed.metaJson);
  return {
    data: {
      userId: parsed.userId,
      productId: parsed.productId ?? null,
      type: parsed.type,
      value: Number.isFinite(parsed.value) ? (parsed.value as number) : 1,
      ts: parseTimestamp(parsed.ts),
      metaText: meta ? JSON.stringify(meta) : null,
      sourceEventId: parsed.sourceEventId ?? null,
      recsysStatus: "pending",
    },
    meta,
  };
}
