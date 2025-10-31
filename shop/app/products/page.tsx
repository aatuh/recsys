import { prisma } from "@/server/db/client";
import { AddToCartButton } from "@/components/AddToCartButton";

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
        {items.map((p: any) => (
          <li key={p.id} className="border rounded p-3">
            <div className="text-sm font-medium">{p.name}</div>
            <div className="text-xs text-gray-600">{p.brand}</div>
            <div className="mt-2 text-sm">
              ${p.price.toFixed(2)} {p.currency}
            </div>
            <div className="mt-2">
              <AddToCartButton productId={p.id} />
            </div>
          </li>
        ))}
      </ul>
    </main>
  );
}
