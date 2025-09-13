import React, { useMemo } from "react";
import {
  RecommendationsSection,
  OverridesSection,
  SimilarItemsSection,
} from "./";
import { useViewState } from "../contexts/ViewStateContext";
import type { internal_http_types_ScoredItem } from "../lib/api-client";

interface RecommendationsPlaygroundViewProps {
  namespace: string;
  generatedUsers: string[];
  generatedItems: string[];
}

export function RecommendationsPlaygroundView({
  namespace,
  generatedUsers,
  generatedItems,
}: RecommendationsPlaygroundViewProps) {
  const { recommendationsPlayground, setRecommendationsPlayground } =
    useViewState();

  const exampleItem = useMemo(() => {
    return generatedItems[0] || "item-0001";
  }, [generatedItems.length]);

  const exampleUser = useMemo(() => {
    return generatedUsers[0] || "user-0001";
  }, [generatedUsers.length]);

  return (
    <div style={{ padding: 16, fontFamily: "system-ui, sans-serif" }}>
      <p style={{ color: "#444", marginBottom: 24 }}>
        Test recommendation algorithms and explore similar items. Adjust blend
        parameters to see how different algorithms contribute to the final
        recommendations.
      </p>

      <OverridesSection
        overrides={recommendationsPlayground.overrides}
        setOverrides={(value) =>
          setRecommendationsPlayground((prev) => ({
            ...prev,
            overrides: value,
          }))
        }
        customProfiles={recommendationsPlayground.customProfiles}
        setCustomProfiles={(value) =>
          setRecommendationsPlayground((prev) => ({
            ...prev,
            customProfiles: value,
          }))
        }
        selectedProfileId={recommendationsPlayground.selectedProfileId}
        setSelectedProfileId={(value) =>
          setRecommendationsPlayground((prev) => ({
            ...prev,
            selectedProfileId: value,
          }))
        }
        isEditingProfile={recommendationsPlayground.isEditingProfile}
        setIsEditingProfile={(value) =>
          setRecommendationsPlayground((prev) => ({
            ...prev,
            isEditingProfile: value,
          }))
        }
      />

      <RecommendationsSection
        recUserId={recommendationsPlayground.recUserId}
        setRecUserId={(value) =>
          setRecommendationsPlayground((prev) => ({
            ...prev,
            recUserId: value,
          }))
        }
        k={recommendationsPlayground.k}
        setK={(value) =>
          setRecommendationsPlayground((prev) => ({ ...prev, k: value }))
        }
        blend={recommendationsPlayground.blend}
        setBlend={(value) =>
          setRecommendationsPlayground((prev) => ({ ...prev, blend: value }))
        }
        namespace={namespace}
        exampleUser={exampleUser}
        recOut={recommendationsPlayground.recOut}
        setRecOut={(value) =>
          setRecommendationsPlayground((prev) => ({ ...prev, recOut: value }))
        }
        recLoading={recommendationsPlayground.recLoading}
        setRecLoading={(value) =>
          setRecommendationsPlayground((prev) => ({
            ...prev,
            recLoading: value,
          }))
        }
        overrides={recommendationsPlayground.overrides}
      />

      <SimilarItemsSection
        simItemId={recommendationsPlayground.simItemId}
        setSimItemId={(value) =>
          setRecommendationsPlayground((prev) => ({
            ...prev,
            simItemId: value,
          }))
        }
        k={recommendationsPlayground.k}
        setK={(value) =>
          setRecommendationsPlayground((prev) => ({ ...prev, k: value }))
        }
        namespace={namespace}
        exampleItem={exampleItem}
        simOut={recommendationsPlayground.simOut}
        setSimOut={(value) =>
          setRecommendationsPlayground((prev) => ({ ...prev, simOut: value }))
        }
        simLoading={recommendationsPlayground.simLoading}
        setSimLoading={(value) =>
          setRecommendationsPlayground((prev) => ({
            ...prev,
            simLoading: value,
          }))
        }
      />
    </div>
  );
}
