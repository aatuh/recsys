import React from "react";
import {
  deriveExplainData,
  summarizeContributions,
  type BlendTriplet,
} from "../utils/explain";

export function Section(props: { title: string; children: React.ReactNode }) {
  return (
    <section style={{ marginBottom: 16 }}>
      <h2 style={{ margin: "8px 0 8px 0" }}>{props.title}</h2>
      <div
        style={{
          border: "1px solid #e0e0e0",
          borderRadius: 8,
          padding: 12,
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
        gap: 8,
        flexWrap: "wrap",
        alignItems: "center",
        marginBottom: 8,
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
  children: React.ReactNode;
}) {
  return (
    <label
      style={{
        display: "flex",
        flexDirection: "column",
        gap: 4,
        minWidth: props.width ?? 160,
      }}
    >
      <span style={{ fontSize: 12, color: "#555" }}>{props.text}</span>
      {props.children}
    </label>
  );
}

export function TextInput(props: React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      {...props}
      style={{
        width: 200,
        padding: "6px 8px",
        borderRadius: 6,
        border: "1px solid #ccc",
        fontFamily: "inherit",
        ...(props.style || {}),
      }}
    />
  );
}

export function NumberInput(
  props: React.InputHTMLAttributes<HTMLInputElement>
) {
  return <TextInput type="number" {...props} />;
}

export function Button(props: React.ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      {...props}
      style={{
        padding: "8px 12px",
        borderRadius: 6,
        border: "1px solid #888",
        background: "#fafafa",
        cursor: "pointer",
        ...(props.style || {}),
      }}
    />
  );
}

export function Code(props: { children: React.ReactNode }) {
  return (
    <pre
      style={{
        background: "#f6f8fa",
        border: "1px solid #e1e4e8",
        borderRadius: 6,
        padding: 12,
        overflowX: "auto",
        maxHeight: 320,
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
        padding: "8px 10px",
        borderBottom: "1px solid #ddd",
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
        padding: "6px 10px",
        borderBottom: "1px solid #eee",
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
                    ? "#fff3cd"
                    : isBlocked
                    ? "#f8d7da"
                    : undefined,
                  borderLeft: isPinned
                    ? "4px solid #ffc107"
                    : isBlocked
                    ? "4px solid #dc3545"
                    : undefined,
                }}
              >
                <Td>
                  <div
                    style={{ display: "flex", alignItems: "center", gap: 8 }}
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
                    style={{ display: "flex", alignItems: "center", gap: 8 }}
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
                            borderRadius: 4,
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
                      gap: 4,
                      alignItems: "center",
                    }}
                  >
                    {(it.reasons ?? []).map((reason, idx) => {
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
                              borderRadius: 4,
                              fontSize: "11px",
                              fontWeight: "bold",
                              border: `1px solid ${decoration.color}`,
                              display: "flex",
                              alignItems: "center",
                              gap: 4,
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
                          border: "1px solid #cbd5f5",
                          background: "#f1f5f9",
                          borderRadius: 999,
                          padding: "2px 8px",
                          fontSize: 12,
                          cursor: "pointer",
                          fontFamily: "monospace",
                          color: "#0f172a",
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
