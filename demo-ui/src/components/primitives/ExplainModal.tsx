import React, { useMemo } from "react";
import type { specs_types_ScoredItem } from "../../lib/api-client";
import {
  deriveExplainData,
  summarizeContributions,
  type BlendTriplet,
  type ExplainBlock,
  type ExplainCaps,
  type ExplainCapUsage,
} from "../../utils/explain";

interface ExplainModalProps {
  open: boolean;
  item: specs_types_ScoredItem | null;
  blend: BlendTriplet;
  onClose: () => void;
}

const BLEND_ROWS = [
  {
    key: "pop" as const,
    label: "Trending now",
    short: "pop",
    color: "#90caf9",
    description:
      "Recent popularity. Time-decayed count of how many users interacted with this item lately.",
    weight: "α",
  },
  {
    key: "cooc" as const,
    label: "Co-visitation",
    short: "co",
    color: "#a5d6a7",
    description:
      "Co-visitation. How often this item appears together with items you interacted with.",
    weight: "β",
  },
  {
    key: "als" as const,
    label: "Personalization",
    short: "emb",
    color: "#ffcc80",
    description:
      "Embeddings/personalization. Vector similarity to your taste profile so we can surface less obvious fits.",
    weight: "γ",
  },
] as const;

const REASON_HELP: Record<string, string> = {
  recent_popularity:
    "This item has recent traction among users (time-decayed events).",
  co_visitation:
    "This item commonly co-occurs with your anchor items (people view or buy them together).",
  personalization:
    "This item is similar to your inferred preferences (embedding match).",
  diversity:
    "MMR and caps ensured a balanced mix so one brand or category does not dominate.",
};

