import { useState } from "react";
import { Section, Row, Label, TextInput, Button } from "../primitives/UIComponents";
import {
  type types_SegmentDryRunRequest,
  type types_SegmentDryRunResponse,
  type specs_types_Segment,
  type specs_types_SegmentProfile,
  ConfigService,
} from "../../lib/api-client";

interface SegmentDryRunToolProps {
  namespace: string;
  segments: specs_types_Segment[];
  profiles: specs_types_SegmentProfile[];
}

interface DryRunResult {
  request: types_SegmentDryRunRequest;
  response: types_SegmentDryRunResponse;
  matchedSegment?: specs_types_Segment;
  matchedProfile?: specs_types_SegmentProfile;
}

export function SegmentDryRunTool({
  namespace,
  segments,
  profiles,
}: SegmentDryRunToolProps) {
  const [userId, setUserId] = useState("");
  const [surface, setSurface] = useState("homepage");
  const [device, setDevice] = useState("mobile");
  const [locale, setLocale] = useState("en-US");
  const [userTier, setUserTier] = useState("");
  const [region, setRegion] = useState("");
  const [customTraits, setCustomTraits] = useState("");
  const [customContext, setCustomContext] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<DryRunResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  const parseJsonSafely = (jsonStr: string): Record<string, any> => {
    if (!jsonStr.trim()) return {};
    try {
      return JSON.parse(jsonStr);
    } catch (e) {
      throw new Error(`Invalid JSON: ${e}`);
    }
  };

  const runDryRun = async () => {
    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const traits: Record<string, any> = {};
      if (userTier) traits.tier = userTier;
      if (region) traits.region = region;

      // Parse custom traits JSON
      if (customTraits.trim()) {
        const customTraitsObj = parseJsonSafely(customTraits);
        Object.assign(traits, customTraitsObj);
      }

      const context: Record<string, any> = {
        surface,
        device,
        locale,
      };

      // Parse custom context JSON
      if (customContext.trim()) {
        const customContextObj = parseJsonSafely(customContext);
        Object.assign(context, customContextObj);
      }

      const request: types_SegmentDryRunRequest = {
        namespace,
        user_id: userId || undefined,
        traits: Object.keys(traits).length > 0 ? traits : undefined,
        context: Object.keys(context).length > 0 ? context : undefined,
      };

      const response = await ConfigService.segmentsDryRun(request);

      // Debug: Log the request and response
      console.log("Dry-run request:", request);
      console.log("Dry-run response:", response);
      console.log(
        "Available segments:",
        segments.map((s) => ({
          id: s.segment_id,
          name: s.name,
          rules: s.rules,
        }))
      );

      // Find matched segment and profile
      const matchedSegment = segments.find(
        (s) => s.segment_id === response.segment_id
      );
      const matchedProfile = profiles.find(
        (p) => p.profile_id === response.profile_id
      );

      setResult({
        request,
        response,
        matchedSegment,
        matchedProfile,
      });
    } catch (e: any) {
      setError(e.message || "Failed to run dry-run");
    } finally {
      setLoading(false);
    }
  };

  const formatProfileConfig = (profile: specs_types_SegmentProfile) => {
    const config: Record<string, any> = {};

    if (profile.blend_alpha !== undefined)
      config["α (popularity)"] = profile.blend_alpha;
    if (profile.blend_beta !== undefined)
      config["β (co-visitation)"] = profile.blend_beta;
    if (profile.blend_gamma !== undefined)
      config["γ (embeddings)"] = profile.blend_gamma;
    if (profile.mmr_lambda !== undefined) config["MMR λ"] = profile.mmr_lambda;
    if (profile.brand_cap !== undefined)
      config["Brand cap"] = profile.brand_cap;
    if (profile.category_cap !== undefined)
      config["Category cap"] = profile.category_cap;
    if (profile.profile_boost !== undefined)
      config["Profile boost"] = profile.profile_boost;
    if (profile.profile_top_n !== undefined)
      config["Profile top-N"] = profile.profile_top_n;
    if (profile.profile_window_days !== undefined)
      config["Profile window (days)"] = profile.profile_window_days;
    if (profile.half_life_days !== undefined)
      config["Half-life (days)"] = profile.half_life_days;
    if (profile.co_vis_window_days !== undefined)
      config["Co-vis window (days)"] = profile.co_vis_window_days;
    if (profile.purchased_window_days !== undefined)
      config["Purchased window (days)"] = profile.purchased_window_days;
    if (profile.popularity_fanout !== undefined)
      config["Popularity fanout"] = profile.popularity_fanout;
    if (profile.rule_exclude_events !== undefined)
      config["Exclude events"] = profile.rule_exclude_events;

    return config;
  };

  return (
    <Section title="Segment Dry-Run Tool">
      <p style={{ color: "#666", marginBottom: 16, fontSize: "14px" }}>
        Test segment matching by providing request context and user traits. See
        which segment matches and what profile configuration would be used.
      </p>

      {/* Show current segments/profiles count */}
      <div
        style={{
          marginBottom: 16,
          padding: 12,
          backgroundColor: "#f0f8ff",
          border: "1px solid #b3d9ff",
          borderRadius: 4,
          fontSize: "14px",
        }}
      >
        <strong>Current Configuration:</strong> {segments.length} segments,{" "}
        {profiles.length} profiles
        {segments.length === 0 && profiles.length === 0 && (
          <div style={{ marginTop: 8, color: "#666" }}>
            No segments or profiles configured. Go to the "Segments" tab to
            create them.
          </div>
        )}
      </div>

      <Row>
        <Label text="User ID (optional)">
          <TextInput
            placeholder="user-0001"
            value={userId}
            onChange={(e) => setUserId(e.target.value)}
          />
        </Label>
        <Label text="Surface">
          <TextInput
            placeholder="homepage"
            value={surface}
            onChange={(e) => setSurface(e.target.value)}
          />
        </Label>
      </Row>

      <Row>
        <Label text="Device">
          <TextInput
            placeholder="mobile"
            value={device}
            onChange={(e) => setDevice(e.target.value)}
          />
        </Label>
        <Label text="Locale">
          <TextInput
            placeholder="en-US"
            value={locale}
            onChange={(e) => setLocale(e.target.value)}
          />
        </Label>
      </Row>

      <Row>
        <Label text="User Tier (trait)">
          <TextInput
            placeholder="VIP (try this to match example segment)"
            value={userTier}
            onChange={(e) => setUserTier(e.target.value)}
          />
        </Label>
        <Label text="Region (trait)">
          <TextInput
            placeholder="FI, US, DE"
            value={region}
            onChange={(e) => setRegion(e.target.value)}
          />
        </Label>
      </Row>

      <div style={{ marginTop: 16 }}>
        <Label text="Custom Traits (JSON)">
          <textarea
            style={{
              width: "100%",
              minHeight: "80px",
              padding: "8px",
              border: "1px solid #ddd",
              borderRadius: "4px",
              fontFamily: "monospace",
              fontSize: "12px",
            }}
            placeholder='{"ltv_eur": 500, "last_play_days": 7}'
            value={customTraits}
            onChange={(e) => setCustomTraits(e.target.value)}
          />
        </Label>
      </div>

      <div style={{ marginTop: 16 }}>
        <Label text="Custom Context (JSON)">
          <textarea
            style={{
              width: "100%",
              minHeight: "80px",
              padding: "8px",
              border: "1px solid #ddd",
              borderRadius: "4px",
              fontFamily: "monospace",
              fontSize: "12px",
            }}
            placeholder='{"time_of_day": "evening", "campaign_id": "summer2024"}'
            value={customContext}
            onChange={(e) => setCustomContext(e.target.value)}
          />
        </Label>
      </div>

      <div style={{ height: 16 }} />
      <Button onClick={runDryRun} disabled={loading}>
        {loading ? "Running..." : "Run Dry-Run"}
      </Button>

      {error && (
        <div
          style={{
            marginTop: 16,
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

      {/* Show available segments for debugging */}
      {segments.length > 0 && (
        <div style={{ marginTop: 24 }}>
          <h4 style={{ margin: "0 0 16px 0", color: "#333" }}>
            Available Segments ({segments.length})
          </h4>
          <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
            {segments.map((segment) => (
              <div
                key={segment.segment_id}
                style={{
                  padding: "8px 12px",
                  backgroundColor: "#f8f9fa",
                  border: "1px solid #dee2e6",
                  borderRadius: 4,
                  fontSize: "12px",
                }}
              >
                <strong>{segment.name}</strong> ({segment.segment_id})
                <br />
                <span style={{ color: "#666" }}>
                  Rules: {JSON.stringify(segment.rules, null, 2)}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {result && (
        <div style={{ marginTop: 24 }}>
          <h4 style={{ margin: "0 0 16px 0", color: "#333" }}>
            Dry-Run Results
          </h4>

          <div
            style={{
              display: "flex",
              gap: 16,
              marginBottom: 16,
              flexWrap: "wrap",
            }}
          >
            <div
              style={{
                padding: "8px 12px",
                backgroundColor: result.response.matched
                  ? "#e8f5e8"
                  : "#f5f5f5",
                border: `1px solid ${
                  result.response.matched ? "#4caf50" : "#ddd"
                }`,
                borderRadius: 4,
                fontSize: "14px",
              }}
            >
              <strong>Matched:</strong> {result.response.matched ? "Yes" : "No"}
            </div>

            {result.response.segment_id && (
              <div
                style={{
                  padding: "8px 12px",
                  backgroundColor: "#e3f2fd",
                  border: "1px solid #2196f3",
                  borderRadius: 4,
                  fontSize: "14px",
                }}
              >
                <strong>Segment:</strong> {result.response.segment_id}
              </div>
            )}

            {result.response.profile_id && (
              <div
                style={{
                  padding: "8px 12px",
                  backgroundColor: "#f3e5f5",
                  border: "1px solid #9c27b0",
                  borderRadius: 4,
                  fontSize: "14px",
                }}
              >
                <strong>Profile:</strong> {result.response.profile_id}
              </div>
            )}
          </div>

          {result.matchedSegment && (
            <div style={{ marginBottom: 16 }}>
              <h5 style={{ margin: "0 0 8px 0", color: "#333" }}>
                Matched Segment Details
              </h5>
              <div
                style={{
                  padding: 12,
                  backgroundColor: "#f9f9f9",
                  border: "1px solid #ddd",
                  borderRadius: 4,
                  fontSize: "14px",
                }}
              >
                <div>
                  <strong>Name:</strong> {result.matchedSegment.name || "N/A"}
                </div>
                <div>
                  <strong>Description:</strong>{" "}
                  {result.matchedSegment.description || "N/A"}
                </div>
                <div>
                  <strong>Priority:</strong>{" "}
                  {result.matchedSegment.priority || "N/A"}
                </div>
                <div>
                  <strong>Active:</strong>{" "}
                  {result.matchedSegment.active ? "Yes" : "No"}
                </div>
                <div>
                  <strong>Rules:</strong>{" "}
                  {result.matchedSegment.rules?.length || 0} rule(s)
                </div>
              </div>
            </div>
          )}

          {result.matchedProfile && (
            <div>
              <h5 style={{ margin: "0 0 8px 0", color: "#333" }}>
                Effective Profile Configuration
              </h5>
              <div
                style={{
                  padding: 12,
                  backgroundColor: "#f9f9f9",
                  border: "1px solid #ddd",
                  borderRadius: 4,
                  fontSize: "14px",
                }}
              >
                <div style={{ marginBottom: 8 }}>
                  <strong>Profile ID:</strong>{" "}
                  {result.matchedProfile.profile_id}
                </div>
                {result.matchedProfile.description && (
                  <div style={{ marginBottom: 8 }}>
                    <strong>Description:</strong>{" "}
                    {result.matchedProfile.description}
                  </div>
                )}
                <div style={{ marginTop: 12 }}>
                  <strong>Configuration:</strong>
                  <div
                    style={{
                      marginTop: 8,
                      display: "grid",
                      gridTemplateColumns:
                        "repeat(auto-fit, minmax(200px, 1fr))",
                      gap: "8px",
                    }}
                  >
                    {Object.entries(
                      formatProfileConfig(result.matchedProfile)
                    ).map(([key, value]) => (
                      <div
                        key={key}
                        style={{
                          padding: "4px 8px",
                          backgroundColor: "#fff",
                          border: "1px solid #eee",
                          borderRadius: 2,
                          fontSize: "12px",
                        }}
                      >
                        <strong>{key}:</strong> {String(value)}
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}
    </Section>
  );
}
