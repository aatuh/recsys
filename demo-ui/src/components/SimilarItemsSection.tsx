import React from "react";
import {
  Section,
  Row,
  Label,
  TextInput,
  NumberInput,
  Button,
  ResultsTable,
} from "./UIComponents";
import { similar } from "../services/apiService";
import type { internal_http_types_ScoredItem } from "../lib/api-client";

interface SimilarItemsSectionProps {
  simItemId: string;
  setSimItemId: (itemId: string) => void;
  k: number;
  setK: (k: number) => void;
  namespace: string;
  exampleItem: string;
  simOut: internal_http_types_ScoredItem[] | null;
  setSimOut: (items: internal_http_types_ScoredItem[] | null) => void;
  simLoading: boolean;
  setSimLoading: (loading: boolean) => void;
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
  async function runSimilar() {
    const id = simItemId || exampleItem;
    setSimLoading(true);
    setSimOut(null);
    try {
      const r = await similar(id, namespace, k);
      setSimOut(r);
    } catch (e: any) {
      setSimOut([{ item_id: `Error: ${e.message}`, score: 0 }]);
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
      <Button onClick={runSimilar} disabled={simLoading}>
        {simLoading ? "Running..." : "Get similar"}
      </Button>

      <div style={{ height: 8 }} />
      <ResultsTable items={simOut} />
    </Section>
  );
}
