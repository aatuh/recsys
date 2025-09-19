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

            return (
              <tr key={`${it.item_id}-${i}`}>
                <Td>{i + 1}</Td>
                <Td mono>{it.item_id}</Td>
                <Td>{it.score?.toFixed(6)}</Td>
                <Td>{(it.reasons ?? []).join(", ")}</Td>
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
