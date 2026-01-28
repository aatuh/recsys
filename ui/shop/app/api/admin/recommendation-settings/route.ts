import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import {
  listAlgorithmProfiles,
  getAlgorithmProfile,
  createAlgorithmProfile,
  updateAlgorithmProfile,
  AlgorithmProfileInput,
  deleteAlgorithmProfile,
  deleteAllAlgorithmProfiles,
} from "@/server/services/recommendationProfiles";
import {
  getBanditFeatureStatus,
  getConfiguredBanditPolicyIds,
} from "@/server/services/recsys";

const ARRAY_DELIMITER = ",";

const ProfileSchema = z.object({
  profileId: z.string().min(1, "profileId is required"),
  name: z.string().min(1, "name is required"),
  description: z.string().optional(),
  surface: z.string().min(1).optional().nullable(),
  isDefault: z.coerce.boolean().optional(),
  blendAlpha: z.coerce.number().min(0),
  blendBeta: z.coerce.number().min(0),
  blendGamma: z.coerce.number().min(0),
  popularityHalflifeDays: z.coerce.number().min(0),
  covisWindowDays: z.coerce.number().min(0),
  popularityFanout: z.coerce.number().int().min(1),
  mmrLambda: z.coerce.number().min(0),
  brandCap: z.coerce.number().int().min(0),
  categoryCap: z.coerce.number().int().min(0),
  ruleExcludeEvents: z.coerce.boolean(),
  purchasedWindowDays: z.coerce.number().min(0),
  profileWindowDays: z.coerce.number().min(0),
  profileTopN: z.coerce.number().int().min(0),
  profileBoost: z.coerce.number().min(0),
  excludeEventTypes: z
    .union([
      z.array(z.string()),
      z
        .string()
        .transform((value) =>
          value
            .split(ARRAY_DELIMITER)
            .map((part) => part.trim())
            .filter((part) => part.length > 0)
        ),
    ])
    .transform((value) =>
      Array.isArray(value)
        ? value.map((part) => part.trim()).filter(Boolean)
        : value
    ),
});

function toProfileInput(parsed: z.infer<typeof ProfileSchema>): AlgorithmProfileInput {
  return {
    profileId: parsed.profileId,
    name: parsed.name,
    description: parsed.description ?? undefined,
    surface: parsed.surface ?? undefined,
    isDefault: parsed.isDefault ?? false,
    blendAlpha: parsed.blendAlpha,
    blendBeta: parsed.blendBeta,
    blendGamma: parsed.blendGamma,
    popularityHalflifeDays: parsed.popularityHalflifeDays,
    covisWindowDays: parsed.covisWindowDays,
    popularityFanout: parsed.popularityFanout,
    mmrLambda: parsed.mmrLambda,
    brandCap: parsed.brandCap,
    categoryCap: parsed.categoryCap,
    ruleExcludeEvents: parsed.ruleExcludeEvents,
    purchasedWindowDays: parsed.purchasedWindowDays,
    profileWindowDays: parsed.profileWindowDays,
    profileTopN: parsed.profileTopN,
    profileBoost: parsed.profileBoost,
    excludeEventTypes: parsed.excludeEventTypes,
  };
}

function errorResponse(message: string, status = 400) {
  return NextResponse.json({ error: message }, { status });
}

export async function GET(req: NextRequest) {
  try {
    const { searchParams } = new URL(req.url);
    const profileId = searchParams.get("profileId") ?? undefined;
    const surface = searchParams.get("surface") ?? undefined;

    let profileResult;
    try {
      profileResult = await getAlgorithmProfile({
        profileId: profileId ?? undefined,
        surface: surface ?? undefined,
      });
    } catch (error) {
      return errorResponse(
        error instanceof Error ? error.message : "Profile not found",
        404
      );
    }

    const profiles = await listAlgorithmProfiles();
    const bandit = await getBanditFeatureStatus();

    return NextResponse.json({
      profile: profileResult.profile,
      profile_source: profileResult.source,
      profiles,
      bandit,
      configuredPolicies: getConfiguredBanditPolicyIds(),
    });
  } catch (error) {
    console.error("Failed to fetch recommendation profiles", error);
    return NextResponse.json(
      { error: "Failed to load recommendation profiles" },
      { status: 500 }
    );
  }
}

export async function POST(req: NextRequest) {
  try {
    const body = await req.json();
    const parsed = ProfileSchema.parse(body);
    const profile = await createAlgorithmProfile(toProfileInput(parsed));
    const profiles = await listAlgorithmProfiles();
    const bandit = await getBanditFeatureStatus();
    return NextResponse.json(
      {
        profile,
        profile_source: "explicit",
        profiles,
        bandit,
        configuredPolicies: getConfiguredBanditPolicyIds(),
      },
      { status: 201 }
    );
  } catch (error) {
    if (error instanceof z.ZodError) {
      return NextResponse.json(
        { error: "Invalid profile payload", details: error.flatten() },
        { status: 400 }
      );
    }
    console.error("Failed to create algorithm profile", error);
    return NextResponse.json(
      { error: "Failed to create profile" },
      { status: 500 }
    );
  }
}

export async function DELETE(req: NextRequest) {
  try {
    const { searchParams } = new URL(req.url);
    let profileId = searchParams.get("profileId");
    if (!profileId) {
      const body = await req.json().catch(() => ({}));
      profileId = body?.profileId ?? body?.id ?? undefined;
    }

    if (!profileId) {
      return NextResponse.json(
        { error: "profileId is required" },
        { status: 400 }
      );
    }

    if (profileId === "all") {
      await deleteAllAlgorithmProfiles();
    } else {
      await deleteAlgorithmProfile(profileId);
    }

    const profiles = await listAlgorithmProfiles();
    const bandit = await getBanditFeatureStatus();
    return NextResponse.json({
      profiles,
      bandit,
      configuredPolicies: getConfiguredBanditPolicyIds(),
    });
  } catch (error) {
    console.error("Failed to delete recommendation profile", error);
    return NextResponse.json(
      { error: "Failed to delete profile" },
      { status: 500 }
    );
  }
}

export async function PUT(req: NextRequest) {
  try {
    const body = await req.json();
    const parsed = ProfileSchema.parse(body);
    const { profileId, ...rest } = toProfileInput(parsed);
    const profile = await updateAlgorithmProfile(profileId, rest);
    const profiles = await listAlgorithmProfiles();
    const bandit = await getBanditFeatureStatus();
    return NextResponse.json({
      profile,
      profile_source: "explicit",
      profiles,
      bandit,
      configuredPolicies: getConfiguredBanditPolicyIds(),
    });
  } catch (error) {
    if (error instanceof z.ZodError) {
      return NextResponse.json(
        { error: "Invalid profile payload", details: error.flatten() },
        { status: 400 }
      );
    }
    if (error instanceof Error && error.message.includes("not found")) {
      return errorResponse(error.message, 404);
    }
    console.error("Failed to update algorithm profile", error);
    return NextResponse.json(
      { error: "Failed to update profile" },
      { status: 500 }
    );
  }
}
