import { prisma } from "@/server/db/client";
import type { RecommendationProfile } from "@prisma/client";

const DEFAULT_PROFILE_ID =
  process.env.SHOP_DEFAULT_PROFILE_ID?.trim() || "default";
const PROFILE_CACHE_TTL_MS = Number(
  process.env.SHOP_PROFILE_CACHE_TTL_MS ?? "10000"
);

export type AlgorithmProfileDTO = {
  profileId: string;
  name: string;
  description?: string | null;
  surface?: string | null;
  isDefault: boolean;
  blendAlpha: number;
  blendBeta: number;
  blendGamma: number;
  popularityHalflifeDays: number;
  covisWindowDays: number;
  popularityFanout: number;
  mmrLambda: number;
  brandCap: number;
  categoryCap: number;
  ruleExcludeEvents: boolean;
  purchasedWindowDays: number;
  profileWindowDays: number;
  profileTopN: number;
  profileBoost: number;
  excludeEventTypes: string[];
  createdAt: string;
  updatedAt: string;
};

export type AlgorithmProfileInput = {
  profileId: string;
  name: string;
  description?: string | null;
  surface?: string | null;
  isDefault?: boolean;
  blendAlpha: number;
  blendBeta: number;
  blendGamma: number;
  popularityHalflifeDays: number;
  covisWindowDays: number;
  popularityFanout: number;
  mmrLambda: number;
  brandCap: number;
  categoryCap: number;
  ruleExcludeEvents: boolean;
  purchasedWindowDays: number;
  profileWindowDays: number;
  profileTopN: number;
  profileBoost: number;
  excludeEventTypes: string[];
};

const DEFAULT_PROFILE_TEMPLATE: AlgorithmProfileDTO = {
  profileId: DEFAULT_PROFILE_ID,
  name: "Default profile",
  description: "Seeded defaults",
  surface: null,
  isDefault: true,
  blendAlpha: 0.25,
  blendBeta: 0.35,
  blendGamma: 0.4,
  popularityHalflifeDays: 4,
  covisWindowDays: 28,
  popularityFanout: 500,
  mmrLambda: 0.3,
  brandCap: 2,
  categoryCap: 3,
  ruleExcludeEvents: true,
  purchasedWindowDays: 180,
  profileWindowDays: 30,
  profileTopN: 64,
  profileBoost: 0.7,
  excludeEventTypes: ["view", "click", "add"],
  createdAt: new Date(0).toISOString(),
  updatedAt: new Date(0).toISOString(),
};

type CachedProfile = {
  value: AlgorithmProfileDTO;
  source: AlgorithmProfileSource;
  expiresAt: number;
};

const profileCache = new Map<string, CachedProfile>();

function toDTO(profile: RecommendationProfile): AlgorithmProfileDTO {
  const tokens = profile.excludeEventTypes
    ? profile.excludeEventTypes
        .split(",")
        .map((token) => token.trim())
        .filter(Boolean)
    : [];
  return {
    profileId: profile.profileId,
    name: profile.name,
    description: profile.description,
    surface: profile.surface,
    isDefault: profile.isDefault,
    blendAlpha: profile.blendAlpha,
    blendBeta: profile.blendBeta,
    blendGamma: profile.blendGamma,
    popularityHalflifeDays: profile.popularityHalflifeDays,
    covisWindowDays: profile.covisWindowDays,
    popularityFanout: profile.popularityFanout,
    mmrLambda: profile.mmrLambda,
    brandCap: profile.brandCap,
    categoryCap: profile.categoryCap,
    ruleExcludeEvents: profile.ruleExcludeEvents,
    purchasedWindowDays: profile.purchasedWindowDays,
    profileWindowDays: profile.profileWindowDays,
    profileTopN: profile.profileTopN,
    profileBoost: profile.profileBoost,
    excludeEventTypes: tokens,
    createdAt: profile.createdAt.toISOString(),
    updatedAt: profile.updatedAt.toISOString(),
  };
}

function serializeExcludeEventTypes(values: string[]): string {
  return values.map((value) => value.trim()).filter(Boolean).join(",");
}

