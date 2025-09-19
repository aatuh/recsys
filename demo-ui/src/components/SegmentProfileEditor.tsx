import React, { useState, useEffect } from "react";
import { Section, Row, Button } from "./UIComponents";
import {
  ConfigService,
  type specs_types_Segment,
  type specs_types_SegmentProfile,
  type specs_types_SegmentRule,
} from "../lib/api-client";

interface SegmentProfileEditorProps {
  namespace: string;
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

export function SegmentProfileEditor({ namespace }: SegmentProfileEditorProps) {
  const [segments, setSegments] = useState<SegmentWithRules[]>([]);
  const [profiles, setProfiles] = useState<specs_types_SegmentProfile[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"segments" | "profiles">(
    "segments"
  );

  const loadData = async () => {
    setLoading(true);
    setError(null);

    try {
      const [segmentsResponse, profilesResponse] = await Promise.all([
        ConfigService.getV1Segments(namespace),
        ConfigService.getV1SegmentProfiles(namespace),
      ]);

      setSegments(segmentsResponse.segments || []);
      setProfiles(profilesResponse.profiles || []);
    } catch (e: any) {
      setError(e.message || "Failed to load segments and profiles");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [namespace]);

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

  const formatProfileTableData = (profiles: specs_types_SegmentProfile[]) => {
    return profiles.map((profile) => ({
      id: profile.profile_id || "N/A",
      description: profile.description || "N/A",
      blend_alpha: profile.blend_alpha?.toFixed(2) || "N/A",
      blend_beta: profile.blend_beta?.toFixed(2) || "N/A",
      blend_gamma: profile.blend_gamma?.toFixed(2) || "N/A",
      mmr_lambda: profile.mmr_lambda?.toFixed(2) || "N/A",
      brand_cap: profile.brand_cap || "N/A",
      category_cap: profile.category_cap || "N/A",
      profile_boost: profile.profile_boost?.toFixed(2) || "N/A",
      created_at: profile.created_at
        ? new Date(profile.created_at).toLocaleDateString()
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

  const profileColumns = [
    { key: "id", label: "Profile ID" },
    { key: "description", label: "Description" },
    { key: "blend_alpha", label: "α (Pop)" },
    { key: "blend_beta", label: "β (Co-vis)" },
    { key: "blend_gamma", label: "γ (Embed)" },
    { key: "mmr_lambda", label: "MMR λ" },
    { key: "brand_cap", label: "Brand Cap" },
    { key: "category_cap", label: "Cat Cap" },
    { key: "profile_boost", label: "Boost" },
    { key: "created_at", label: "Created" },
  ];

  const ruleColumns = [
    { key: "index", label: "#" },
    { key: "enabled", label: "Enabled" },
    { key: "rule_expr", label: "Rule Expression" },
  ];

  return (
    <Section title="Segment Profiles & Rules">
      <p style={{ color: "#666", marginBottom: 16, fontSize: "14px" }}>
        View and manage segment definitions and their associated profile
        configurations. This is currently read-only - use the API directly for
        modifications.
      </p>

      <Row>
        <div style={{ display: "flex", gap: 8, marginBottom: 16 }}>
          <Button
            onClick={() => setActiveTab("segments")}
            style={{
              backgroundColor: activeTab === "segments" ? "#2196f3" : "#f5f5f5",
              color: activeTab === "segments" ? "white" : "#333",
            }}
          >
            Segments ({segments.length})
          </Button>
          <Button
            onClick={() => setActiveTab("profiles")}
            style={{
              backgroundColor: activeTab === "profiles" ? "#2196f3" : "#f5f5f5",
              color: activeTab === "profiles" ? "white" : "#333",
            }}
          >
            Profiles ({profiles.length})
          </Button>
        </div>

        <Button onClick={loadData} disabled={loading}>
          {loading ? "Loading..." : "Refresh"}
        </Button>
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

      {activeTab === "segments" && (
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
              No segments found. Create segments using the API.
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
      )}

      {activeTab === "profiles" && (
        <div>
          <h4 style={{ margin: "0 0 16px 0", color: "#333" }}>
            Profiles ({profiles.length})
          </h4>

          {profiles.length === 0 ? (
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
              No profiles found. Create profiles using the API.
            </div>
          ) : (
            <SimpleTable
              data={formatProfileTableData(profiles)}
              columns={profileColumns}
            />
          )}
        </div>
      )}
    </Section>
  );
}
