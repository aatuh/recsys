import { NextRequest, NextResponse } from "next/server";
import { randomUUID } from "crypto";
import { prisma } from "@/server/db/client";
import { normalizeEventPayload } from "@/server/normalizers/event";

const DEFAULT_TYPES = ["view", "click", "add", "purchase"] as const;
const DEFAULT_SURFACES = ["home", "pdp", "cart", "checkout"];
const DEFAULT_WIDGETS = ["home_top_picks", "similar_items", "frequently_bought"];

function pickRandom<T>(array: T[]): T {
  return array[Math.floor(Math.random() * array.length)];
}

export async function POST(req: NextRequest) {
  const {
    count = 40,
    types,
    surfaces,
    widgets,
    includeBandit = true,
    includeColdStart = true,
  } = (await req.json().catch(() => ({}))) as {
    count?: number;
    types?: string[];
    surfaces?: string[];
    widgets?: string[];
    includeBandit?: boolean;
    includeColdStart?: boolean;
  };

  const limit = Math.max(1, Math.min(500, Math.floor(count)));

  const [users, products] = await Promise.all([
    prisma.user.findMany({
      select: { id: true },
      orderBy: { createdAt: "desc" },
      take: 200,
    }),
    prisma.product.findMany({
      select: { id: true },
      orderBy: { createdAt: "desc" },
      take: 200,
    }),
  ]);

  if (users.length === 0 || products.length === 0) {
    return NextResponse.json(
      { error: "Users and products are required to seed events" },
      { status: 400 }
    );
  }

  const requestedTypes =
    Array.isArray(types) && types.length > 0
      ? types
          .map((entry) => entry.toLowerCase())
          .filter((entry) =>
            DEFAULT_TYPES.includes(entry as (typeof DEFAULT_TYPES)[number])
          )
      : [];
  const typePool =
    requestedTypes.length > 0 ? requestedTypes : [...DEFAULT_TYPES];

  const surfacePool =
    Array.isArray(surfaces) && surfaces.length > 0
      ? surfaces
      : DEFAULT_SURFACES;

  const widgetPool =
    Array.isArray(widgets) && widgets.length > 0 ? widgets : DEFAULT_WIDGETS;

  const now = Date.now();
  const events = Array.from({ length: limit }).map(() => {
    const user = pickRandom(users).id;
    const product = pickRandom(products).id;
    const type = pickRandom(typePool) as (typeof DEFAULT_TYPES)[number];
    const value =
      type === "purchase" ? Number((1 + Math.random() * 3).toFixed(2)) : 1;
    const timestamp = new Date(
      now - Math.floor(Math.random() * 1000 * 60 * 60 * 24 * 7)
    ).toISOString();
    const meta: Record<string, unknown> = {
      surface: pickRandom(surfacePool),
      widget: pickRandom(widgetPool),
      request_id: `seed-${randomUUID()}`,
      rank: Math.floor(Math.random() * 12) + 1,
      recommended: true,
    };
    if (includeBandit) {
      meta.bandit_policy_id = "manual_explore_default";
      meta.bandit_request_id = meta.request_id;
      meta.bandit_algorithm = "epsilon_greedy";
      meta.bandit_explore = Math.random() < 0.3;
    }
    if (includeColdStart && type === "view" && Math.random() < 0.25) {
      meta.cold_start = true;
    }
    return {
      userId: user,
      productId: product,
      type,
      value,
      ts: timestamp,
      meta,
    };
  });

  const normalized = events.map((event) => normalizeEventPayload(event));

  await prisma.event.createMany({
    data: normalized.map(({ data }) => ({
      userId: data.userId,
      productId: data.productId,
      type: data.type,
      value: data.value,
      ts: data.ts,
      metaText: data.metaText,
      sourceEventId: data.sourceEventId,
      recsysStatus: data.recsysStatus,
    })),
  });

  return NextResponse.json({ inserted: normalized.length });
}
