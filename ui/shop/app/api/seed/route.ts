import { NextResponse } from "next/server";
import { PrismaClient } from "@prisma/client";

export async function POST() {
  const prisma = new PrismaClient();
  // Reset all tables
  await prisma.cartItem.deleteMany({});
  await prisma.cart.deleteMany({});
  await prisma.orderItem.deleteMany({});
  await prisma.order.deleteMany({});
  await prisma.event.deleteMany({});
  await prisma.product.deleteMany({});
  await prisma.user.deleteMany({});

  // Seed users
  await prisma.user.createMany({
    data: Array.from({ length: 10 }).map((_, i) => ({
      displayName: `User ${i + 1}`,
    })),
  });

  // Seed products
  const brands = ["Acme", "Globex", "Umbrella", "Stark", "Wayne"];
  const categories = ["Shoes", "Electronics", "Home", "Books", "Toys"];
  await prisma.product.createMany({
    data: Array.from({ length: 100 }).map((_, i) => ({
      sku: `SKU-${i + 1}`,
      name: `Product ${i + 1}`,
      description: "Demo product",
      price: Math.round((10 + Math.random() * 190) * 100) / 100,
      currency: "USD",
      brand: brands[i % brands.length],
      category: categories[i % categories.length],
      imageUrl: `https://picsum.photos/seed/${i}/600/600`,
      stockCount: Math.floor(Math.random() * 50) + 1,
      tagsCsv: categories[i % categories.length].toLowerCase(),
    })),
  });

  await prisma.$disconnect();
  return NextResponse.json({ status: "seeded" });
}
