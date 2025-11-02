import { NextRequest, NextResponse } from "next/server";
import { getRecommendations } from "@/server/services/recsys";
import { RecommendationConstraints } from "@/lib/recommendations/constraints";
import { getSeenItemIds, summarizeExclusion } from "@/server/services/seenItems";
import { applyDiversityCaps } from "@/server/services/diversity";
import { getFreshItemIds } from "@/server/services/freshItems";
import { logColdStart } from "@/server/logging/coldStart";
import {
  getAlgorithmProfileCached,
  defaultAlgorithmProfileDTO,
} from "@/server/services/recommendationProfiles";

const FALLBACK_BRAND_CAP = 2;
const FALLBACK_CATEGORY_CAP = 3;

function envNumber(key: string): number | undefined {
  const raw = process.env[key];
  if (raw === undefined) return undefined;
  const parsed = Number(raw);
  return Number.isFinite(parsed) ? parsed : undefined;
}

const ENV_DEFAULT_BRAND_CAP = envNumber("SHOP_DIVERSITY_BRAND_CAP");
const ENV_DEFAULT_CATEGORY_CAP = envNumber("SHOP_DIVERSITY_CATEGORY_CAP");
const DEFAULT_FRESH_SLOTS = Number(
  process.env.SHOP_FRESH_ITEM_SLOTS ?? "2"
);
const DEFAULT_FRESH_MAX_AGE_DAYS = Number(
  process.env.SHOP_FRESH_ITEM_MAX_AGE_DAYS ?? "7"
);

function resolveSurfaceCaps(
  surface: string,
  defaults: { brandCap: number; categoryCap: number }
) {
  const key = surface.toUpperCase();
  const brandCap =
    envNumber(`SHOP_DIVERSITY_BRAND_CAP_${key}`) ?? defaults.brandCap;
  const categoryCap =
    envNumber(`SHOP_DIVERSITY_CATEGORY_CAP_${key}`) ?? defaults.categoryCap;
  return { brandCap, categoryCap };
}

