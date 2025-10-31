import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import {
  forwardEventsBatch,
  mapEventTypeToCode,
} from "@/server/services/recsys";

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
  const created = await prisma.$transaction(
    events.map((e) =>
      prisma.event.create({
        data: {
          userId: e.userId,
          productId: e.productId ?? null,
          type: e.type,
          value: e.value ?? 1,
          ts: e.ts ? new Date(e.ts) : undefined,
          metaText: e.meta ? JSON.stringify(e.meta) : null,
          sourceEventId: e.sourceEventId ?? null,
          recsysStatus: "pending",
        },
      })
    )
  );

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

  const payload = pending.map((e: any) => ({
    user_id: e.userId,
    item_id: e.productId ?? undefined,
    type: mapEventTypeToCode(e.type as any),
    value: e.value,
    ts: e.ts.toISOString(),
    meta: e.metaText ? safeParse(e.metaText) : undefined,
    source_event_id: e.id,
  }));

  try {
    await forwardEventsBatch(payload);
    const ids = pending.map((p: any) => p.id);
    await prisma.event.updateMany({
      where: { id: { in: ids } },
      data: { recsysStatus: "sent", sentAt: new Date() },
    });
    return NextResponse.json({ forwarded: pending.length });
  } catch {
    await prisma.event.updateMany({
      where: { id: { in: pending.map((p: any) => p.id) } },
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
