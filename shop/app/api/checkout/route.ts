import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import {
  forwardEventsBatch,
  mapEventTypeToCode,
} from "@/server/services/recsys";

export async function POST(req: NextRequest) {
  const body = await req.json();
  const { userId } = body as { userId: string };
  if (!userId)
    return NextResponse.json({ error: "userId required" }, { status: 400 });

  const cart = await prisma.cart.findFirst({
    where: { userId },
    include: { items: { include: { product: true } } },
  });
  if (!cart || cart.items.length === 0) {
    return NextResponse.json({ error: "Cart empty" }, { status: 400 });
  }

  const order = await prisma.$transaction(async (tx: any) => {
    const total = cart.items.reduce(
      (s: number, i: any) => s + i.qty * i.unitPrice,
      0
    );
    const o = await tx.order.create({
      data: { userId, total, currency: "USD" },
    });
    for (const i of cart.items as any[]) {
      await tx.orderItem.create({
        data: {
          orderId: o.id,
          productId: i.productId,
          qty: i.qty,
          unitPrice: i.unitPrice,
        },
      });
      await tx.product.update({
        where: { id: i.productId },
        data: { stockCount: { decrement: i.qty } },
      });
    }
    await tx.cartItem.deleteMany({ where: { cartId: cart.id } });
    return o;
  });

  // Record purchase events and forward
  const events = (cart.items as any[]).map((i: any) => ({
    userId,
    productId: i.productId,
    type: "purchase",
    value: i.qty,
    ts: new Date().toISOString(),
  }));

  const created = await prisma.$transaction(
    events.map((e: any) =>
      prisma.event.create({
        data: {
          userId: e.userId,
          productId: e.productId,
          type: "purchase",
          value: e.value,
          ts: new Date(e.ts),
          recsysStatus: "pending",
        },
      })
    )
  );

  const payload = created.map((e: any) => ({
    user_id: e.userId,
    item_id: e.productId ?? undefined,
    type: mapEventTypeToCode("purchase"),
    value: e.value,
    ts: e.ts.toISOString(),
    source_event_id: e.id,
  }));
  await forwardEventsBatch(payload).catch(() => null);

  return NextResponse.json({ orderId: order.id, total: order.total });
}
