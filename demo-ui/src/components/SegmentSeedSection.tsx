import React, { useState } from "react";
import { Section, Row, Button } from "./UIComponents";
import { ConfigService } from "../lib/api-client";

interface SegmentSeedSectionProps {
  namespace: string;
  onLog: (message: string) => void;
}

export function SegmentSeedSection({
  namespace,
  onLog,
}: SegmentSeedSectionProps) {
  const [creating, setCreating] = useState(false);

  const createExampleSegments = async () => {
    setCreating(true);
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
          {
            profile_id: "example-casual",
            description: "Example casual user profile",
            blend_alpha: 0.6,
            blend_beta: 0.3,
            blend_gamma: 0.1,
            mmr_lambda: 0.8,
            brand_cap: 3,
            category_cap: 5,
            profile_boost: 0.1,
            profile_window_days: 14,
            profile_top_n: 20,
            half_life_days: 7,
            co_vis_window_days: 21,
            purchased_window_days: 14,
            rule_exclude_events: false,
          },
        ],
      });
      onLog("✔ Created example segment profiles");

      // Create example segments
      await ConfigService.segmentsUpsert({
        namespace,
        segment: {
          segment_id: "example-vip",
          name: "VIP Users",
          description: "High-value users with premium preferences",
          priority: 100,
          active: true,
          profile_id: "example-vip",
          rules: [
            {
              enabled: true,
              rule: {
                any: [
                  { eq: ["user.traits.tier", "VIP"] },
                  { eq: ["user.traits.tier", "vip"] },
                  { eq: ["user.traits.tier", "premium"] },
                ],
              },
            },
          ],
        },
      });

      await ConfigService.segmentsUpsert({
        namespace,
        segment: {
          segment_id: "example-casual",
          name: "Casual Users",
          description: "Regular users with standard preferences",
          priority: 50,
          active: true,
          profile_id: "example-casual",
          rules: [
            {
              enabled: true,
              rule: {
                any: [
                  { eq: ["user.traits.tier", "free"] },
                  { eq: ["user.traits.tier", "basic"] },
                ],
              },
            },
          ],
        },
      });
      onLog("✔ Created example segments");

      onLog("✅ Segment seeding completed successfully!");
    } catch (error: any) {
      onLog(`❌ Failed to create segments: ${error.message}`);
    } finally {
      setCreating(false);
    }
  };

  return (
    <Section title="Segment Seeding">
      <p style={{ color: "#666", marginBottom: 16, fontSize: "14px" }}>
        Create example segments and profiles for testing. Segments use rules to
        match users and map them to specific recommendation profiles.
      </p>

      <Row>
        <div style={{ display: "flex", gap: 8, marginBottom: 16 }}>
          <Button
            onClick={createExampleSegments}
            disabled={creating}
            style={{
              backgroundColor: "#4caf50",
              color: "white",
            }}
          >
            {creating ? "Creating..." : "Create Example Segments"}
          </Button>
        </div>
      </Row>

      <div
        style={{
          backgroundColor: "#f9f9f9",
          border: "1px solid #ddd",
          borderRadius: 4,
          padding: 12,
          fontSize: "12px",
          color: "#666",
        }}
      >
        <strong>What this creates:</strong>
        <ul style={{ margin: "8px 0", paddingLeft: "20px" }}>
          <li>
            <strong>VIP Profile:</strong> High novelty (γ=0.4), low diversity
            (λ=0.6), brand cap=2
          </li>
          <li>
            <strong>Casual Profile:</strong> Low novelty (γ=0.1), high diversity
            (λ=0.8), brand cap=3
          </li>
          <li>
            <strong>VIP Segment:</strong> Matches users with tier="VIP", "vip",
            or "premium"
          </li>
          <li>
            <strong>Casual Segment:</strong> Matches users with tier="free",
            "basic", or surface="homepage"
          </li>
        </ul>
      </div>
    </Section>
  );
}
