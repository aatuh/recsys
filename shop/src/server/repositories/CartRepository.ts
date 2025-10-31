import { prisma } from "@/server/db/client";

export const CartRepository = {
  async getOrCreate(userId: string) {
    let cart = await prisma.cart.findFirst({ where: { userId } });
    if (!cart) cart = await prisma.cart.create({ data: { userId } });
    return cart;
  },

  async addItem(
    cartId: string,
    productId: string,
    qty: number,
    unitPrice: number
  ) {
    const existing = await prisma.cartItem.findFirst({
      where: { cartId, productId },
    });
    if (existing) {
      return prisma.cartItem.update({
        where: { id: existing.id },
        data: { qty: existing.qty + qty },
      });
    }
    return prisma.cartItem.create({
      data: { cartId, productId, qty, unitPrice },
    });
  },

  async updateQty(cartId: string, productId: string, qty: number) {
    const item = await prisma.cartItem.findFirst({
      where: { cartId, productId },
    });
    if (!item) return null;
    if (qty <= 0) {
      await prisma.cartItem.delete({ where: { id: item.id } });
      return null;
    }
    return prisma.cartItem.update({ where: { id: item.id }, data: { qty } });
  },

  async clear(cartId: string) {
    await prisma.cartItem.deleteMany({ where: { cartId } });
  },
};
