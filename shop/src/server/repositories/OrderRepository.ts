import { prisma } from "@/server/db/client";

export const OrderRepository = {
  async createFromCart(userId: string) {
    const cart = await prisma.cart.findFirst({
      where: { userId },
      include: { items: true },
    });
    if (!cart || cart.items.length === 0) return null;
    return prisma.$transaction(async (tx: any) => {
      const total = cart.items.reduce(
        (s: number, i: any) => s + i.qty * i.unitPrice,
        0
      );
      const order = await tx.order.create({
        data: { userId, total, currency: "USD" },
      });
      for (const i of cart.items as any[]) {
        await tx.orderItem.create({
          data: {
            orderId: order.id,
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
      return order;
    });
  },
};
