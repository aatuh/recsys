import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";

export async function GET(req: NextRequest) {
  const { searchParams } = new URL(req.url);
  const userId = searchParams.get("userId");
  if (!userId)
    return NextResponse.json({ error: "userId required" }, { status: 400 });

  let cart = await prisma.cart.findFirst({
    where: { userId },
    include: { items: { include: { product: true } } },
  });
  if (!cart) {
    cart = await prisma.cart.create({
      data: { userId },
      include: { items: { include: { product: true } } },
    });
  }

  // Ensure we have items included (if created path returned without include)
  if (!cart.items) {
    cart = (await prisma.cart.findFirst({
      where: { userId },
      include: { items: { include: { product: true } } },
    }))!;
  }

  const total = cart.items.reduce(
    (sum: number, ci: any) => sum + ci.qty * ci.unitPrice,
    0
  );
  return NextResponse.json({ cart, total });
}

export async function POST(req: NextRequest) {
  const body = await req.json();
  const { userId, productId, qty } = body as {
    userId: string;
    productId: string;
    qty: number;
  };
  if (!userId || !productId || !qty) {
    return NextResponse.json(
      { error: "userId, productId, qty required" },
      { status: 400 }
    );
  }

  const product = await prisma.product.findUnique({ where: { id: productId } });
  if (!product)
    return NextResponse.json({ error: "Product not found" }, { status: 404 });
  if (qty < 1)
    return NextResponse.json({ error: "Invalid qty" }, { status: 400 });

  // Upsert by userId requires a unique constraint; emulate with find/create
  let cart = await prisma.cart.findFirst({ where: { userId } });
  if (!cart) {
    cart = await prisma.cart.create({ data: { userId } });
  }

  const existing = await prisma.cartItem.findFirst({
    where: { cartId: cart.id, productId },
  });
  if (existing) {
    const updated = await prisma.cartItem.update({
      where: { id: existing.id },
      data: { qty: existing.qty + qty },
    });
    return NextResponse.json(updated, { status: 200 });
  }

  const created = await prisma.cartItem.create({
    data: { cartId: cart.id, productId, qty, unitPrice: product.price },
  });
  return NextResponse.json(created, { status: 201 });
}

export async function PATCH(req: NextRequest) {
  const body = await req.json();
  const { userId, productId, qty } = body as {
    userId: string;
    productId: string;
    qty: number;
  };
  const cart = await prisma.cart.findFirst({ where: { userId } });
  if (!cart)
    return NextResponse.json({ error: "Cart not found" }, { status: 404 });
  const item = await prisma.cartItem.findFirst({
    where: { cartId: cart.id, productId },
  });
  if (!item)
    return NextResponse.json({ error: "Item not found" }, { status: 404 });
  if (qty <= 0) {
    await prisma.cartItem.delete({ where: { id: item.id } });
    return NextResponse.json({ status: "deleted" });
  }
  const updated = await prisma.cartItem.update({
    where: { id: item.id },
    data: { qty },
  });
  return NextResponse.json(updated);
}

export async function DELETE(req: NextRequest) {
  const { searchParams } = new URL(req.url);
  const userId = searchParams.get("userId");
  const productId = searchParams.get("productId");
  const cart = userId
    ? await prisma.cart.findFirst({ where: { userId } })
    : null;
  if (!cart)
    return NextResponse.json({ error: "Cart not found" }, { status: 404 });
  if (!productId) {
    await prisma.cartItem.deleteMany({ where: { cartId: cart.id } });
    return NextResponse.json({ status: "cleared" });
  }
  await prisma.cartItem.deleteMany({ where: { cartId: cart.id, productId } });
  return NextResponse.json({ status: "deleted" });
}
