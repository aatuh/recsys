import { z } from "zod";

// User contract for Recsys API
export const UserContract = z.object({
  user_id: z.string(),
  traits: z
    .object({
      display_name: z.string().optional(),
      locale: z.string().optional(),
      country: z.string().optional(),
      device: z.enum(["mobile", "desktop"]).optional(),
      signup_ts: z.string().optional(),
      last_seen_ts: z.string().optional(),
      loyalty_tier: z.enum(["bronze", "silver", "gold", "vip"]).optional(),
      newsletter: z.boolean().optional(),
      preferred_categories: z.array(z.string()).optional(),
      brand_affinity: z.record(z.number()).optional(),
      price_sensitivity: z.enum(["low", "mid", "high"]).optional(),
      lifetime_value_bucket: z.enum(["L", "M", "H"]).optional(),
    })
    .optional(),
});

export type UserContract = z.infer<typeof UserContract>;

export function buildUserContract(user: {
  id: string;
  displayName: string;
  traitsText?: string | null;
}): UserContract {
  let traits: Record<string, unknown> = {};

  // Parse existing traits if available
  if (user.traitsText) {
    try {
      traits = JSON.parse(user.traitsText);
    } catch {
      // Ignore parse errors, use defaults
    }
  }

  // Set defaults and ensure non-PII
  return {
    user_id: user.id,
    traits: {
      display_name: user.displayName,
      locale: (traits.locale as string) || "en-US",
      country: (traits.country as string) || "US",
      device: (traits.device as "mobile" | "desktop") || "desktop",
      last_seen_ts: new Date().toISOString(),
      loyalty_tier:
        (traits.loyalty_tier as "bronze" | "silver" | "gold" | "vip") ||
        "bronze",
      newsletter: (traits.newsletter as boolean) || false,
      preferred_categories: (traits.preferred_categories as string[]) || [],
      price_sensitivity:
        (traits.price_sensitivity as "low" | "mid" | "high") || "mid",
      lifetime_value_bucket:
        (traits.lifetime_value_bucket as "L" | "M" | "H") || "M",
      ...traits,
    },
  };
}
