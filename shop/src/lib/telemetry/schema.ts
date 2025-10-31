import { z } from "zod";

export const telemetryEventSchema = z.object({
  userId: z.string().min(1),
  productId: z.string().min(1).optional(),
  type: z.enum(["view", "click", "add", "purchase", "custom"]),
  value: z.number().optional(),
  ts: z.string().datetime().optional(),
  meta: z.unknown().optional(),
  sourceEventId: z.string().optional(),
});

export type TelemetryEvent = z.infer<typeof telemetryEventSchema>;