async function ensureDefaultProfile(): Promise<RecommendationProfile> {
  const existing = await prisma.recommendationProfile.findUnique({
    where: { profileId: DEFAULT_PROFILE_ID },
  });
  if (existing) {
    if (!existing.isDefault) {
      await prisma.recommendationProfile.update({
        where: { profileId: DEFAULT_PROFILE_ID },
        data: { isDefault: true },
      });
    }
    return existing;
  }

  const created = await prisma.recommendationProfile.create({
    data: {
      profileId: DEFAULT_PROFILE_TEMPLATE.profileId,
      name: DEFAULT_PROFILE_TEMPLATE.name,
      description: DEFAULT_PROFILE_TEMPLATE.description ?? undefined,
      surface: DEFAULT_PROFILE_TEMPLATE.surface ?? undefined,
      isDefault: true,
      blendAlpha: DEFAULT_PROFILE_TEMPLATE.blendAlpha,
      blendBeta: DEFAULT_PROFILE_TEMPLATE.blendBeta,
      blendGamma: DEFAULT_PROFILE_TEMPLATE.blendGamma,
      popularityHalflifeDays: DEFAULT_PROFILE_TEMPLATE.popularityHalflifeDays,
      covisWindowDays: DEFAULT_PROFILE_TEMPLATE.covisWindowDays,
      popularityFanout: DEFAULT_PROFILE_TEMPLATE.popularityFanout,
      mmrLambda: DEFAULT_PROFILE_TEMPLATE.mmrLambda,
      brandCap: DEFAULT_PROFILE_TEMPLATE.brandCap,
      categoryCap: DEFAULT_PROFILE_TEMPLATE.categoryCap,
      ruleExcludeEvents: DEFAULT_PROFILE_TEMPLATE.ruleExcludeEvents,
      purchasedWindowDays: DEFAULT_PROFILE_TEMPLATE.purchasedWindowDays,
      profileWindowDays: DEFAULT_PROFILE_TEMPLATE.profileWindowDays,
      profileTopN: DEFAULT_PROFILE_TEMPLATE.profileTopN,
      profileBoost: DEFAULT_PROFILE_TEMPLATE.profileBoost,
      excludeEventTypes: serializeExcludeEventTypes(
        DEFAULT_PROFILE_TEMPLATE.excludeEventTypes
      ),
    },
  });
  return created;
}

function cacheKey(args: { profileId?: string; surface?: string | null }) {
  if (args.profileId) {
    return `id:${args.profileId}`;
  }
  if (args.surface) {
    return `surface:${args.surface}`;
  }
  return "default";
}

function cacheProfile(
  key: string,
  value: AlgorithmProfileDTO,
  source: AlgorithmProfileSource
) {
  profileCache.set(key, {
    value,
    source,
    expiresAt: Date.now() + Math.max(PROFILE_CACHE_TTL_MS, 1000),
  });
}

export function clearAlgorithmProfileCache() {
  profileCache.clear();
}

export function defaultAlgorithmProfileDTO(): AlgorithmProfileDTO {
  return { ...DEFAULT_PROFILE_TEMPLATE };
}

export async function listAlgorithmProfiles(): Promise<AlgorithmProfileDTO[]> {
  await ensureDefaultProfile();
  const rows = await prisma.recommendationProfile.findMany({
    orderBy: [{ isDefault: "desc" }, { name: "asc" }],
  });
  return rows.map(toDTO);
}

export type AlgorithmProfileSource =
  | "explicit"
  | "surface"
  | "default"
  | "fallback";

export async function getAlgorithmProfile(options: {
  profileId?: string;
  surface?: string | null;
} = {}): Promise<{ profile: AlgorithmProfileDTO; source: AlgorithmProfileSource }> {
  await ensureDefaultProfile();
  const { profileId, surface } = options;

  if (profileId) {
    const explicit = await prisma.recommendationProfile.findUnique({
      where: { profileId },
    });
    if (!explicit) {
      throw new Error(`Profile '${profileId}' not found`);
    }
    return { profile: toDTO(explicit), source: "explicit" };
  }

  if (surface) {
    const surfaceDefault = await prisma.recommendationProfile.findFirst({
      where: { surface, isDefault: true },
    });
    if (surfaceDefault) {
      return { profile: toDTO(surfaceDefault), source: "surface" };
    }
  }

  const globalDefault = await prisma.recommendationProfile.findFirst({
    where: { isDefault: true },
    orderBy: { updatedAt: "desc" },
  });
  if (globalDefault) {
    return { profile: toDTO(globalDefault), source: "default" };
  }

  const fallback = await ensureDefaultProfile();
  return { profile: toDTO(fallback), source: "fallback" };
}

