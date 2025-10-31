import { prisma } from "@/server/db/client";
import { AddToCartButton } from "@/components/AddToCartButton";
import ProductPlaceholder from "@/components/ProductPlaceholder";

export default async function HomePage() {
  const products = await prisma.product.findMany({
    orderBy: { createdAt: "desc" },
    take: 20,
  });
  return (
    <main className="p-6 space-y-4">
      <h1 className="text-2xl font-semibold">Recsys Shop</h1>
      <p className="text-sm text-muted-foreground">Browse our catalog</p>
      <ul className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {products.map((p) => (
          <li key={p.id} className="border rounded p-3">
            {p.imageUrl ? (
              <img
                src={p.imageUrl}
                alt={p.name}
                className="w-full h-36 object-cover border rounded mb-2"
              />
            ) : (
              <ProductPlaceholder
                seed={p.id}
                label={p.name}
                className="w-full h-36 border rounded mb-2"
              />
            )}
            <a
              href={`/products/${p.id}`}
              data-product-id={p.id}
              className="text-sm font-medium"
            >
              {p.name}
            </a>
            <div className="text-xs text-gray-600">{p.brand}</div>
            <div className="mt-1 text-sm">
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