export function ExplainModal({
  open,
  item,
  blend,
  onClose,
}: ExplainModalProps) {
  const {
    shares,
    contributions,
    anchors,
    notes,
    reasons,
    explainBlock,
    usingExplainShares,
    blendWeights,
    blendNorms,
    rawSignals,
  } = useMemo(() => deriveExplainData(item, blend), [item, blend]);

  if (!open || !item) return null;

  const personalization = explainBlock?.personalization ?? null;
  const mmr = explainBlock?.mmr ?? null;
  const caps = explainBlock?.caps ?? null;
  const showPersonalization = hasPersonalization(personalization);
  const mmrChips = buildMmrChips(mmr);
  const capChips = buildCapChips(caps);
  const hasNorms =
    blendNorms.pop > 0 || blendNorms.cooc > 0 || blendNorms.als > 0;
  const hasRawSignals = Boolean(
    rawSignals &&
      (rawSignals.pop !== undefined ||
        rawSignals.cooc !== undefined ||
        rawSignals.emb !== undefined)
  );
  const isNotesDuplicate = notesEqual(reasons, notes) && reasons.length > 0;
  const sortedReasons = sortReasons(reasons);
  const summaryLine = summarizeContributions(contributions);

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
        onClick={(event) => event.stopPropagation()}
        style={{
          width: 640,
          maxWidth: "95vw",
          background: "white",
          borderRadius: 8,
          border: "1px solid #e2e8f0",
          boxShadow: "0 14px 45px rgba(15,23,42,0.18)",
          padding: 18,
          fontSize: 14,
        }}
      >
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            gap: 8,
            alignItems: "baseline",
            marginBottom: 12,
          }}
        >
          <h3 style={{ margin: 0, fontSize: 18 }}>
            Why this item?{" "}
            <span style={{ color: "#64748b", fontWeight: 400 }}>
              {item.item_id}
            </span>
          </h3>
          <button
            onClick={onClose}
            title="Close"
            style={{
              padding: "4px 10px",
              borderRadius: 6,
              border: "1px solid #cbd5f5",
              background: "#f8fafc",
              cursor: "pointer",
            }}
          >
            Close
          </button>
        </div>

        <section style={{ marginBottom: 16, color: "#0f172a" }}>
          <div style={{ marginBottom: 8, fontWeight: 600 }}>
            Blend contributions
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
            {BLEND_ROWS.map((row) => (
              <BlendRow
                key={row.key}
                row={row}
                shares={shares}
                contributions={contributions}
              />
            ))}
          </div>
          <div
            style={{
              marginTop: 8,
              color: "#475569",
              fontFamily: "monospace",
              fontSize: 13,
            }}
          >
            {summaryLine}
          </div>
          <div
            style={{
              marginTop: 6,
              fontSize: 12,
              color: "#64748b",
            }}
          >
            Weights{" "}
            {BLEND_ROWS.map(
              (row) => `${row.weight} ${formatNumber(blendWeights[row.key])}`
            ).join(" · ")}
          </div>
          {hasNorms && (
            <div
              style={{
                marginTop: 4,
                fontSize: 12,
                color: "#64748b",
              }}
            >
              Norms pop {formatNumber(blendNorms.pop)}, co-vis{" "}
              {formatNumber(blendNorms.cooc)}, emb{" "}
              {formatNumber(blendNorms.als)}
            </div>
          )}
          {hasRawSignals && (
            <div
              style={{
                marginTop: 4,
                fontSize: 12,
                color: "#64748b",
              }}
            >
              Raw pop {formatNumber(rawSignals?.pop)}, co-vis{" "}
              {formatNumber(rawSignals?.cooc)}, emb{" "}
              {formatNumber(rawSignals?.emb)}
            </div>
          )}
          {!usingExplainShares && (
            <div style={{ marginTop: 6, color: "#9333ea", fontSize: 12 }}>
              Per-item explain data was not returned, so we fell back to blend
              weights for this view.
            </div>
          )}
        </section>

        {showPersonalization && personalization && (
          <section style={{ marginBottom: 16 }}>
            <div style={{ marginBottom: 6, fontWeight: 600 }}>
              Personalization boost
            </div>
            <div
              style={{
                display: "flex",
                gap: 12,
                flexWrap: "wrap",
                color: "#475569",
              }}
            >
              <span>Overlap {formatPercent(personalization.overlap)}</span>
              <span>
                Boost {formatMultiplier(personalization.boost_multiplier)}
              </span>
              {personalization.raw?.profile_boost !== undefined && (
                <span>
                  profile_boost{" "}
                  {formatNumber(personalization.raw.profile_boost)}
                </span>
              )}
            </div>
          </section>
        )}

        {(mmrChips.length > 0 || capChips.length > 0) && (
          <section style={{ marginBottom: 16 }}>
            <div style={{ marginBottom: 6, fontWeight: 600 }}>
              Diversity & caps
            </div>
            <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
              {mmrChips.map((chip) => (
                <Chip key={`mmr-${chip}`} title="MMR detail">
                  {chip}
                </Chip>
              ))}
              {capChips.map((chip) => (
                <Chip key={`cap-${chip}`} title="Cap applied">
                  {chip}
                </Chip>
              ))}
            </div>
          </section>
        )}

        {anchors.length > 0 && (
          <section style={{ marginBottom: 16 }}>
            <div style={{ marginBottom: 6, fontWeight: 600 }}>Anchors</div>
            <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
              {anchors.map((anchor) => (
                <Badge key={anchor} title="Anchor item">
                  {anchor}
                </Badge>
              ))}
            </div>
          </section>
        )}

        {sortedReasons.length > 0 && (
          <section style={{ marginBottom: 16 }}>
            <div style={{ marginBottom: 4, fontWeight: 600 }}>Reason tags</div>
            <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
              {sortedReasons.map((reason) => {
                // Check if this is a rule-related reason
                const isRuleReason = reason.startsWith("rule.");
                let chipStyle = {};
                let chipContent = reason;

                if (isRuleReason) {
                  if (reason.startsWith("rule.pin")) {
                    chipStyle = {
                      backgroundColor: "#fff3cd",
                      color: "#856404",
                      border: "1px solid #ffc107",
                    };
                    chipContent = `⭐ ${reason}`;
                  } else if (reason.startsWith("rule.block")) {
                    chipStyle = {
                      backgroundColor: "#f8d7da",
                      color: "#721c24",
                      border: "1px solid #dc3545",
                    };
                    chipContent = `🚫 ${reason}`;
                  } else if (reason.startsWith("rule.boost")) {
                    chipStyle = {
                      backgroundColor: "#d4edda",
                      color: "#155724",
                      border: "1px solid #28a745",
                    };
                    chipContent = `⬆️ ${reason}`;
                  }
                }

                return (
                  <Chip
                    key={`reason-${reason}`}
                    title={describeReason(reason)}
                    style={chipStyle}
                  >
                    {chipContent}
                  </Chip>
                );
              })}
            </div>
            <ul
              style={{
                margin: 8,
                marginLeft: 18,
                color: "#475569",
                lineHeight: 1.4,
              }}
            >
              {sortedReasons.map((reason) => (
                <li key={`reason-explain-${reason}`}>
                  <strong>{titleCase(reason)}</strong>: {describeReason(reason)}
                </li>
              ))}
            </ul>
          </section>
        )}

        {/* Rule Effects Section */}
        {sortedReasons.some((r) => r.startsWith("rule.")) && (
          <section style={{ marginBottom: 16 }}>
            <div style={{ marginBottom: 8, fontWeight: 600, color: "#0f172a" }}>
              Rule Engine Effects
            </div>
            <div
              style={{
                backgroundColor: "#f8f9fa",
                border: "1px solid #e9ecef",
                borderRadius: 6,
                padding: 12,
              }}
            >
              {sortedReasons
                .filter((r) => r.startsWith("rule."))
                .map((reason) => {
                  if (reason.startsWith("rule.pin")) {
                    return (
                      <div
                        key={reason}
                        style={{
                          marginBottom: 8,
                          display: "flex",
                          alignItems: "center",
                          gap: 8,
                        }}
                      >
                        <span style={{ fontSize: "16px" }}>⭐</span>
                        <div>
                          <strong style={{ color: "#856404" }}>
                            Pinned Item
                          </strong>
                          <div style={{ fontSize: "12px", color: "#6c757d" }}>
                            This item was pinned to the top of the results by a
                            rule.
                          </div>
                        </div>
                      </div>
                    );
                  } else if (reason.startsWith("rule.block")) {
                    return (
                      <div
                        key={reason}
                        style={{
                          marginBottom: 8,
                          display: "flex",
                          alignItems: "center",
                          gap: 8,
                        }}
                      >
                        <span style={{ fontSize: "16px" }}>🚫</span>
                        <div>
                          <strong style={{ color: "#721c24" }}>
                            Blocked Item
                          </strong>
                          <div style={{ fontSize: "12px", color: "#6c757d" }}>
                            This item was blocked from appearing in results by a
                            rule.
                          </div>
                        </div>
                      </div>
                    );
                  } else if (reason.startsWith("rule.boost")) {
                    const boostMatch = reason.match(
                      /rule\.boost:([+-]?\d*\.?\d+)/
                    );
                    const boostValue = boostMatch ? boostMatch[1] : "unknown";
                    return (
                      <div
                        key={reason}
                        style={{
                          marginBottom: 8,
                          display: "flex",
                          alignItems: "center",
                          gap: 8,
                        }}
                      >
                        <span style={{ fontSize: "16px" }}>⬆️</span>
                        <div>
                          <strong style={{ color: "#155724" }}>
                            Boosted Score
                          </strong>
                          <div style={{ fontSize: "12px", color: "#6c757d" }}>
                            This item's score was boosted by +{boostValue} by a
                            rule.
                          </div>
                        </div>
                      </div>
                    );
                  }
                  return null;
                })}
            </div>
          </section>
        )}

        {!isNotesDuplicate && notes.length > 0 && (
          <section>
            <div style={{ marginBottom: 4, fontWeight: 600 }}>Raw notes</div>
            <ul style={{ margin: 0, paddingLeft: 18 }}>
              {notes.map((note, idx) => (
                <li key={`${idx}-${note}`}>{note}</li>
              ))}
            </ul>
          </section>
        )}
      </div>
    </div>
  );
}

