import React from "react";
import { color, spacing, radius, text } from "../../ui/tokens";

interface Metric {
  label: string;
  value: string | number | undefined | null;
}

interface WhyItWorksProps {
  title?: string;
  metrics: Metric[];
}

export function WhyItWorks({
  title = "Why it works",
  metrics,
}: WhyItWorksProps) {
  const filtered = metrics.filter(
    (m) => m.value !== undefined && m.value !== null && m.value !== ""
  );
  if (filtered.length === 0) return null;

  return (
    <aside
      aria-label="Why it works"
      style={{
        backgroundColor: color.panelSubtle,
        border: `1px solid ${color.panelBorder}`,
        borderRadius: radius.md,
        padding: spacing.md,
        margin: `${spacing.md}px 0`,
      }}
    >
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: spacing.sm,
        }}
      >
        <strong style={{ color: color.text, fontSize: text.md }}>
          {title}
        </strong>
      </div>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(140px, 1fr))",
          gap: spacing.md,
        }}
      >
        {filtered.map((m) => (
          <div
            key={m.label}
            style={{
              display: "flex",
              flexDirection: "column",
              gap: spacing.xs,
            }}
          >
            <span style={{ fontSize: 11, color: color.textMuted }}>
              {m.label}
            </span>
            <span style={{ fontSize: text.md, color: color.text }}>
              {String(m.value)}
            </span>
          </div>
        ))}
      </div>
    </aside>
  );
}