export async function GET(req: NextRequest) {
  try {
    const { searchParams } = new URL(req.url);
    const userId = searchParams.get("userId");
    const k = Number(searchParams.get("k") ?? "8");
    const includeReasons = searchParams.get("includeReasons") === "true";
    const surface = searchParams.get("surface") ?? "home";
    const widget = searchParams.get("widget") ?? undefined;
    const variant = searchParams.get("variant") ?? undefined;
    const requestedProfileId = searchParams.get("profileId") ?? undefined;

    if (!userId) {
      return NextResponse.json({ error: "userId required" }, { status: 400 });
    }

    let settings = defaultAlgorithmProfileDTO();
    let profileSource = "fallback";
    try {
      const result = await getAlgorithmProfileCached({
        profileId: requestedProfileId ?? undefined,
        surface,
      });
      settings = result.profile;
      profileSource = result.source;
    } catch (err) {
      if (requestedProfileId) {
        const message =
          err instanceof Error ? err.message : "Requested profile not found";
        return NextResponse.json({ error: message }, { status: 404 });
      }
      console.warn(
        "[recsys] defaulting algorithm profile due to load failure",
        err
      );
    }

    // Parse constraints from query parameters
    const constraints: RecommendationConstraints = {};
    
    const minPrice = searchParams.get("minPrice");
    const maxPrice = searchParams.get("maxPrice");
    if (minPrice && maxPrice) {
      constraints.price_between = [Number(minPrice), Number(maxPrice)];
    }
    
    const includeTags = searchParams.get("includeTags");
    if (includeTags) {
      constraints.include_tags_any = includeTags.split(",").map(tag => tag.trim());
    }

    const excludeTags = searchParams.get("excludeTags");
    if (excludeTags) {
      constraints.exclude_tags_any = excludeTags.split(",").map(tag => tag.trim());
    }

    const brandCapParam = searchParams.get("brandCap");
    const categoryCapParam = searchParams.get("categoryCap");
    const baseBrandCap =
      Number.isFinite(settings.brandCap) && settings.brandCap >= 0
        ? settings.brandCap
        : ENV_DEFAULT_BRAND_CAP ?? FALLBACK_BRAND_CAP;
    const baseCategoryCap =
      Number.isFinite(settings.categoryCap) && settings.categoryCap >= 0
        ? settings.categoryCap
        : ENV_DEFAULT_CATEGORY_CAP ?? FALLBACK_CATEGORY_CAP;

    const { brandCap: surfaceBrandCap, categoryCap: surfaceCategoryCap } =
      resolveSurfaceCaps(surface, {
        brandCap: baseBrandCap,
        categoryCap: baseCategoryCap,
      });

    const parsedBrandCap =
      brandCapParam !== null ? Number(brandCapParam) : surfaceBrandCap;
    const parsedCategoryCap =
      categoryCapParam !== null
        ? Number(categoryCapParam)
        : surfaceCategoryCap;
    
    const resolvedBrandCap = Number.isFinite(parsedBrandCap)
      ? parsedBrandCap
      : baseBrandCap;
    const resolvedCategoryCap = Number.isFinite(parsedCategoryCap)
      ? parsedCategoryCap
      : baseCategoryCap;

    if (brandCapParam) {
      constraints.brand_cap = resolvedBrandCap;
    }
    
    if (categoryCapParam) {
      constraints.category_cap = resolvedCategoryCap;
    }

    const extraContext: Record<string, string> = {};
    if (variant) {
      extraContext.variant = variant;
    }

    const response = await getRecommendations({
      userId,
      k,
      includeReasons,
      constraints: Object.keys(constraints).length > 0 ? constraints : undefined,
      surface,
      widget,
      context: Object.keys(extraContext).length > 0 ? extraContext : undefined,
      profileId: settings.profileId,
    });
    const seenItemIds = await getSeenItemIds({ userId });
    const items = response.items ?? [];
    const filtered = items.filter((item) => !seenItemIds.has(item.item_id));
    const excludedCount = items.length - filtered.length;

    if (excludedCount > 0) {
      console.info(
        `[recsys] seen-filter user=${userId} ${summarizeExclusion(
          seenItemIds,
          excludedCount
        )}`
      );
    }

    let workingItems = filtered.length > 0 ? filtered : items;
    const payload: Record<string, unknown> = {
      ...response,
      surface,
      widget,
      settings_version: settings.updatedAt,
      profile_id: settings.profileId,
      profile_source: profileSource,
    };

    if (response.profile_id) {
      payload.profile_id = response.profile_id;
    }
    if (response.profile_source) {
      payload.profile_source = response.profile_source;
    }

    // resolved caps already calculated above, reuse for post-processing
    const shouldApplyDiversity =
      (resolvedBrandCap ?? 0) > 0 || (resolvedCategoryCap ?? 0) > 0;

    if (shouldApplyDiversity && workingItems.length > 0) {
      const { kept, excluded: diversityExcluded } = await applyDiversityCaps(
        workingItems,
        {
          brandCap: resolvedBrandCap,
          categoryCap: resolvedCategoryCap,
        }
      );

      if (kept.length > 0) {
        workingItems = kept;
      }

      const excludedTotal = Object.values(diversityExcluded).reduce(
        (acc, value) => acc + value,
        0
      );

      if (excludedTotal > 0) {
        console.info(
          `[recsys] diversity-filter user=${userId} excluded=${JSON.stringify(
            diversityExcluded
          )}`
        );
        payload.filters = {
          ...(payload.filters as Record<string, unknown>),
          diversity: {
            excluded: diversityExcluded,
            applied_caps: {
              brand: resolvedBrandCap,
              category: resolvedCategoryCap,
            },
          },
        };
      }
    }

    const freshSlots = DEFAULT_FRESH_SLOTS;
    if (freshSlots > 0) {
      const existingIds = new Set(workingItems.map((item) => item.item_id));
      const exclusionSet = new Set<string>([...seenItemIds, ...existingIds]);

      const freshIds = await getFreshItemIds({
        limit: freshSlots,
        maxAgeDays: DEFAULT_FRESH_MAX_AGE_DAYS,
        excludeIds: exclusionSet,
      });

      if (freshIds.length > 0) {
        const freshItems = freshIds.map((id, index) => ({
          item_id: id,
          score: 0,
          reasons: ["cold_start", `fresh_rank_${index + 1}`],
          metadata: { cold_start: true },
        }));

        workingItems = [...workingItems, ...freshItems];

        console.info(
          `[recsys] cold-start user=${userId} inserted=${freshIds.length} items`
        );

        const baseRank = workingItems.length - freshItems.length;
        freshItems.forEach((item, idx) => {
          logColdStart({
            type: "impression",
            userId,
            itemId: item.item_id,
            surface,
            widget,
            rank: baseRank + idx + 1,
          });
        });

        payload.filters = {
          ...(payload.filters as Record<string, unknown>),
          cold_start: {
            inserted_count: freshIds.length,
            slots: freshSlots,
            max_age_days: DEFAULT_FRESH_MAX_AGE_DAYS,
          },
        };
      }
    }

    payload.items = workingItems.slice(0, k);

    if (excludedCount > 0) {
      payload.filters = {
        ...(payload.filters as Record<string, unknown>),
        seen: {
          excluded_count: excludedCount,
          cache_size: seenItemIds.size,
        },
      };
    }

    return NextResponse.json(payload);
  } catch (error) {
    console.error("Recommendation API error:", error);
    return NextResponse.json(
      { error: "Failed to get recommendations" },
      { status: 500 }
    );
  }
}
