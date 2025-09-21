import React from "react";
import {
  Section,
  Row,
  Label,
  TextInput,
  NumberInput,
  Button,
  ResultsTable,
} from "../primitives/UIComponents";
import { similar } from "../../services/apiService";
import { useToast } from "../../ui/Toast";
import { logger } from "../../utils/logger";

import type { specs_types_ScoredItem } from "../../lib/api-client";

interface SimilarItemsSectionProps {
  simItemId: string;
  setSimItemId: (value: string) => void;
  k: number;
  setK: (value: number) => void;
  namespace: string;
  exampleItem: string;
  simOut: specs_types_ScoredItem[] | null;
  setSimOut: (value: specs_types_ScoredItem[] | null) => void;
  simLoading: boolean;
  setSimLoading: (value: boolean) => void;
}

export function SimilarItemsSection({
  simItemId,
  setSimItemId,
  k,
  setK,
  namespace,
  exampleItem,
  simOut,
  setSimOut,
  simLoading,
  setSimLoading,
}: SimilarItemsSectionProps) {
  const toast = useToast();
  async function runSimilar() {
    const id = simItemId || exampleItem;
    logger.info("similar.start", { id, k, namespace });
    setSimLoading(true);
    setSimOut(null);
    try {
      const r = await similar(id, namespace, k);
      setSimOut(r);
      logger.info("similar.success", { id, k, numItems: r?.length ?? 0 });
    } catch (e: any) {
      const msg = e?.message || "Failed to get similar items";
      setSimOut([{ item_id: `Error: ${msg}`, score: 0 }]);
      toast.error(msg);
      logger.error("similar.error", { id, k, error: msg });
    } finally {
      setSimLoading(false);
    }
  }

  return (
    <Section title="Similar items">
      <Row>
        <Label text="Item ID (leave blank to use first generated)">
          <TextInput
            placeholder={exampleItem}
            value={simItemId}
            onChange={(e) => setSimItemId(e.target.value)}
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
      <Button
        onClick={runSimilar}
        disabled={simLoading}
        aria-label="Get similar items"
      >
        {simLoading ? "Running..." : "Get similar"}
      </Button>

      <div style={{ height: 8 }} />
      {!simLoading && (!simOut || simOut.length === 0) ? (
        <div
          role="status"
          aria-live="polite"
          style={{ fontSize: 12, opacity: 0.7 }}
        >
          No results yet. Run a similar-items query to see items.
        </div>
      ) : null}

      <ResultsTable items={simOut} />
    </Section>
  );
}
