import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import {
  deleteAllItemsInNamespace,
  deleteAllUsersInNamespace,
  deleteAllEventsInNamespace,
} from "@/server/services/recsys";

export async function POST(req: NextRequest) {
  const body = (await req.json().catch(() => ({}))) as { tables?: string[] };
  const requested = new Set(
    (
      body.tables ?? [
        "events",
        "orderItem",
        "order",
        "cartItem",
        "cart",
        "product",
        "user",
      ]
    ).map((t) => t)
  );
  // Expand dependencies so FK constraints won't fail even if caller passes a subset
  if (requested.has("user")) {
    requested.add("events");
    requested.add("orderItem");
    requested.add("order");
    requested.add("cartItem");
    requested.add("cart");
  }
  if (requested.has("product")) {
    requested.add("orderItem");
    requested.add("cartItem");
  }
  if (requested.has("order")) {
    requested.add("orderItem");
  }
  if (requested.has("cart")) {
    requested.add("cartItem");
  }
  // Enforce safe deletion order (children first)
  const order = [
    "events",
    "orderItem",
    "order",
    "cartItem",
    "cart",
    "product",
    "user",
  ];
  for (const t of order) {
    if (!requested.has(t)) continue;
    if (t === "events") await prisma.event.deleteMany({});
    else if (t === "orderItem") await prisma.orderItem.deleteMany({});
    else if (t === "order") await prisma.order.deleteMany({});
    else if (t === "cartItem") await prisma.cartItem.deleteMany({});
    else if (t === "cart") await prisma.cart.deleteMany({});
    else if (t === "product") await prisma.product.deleteMany({});
    else if (t === "user") await prisma.user.deleteMany({});
  }

  // Sync deletions with Recsys
  try {
    if (requested.has("events")) {
      await deleteAllEventsInNamespace();
    }
    if (requested.has("product")) {
      // Namespace-wide items delete ensures full cleanup
      await deleteAllItemsInNamespace();
    }

    if (requested.has("user")) {
      // Delete all events first, then all users in namespace
      await deleteAllUsersInNamespace();
    }
  } catch (err) {
    console.error("Recsys sync during nuke failed:", err);
  }

  return NextResponse.json({ status: "nuked", tables: Array.from(requested) });
}
