import type {
  specs_types_ScoredItem,
  types_ExplainBlock,
  types_ExplainBlend,
  types_ExplainPersonalization,
  types_ExplainMMR,
  types_ExplainCaps,
  types_ExplainCapUsage,
} from "../lib/api-client";

export type BlendTriplet = { pop: number; cooc: number; als: number };

// Re-export API types for convenience
export type ExplainBlend = types_ExplainBlend;
export type ExplainPersonalization = types_ExplainPersonalization;
export type ExplainMMR = types_ExplainMMR;
export type ExplainCapUsage = types_ExplainCapUsage;
export type ExplainCaps = types_ExplainCaps;
export type ExplainBlock = types_ExplainBlock;

type ParsedReasonData = {
  contrib: Partial<Record<keyof BlendTriplet, number>>;
  anchors: string[];
  notes: string[];
};

export type ExplainDerived = {
  shares: BlendTriplet;
  contributions: BlendTriplet;
  anchors: string[];
  notes: string[];
  reasons: string[];
  explainBlock: ExplainBlock | null;
  usingExplainShares: boolean;
  blendWeights: BlendTriplet;
  blendNorms: BlendTriplet;
  rawSignals: ExplainBlend["raw"] | null | undefined;
};

const ZERO: BlendTriplet = { pop: 0, cooc: 0, als: 0 };

function safeNumber(value: number | undefined | null): number {
  return typeof value === "number" && Number.isFinite(value) ? value : 0;
}

function parseReasons(reasons: string[] | undefined): ParsedReasonData {
  const contrib: Partial<Record<keyof BlendTriplet, number>> = {};
  const anchors: string[] = [];
  const notes: string[] = [];

  const rx = {
    pop: /\bpop(?:ularity)?\s*[:=]\s*([0-9.]+)/i,
    cooc: /\b(?:co[-\s]?vis(?:itation)?|cooc)\s*[:=]\s*([0-9.]+)/i,
    als: /\b(?:als|embed(?:ding)?|vec(?:tor)?)\s*[:=]\s*([0-9.]+)/i,
    anchor: /\banchor\s*[:=]\s*([A-Za-z0-9_\-:.]+)/gi,
  } as const;

  for (const reason of reasons ?? []) {
    const pop = reason.match(rx.pop);
    if (pop) contrib.pop = Number(pop[1]);

    const cooc = reason.match(rx.cooc);
    if (cooc) contrib.cooc = Number(cooc[1]);

    const als = reason.match(rx.als);
    if (als) contrib.als = Number(als[1]);

    let match: RegExpExecArray | null;
    while ((match = rx.anchor.exec(reason)) !== null) {
      if (match[1]) anchors.push(match[1]);
    }

    if (!pop && !cooc && !als) {
      notes.push(reason);
    }
  }

  return { contrib, anchors, notes };
}

export function deriveExplainData(
  item: specs_types_ScoredItem | null | undefined,
  fallbackBlend: BlendTriplet
): ExplainDerived {
  const explain =
    (item as { explain?: ExplainBlock | null } | null)?.explain ?? null;
  const parsed = parseReasons(item?.reasons);

  const blendWeights: BlendTriplet = {
    pop: safeNumber(explain?.blend?.alpha ?? fallbackBlend.pop),
    cooc: safeNumber(explain?.blend?.beta ?? fallbackBlend.cooc),
    als: safeNumber(explain?.blend?.gamma ?? fallbackBlend.als),
  };

  const blendNorms: BlendTriplet = {
    pop: safeNumber(explain?.blend?.pop_norm),
    cooc: safeNumber(explain?.blend?.cooc_norm),
    als: safeNumber(explain?.blend?.emb_norm),
  };

  const explainContrib: BlendTriplet = {
    pop: safeNumber(explain?.blend?.contrib?.pop),
    cooc: safeNumber(explain?.blend?.contrib?.cooc),
    als: safeNumber(explain?.blend?.contrib?.emb),
  };

  let contributions: BlendTriplet = { ...ZERO };
  let usingExplainShares = false;

  const explainSum =
    explainContrib.pop + explainContrib.cooc + explainContrib.als;
  if (explainSum > 0) {
    contributions = explainContrib;
    usingExplainShares = true;
  }

  if (!usingExplainShares) {
    const normBased: BlendTriplet = {
      pop: blendWeights.pop * blendNorms.pop,
      cooc: blendWeights.cooc * blendNorms.cooc,
      als: blendWeights.als * blendNorms.als,
    };
    const normSum = normBased.pop + normBased.cooc + normBased.als;
    if (normSum > 0) {
      contributions = normBased;
      usingExplainShares = Boolean(explain?.blend);
    }
  }

  if (!usingExplainShares) {
    contributions = {
      pop: safeNumber(parsed.contrib.pop),
      cooc: safeNumber(parsed.contrib.cooc),
      als: safeNumber(parsed.contrib.als),
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

  const shares: BlendTriplet = {
    pop: contributions.pop / sum,
    cooc: contributions.cooc / sum,
    als: contributions.als / sum,
  };

  const anchorSource = (explain?.anchors ?? parsed.anchors ?? []).filter(
    (value): value is string => Boolean(value)
  );
  const anchors = Array.from(new Set(anchorSource));

  return {
    shares,
    contributions,
    anchors,
    notes: parsed.notes,
    reasons: item?.reasons ?? [],
    explainBlock: explain,
    usingExplainShares,
    blendWeights,
    blendNorms,
    rawSignals: explain?.blend?.raw,
  };
}

export function summarizeContributions(contributions: BlendTriplet): string {
  const total = contributions.pop + contributions.cooc + contributions.als;
  if (total === 0) {
    return "pop 0.00 · co 0.00 · emb 0.00";
  }
  const fmt = (value: number) => value.toFixed(2);
  return `pop ${fmt(contributions.pop)} · co ${fmt(
    contributions.cooc
  )} · emb ${fmt(contributions.als)}`;
}
