import React, { useState } from "react";
import {
  Section,
  Row,
  Label,
  TextInput,
  NumberInput,
  Button,
  ResultsTable,
} from "../primitives/UIComponents";
import { recommend } from "../../services/api";
import type {
  specs_types_ScoredItem,
  types_RecommendResponse,
  types_Overrides,
} from "../../lib/api-client";
import { ExplainModal } from "../primitives/ExplainModal";
import { SegmentProfileBadge } from "../primitives/SegmentProfileBadge";
import { DiffBlock, useDiffGenerator } from "../primitives/DiffBlock";
import { useToast } from "../../contexts/ToastContext";
import { logger } from "../../utils/logger";
import { WhyItWorks } from "../primitives/WhyItWorks";
import type { AlgorithmBlend } from "../../types/ui";

interface RecommendationsSectionProps {
  recUserId: string;
  setRecUserId: (userId: string) => void;
  k: number;
  setK: (k: number) => void;
  blend: AlgorithmBlend;
  setBlend: (blend: AlgorithmBlend) => void;
  namespace: string;
  exampleUser: string;
  recOut: specs_types_ScoredItem[] | null;
  setRecOut: (items: specs_types_ScoredItem[] | null) => void;
  recLoading: boolean;
  setRecLoading: (loading: boolean) => void;
  overrides: types_Overrides | null;
  recResponse?: types_RecommendResponse | null;
}

export function RecommendationsSection({
  recUserId,
  setRecUserId,
  k,
  setK,
  blend,
  setBlend,
  namespace,
  exampleUser,
  recOut,
  setRecOut,
  recLoading,
  setRecLoading,
  overrides,
  recResponse: _recResponse,
}: RecommendationsSectionProps) {
  const [explainItem, setExplainItem] = useState<specs_types_ScoredItem | null>(
    null
  );
  const [lastResponse, setLastResponse] =
    useState<types_RecommendResponse | null>(null);
  const [baselineBlend, setBaselineBlend] = useState<AlgorithmBlend | null>(
    null
  );
  const toast = useToast();
  const { generateOverrideDiffs } = useDiffGenerator();

  async function runRecommend() {
    const id = recUserId || exampleUser;
    logger.info("recommend.start", { id, k, blend, namespace });
    setRecLoading(true);
    setRecOut(null);
    setLastResponse(null);

    // Capture baseline blend for diff comparison
    if (!baselineBlend) {
      setBaselineBlend({ ...blend });
    }

    try {
      const r: types_RecommendResponse = await recommend({
        user_id: id,
        namespace,
        k,
        blend,
        overrides: overrides || undefined,
      });
      setRecOut(r.items ?? []);
      setLastResponse(r);
      logger.info("recommend.success", {
        id,
        k,
        numItems: r.items?.length ?? 0,
        profile: r.profile_id,
        segment: r.segment_id,
      });
    } catch (e: any) {
      const msg = e?.message || "Failed to get recommendations";
      setRecOut([{ item_id: `Error: ${msg}`, score: 0 }]);
      toast.showError(msg);
      logger.error("recommend.error", { id, k, error: msg });
    } finally {
      setRecLoading(false);
    }
  }

  return (
    <Section title="Recommendations">
      <Row>
        <Label text="User ID (leave blank to use first generated)">
          <TextInput
            placeholder={exampleUser}
            value={recUserId}
            onChange={(e) => setRecUserId(e.target.value)}
          />
        </Label>
        <Label text="Top-K">
          <NumberInput
            min={1}
            value={k}
            onChange={(e) => setK(Number(e.target.value))}
          />
        </Label>
      </Row>

      <div style={{ height: 8 }} />
      {lastResponse && (
        <WhyItWorks
          metrics={[
            { label: "Profile", value: lastResponse.profile_id },
            { label: "Segment", value: lastResponse.segment_id },
            { label: "Items Returned", value: lastResponse.items?.length },
          ]}
        />
      )}

      <Row>
        <Label text="Blend: popularity">
          <NumberInput
            step="0.1"
            value={blend.pop}
            onChange={(e) =>
              setBlend({ ...blend, pop: Number(e.target.value) })
            }
          />
        </Label>
        <Label text="Blend: co-visitation">
          <NumberInput
            step="0.1"
            value={blend.cooc}
            onChange={(e) =>
              setBlend({ ...blend, cooc: Number(e.target.value) })
            }
          />
        </Label>
        <Label text="Blend: embeddings (als)">
          <NumberInput
            step="0.1"
            value={blend.als}
            onChange={(e) =>
              setBlend({ ...blend, als: Number(e.target.value) })
            }
          />
        </Label>
      </Row>

      <div style={{ height: 8 }} />
      <Button
        onClick={runRecommend}
        disabled={recLoading}
        aria-label="Get recommendations"
      >
        {recLoading ? "Running..." : "Get recommendations"}
      </Button>

      {/* Show segment/profile badges if available */}
      {(lastResponse?.segment_id || lastResponse?.profile_id) && (
        <div style={{ marginTop: 12, marginBottom: 8 }}>
          <SegmentProfileBadge
            segmentId={lastResponse.segment_id}
            profileId={lastResponse.profile_id}
          />
        </div>
      )}

      {/* Show diff block if overrides were applied */}
      {overrides && baselineBlend && (
        <DiffBlock
          title="Override Changes Applied"
          diffs={generateOverrideDiffs(
            baselineBlend,
            blend,
            "Algorithm Parameters"
          )}
          compact={true}
          showExport={true}
          onExport={(diffs) => {
            const exportData = {
              timestamp: new Date().toISOString(),
              namespace,
              user_id: recUserId || exampleUser,
              changes: diffs,
              original_blend: baselineBlend,
              modified_blend: blend,
            };
            const blob = new Blob([JSON.stringify(exportData, null, 2)], {
              type: "application/json",
            });
            const url = URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = url;
            a.download = `override-changes-${
              new Date().toISOString().split("T")[0]
            }.json`;
            a.click();
            URL.revokeObjectURL(url);
            toast.showSuccess("Changes exported successfully");
          }}
        />
      )}

      <div style={{ height: 8 }} />
      {!recLoading && (!recOut || recOut.length === 0) ? (
        <div
          role="status"
          aria-live="polite"
          style={{ fontSize: 12, opacity: 0.7 }}
        >
          No results yet. Run a recommendation to see items.
        </div>
      ) : null}

      <ResultsTable
        items={recOut}
        showExplain
        onExplain={(it) => setExplainItem(it)}
        blend={blend}
      />

      <ExplainModal
        open={!!explainItem}
        item={explainItem}
        blend={blend}
        onClose={() => setExplainItem(null)}
      />
    </Section>
  );
}
