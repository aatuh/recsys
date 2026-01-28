import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { deleteItems } from "@/server/services/recsys";
import { normalizeProductPatch } from "@/server/normalizers/product";

type BatchPayload = {
  action: "update" | "delete";
  ids: string[];
  data?: Record<string, unknown>;
};

export async function POST(req: NextRequest) {
  const body = (await req.json()) as BatchPayload;
  if (!body.ids?.length)
    return NextResponse.json({ error: "ids required" }, { status: 400 });
  if (body.action === "delete") {
    // Delete dependent records first to avoid foreign key constraints
    await prisma.cartItem.deleteMany({
      where: { productId: { in: body.ids } },
    });
    await prisma.orderItem.deleteMany({
      where: { productId: { in: body.ids } },
    });

    // Now delete the products
    const res = await prisma.product.deleteMany({
      where: { id: { in: body.ids } },
    });

    // Sync deletion with Recsys
    void deleteItems(body.ids).catch(() => null);
    return NextResponse.json({ deleted: res.count });
  }
  if (body.action === "update") {
    const { data } = body;
    if (!data)
      return NextResponse.json({ error: "data required" }, { status: 400 });
    const normalized = normalizeProductPatch(data);
    if (Object.keys(normalized).length === 0) {
      return NextResponse.json(
        { error: "no valid fields to update" },
        { status: 400 }
      );
    }
    const tx = body.ids.map((id) =>
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      prisma.product.update({ where: { id }, data: normalized as any })
    );
    await prisma.$transaction(tx);
    return NextResponse.json({ updated: body.ids.length });
  }
  return NextResponse.json({ error: "invalid action" }, { status: 400 });
}
