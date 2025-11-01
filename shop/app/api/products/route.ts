import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { upsertItems } from "@/server/services/recsys";
import { buildItemContract } from "@/lib/contracts/item";

export async function GET(req: NextRequest) {
  const { searchParams } = new URL(req.url);
  const limit = Number(searchParams.get("limit") ?? "20");
  const offset = Number(searchParams.get("offset") ?? "0");
  const q = searchParams.get("q") ?? undefined;
  const ids = searchParams.get("ids") ?? undefined;

  let where: Record<string, unknown> = {};

  // If specific IDs are requested, fetch only those
  if (ids) {
    const idList = ids.split(",").filter(Boolean);
    where = { id: { in: idList } };
  } else if (q) {
    // Otherwise, use search query
    where = {
      OR: [{ name: { contains: q } }, { description: { contains: q } }],
    };
  }

  const [items, total] = await Promise.all([
    prisma.product.findMany({
      where,
      orderBy: ids ? undefined : { createdAt: "desc" }, // Don't order when fetching specific IDs
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
  void upsertItems([buildItemContract(item)]).catch((error) => {
    console.error("Failed to sync product to recsys:", error);
  });
  return NextResponse.json(item, { status: 201 });
}
