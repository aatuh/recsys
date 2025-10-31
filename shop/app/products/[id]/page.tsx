import { prisma } from "@/server/db/client";
import { getSimilar } from "@/server/services/recsys";
import { AddToCartButton } from "@/components/AddToCartButton";
import { ViewEventOnMount } from "@/components/ViewEventOnMount";
import ProductPlaceholder from "@/components/ProductPlaceholder";

export default async function ProductDetail({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const product = await prisma.product.findUnique({ where: { id } });
  if (!product) return <div className="p-6">Not found</div>;

  let similar: any[] = [];
  try {
    const s = await getSimilar({ itemId: product.id, k: 8 });
    similar = s as any[];
  } catch {}

  return (
    <main className="space-y-6">
      <ViewEventOnMount productId={product.id} />
      <section className="flex gap-6">
        {product.imageUrl ? (
          <img
            src={product.imageUrl}
            alt={product.name}
            className="w-60 h-60 object-cover border rounded"
          />
        ) : (
          <ProductPlaceholder
            seed={product.id}
            label={product.name}
            className="w-60 h-60 border rounded"
          />
        )}
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold">{product.name}</h1>
          <div className="text-sm text-gray-700">{product.brand}</div>
          <div className="text-sm text-gray-700">{product.category}</div>
          <div className="text-lg font-medium">
            ${product.price.toFixed(2)} {product.currency}
          </div>
          <AddToCartButton productId={product.id} />
        </div>
      </section>

      <section className="space-y-3">
        <h2 className="text-lg font-semibold">Similar items</h2>
        <ul className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {similar.map((it) => (
            <li key={it.item_id} className="border rounded p-3">
              <a
                href={`/products/${encodeURIComponent(it.item_id)}?rec=1`}
                className="text-sm underline"
              >
                {it.item_id}
              </a>
              <div className="text-xs text-gray-600">score {it.score}</div>
            </li>
          ))}
        </ul>
      </section>
    </main>
  );
}
