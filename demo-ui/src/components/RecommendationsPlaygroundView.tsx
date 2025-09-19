import React, { useMemo, useState, useEffect } from "react";
import {
  RecommendationsSection,
  OverridesSection,
  SimilarItemsSection,
} from "./";
import { SegmentDryRunTool } from "./SegmentDryRunTool";
import { SegmentProfileEditor } from "./SegmentProfileEditor";
import { useViewState } from "../contexts/ViewStateContext";
import {
  ConfigService,
  type specs_types_Segment,
  type specs_types_SegmentProfile,
  type types_RecommendResponse,
} from "../lib/api-client";

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

  const [segments, setSegments] = useState<specs_types_Segment[]>([]);
  const [profiles, setProfiles] = useState<specs_types_SegmentProfile[]>([]);
  const [lastRecResponse] = useState<types_RecommendResponse | null>(null);
  const [activeTab, setActiveTab] = useState<
    "recommendations" | "profiles" | "dry-run"
  >("recommendations");

  const exampleItem = useMemo(() => {
    return generatedItems[0] || "item-0001";
  }, [generatedItems.length]);

  const exampleUser = useMemo(() => {
    return generatedUsers[0] || "user-0001";
  }, [generatedUsers.length]);

  // Load segments and profiles on mount
  const loadSegmentsAndProfiles = async () => {
    try {
      const [segmentsResponse, profilesResponse] = await Promise.all([
        ConfigService.getV1Segments(namespace),
        ConfigService.getV1SegmentProfiles(namespace),
      ]);
      setSegments(segmentsResponse.segments || []);
      setProfiles(profilesResponse.profiles || []);
    } catch (e) {
      console.warn("Failed to load segments/profiles:", e);
    }
  };

  useEffect(() => {
    loadSegmentsAndProfiles();
  }, [namespace]);

  return (
    <div style={{ padding: 16, fontFamily: "system-ui, sans-serif" }}>
      <p style={{ color: "#444", marginBottom: 24 }}>
        Test recommendation algorithms and explore similar items. Adjust blend
        parameters to see how different algorithms contribute to the final
        recommendations. Use the segment tools to test segment matching and view
        profile configurations.
      </p>

      {/* Tab Navigation */}
      <div
        style={{
          display: "flex",
          gap: 8,
          marginBottom: 24,
          borderBottom: "1px solid #ddd",
          paddingBottom: 8,
        }}
      >
        <button
          onClick={() => setActiveTab("recommendations")}
          style={{
            padding: "8px 16px",
            border: "none",
            backgroundColor:
              activeTab === "recommendations" ? "#2196f3" : "#f5f5f5",
            color: activeTab === "recommendations" ? "white" : "#333",
            borderRadius: 4,
            cursor: "pointer",
            fontSize: "14px",
          }}
        >
          Recommendations
        </button>
        <button
          onClick={() => setActiveTab("profiles")}
          style={{
            padding: "8px 16px",
            border: "none",
            backgroundColor: activeTab === "profiles" ? "#2196f3" : "#f5f5f5",
            color: activeTab === "profiles" ? "white" : "#333",
            borderRadius: 4,
            cursor: "pointer",
            fontSize: "14px",
          }}
        >
          Profiles & Rules
        </button>
        <button
          onClick={() => setActiveTab("dry-run")}
          style={{
            padding: "8px 16px",
            border: "none",
            backgroundColor: activeTab === "dry-run" ? "#2196f3" : "#f5f5f5",
            color: activeTab === "dry-run" ? "white" : "#333",
            borderRadius: 4,
            cursor: "pointer",
            fontSize: "14px",
          }}
        >
          Dry-Run Tool
        </button>
      </div>

      {activeTab === "recommendations" && (
        <>
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
              setRecommendationsPlayground((prev) => ({
                ...prev,
                blend: value,
              }))
            }
            namespace={namespace}
            exampleUser={exampleUser}
            recOut={recommendationsPlayground.recOut}
            setRecOut={(value) =>
              setRecommendationsPlayground((prev) => ({
                ...prev,
                recOut: value,
              }))
            }
            recLoading={recommendationsPlayground.recLoading}
            setRecLoading={(value) =>
              setRecommendationsPlayground((prev) => ({
                ...prev,
                recLoading: value,
              }))
            }
            overrides={recommendationsPlayground.overrides}
            recResponse={lastRecResponse}
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
              setRecommendationsPlayground((prev) => ({
                ...prev,
                simOut: value,
              }))
            }
            simLoading={recommendationsPlayground.simLoading}
            setSimLoading={(value) =>
              setRecommendationsPlayground((prev) => ({
                ...prev,
                simLoading: value,
              }))
            }
          />
        </>
      )}

      {activeTab === "profiles" && (
        <SegmentProfileEditor namespace={namespace} />
      )}

      {activeTab === "dry-run" && (
        <SegmentDryRunTool
          namespace={namespace}
          segments={segments}
          profiles={profiles}
        />
      )}
    </div>
  );
}
