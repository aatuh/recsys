import React from "react";
import { Button } from "../primitives/UIComponents";
import { DiffBlock } from "../primitives/DiffBlock";
import type {
  RuleDryRunRequest,
  RuleDryRunResult,
  ScoredItem,
} from "../../types/ui";

export function RuleDryRunModal(props: {
  open: boolean;
  loading: boolean;
  data: RuleDryRunResult | null;
  form: RuleDryRunRequest;
  setForm: (v: RuleDryRunRequest) => void;
  onRun: () => void;
  onClose: () => void;
}) {
  const { form, setForm, data } = props;

  if (!props.open) return null;

  // Generate diffs from dry run results
  const generateDiffsFromResults = () => {
    if (!data || !data.original_items || !data.filtered_items) return [];

    const originalItems = data.original_items;
    const filteredItems = data.filtered_items;

    const diffs = [];

    // Check for blocked items
    const blockedItems = originalItems.filter(
      (item: ScoredItem) =>
        !filteredItems.some(
          (filtered: ScoredItem) => filtered.item_id === item.item_id
        )
    );

    if (blockedItems.length > 0) {
      diffs.push({
        field: "Blocked Items",
        before: blockedItems.map((item: ScoredItem) => item.item_id).join(", "),
        after: "—",
        reason: "Items blocked by rules",
      });
    }

    // Check for boosted items
    const boostedItems = filteredItems.filter(
      (item: ScoredItem) => item.boost_value && item.boost_value > 0
    );

    if (boostedItems.length > 0) {
      boostedItems.forEach((item: ScoredItem) => {
        diffs.push({
          field: `Boosted: ${item.item_id}`,
          before: "0.00",
          after: item.boost_value?.toFixed(2) || "0.00",
          reason: "Item boosted by rules",
        });
      });
    }

    // Check for pinned items
    const pinnedItems = filteredItems.filter(
      (item: ScoredItem) => item.pinned === true
    );

    if (pinnedItems.length > 0) {
      diffs.push({
        field: "Pinned Items",
        before: "—",
        after: pinnedItems.map((item: ScoredItem) => item.item_id).join(", "),
        reason: "Items pinned by rules",
      });
    }

    return diffs;
  };
  return (
    <div
      style={{
        position: "fixed",
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: "rgba(0,0,0,0.5)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 1000,
      }}
    >
      <div
        style={{
          backgroundColor: "white",
          padding: 24,
          borderRadius: 8,
          width: "90%",
          maxWidth: 800,
          maxHeight: "90vh",
          overflowY: "auto",
        }}
      >
        <h3 style={{ margin: "0 0 16px 0" }}>Rule Dry Run</h3>
        <div style={{ display: "grid", gap: 12, marginBottom: 16 }}>
          <div>
            <label
              style={{ display: "block", marginBottom: 4, fontWeight: "bold" }}
            >
              Surface *
            </label>
            <input
              type="text"
              value={form.surface}
              onChange={(e) => setForm({ ...form, surface: e.target.value })}
              style={{
                width: "100%",
                padding: 8,
                border: "1px solid #ddd",
                borderRadius: 4,
              }}
              placeholder="e.g., home, gamepage"
            />
          </div>
          <div>
            <label
              style={{ display: "block", marginBottom: 4, fontWeight: "bold" }}
            >
              Segment ID
            </label>
            <input
              type="text"
              value={form.segment_id}
              onChange={(e) => setForm({ ...form, segment_id: e.target.value })}
              style={{
                width: "100%",
                padding: 8,
                border: "1px solid #ddd",
                borderRadius: 4,
              }}
              placeholder="Optional"
            />
          </div>
          <div>
            <label
              style={{ display: "block", marginBottom: 4, fontWeight: "bold" }}
            >
              Candidate Item IDs *
            </label>
            <input
              type="text"
              value={form.items.join(", ")}
              onChange={(e) =>
                setForm({
                  ...form,
                  items: e.target.value
                    .split(",")
                    .map((s) => s.trim())
                    .filter((s) => s),
                })
              }
              style={{
                width: "100%",
                padding: 8,
                border: "1px solid #ddd",
                borderRadius: 4,
              }}
              placeholder="item1, item2, item3"
            />
          </div>
        </div>
        <div style={{ display: "flex", gap: 8, marginBottom: 16 }}>
          <Button
            onClick={props.onRun}
            disabled={props.loading || !form.surface || form.items.length === 0}
            style={{ backgroundColor: "#17a2b8", color: "white" }}
          >
            {props.loading ? "Running..." : "Run Dry Run"}
          </Button>
          <Button
            onClick={props.onClose}
            style={{ backgroundColor: "#6c757d", color: "white" }}
          >
            Close
          </Button>
        </div>
        {data && (
          <div>
            <h4 style={{ margin: "0 0 12px 0" }}>Dry Run Results</h4>

            {/* Show diff block for rule changes */}
            <DiffBlock
              title="Rule Changes Applied"
              diffs={generateDiffsFromResults()}
              showReasons={true}
              compact={false}
              showExport={true}
              onExport={(diffs) => {
                const exportData = {
                  timestamp: new Date().toISOString(),
                  surface: form.surface,
                  segment_id: form.segment_id,
                  candidate_items: form.items,
                  changes: diffs,
                  raw_results: data,
                };
                const blob = new Blob([JSON.stringify(exportData, null, 2)], {
                  type: "application/json",
                });
                const url = URL.createObjectURL(blob);
                const a = document.createElement("a");
                a.href = url;
                a.download = `rule-dry-run-${
                  new Date().toISOString().split("T")[0]
                }.json`;
                a.click();
                URL.revokeObjectURL(url);
              }}
            />

            {/* Raw data for debugging */}
            <details style={{ marginTop: 16 }}>
              <summary
                style={{
                  cursor: "pointer",
                  fontWeight: "bold",
                  marginBottom: 8,
                }}
              >
                Raw Results (Debug)
              </summary>
              <pre
                style={{
                  fontSize: 12,
                  background: "#f6f8fa",
                  border: "1px solid #e1e4e8",
                  padding: 12,
                  borderRadius: 6,
                  overflowX: "auto",
                }}
              >
                {JSON.stringify(data, null, 2)}
              </pre>
            </details>
          </div>
        )}
      </div>
    </div>
  );
}
