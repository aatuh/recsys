import React, { useState } from "react";
import { Section, Row, Button } from "./UIComponents";
import {
  ConfigService,
  type specs_types_Segment,
  type specs_types_SegmentProfile,
  type specs_types_SegmentRule,
} from "../lib/api-client";

interface SegmentsSectionProps {
  namespace: string;
  segments: specs_types_Segment[];
  profiles: specs_types_SegmentProfile[];
  onSegmentsChange: () => void;
}

interface SegmentWithRules extends specs_types_Segment {
  rules?: specs_types_SegmentRule[];
}

// Simple table component for segment data
function SimpleTable({
  data,
  columns,
}: {
  data: any[];
  columns: { key: string; label: string }[];
}) {
  if (data.length === 0) return null;

  return (
    <div style={{ overflowX: "auto" }}>
      <table
        style={{
          borderCollapse: "collapse",
          width: "100%",
          border: "1px solid #ddd",
        }}
      >
        <thead>
          <tr style={{ backgroundColor: "#f5f5f5" }}>
            {columns.map((col) => (
              <th
                key={col.key}
                style={{
                  padding: "8px 12px",
                  textAlign: "left",
                  border: "1px solid #ddd",
                  fontWeight: "bold",
                }}
              >
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((row, i) => (
            <tr key={i}>
              {columns.map((col) => (
                <td
                  key={col.key}
                  style={{
                    padding: "8px 12px",
                    border: "1px solid #ddd",
                    fontSize: "14px",
                  }}
                >
                  {row[col.key]}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export function SegmentsSection({
  namespace,
  segments,
  profiles: _profiles,
  onSegmentsChange,
}: SegmentsSectionProps) {
  const [loading, _setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [creatingExamples, setCreatingExamples] = useState(false);

  const createExampleSegments = async () => {
    setCreatingExamples(true);
    setError(null);

    try {
      // Create example profile
      await ConfigService.segmentProfilesUpsert({
        profiles: [
          {
            profile_id: "example-vip",
            description: "Example VIP profile with high novelty",
            blend_alpha: 0.8,
            blend_beta: 0.2,
            blend_gamma: 0.4,
            mmr_lambda: 0.6,
            brand_cap: 2,
            category_cap: 3,
            profile_boost: 0.25,
            profile_window_days: 30,
            profile_top_n: 12,
            half_life_days: 3,
            co_vis_window_days: 14,
            purchased_window_days: 7,
            rule_exclude_events: true,
          },
        ],
      });

      // Create example segment
      await ConfigService.segmentsUpsert({
        namespace,
        segment: {
          segment_id: "example-vip",
          name: "Example VIP Users",
          description: "Example segment for VIP users",
          priority: 100,
          active: true,
          profile_id: "example-vip",
          rules: [
            {
              enabled: true,
              rule: {
                any: [
                  { eq: ["user.tier", "VIP"] },
                  { eq: ["user.tier", "vip"] },
                  { eq: ["ctx.surface", "homepage"] },
                ],
              },
            },
          ],
        },
      });

      // Refresh the data
      onSegmentsChange();
    } catch (e: any) {
      setError(`Failed to create examples: ${e.message}`);
    } finally {
      setCreatingExamples(false);
    }
  };

  const formatSegmentTableData = (segments: SegmentWithRules[]) => {
    return segments.map((segment) => ({
      id: segment.segment_id || "N/A",
      name: segment.name || "N/A",
      description: segment.description || "N/A",
      priority: segment.priority || "N/A",
      active: segment.active ? "Yes" : "No",
      profile_id: segment.profile_id || "N/A",
      rules_count: segment.rules?.length || 0,
      created_at: segment.created_at
        ? new Date(segment.created_at).toLocaleDateString()
        : "N/A",
    }));
  };

  const formatRuleTableData = (rules: specs_types_SegmentRule[]) => {
    return rules.map((rule, index) => ({
      index: index + 1,
      enabled: rule.enabled ? "Yes" : "No",
      rule_expr: rule.rule ? JSON.stringify(rule.rule, null, 2) : "N/A",
    }));
  };

  const segmentColumns = [
    { key: "id", label: "Segment ID" },
    { key: "name", label: "Name" },
    { key: "description", label: "Description" },
    { key: "priority", label: "Priority" },
    { key: "active", label: "Active" },
    { key: "profile_id", label: "Profile ID" },
    { key: "rules_count", label: "Rules" },
    { key: "created_at", label: "Created" },
  ];

  const ruleColumns = [
    { key: "index", label: "#" },
    { key: "enabled", label: "Enabled" },
    { key: "rule_expr", label: "Rule Expression" },
  ];

  return (
    <Section title="Segments Management">
      <p style={{ color: "#666", marginBottom: 16, fontSize: "14px" }}>
        Create and manage segment definitions. Segments use rules to match users
        and map them to specific profile configurations.
      </p>

      <Row>
        <div style={{ display: "flex", gap: 8, marginBottom: 16 }}>
          <Button
            onClick={createExampleSegments}
            disabled={creatingExamples}
            style={{
              backgroundColor: "#4caf50",
              color: "white",
            }}
          >
            {creatingExamples ? "Creating..." : "Create Example Segments"}
          </Button>
          <Button onClick={onSegmentsChange} disabled={loading}>
            {loading ? "Loading..." : "Refresh"}
          </Button>
        </div>
      </Row>

      {error && (
        <div
          style={{
            marginBottom: 16,
            padding: 12,
            backgroundColor: "#fee",
            border: "1px solid #fcc",
            borderRadius: 4,
            color: "#c33",
          }}
        >
          Error: {error}
        </div>
      )}

      <div>
        <h4 style={{ margin: "0 0 16px 0", color: "#333" }}>
          Segments ({segments.length})
        </h4>

        {segments.length === 0 ? (
          <div
            style={{
              padding: 24,
              textAlign: "center",
              color: "#666",
              backgroundColor: "#f9f9f9",
              border: "1px solid #ddd",
              borderRadius: 4,
            }}
          >
            No segments found. Click "Create Example Segments" to get started.
          </div>
        ) : (
          <SimpleTable
            data={formatSegmentTableData(segments)}
            columns={segmentColumns}
          />
        )}

        {/* Show rules for each segment */}
        {segments.map(
          (segment) =>
            segment.rules &&
            segment.rules.length > 0 && (
              <div key={segment.segment_id} style={{ marginTop: 24 }}>
                <h5 style={{ margin: "0 0 8px 0", color: "#333" }}>
                  Rules for Segment: {segment.name || segment.segment_id}
                </h5>
                <SimpleTable
                  data={formatRuleTableData(segment.rules)}
                  columns={ruleColumns}
                />
              </div>
            )
        )}
      </div>
    </Section>
  );
}
