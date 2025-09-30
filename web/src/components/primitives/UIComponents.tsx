import React from "react";
import { color, radius, spacing, text } from "../../ui/tokens";
import {
  deriveExplainData,
  summarizeContributions,
  type BlendTriplet,
} from "../../utils/explain";

export function Section(props: { title: string; children: React.ReactNode }) {
  return (
    <section style={{ marginBottom: spacing.lg }}>
      <h2 style={{ margin: "0 0 12px 0", fontSize: text.lg, fontWeight: 600 }}>
        {props.title}
      </h2>
      <div
        style={{
          border: `1px solid ${color.panelBorder}`,
          borderRadius: radius.lg,
          padding: spacing.lg,
          backgroundColor: color.panelBg,
        }}
      >
        {props.children}
      </div>
    </section>
  );
}

export function Row(props: {
  children: React.ReactNode;
  style?: React.CSSProperties;
}) {
  return (
    <div
      style={{
        display: "flex",
        gap: spacing.md,
        flexWrap: "wrap",
        alignItems: "center",
        marginBottom: spacing.md,
        ...props.style,
      }}
    >
      {props.children}
    </div>
  );
}

export function Label(props: {
  text: string;
  width?: number;
  required?: boolean;
  children: React.ReactNode;
  error?: string;
}) {
  return (
    <label
      style={{
        display: "flex",
        flexDirection: "column",
        gap: spacing.xs,
        minWidth: props.width ?? 160,
        flex: "1 1 200px", // Allow labels to grow and shrink on mobile
      }}
    >
      <span
        style={{ fontSize: text.sm, color: color.textMuted, fontWeight: 500 }}
      >
        {props.text}
        {props.required && (
          <span style={{ color: color.danger, marginLeft: spacing.xs }}>*</span>
        )}
      </span>
      {props.children}
      {props.error && (
        <span
          style={{
            fontSize: text.xs,
            color: color.danger,
            marginTop: spacing.xs,
          }}
        >
          {props.error}
        </span>
      )}
    </label>
  );
}

export function TextInput(
  props: React.InputHTMLAttributes<HTMLInputElement> & {
    error?: boolean;
  }
) {
  const { error, ...inputProps } = props;
  return (
    <input
      {...inputProps}
      style={{
        width: "min(280px, 100%)",
        padding: `${spacing.sm}px ${spacing.md}px`,
        borderRadius: radius.md,
        border: `1px solid ${error ? color.danger : color.border}`,
        fontFamily: "inherit",
        backgroundColor: error ? color.dangerBg : "transparent",
        ...(props.style || {}),
      }}
    />
  );
}

export function NumberInput(
  props: React.InputHTMLAttributes<HTMLInputElement> & {
    error?: boolean;
    min?: number;
    max?: number;
  }
) {
  const { error, min, max, ...inputProps } = props;
  return (
    <TextInput
      type="number"
      min={min}
      max={max}
      error={error}
      {...inputProps}
    />
  );
}

export function Button(props: React.ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      {...props}
      style={{
        padding: `${spacing.md}px ${spacing.lg}px`,
        borderRadius: radius.md,
        border: `1px solid ${color.buttonBorder}`,
        background: color.buttonBg,
        cursor: "pointer",
        color: color.text,
        transition: "all 120ms ease",
        fontSize: text.sm,
        fontWeight: 500,
        minHeight: 40, // Ensure touch targets are large enough
        ...(props.style || {}),
      }}
      onMouseEnter={(e) => {
        if (!props.disabled) {
          e.currentTarget.style.backgroundColor = color.buttonHover;
        }
      }}
      onMouseLeave={(e) => {
        if (!props.disabled) {
          e.currentTarget.style.backgroundColor = color.buttonBg;
        }
      }}
    />
  );
}

export function Code(props: { children: React.ReactNode }) {
  return (
    <pre
      style={{
        background: color.codeBg,
        border: `1px solid ${color.codeBorder}`,
        borderRadius: radius.md,
        padding: spacing.lg,
        overflowX: "auto",
        maxHeight: 320,
        color: color.text,
      }}
    >
      <code>{props.children}</code>
    </pre>
  );
}

export function Th(props: { children: React.ReactNode }) {
  return (
    <th
      style={{
        textAlign: "left",
        padding: `${spacing.md}px ${spacing.lg}px`,
        borderBottom: `1px solid ${color.border}`,
        fontWeight: 600,
        fontSize: 13,
      }}
    >
      {props.children}
    </th>
  );
}

export function Td(props: { children: React.ReactNode; mono?: boolean }) {
  return (
    <td
      style={{
        padding: `${spacing.sm}px ${spacing.lg}px`,
        borderBottom: `1px solid ${color.panelBorder}`,
        fontFamily: props.mono ? "monospace" : undefined,
      }}
    >
      {props.children}
    </td>
  );
}

export interface ResultsTableProps {
  items: any[] | null;
  showExplain?: boolean;
  onExplain?: (value: any) => void;
  blend?: BlendTriplet;
}

