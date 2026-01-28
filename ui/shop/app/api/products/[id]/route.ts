import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { upsertItems, deleteItems } from "@/server/services/recsys";
import { buildItemContract } from "@/lib/contracts/item";

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

  // Sync changes to Recsys
  void upsertItems([buildItemContract(item)]).catch((error) => {
    console.error("Failed to sync product update to recsys:", error);
  });

  return NextResponse.json(item);
}

export async function DELETE(
  _req: NextRequest,
  context: { params: Promise<{ id: string }> }
) {
  const { id } = await context.params;

  // Delete dependent records first to avoid foreign key constraints
  await prisma.cartItem.deleteMany({
    where: { productId: id },
  });
  await prisma.orderItem.deleteMany({
    where: { productId: id },
  });

  // Now delete the product
  await prisma.product.delete({ where: { id } });
  void deleteItems([id]).catch(() => null);
  return NextResponse.json({ status: "deleted" });
}
