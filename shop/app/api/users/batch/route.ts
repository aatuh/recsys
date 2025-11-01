import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { deleteUsers } from "@/server/services/recsys";

export async function POST(req: NextRequest) {
  const body = (await req.json()) as {
    action: "update" | "delete";
    ids: string[];
    data?: Record<string, unknown>;
  };
  if (!body.ids?.length)
    return NextResponse.json({ error: "ids required" }, { status: 400 });
  if (body.action === "delete") {
    const res = await prisma.user.deleteMany({
      where: { id: { in: body.ids } },
    });
    // Sync deletion with Recsys
    void deleteUsers(body.ids).catch(() => null);
    return NextResponse.json({ deleted: res.count });
  }
  if (body.action === "update") {
    if (!body.data)
      return NextResponse.json({ error: "data required" }, { status: 400 });
    const tx = body.ids.map((id) =>
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      prisma.user.update({ where: { id }, data: body.data as any })
    );
    await prisma.$transaction(tx);
    return NextResponse.json({ updated: body.ids.length });
  }
  return NextResponse.json({ error: "invalid action" }, { status: 400 });
}
