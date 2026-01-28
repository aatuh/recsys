import pkg from "@prisma/client";
const { PrismaClient } = pkg as any;
const prisma = new PrismaClient();

async function main() {
  const users = await prisma.user.createMany({
    data: Array.from({ length: 10 }).map((_, i) => ({
      displayName: `User ${i + 1}`,
    })),
    skipDuplicates: true as any,
  });

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
    skipDuplicates: true as any,
  });

  console.log("Seed complete", users);
}

main()
  .catch((e) => {
    console.error(e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