function BlendRow(props: {
  row: (typeof BLEND_ROWS)[number];
  shares: BlendTriplet;
  contributions: BlendTriplet;
}) {
  const { row, shares, contributions } = props;
  const share = shares[row.key] ?? 0;
  const value = contributions[row.key] ?? 0;
  const width = Math.max(share * 100, share > 0 ? 6 : 0);

  return (
    <div
      title={row.description}
      style={{
        display: "flex",
        alignItems: "center",
        gap: 10,
        color: "#0f172a",
      }}
    >
      <span style={{ width: 124 }}>
        {row.label} ({row.short})
      </span>
      <div
        style={{
          flex: 1,
          height: 6,
          borderRadius: 6,
          background: "#e2e8f0",
          overflow: "hidden",
          position: "relative",
        }}
      >
        <div
          style={{
            width: `${width}%`,
            minWidth: share > 0 ? 6 : 0,
            height: "100%",
            background: row.color,
          }}
        />
      </div>
      <span
        style={{
          width: 52,
          textAlign: "right",
          fontVariantNumeric: "tabular-nums",
          color: "#475569",
        }}
      >
        {formatPercent(share)}
      </span>
      <span
        style={{
          width: 70,
          textAlign: "right",
          fontVariantNumeric: "tabular-nums",
          color: "#1e293b",
        }}
      >
        {formatNumber(value)}
      </span>
    </div>
  );
}

