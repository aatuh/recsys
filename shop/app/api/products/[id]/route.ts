import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { upsertItems, deleteItems } from "@/server/services/recsys";

export async function GET(
  _req: NextRequest,
  context: { params: Promise<{ id: string }> }
) {
  const { id } = await context.params;
  const item = await prisma.product.findUnique({ where: { id } });
  if (!item) return NextResponse.json({ error: "Not found" }, { status: 404 });
  return NextResponse.json(item);
}

export async function PATCH(
  req: NextRequest,
  context: { params: Promise<{ id: string }> }
) {
  const { id } = await context.params;
  const body = await req.json();
  const item = await prisma.product.update({
    where: { id },
    data: body,
  });
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
  return NextResponse.json(item);
}

export async function DELETE(
  _req: NextRequest,
  context: { params: Promise<{ id: string }> }
) {
  const { id } = await context.params;
  await prisma.product.delete({ where: { id } });
  void deleteItems([id]).catch(() => null);
  return NextResponse.json({ status: "deleted" });
}
