import React, { useMemo } from "react";
import type { internal_http_types_ScoredItem } from "../lib/api-client";

type Blend = { pop: number; cooc: number; als: number };

type ExplainBlend = {
  alpha?: number;
  beta?: number;
  gamma?: number;
  pop_norm?: number;
  cooc_norm?: number;
  emb_norm?: number;
  contrib?: {
    pop?: number;
    cooc?: number;
    emb?: number;
  };
  raw?: {
    pop?: number;
    cooc?: number;
    emb?: number;
  };
};

type ExplainPersonalization = {
  overlap?: number;
  boost_multiplier?: number;
  raw?: {
    profile_boost?: number;
  };
};

type ExplainMMR = {
  lambda?: number;
  max_sim?: number;
  penalty?: number;
  relevance?: number;
  rank?: number;
};

type ExplainCapUsage = {
  applied?: boolean;
  limit?: number | null;
  count?: number | null;
  value?: string | null;
};

type ExplainCaps = {
  brand?: ExplainCapUsage | null;
  category?: ExplainCapUsage | null;
};

type ExplainBlock = {
  blend?: ExplainBlend | null;
  personalization?: ExplainPersonalization | null;
  mmr?: ExplainMMR | null;
  caps?: ExplainCaps | null;
  anchors?: string[] | null;
};