export async function getAlgorithmProfileCached(options: {
  profileId?: string;
  surface?: string | null;
} = {}): Promise<{ profile: AlgorithmProfileDTO; source: AlgorithmProfileSource }> {
  const key = cacheKey(options);
  const cached = profileCache.get(key);
  const now = Date.now();
  if (cached && cached.expiresAt > now) {
    return { profile: cached.value, source: cached.source };
  }
  const result = await getAlgorithmProfile(options);
  cacheProfile(key, result.profile, result.source);
  return result;
}

async function unsetOtherDefaults(
  profileId: string,
  surface: string | null | undefined
) {
  await prisma.recommendationProfile.updateMany({
    where: {
      profileId: { not: profileId },
      OR: surface
        ? [{ surface }, { surface: null }]
        : [{ surface: null }, { surface: { not: null } }],
    },
    data: { isDefault: false },
  });
}

export async function createAlgorithmProfile(
  input: AlgorithmProfileInput
): Promise<AlgorithmProfileDTO> {
  await ensureDefaultProfile();
  const created = await prisma.recommendationProfile.create({
    data: {
      profileId: input.profileId,
      name: input.name,
      description: input.description,
      surface: input.surface,
      isDefault: Boolean(input.isDefault),
      blendAlpha: input.blendAlpha,
      blendBeta: input.blendBeta,
      blendGamma: input.blendGamma,
      popularityHalflifeDays: input.popularityHalflifeDays,
      covisWindowDays: input.covisWindowDays,
      popularityFanout: input.popularityFanout,
      mmrLambda: input.mmrLambda,
      brandCap: input.brandCap,
      categoryCap: input.categoryCap,
      ruleExcludeEvents: input.ruleExcludeEvents,
      purchasedWindowDays: input.purchasedWindowDays,
      profileWindowDays: input.profileWindowDays,
      profileTopN: input.profileTopN,
      profileBoost: input.profileBoost,
      excludeEventTypes: serializeExcludeEventTypes(input.excludeEventTypes),
    },
  });

  if (created.isDefault) {
    await unsetOtherDefaults(created.profileId, created.surface);
  }

  clearAlgorithmProfileCache();
  return toDTO(created);
}

export async function updateAlgorithmProfile(
  profileId: string,
  input: Omit<AlgorithmProfileInput, "profileId">
): Promise<AlgorithmProfileDTO> {
  const updated = await prisma.recommendationProfile.update({
    where: { profileId },
    data: {
      name: input.name,
      description: input.description,
      surface: input.surface,
      isDefault: Boolean(input.isDefault),
      blendAlpha: input.blendAlpha,
      blendBeta: input.blendBeta,
      blendGamma: input.blendGamma,
      popularityHalflifeDays: input.popularityHalflifeDays,
      covisWindowDays: input.covisWindowDays,
      popularityFanout: input.popularityFanout,
      mmrLambda: input.mmrLambda,
      brandCap: input.brandCap,
      categoryCap: input.categoryCap,
      ruleExcludeEvents: input.ruleExcludeEvents,
      purchasedWindowDays: input.purchasedWindowDays,
      profileWindowDays: input.profileWindowDays,
      profileTopN: input.profileTopN,
      profileBoost: input.profileBoost,
      excludeEventTypes: serializeExcludeEventTypes(input.excludeEventTypes),
    },
  });

  if (updated.isDefault) {
    await unsetOtherDefaults(updated.profileId, updated.surface);
  }

  clearAlgorithmProfileCache();
  return toDTO(updated);
}

export async function deleteAlgorithmProfile(profileId: string): Promise<void> {
  await prisma.recommendationProfile.delete({ where: { profileId } });
  clearAlgorithmProfileCache();
  await ensureDefaultProfile();
}

export async function deleteAllAlgorithmProfiles(): Promise<void> {
  await prisma.recommendationProfile.deleteMany({});
  clearAlgorithmProfileCache();
  await ensureDefaultProfile();
}
