import React, { useMemo } from "react";
import type { types_ScoredItem } from "../lib/api-client";

type Blend = { pop: number; cooc: number; als: number };

interface ExplainModalProps {
  open: boolean;
  item: types_ScoredItem | null;
  blend: Blend;
  onClose: () => void;
}

/* Human explanations for the three components. Keep short and clear. */
const TERM_HELP: Record<keyof Blend, string> = {
  pop:
    "Recent popularity. Time-decayed count of how many users interacted " +
    "with this item lately.",
  cooc:
    "Co-visitation. How often this item appears together with items you " +
    "interacted with (item-to-item co-occurrence).",
  als:
    "Embeddings/personalization. Vector-space similarity to your taste " +
    "profile; can surface less obvious but relevant items.",
};

/* Known reason tag explanations (fallback is auto-formatted). */
const REASON_HELP: Record<string, string> = {
  recent_popularity:
    "This item has recent traction among users (time-decayed events).",
  co_visitation:
    "This item commonly co-occurs with your anchor items (you viewed/" +
    "bought similar things).",
  personalization:
    "This item is similar to your inferred preferences (embedding match).",
};

/**
 * Parses free-form reasons and tries to extract numeric contributions:
 *   "pop:0.62", "cooc=0.28", "als:0.10"
 * It also extracts anchors like "anchor:item-0007".
 * Non-numeric lines are kept as notes.
 */
function parseReasons(reasons: string[] | undefined): {
  contrib: Partial<Record<keyof Blend, number>>;
  anchors: string[];
  notes: string[];
} {
  const c: Partial<Record<keyof Blend, number>> = {};
  const anchors: string[] = [];
  const notes: string[] = [];

  const rx = {
    pop: /\bpop(?:ularity)?\s*[:=]\s*([0-9.]+)/i,
    cooc: /\b(?:co[-\s]?vis(?:itation)?|cooc)\s*[:=]\s*([0-9.]+)/i,
    als: /\b(?:als|embed(?:ding)?|vec(?:tor)?)\s*[:=]\s*([0-9.]+)/i,
    anchor: /\banchor\s*[:=]\s*([A-Za-z0-9_\-:.]+)/gi,
  };

  for (const r of reasons ?? []) {
    const p = r.match(rx.pop);
    if (p) c.pop = Number(p[1]);

    const k = r.match(rx.cooc);
    if (k) c.cooc = Number(k[1]);

    const a = r.match(rx.als);
    if (a) c.als = Number(a[1]);

    let m: RegExpExecArray | null;
    while ((m = rx.anchor.exec(r)) !== null) anchors.push(m[1] ?? "");

    if (!p && !k && !a) notes.push(r);
  }

  return { contrib: c, anchors, notes };
}

function pct(n: number) {
  return `${Math.round(n * 100)}%`;
}

function toTitleWords(s: string) {
  return s
    .replace(/[_\-]+/g, " ")
    .trim()
    .replace(/\s+/g, " ")
    .replace(/\b\w/g, (m) => m.toUpperCase());
}

/* Simple stacked bar without external libs. */
function StackedBar(props: { shares: Blend }) {
  const total = props.shares.pop + props.shares.cooc + props.shares.als;
  const t = Math.max(total, 1);

  const parts = [
    { key: "pop", label: "pop", value: props.shares.pop / t },
    { key: "cooc", label: "cooc", value: props.shares.cooc / t },
    { key: "als", label: "als", value: props.shares.als / t },
  ];

  return (
    <div
      style={{
        border: "1px solid #ccc",
        borderRadius: 6,
        width: "100%",
        height: 16,
        display: "flex",
        overflow: "hidden",
      }}
      aria-label="contribution-bar"
    >
      {parts.map((p) => (
        <div
          key={p.key}
          style={{
            width: pct(p.value),
            height: "100%",
            background:
              p.key === "pop"
                ? "#90caf9"
                : p.key === "cooc"
                ? "#a5d6a7"
                : "#ffcc80",
          }}
          title={`${p.label}: ${pct(p.value)}`}
        />
      ))}
    </div>
  );
}