interface ExplainModalProps {
  open: boolean;
  item: internal_http_types_ScoredItem | null;
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
  diversity:
    "MMR and caps ensured a balanced mix, preventing one brand or category from dominating.",
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

function formatNumber(value: number | null | undefined, digits = 2) {
  if (value === undefined || value === null || Number.isNaN(value)) {
    return "—";
  }
  return Number(value).toFixed(digits);
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
    usingExplainShares,
    isNotesDuplicate,
    sortedReasons,
    explainBlock,
    contributions,
    blendWeights,
    blendNorms,
    rawSignals,
  } = useMemo(() => {
    const explain = (item as { explain?: ExplainBlock | null } | null)?.explain;
    const parsed = parseReasons(item?.reasons);

    const safe = (n: number | undefined | null): number =>
      typeof n === "number" && Number.isFinite(n) ? n : 0;

    const fallbackBlend = { ...blend };

    const blendWeights = {
      pop: safe(explain?.blend?.alpha ?? fallbackBlend.pop),
      cooc: safe(explain?.blend?.beta ?? fallbackBlend.cooc),
      als: safe(explain?.blend?.gamma ?? fallbackBlend.als),
    } satisfies Blend;

    const blendNorms = {
      pop: safe(explain?.blend?.pop_norm),
      cooc: safe(explain?.blend?.cooc_norm),
      als: safe(explain?.blend?.emb_norm),
    } satisfies Blend;

    let contributions: Blend = { pop: 0, cooc: 0, als: 0 };
    let usingExplainShares = false;

    const contribFromExplain = {
      pop: safe(explain?.blend?.contrib?.pop),
      cooc: safe(explain?.blend?.contrib?.cooc),
      als: safe(explain?.blend?.contrib?.emb),
    } satisfies Blend;
    const contribExplainSum =
      contribFromExplain.pop + contribFromExplain.cooc + contribFromExplain.als;

    if (contribExplainSum > 0) {
      contributions = contribFromExplain;
      usingExplainShares = true;
    }

    if (!usingExplainShares) {
      const normBased = {
        pop: blendWeights.pop * blendNorms.pop,
        cooc: blendWeights.cooc * blendNorms.cooc,
        als: blendWeights.als * blendNorms.als,
      } satisfies Blend;
      const normSum = normBased.pop + normBased.cooc + normBased.als;
      if (normSum > 0) {
        contributions = normBased;
        usingExplainShares = Boolean(explain?.blend);
      }
    }

    if (
      !usingExplainShares &&
      (parsed.contrib.pop !== undefined ||
        parsed.contrib.cooc !== undefined ||
        parsed.contrib.als !== undefined)
    ) {
      contributions = {
        pop: safe(parsed.contrib.pop),
        cooc: safe(parsed.contrib.cooc),
        als: safe(parsed.contrib.als),
      };
    }

    if (
      contributions.pop === 0 &&
      contributions.cooc === 0 &&
      contributions.als === 0
    ) {
      contributions = { ...blendWeights };
    }

    const sum =
      contributions.pop + contributions.cooc + contributions.als ||
      blendWeights.pop + blendWeights.cooc + blendWeights.als ||
      1;

    const shares = {
      pop: sum > 0 ? contributions.pop / sum : 0,
      cooc: sum > 0 ? contributions.cooc / sum : 0,
      als: sum > 0 ? contributions.als / sum : 0,
    } satisfies Blend;

    const anchorSource = (explain?.anchors ?? parsed.anchors ?? []).filter(
      (a): a is string => Boolean(a)
    );
    const anchors = Array.from(new Set(anchorSource));

    const reasons = item?.reasons ?? [];
    const dup = notesEqual(reasons, parsed.notes) && reasons.length > 0;

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
      shares,
      contributions,
      anchors,
      notes: parsed.notes,
      usingExplainShares,
      isNotesDuplicate: dup,
      sortedReasons: sorted,
      explainBlock: explain ?? null,
      blendWeights,
      blendNorms,
      rawSignals: explain?.blend?.raw ?? null,
    };
  }, [item, blend]);

  if (!open || !item) return null;

  const personalization = explainBlock?.personalization ?? null;
  const mmr = explainBlock?.mmr ?? null;
  const caps = explainBlock?.caps ?? null;
  const hasNorms =
    blendNorms.pop > 0 || blendNorms.cooc > 0 || blendNorms.als > 0;
  const hasRawSignals = Boolean(
    rawSignals &&
      (rawSignals.pop !== undefined ||
        rawSignals.cooc !== undefined ||
        rawSignals.emb !== undefined)
  );
  const boostMultiplier = personalization?.boost_multiplier ?? 1;
  const boostDelta = (boostMultiplier - 1) * 100;
  const boostPercentText =
    Math.abs(boostDelta) > 0.01
      ? ` (${boostDelta >= 0 ? "+" : ""}${boostDelta.toFixed(1)}%)`
      : "";
  const showMmr =
    !!mmr &&
    (mmr.lambda !== undefined ||
      mmr.penalty !== undefined ||
      mmr.max_sim !== undefined ||
      mmr.relevance !== undefined ||
      mmr.rank !== undefined);
  const capSummaries = caps
    ? [
        describeCap("Brand", caps.brand),
        describeCap("Category", caps.category),
      ].filter((entry): entry is string => Boolean(entry))
    : [];

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
          </div>
          <div
            style={{
              display: "flex",
              gap: 12,
              marginTop: 6,
              color: "#555",
              flexWrap: "wrap",
            }}
          >
            <span>contrib pop: {formatNumber(contributions.pop)}</span>
            <span>contrib cooc: {formatNumber(contributions.cooc)}</span>
            <span>contrib als: {formatNumber(contributions.als)}</span>
          </div>
          <div
            style={{
              display: "flex",
              gap: 12,
              marginTop: 6,
              color: "#555",
              flexWrap: "wrap",
            }}
          >
            <span title="Alpha (popularity weight)">
              α: {formatNumber(blendWeights.pop)}
            </span>
            <span title="Beta (co-visitation weight)">
              β: {formatNumber(blendWeights.cooc)}
            </span>
            <span title="Gamma (embedding weight)">
              γ: {formatNumber(blendWeights.als)}
            </span>
          </div>
          {hasNorms && (
            <div
              style={{
                display: "flex",
                gap: 12,
                marginTop: 6,
                color: "#555",
                flexWrap: "wrap",
              }}
            >
              <span>pop_norm: {formatNumber(blendNorms.pop)}</span>
              <span>cooc_norm: {formatNumber(blendNorms.cooc)}</span>
              <span>emb_norm: {formatNumber(blendNorms.als)}</span>
            </div>
          )}
          {hasRawSignals && (
            <div
              style={{
                display: "flex",
                gap: 12,
                marginTop: 6,
                color: "#555",
                flexWrap: "wrap",
              }}
            >
              <span>raw pop: {formatNumber(rawSignals?.pop)}</span>
              <span>raw cooc: {formatNumber(rawSignals?.cooc)}</span>
              <span>raw emb: {formatNumber(rawSignals?.emb)}</span>
            </div>
          )}
          {!usingExplainShares && (
            <span
              style={{
                color: "#9c27b0",
                display: "block",
                marginTop: 6,
              }}
            >
              Per-item explain data was not returned, so these shares fall back
              to the current blend weights.
            </span>
          )}

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
                    {REASON_HELP[r] || `System hint: ${  toTitleWords(r)}`}
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

        {personalization && (
          <div style={{ marginBottom: 10 }}>
            <div style={{ marginBottom: 6, fontWeight: 600 }}>
              Personalization boost
            </div>
            <div
              style={{
                display: "flex",
                gap: 12,
                flexWrap: "wrap",
                color: "#555",
              }}
            >
              <span>overlap: {pct(personalization.overlap ?? 0)}</span>
              <span>
                boost: ×{formatNumber(boostMultiplier, 2)}
                {boostPercentText}
              </span>
              {personalization.raw?.profile_boost !== undefined && (
                <span>
                  profile_boost: {formatNumber(personalization.raw.profile_boost)}
                </span>
              )}
            </div>
          </div>
        )}

        {showMmr && mmr && (
          <div style={{ marginBottom: 10 }}>
            <div style={{ marginBottom: 6, fontWeight: 600 }}>
              Diversity (MMR)
            </div>
            <div
              style={{
                display: "flex",
                gap: 12,
                flexWrap: "wrap",
                color: "#555",
              }}
            >
              {mmr.lambda !== undefined && (
                <span>λ: {formatNumber(mmr.lambda)}</span>
              )}
              {mmr.penalty !== undefined && (
                <span>penalty: {formatNumber(mmr.penalty)}</span>
              )}
              {mmr.max_sim !== undefined && (
                <span>max_sim: {formatNumber(mmr.max_sim)}</span>
              )}
              {mmr.relevance !== undefined && (
                <span>relevance: {formatNumber(mmr.relevance)}</span>
              )}
              {mmr.rank !== undefined && <span>pick order: {mmr.rank}</span>}
            </div>
          </div>
        )}

        {capSummaries.length > 0 && (
          <div style={{ marginBottom: 10 }}>
            <div style={{ marginBottom: 6, fontWeight: 600 }}>Caps</div>
            <ul style={{ margin: 0, paddingLeft: 18, color: "#555" }}>
              {capSummaries.map((summary) => (
                <li key={summary}>{summary}</li>
              ))}
            </ul>
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

function describeCap(label: string, cap?: ExplainCapUsage | null): string | null {
  if (!cap) return null;
  const parts: string[] = [];
  parts.push(cap.applied ? "applied" : "not applied");
  if (cap.count !== undefined && cap.count !== null && cap.limit !== undefined && cap.limit !== null) {
    parts.push(`${cap.count}/${cap.limit}`);
  } else if (cap.limit !== undefined && cap.limit !== null) {
    parts.push(`limit ${cap.limit}`);
  }
  if (cap.value) {
    parts.push(cap.value);
  }
  return `${label}: ${parts.join(", ")}`;
}