export function ResultsTable(props: ResultsTableProps) {
  const { items, showExplain = false, onExplain, blend } = props;
  if (!items) return null;

  // Helper function to detect rule-related reasons
  const getRuleDecorations = (reasons: string[]) => {
    const decorations = [];

    for (const reason of reasons) {
      if (reason.startsWith("rule.pin")) {
        decorations.push({
          type: "pinned",
          icon: "⭐",
          label: "Pinned",
          color: "#ffc107",
          backgroundColor: "#fff3cd",
        });
      } else if (reason.startsWith("rule.block")) {
        decorations.push({
          type: "blocked",
          icon: "🚫",
          label: "Blocked",
          color: "#dc3545",
          backgroundColor: "#f8d7da",
        });
      } else if (reason.startsWith("rule.boost")) {
        const boostMatch = reason.match(/rule\.boost:([+-]?\d*\.?\d+)/);
        if (boostMatch) {
          decorations.push({
            type: "boosted",
            icon: "⬆️",
            label: `+${boostMatch[1]}`,
            color: "#28a745",
            backgroundColor: "#d4edda",
          });
        }
      }
    }

    return decorations;
  };

  return (
    <div style={{ overflowX: "auto" }}>
      <table
        style={{
          borderCollapse: "collapse",
          width: "100%",
          minWidth: 560,
        }}
      >
        <thead>
          <tr>
            <Th>rank</Th>
            <Th>item_id</Th>
            <Th>score</Th>
            <Th>reasons</Th>
            {showExplain && <Th>why</Th>}
          </tr>
        </thead>
        <tbody>
          {items.map((it, i) => {
            const explainData =
              showExplain && onExplain && blend
                ? deriveExplainData(it, blend)
                : null;
            const pillText = explainData
              ? summarizeContributions(explainData.contributions)
              : null;

            const ruleDecorations = getRuleDecorations(it.reasons ?? []);
            const isPinned = ruleDecorations.some((d) => d.type === "pinned");
            const isBlocked = ruleDecorations.some((d) => d.type === "blocked");

            return (
              <tr
                key={`${it.item_id}-${i}`}
                style={{
                  backgroundColor: isPinned
                    ? color.warningBg
                    : isBlocked
                    ? color.dangerBg
                    : undefined,
                  borderLeft: isPinned
                    ? `4px solid ${color.warning}`
                    : isBlocked
                    ? `4px solid ${color.danger}`
                    : undefined,
                }}
              >
                <Td>
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: spacing.md,
                    }}
                  >
                    <span>{i + 1}</span>
                    {isPinned && (
                      <span style={{ fontSize: "16px" }} title="Pinned item">
                        ⭐
                      </span>
                    )}
                    {isBlocked && (
                      <span style={{ fontSize: "16px" }} title="Blocked item">
                        🚫
                      </span>
                    )}
                  </div>
                </Td>
                <Td mono>{it.item_id}</Td>
                <Td>
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: spacing.md,
                    }}
                  >
                    <span>{it.score?.toFixed(6)}</span>
                    {ruleDecorations
                      .filter((d) => d.type === "boosted")
                      .map((decoration, idx) => (
                        <span
                          key={idx}
                          style={{
                            backgroundColor: decoration.backgroundColor,
                            color: decoration.color,
                            padding: "2px 6px",
                            borderRadius: radius.sm,
                            fontSize: "11px",
                            fontWeight: "bold",
                            border: `1px solid ${decoration.color}`,
                          }}
                          title={`Boosted by rule: ${decoration.label}`}
                        >
                          {decoration.label}
                        </span>
                      ))}
                  </div>
                </Td>
                <Td>
                  <div
                    style={{
                      display: "flex",
                      flexWrap: "wrap",
                      gap: spacing.xs,
                      alignItems: "center",
                    }}
                  >
                    {(it.reasons ?? []).map((reason: string, idx: number) => {
                      const decoration = ruleDecorations.find((d) =>
                        reason.includes(d.type)
                      );
                      if (decoration) {
                        return (
                          <span
                            key={idx}
                            style={{
                              backgroundColor: decoration.backgroundColor,
                              color: decoration.color,
                              padding: "2px 6px",
                              borderRadius: radius.sm,
                              fontSize: "11px",
                              fontWeight: "bold",
                              border: `1px solid ${decoration.color}`,
                              display: "flex",
                              alignItems: "center",
                              gap: spacing.xs,
                            }}
                            title={`Rule effect: ${reason}`}
                          >
                            <span>{decoration.icon}</span>
                            <span>{decoration.label}</span>
                          </span>
                        );
                      }
                      return (
                        <span key={idx} style={{ fontSize: "12px" }}>
                          {reason}
                        </span>
                      );
                    })}
                  </div>
                </Td>
                {showExplain && (
                  <Td>
                    {explainData ? (
                      <button
                        type="button"
                        onClick={() => onExplain?.(it)}
                        title="View structured explain"
                        style={{
                          border: `1px solid ${color.border}`,
                          background: color.panelSubtle,
                          borderRadius: radius.pill,
                          padding: "2px 8px",
                          fontSize: 12,
                          cursor: "pointer",
                          fontFamily: "monospace",
                          color: color.text,
                        }}
                      >
                        {pillText}
                      </button>
                    ) : (
                      <Button
                        onClick={() => onExplain && onExplain(it)}
                        title="Explain"
                        style={{ padding: "4px 8px" }}
                      >
                        Why?
                      </Button>
                    )}
                  </Td>
                )}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
