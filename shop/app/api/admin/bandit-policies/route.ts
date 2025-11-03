import { NextRequest, NextResponse } from "next/server";
import { BanditService } from "@/lib/api-client/services/BanditService";
import { loadConfig } from "@/server/services/config";
import {
  initRecsysClient,
  resetBanditFeatureStatus,
} from "@/server/services/recsys";

function toNumber(value: unknown): number | undefined {
  if (value === null || value === undefined || value === "") {
    return undefined;
  }
  const num = Number(value);
  return Number.isFinite(num) ? num : undefined;
}

export async function GET() {
  initRecsysClient();
  const cfg = loadConfig();
  const policies = await BanditService.getV1BanditPolicies(cfg.recsysNamespace);
  return NextResponse.json({ policies });
}

export async function POST(req: NextRequest) {
  try {
    initRecsysClient();
    const body = await req.json();
    const payload = body?.policy ?? body;
    if (!payload?.policy_id || typeof payload.policy_id !== "string") {
      return NextResponse.json(
        { error: "policy_id is required" },
        { status: 400 }
      );
    }
    const cfg = loadConfig();
    await BanditService.upsertBanditPolicies({
      namespace: cfg.recsysNamespace,
      policies: [
        {
          policy_id: payload.policy_id,
          name: payload.name ?? payload.policy_id,
          active: Boolean(payload.active ?? true),
          blend_alpha: toNumber(payload.blend_alpha),
          blend_beta: toNumber(payload.blend_beta),
          blend_gamma: toNumber(payload.blend_gamma),
          mmr_lambda: toNumber(payload.mmr_lambda),
          brand_cap: toNumber(payload.brand_cap),
          category_cap: toNumber(payload.category_cap),
        },
      ],
    });
    resetBanditFeatureStatus();
    return NextResponse.json({ status: "success" });
  } catch (error) {
    console.error("Failed to upsert bandit policy", error);
    return NextResponse.json(
      { error: "Failed to upsert bandit policy" },
      { status: 500 }
    );
  }
}