export function ExplainModal({
  open,
  item,
  blend,
  onClose,
}: ExplainModalProps) {
  const {
    shares,
    anchors,
    notes,
    hasExtracted,
    isNotesDuplicate,
    sortedReasons,
  } = useMemo(() => {
    const parsed = parseReasons(item?.reasons);
    const b = { ...blend };

    const hasNums =
      parsed.contrib.pop !== undefined ||
      parsed.contrib.cooc !== undefined ||
      parsed.contrib.als !== undefined;

    const raw = {
      pop: parsed.contrib.pop ?? 0,
      cooc: parsed.contrib.cooc ?? 0,
      als: parsed.contrib.als ?? 0,
    };

    const weighted = hasNums
      ? {
          pop: raw.pop * b.pop,
          cooc: raw.cooc * b.cooc,
          als: raw.als * b.als,
        }
      : { pop: b.pop, cooc: b.cooc, als: b.als };

    const sum = weighted.pop + weighted.cooc + weighted.als || 1;
    const norm = {
      pop: weighted.pop / sum,
      cooc: weighted.cooc / sum,
      als: weighted.als / sum,
    };

    const reasons = item?.reasons ?? [];
    const dup = notesEqual(reasons, parsed.notes) && reasons.length > 0;

    /* Keep reasons in a stable, readable order. */
    const order = ["recent_popularity", "co_visitation", "personalization"];
    const sorted = [...reasons].sort((a, b2) => {
      const ia = order.indexOf(a);
      const ib = order.indexOf(b2);
      if (ia === -1 && ib === -1) return a.localeCompare(b2);
      if (ia === -1) return 1;
      if (ib === -1) return -1;
      return ia - ib;
    });

    return {
      shares: norm,
      anchors: parsed.anchors,
      notes: parsed.notes,
      hasExtracted: hasNums,
      isNotesDuplicate: dup,
      sortedReasons: sorted,
    };
  }, [item, blend]);

  if (!open || !item) return null;

  return (
    <div
      role="dialog"
      aria-modal="true"
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(0,0,0,0.35)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: 16,
        zIndex: 1000,
      }}
      onClick={onClose}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          width: 640,
          maxWidth: "95vw",
          background: "white",
          borderRadius: 8,
          border: "1px solid #ddd",
          boxShadow: "0 8px 30px rgba(0,0,0,0.2)",
          padding: 16,
          fontSize: 14,
        }}
      >
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            gap: 8,
            alignItems: "center",
            marginBottom: 8,
          }}
        >
          <h3 style={{ margin: 0 }}>
            Why this item?{" "}
            <span style={{ color: "#666", fontWeight: 400 }}>
              {item.item_id}
            </span>
          </h3>
          <button
            onClick={onClose}
            title="Close"
            style={{
              padding: "4px 8px",
              borderRadius: 6,
              border: "1px solid #aaa",
              background: "#fafafa",
              cursor: "pointer",
            }}
          >
            Close
          </button>
        </div>

        {/* Contribution section */}
        <div style={{ marginBottom: 10, color: "#333" }}>
          <div style={{ marginBottom: 6, fontWeight: 600 }}>
            Contribution breakdown
          </div>
          <StackedBar shares={shares} />
          <div
            style={{
              display: "flex",
              gap: 12,
              marginTop: 6,
              color: "#555",
              flexWrap: "wrap",
            }}
          >
            <span title={TERM_HELP.pop}>pop: {pct(shares.pop)}</span>
            <span title={TERM_HELP.cooc}>cooc: {pct(shares.cooc)}</span>
            <span title={TERM_HELP.als}>als: {pct(shares.als)}</span>
            {!hasExtracted && (
              <span style={{ color: "#9c27b0" }}>
                (Based on current blend only. The API did not include per-item
                numeric attributions for this result, so these shares are
                estimated from the UI blend and will look the same for other
                items.)
              </span>
            )}
          </div>

          {/* Mini glossary for plain-English meaning. */}
          <ul
            style={{
              margin: "8px 0 0 0",
              paddingLeft: 18,
              color: "#555",
              lineHeight: 1.35,
            }}
          >
            <li>
              <strong>pop</strong>: {TERM_HELP.pop}
            </li>
            <li>
              <strong>cooc</strong>: {TERM_HELP.cooc}
            </li>
            <li>
              <strong>als</strong>: {TERM_HELP.als}
            </li>
          </ul>
        </div>

        {/* Reasons section with plain-English lines */}
        <div style={{ marginBottom: 10 }}>
          <div style={{ marginBottom: 6, fontWeight: 600 }}>Raw reasons</div>
          <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
            {sortedReasons.map((r, i) => (
              <span
                key={`${i}-${r}`}
                title={REASON_HELP[r] || toTitleWords(r)}
                style={{
                  background: "#f3f4f6",
                  border: "1px solid #e5e7eb",
                  borderRadius: 12,
                  padding: "2px 8px",
                  fontSize: 12,
                }}
              >
                {r}
              </span>
            ))}
            {sortedReasons.length === 0 && (
              <span style={{ color: "#777" }}>(no reasons)</span>
            )}
          </div>

          {sortedReasons.length > 0 && (
            <div style={{ marginTop: 8 }}>
              <div style={{ marginBottom: 4, fontWeight: 600 }}>
                Reasons explained
              </div>
              <ul
                style={{
                  margin: 0,
                  paddingLeft: 18,
                  color: "#555",
                  lineHeight: 1.35,
                }}
              >
                {sortedReasons.map((r, i) => (
                  <li key={`explain-${i}-${r}`}>
                    <strong>{toTitleWords(r)}</strong>:{" "}
                    {REASON_HELP[r] || "System hint: " + toTitleWords(r)}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>

        {anchors.length > 0 && (
          <div style={{ marginBottom: 10 }}>
            <div style={{ marginBottom: 6, fontWeight: 600 }}>
              Anchors referenced
            </div>
            <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
              {anchors.map((a) => (
                <code
                  key={a}
                  style={{
                    background: "#fff8e1",
                    border: "1px solid #ffe0b2",
                    borderRadius: 6,
                    padding: "2px 6px",
                    fontSize: 12,
                  }}
                >
                  {a}
                </code>
              ))}
            </div>
          </div>
        )}

        {!isNotesDuplicate && notes.length > 0 && (
          <div style={{ marginBottom: 8 }}>
            <div style={{ marginBottom: 6, fontWeight: 600 }}>Notes</div>
            <ul style={{ margin: 0, paddingLeft: 18 }}>
              {notes.map((n, i) => (
                <li key={`${i}-${n}`}>{n}</li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </div>
  );
}

/* Returns true if both arrays contain the same strings, ignoring order. */
function notesEqual(a: string[], b: string[]) {
  if (a.length !== b.length) return false;
  const sa = new Set(a);
  const sb = new Set(b);
  if (sa.size !== sb.size) return false;
  for (const x of sa) if (!sb.has(x)) return false;
  return true;
}
