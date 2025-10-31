import { NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import {
  forwardEventsBatch,
  mapEventTypeToCode,
} from "@/server/services/recsys";

export async function POST() {
  const pending = await prisma.event.findMany({
    where: { recsysStatus: "pending" },
    orderBy: { ts: "asc" },
    take: 500,
  });
  if (pending.length === 0) return NextResponse.json({ forwarded: 0 });

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
    await prisma.event.updateMany({
      where: { id: { in: pending.map((p: any) => p.id) } },
      data: { recsysStatus: "sent", sentAt: new Date() },
    });
    return NextResponse.json({ forwarded: pending.length });
  } catch {
    await prisma.event.updateMany({
      where: { id: { in: pending.map((p: any) => p.id) } },
      data: { recsysStatus: "failed" },
    });
    return NextResponse.json({ error: "forward_failed" }, { status: 502 });
  }
}

function safeParse(s: string): unknown | undefined {
  try {
    return JSON.parse(s);
  } catch {
    return undefined;
  }
}
