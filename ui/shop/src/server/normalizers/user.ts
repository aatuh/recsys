import { z } from "zod";

const UserCreateSchema = z.object({
  displayName: z.string().min(1),
  traits: z.union([z.string(), z.record(z.any())]).optional().nullable(),
  traitsText: z.string().optional().nullable(),
});

const UserUpdateSchema = UserCreateSchema.partial();

function sanitizeName(value: string | null | undefined): string {
  if (!value) return "";
  return value.trim();
}

function toTraitsText(
  traits: unknown,
  fallback?: string | null
): string | null {
  if (typeof traits === "string") {
    const trimmed = traits.trim();
    if (!trimmed) {
      return null;
    }
    try {
      JSON.parse(trimmed);
      return trimmed;
    } catch {
      return JSON.stringify({ notes: trimmed });
    }
  }

  if (traits && typeof traits === "object") {
    return JSON.stringify(traits);
  }

  if (typeof fallback === "string" && fallback.trim()) {
    try {
      JSON.parse(fallback);
      return fallback;
    } catch {
      return JSON.stringify({ notes: fallback });
    }
  }

  return null;
}

export type NormalizedUserInput = {
  displayName: string;
  traitsText: string | null;
};

export function normalizeUserPayload(payload: unknown): NormalizedUserInput {
  const parsed = UserCreateSchema.parse(payload ?? {});
  return {
    displayName: sanitizeName(parsed.displayName),
    traitsText: toTraitsText(parsed.traits, parsed.traitsText),
  };
}

export function normalizeUserPatch(
  payload: unknown
): Partial<NormalizedUserInput> {
  const parsed = UserUpdateSchema.parse(payload ?? {});
  const result: Partial<NormalizedUserInput> = {};
  if (parsed.displayName !== undefined) {
    result.displayName = sanitizeName(parsed.displayName);
  }
  if (parsed.traits !== undefined || parsed.traitsText !== undefined) {
    result.traitsText = toTraitsText(parsed.traits, parsed.traitsText);
  }
  return result;
}
