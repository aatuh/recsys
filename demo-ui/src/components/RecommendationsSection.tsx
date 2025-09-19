import React, { useState } from "react";
import {
  Section,
  Row,
  Label,
  TextInput,
  NumberInput,
  Button,
  ResultsTable,
} from "./UIComponents";
import { recommend } from "../services/apiService";
import type {
  specs_types_ScoredItem,
  types_RecommendResponse,
  types_Overrides,
} from "../lib/api-client";
import { ExplainModal } from "./ExplainModal";
import { SegmentProfileBadge } from "./SegmentProfileBadge";

interface RecommendationsSectionProps {
  recUserId: string;
  setRecUserId: (userId: string) => void;
  k: number;
  setK: (k: number) => void;
  blend: { pop: number; cooc: number; als: number };
  setBlend: (blend: { pop: number; cooc: number; als: number }) => void;
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

  async function runRecommend() {
    const id = recUserId || exampleUser;
    setRecLoading(true);
    setRecOut(null);
    setLastResponse(null);
    try {
      const r: types_RecommendResponse = await recommend(
        id,
        namespace,
        k,
        blend,
        overrides
      );
      setRecOut(r.items ?? []);
      setLastResponse(r);
    } catch (e: any) {
      setRecOut([{ item_id: `Error: ${e.message}`, score: 0 }]);
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
      <Button onClick={runRecommend} disabled={recLoading}>
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

      <div style={{ height: 8 }} />
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
