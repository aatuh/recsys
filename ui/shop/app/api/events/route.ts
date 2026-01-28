import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { forwardEventsBatch } from "@/server/services/recsys";
import { buildEventContract } from "@/lib/contracts/event";
import { maybeLogColdStart } from "@/server/logging/coldStart";
import { normalizeEventPayload } from "@/server/normalizers/event";

export async function GET(req: NextRequest) {
  const { searchParams } = new URL(req.url);
  const limit = Number(searchParams.get("limit") ?? "50");
  const offset = Number(searchParams.get("offset") ?? "0");
  const type = searchParams.get("type") as
    | "view"
    | "click"
    | "add"
    | "purchase"
    | "custom"
    | null;

  const where = type ? { type } : {};
  const [items, total] = await Promise.all([
    prisma.event.findMany({
      where,
      orderBy: { ts: "desc" },
      skip: offset,
      take: limit,
    }),
    prisma.event.count({ where }),
  ]);

  return NextResponse.json({ items, total, limit, offset });
}

export async function POST(req: NextRequest) {
  const body = await req.json();
  const events = Array.isArray(body) ? body : [body];
  const normalized = events.map((event) => normalizeEventPayload(event));
  const created = await prisma.$transaction(
    normalized.map((item) =>
      prisma.event.create({
        data: item.data,
      })
    )
  );

  const validTypes = new Set(["view", "click", "add", "purchase"]);
  normalized.forEach(({ data, meta }) => {
    if (!validTypes.has(data.type) || !data.productId) {
      return;
    }
    // Map "view" to "impression" for cold start logging
    const coldStartType =
      data.type === "view"
        ? "impression"
        : (data.type as "click" | "add" | "purchase");
    maybeLogColdStart(
      coldStartType,
      data.userId,
      data.productId,
      meta as Record<string, unknown> | undefined
    );
  });

  return NextResponse.json({ inserted: created.length });
}

export async function PUT() {
  // Flush pending events to recsys
  const pending = await prisma.event.findMany({
    where: { recsysStatus: "pending" },
    orderBy: { ts: "asc" },
    take: 200,
  });

  if (pending.length === 0) {
    return NextResponse.json({ forwarded: 0 });
  }

  const payload = pending.map(
    (e: {
      userId: string;
      productId: string | null;
      type: string;
      value: number;
      ts: Date;
      metaText: string | null;
      id: string;
    }) =>
      buildEventContract({
        userId: e.userId,
        productId: e.productId,
        type: e.type as "view" | "click" | "add" | "purchase" | "custom",
        value: e.value,
        ts: e.ts.toISOString(),
        meta: e.metaText
          ? (safeParse(e.metaText) as Record<string, unknown>)
          : undefined,
        sourceEventId: e.id,
      })
  );

  try {
    await forwardEventsBatch(payload);
    const ids = pending.map((p: { id: string }) => p.id);
    await prisma.event.updateMany({
      where: { id: { in: ids } },
      data: { recsysStatus: "sent", sentAt: new Date() },
    });
    return NextResponse.json({ forwarded: pending.length });
  } catch {
    await prisma.event.updateMany({
      where: { id: { in: pending.map((p: { id: string }) => p.id) } },
      data: { recsysStatus: "failed" },
    });
    return NextResponse.json(
      { error: "forward_failed", count: pending.length },
      { status: 502 }
    );
  }
}

function safeParse(s: string): unknown | undefined {
  try {
    return JSON.parse(s);
  } catch {
    return undefined;
  }
}
