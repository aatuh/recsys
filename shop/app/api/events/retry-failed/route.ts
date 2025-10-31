import { NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import {
  forwardEventsBatch,
  mapEventTypeToCode,
} from "@/server/services/recsys";

export async function POST() {
  const failed = await prisma.event.findMany({
    where: { recsysStatus: "failed" },
    orderBy: { ts: "asc" },
    take: 200,
  });
  if (failed.length === 0) return NextResponse.json({ retried: 0 });
  const payload = failed.map((e: any) => ({
    user_id: e.userId,
    item_id: e.productId ?? undefined,
    type: mapEventTypeToCode(e.type as any),
    value: e.value,
    ts: e.ts.toISOString(),
    source_event_id: e.id,
  }));
  try {
    await forwardEventsBatch(payload);
    await prisma.event.updateMany({
      where: { id: { in: failed.map((f: any) => f.id) } },
      data: { recsysStatus: "sent", sentAt: new Date() },
    });
    return NextResponse.json({ retried: failed.length });
  } catch {
    return NextResponse.json({ error: "retry_failed" }, { status: 502 });
  }
}
