import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";

export async function POST(req: NextRequest) {
  const body = (await req.json()) as { action: "delete"; ids: string[] };
  if (body.action !== "delete" || !body.ids?.length) {
    return NextResponse.json({ error: "invalid" }, { status: 400 });
  }
  const res = await prisma.event.deleteMany({
    where: { id: { in: body.ids } },
  });
  return NextResponse.json({ deleted: res.count });
}