function Chip({
  children,
  title,
  style,
}: {
  children: React.ReactNode;
  title?: string;
  style?: React.CSSProperties;
}) {
  return (
    <span
      title={title}
      style={{
        background: "#f1f5f9",
        borderRadius: 999,
        padding: "2px 8px",
        fontSize: 12,
        color: "#0f172a",
        border: "1px solid #cbd5f5",
        ...style,
      }}
    >
      {children}
    </span>
  );
}

function Badge({
  children,
  title,
}: {
  children: React.ReactNode;
  title?: string;
}) {
  return (
    <code
      title={title}
      style={{
        background: "#f8fafc",
        border: "1px solid #cbd5f5",
        borderRadius: 6,
        padding: "2px 6px",
        fontSize: 12,
        color: "#0f172a",
      }}
    >
      {children}
    </code>
  );
}

function formatNumber(value: number | null | undefined, digits = 2) {
  if (value === undefined || value === null || Number.isNaN(value)) {
    return "—";
  }
  return Number(value).toFixed(digits);
}

function formatPercent(value: number | null | undefined) {
  if (value === undefined || value === null || Number.isNaN(value)) {
    return "—";
  }
  return `${Math.round(value * 100)}%`;
}

function formatMultiplier(value: number | null | undefined) {
  if (value === undefined || value === null || Number.isNaN(value)) {
    return "×1.00";
  }
  const delta = (value - 1) * 100;
  const suffix =
    Math.abs(delta) > 0.01
      ? ` (${delta >= 0 ? "+" : ""}${delta.toFixed(1)}%)`
      : "";
  return `×${value.toFixed(2)}${suffix}`;
}

function notesEqual(a: string[], b: string[]) {
  if (a.length !== b.length) return false;
  const sa = new Set(a);
  const sb = new Set(b);
  if (sa.size !== sb.size) return false;
  for (const value of sa) {
    if (!sb.has(value)) return false;
  }
  return true;
}

function sortReasons(reasons: string[]): string[] {
  const order = ["recent_popularity", "co_visitation", "personalization"];
  return [...reasons].sort((a, b) => {
    const ia = order.indexOf(a);
    const ib = order.indexOf(b);
    if (ia === -1 && ib === -1) return a.localeCompare(b);
    if (ia === -1) return 1;
    if (ib === -1) return -1;
    return ia - ib;
  });
}

function describeReason(reason: string): string {
  return REASON_HELP[reason] ?? `System hint: ${titleCase(reason)}`;
}

function titleCase(text: string) {
  return text
    .replace(/[_-]+/g, " ")
    .trim()
    .replace(/\s+/g, " ")
    .replace(/\b\w/g, (match) => match.toUpperCase());
}

function hasPersonalization(
  personalization: ExplainBlock["personalization"] | null
): boolean {
  if (!personalization) return false;
  const { overlap, boost_multiplier, raw } = personalization;
  return (
    (overlap !== undefined && overlap !== null) ||
    (boost_multiplier !== undefined && boost_multiplier !== null) ||
    (raw?.profile_boost !== undefined && raw.profile_boost !== null)
  );
}

function buildCapChips(caps: ExplainCaps | null | undefined): string[] {
  if (!caps) return [];
  return [
    capChip("Brand", caps.brand),
    capChip("Category", caps.category),
  ].filter((value): value is string => Boolean(value));
}

function capChip(label: string, cap?: ExplainCapUsage | null): string | null {
  if (!cap || !cap.applied) return null;
  const parts: string[] = [];
  if (
    cap.count !== undefined &&
    cap.count !== null &&
    cap.limit !== undefined &&
    cap.limit !== null
  ) {
    parts.push(`${cap.count}/${cap.limit}`);
  } else if (cap.limit !== undefined && cap.limit !== null) {
    parts.push(`limit ${cap.limit}`);
  }
  if (cap.value) {
    parts.push(cap.value);
  }
  const suffix = parts.length > 0 ? ` (${parts.join(" · ")})` : "";
  return `${label} cap${suffix}`;
}

function buildMmrChips(mmr: ExplainBlock["mmr"] | null | undefined): string[] {
  if (!mmr) return [];
  const chips: string[] = [];
  if (mmr.lambda !== undefined) chips.push(`λ ${formatNumber(mmr.lambda)}`);
  if (mmr.max_sim !== undefined)
    chips.push(`max_sim ${formatNumber(mmr.max_sim)}`);
  if (mmr.penalty !== undefined)
    chips.push(`penalty ${formatNumber(mmr.penalty)}`);
  if (mmr.relevance !== undefined)
    chips.push(`relevance ${formatNumber(mmr.relevance)}`);
  if (mmr.rank !== undefined) chips.push(`pick #${mmr.rank}`);
  return chips;
}
