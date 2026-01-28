import { NextRequest, NextResponse } from "next/server";
import { BanditService } from "@/lib/api-client/services/BanditService";
import { loadConfig } from "@/server/services/config";
import {
  initRecsysClient,
  resetBanditFeatureStatus,
} from "@/server/services/recsys";

export async function DELETE(
  _req: NextRequest,
  { params }: { params: Promise<{ policyId: string }> }
) {
  const { policyId } = await params;
  if (!policyId) {
    return NextResponse.json(
      { error: "policyId is required" },
      { status: 400 }
    );
  }
  initRecsysClient();
  const cfg = loadConfig();
  try {
    const policies = await BanditService.getV1BanditPolicies(
      cfg.recsysNamespace
    );
    const targetPolicies =
      policyId === "all"
        ? policies
        : policies.filter((policy) => policy.policy_id === policyId);

    if (!targetPolicies.length) {
      return NextResponse.json({ error: "Policy not found" }, { status: 404 });
    }

    await BanditService.upsertBanditPolicies({
      namespace: cfg.recsysNamespace,
      policies: targetPolicies.map((policy) => ({
        ...policy,
        active: false,
      })),
    });
    resetBanditFeatureStatus();
    return NextResponse.json({ status: "success" });
  } catch (error) {
    console.error("Failed to deactivate bandit policy", error);
    return NextResponse.json(
      { error: "Failed to deactivate bandit policy" },
      { status: 500 }
    );
  }
}
