import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";

export async function GET(req: NextRequest) {
  const { searchParams } = new URL(req.url);
  const limit = Number(searchParams.get("limit") ?? "20");
  const offset = Number(searchParams.get("offset") ?? "0");

  // Find carts with items and aggregate item count per cart
  const [carts, total] = await Promise.all([
    prisma.cart.findMany({
      orderBy: { updatedAt: "desc" },
      skip: offset,
      take: limit,
      include: { items: true },
    }),
    prisma.cart.count(),
  ]);

  const items = carts.map((c) => ({
    id: c.id,
    userId: c.userId,
    items: c.items.reduce((sum, it) => sum + it.qty, 0),
    updatedAt: c.updatedAt,
  }));

  return NextResponse.json({ items, total });
}
