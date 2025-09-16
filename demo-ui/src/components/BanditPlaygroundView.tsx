import React, { useState } from "react";
import { PolicyManager } from "./PolicyManager";
import { DecisionSection } from "./DecisionSection";
import { OneShotRecommendationsSection } from "./OneShotRecommendationsSection";
import { RewardFeedbackSection } from "./RewardFeedbackSection";
import { BanditDashboards } from "./BanditDashboards";
// Define BanditPolicy type locally since it's not exported from api-client
interface types_BanditPolicy {
  policy_id?: string;
  name: string;
  notes?: string;
  active: boolean;
  [key: string]: any; // Allow additional properties
}

interface BanditPlaygroundViewProps {
  namespace: string;
  generatedUsers: string[];
}

export function BanditPlaygroundView({
  namespace,
  generatedUsers,
}: BanditPlaygroundViewProps) {
  const [availablePolicies, setAvailablePolicies] = useState<
    types_BanditPolicy[]
  >([]);

  return (
    <div style={{ padding: 16, fontFamily: "system-ui, sans-serif" }}>
      <p style={{ color: "#444", marginBottom: 24 }}>
        Manage bandit policies and test multi-armed bandit algorithms. Create,
        edit, and copy policies with different algorithm parameters to optimize
        recommendation performance. Simulate decisions to explore vs exploit
        behavior and provide reward feedback to help the system learn.
      </p>

      <PolicyManager
        namespace={namespace}
        onPoliciesChange={setAvailablePolicies}
      />

      <DecisionSection
        namespace={namespace}
        availablePolicies={availablePolicies}
      />

      <OneShotRecommendationsSection
        namespace={namespace}
        availablePolicies={availablePolicies}
        generatedUsers={generatedUsers}
      />

      <RewardFeedbackSection namespace={namespace} />

      <BanditDashboards
        namespace={namespace}
        availablePolicies={availablePolicies}
      />
    </div>
  );
}
