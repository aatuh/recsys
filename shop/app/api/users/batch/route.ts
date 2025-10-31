import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";

export async function POST(req: NextRequest) {
  const body = (await req.json()) as {
    action: "update" | "delete";
    ids: string[];
    data?: any;
  };
  if (!body.ids?.length)
    return NextResponse.json({ error: "ids required" }, { status: 400 });
  if (body.action === "delete") {
    const res = await prisma.user.deleteMany({
      where: { id: { in: body.ids } },
    });
    return NextResponse.json({ deleted: res.count });
  }
  if (body.action === "update") {
    const tx = body.ids.map((id) =>
      prisma.user.update({ where: { id }, data: body.data })
    );
    await prisma.$transaction(tx);
    return NextResponse.json({ updated: body.ids.length });
  }
  return NextResponse.json({ error: "invalid action" }, { status: 400 });
}
