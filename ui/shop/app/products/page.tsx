import { prisma } from "@/server/db/client";
import { ProductCard } from "@/components/ProductCard";

export default async function ProductsPage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string; offset?: string; limit?: string }>;
}) {
  const sp = await searchParams;
  const limit = Number(sp.limit ?? "20");
  const offset = Number(sp.offset ?? "0");
  const q = sp.q;

  const where = q
    ? {
        OR: [
          { name: { contains: q, mode: "insensitive" } },
          { description: { contains: q, mode: "insensitive" } },
        ],
      }
    : {};

  const [items, total] = await Promise.all([
    prisma.product.findMany({
      where,
      orderBy: { createdAt: "desc" },
      skip: offset,
      take: limit,
    }),
    prisma.product.count({ where }),
  ]);

  return (
    <main className="space-y-4">
      <h1 className="text-xl font-semibold">Products</h1>
      <p className="text-sm text-gray-600">Total: {total}</p>
      <ul className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {items.map(
          (p: {
            id: string;
            name: string;
            brand?: string | null;
            category?: string | null;
            price: number;
            currency: string;
            imageUrl?: string | null;
          }) => (
            <li key={p.id}>
              <ProductCard
                product={{
                  id: p.id,
                  name: p.name,
                  brand: p.brand || undefined,
                  category: p.category || undefined,
                  price: p.price,
                  currency: p.currency,
                  imageUrl: p.imageUrl || undefined,
                }}
                surface="products"
                widget="products_page"
                showId={true}
              />
            </li>
          )
        )}
      </ul>
    </main>
  );
}
