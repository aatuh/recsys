import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { upsertItems } from "@/server/services/recsys";

export async function GET(req: NextRequest) {
  const { searchParams } = new URL(req.url);
  const limit = Number(searchParams.get("limit") ?? "20");
  const offset = Number(searchParams.get("offset") ?? "0");
  const q = searchParams.get("q") ?? undefined;

  const where = q
    ? {
        OR: [{ name: { contains: q } }, { description: { contains: q } }],
      }
    : {};

  const [items, total] = await Promise.all([
    prisma.product.findMany({
      where,
      orderBy: { createdAt: "desc" },
      skip: offset,
      take: limit,
    }),
    prisma.product.count({ where }),
  ]);

  return NextResponse.json({ items, total, limit, offset });
}

export async function POST(req: NextRequest) {
  const body = await req.json();
  const item = await prisma.product.create({ data: body });
  // Fire-and-forget upsert to recsys (non-blocking)
  void upsertItems([
    {
      item_id: item.id,
      available: item.stockCount > 0,
      price: item.price,
      tags: item.tagsCsv
        ? item.tagsCsv
            .split(",")
            .map((s: string) => s.trim())
            .filter(Boolean)
        : undefined,
    },
  ]).catch(() => null);
  return NextResponse.json(item, { status: 201 });
}
