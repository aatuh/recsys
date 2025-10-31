import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";

export async function GET(req: NextRequest) {
  const { searchParams } = new URL(req.url);
  const userId = searchParams.get("userId") ?? undefined;
  const limit = Number(searchParams.get("limit") ?? "20");
  const offset = Number(searchParams.get("offset") ?? "0");

  const where = userId ? { userId } : {};
  const [items, total] = await Promise.all([
    prisma.order.findMany({
      where,
      include: { items: true },
      orderBy: { createdAt: "desc" },
      skip: offset,
      take: limit,
    }),
    prisma.order.count({ where }),
  ]);
  return NextResponse.json({ items, total, limit, offset });
}
