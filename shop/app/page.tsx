import { prisma } from "@/server/db/client";
import { ProductCard } from "@/components/ProductCard";
import { ClientRecommendations } from "@/components/ClientRecommendations";

export default async function HomePage() {
  const products = await prisma.product.findMany({
    orderBy: { createdAt: "desc" },
    take: 20,
  });

  return (
    <main className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Recsys Shop</h1>
        <p className="text-sm text-muted-foreground">Browse our catalog</p>
      </div>

      {/* Recommendations */}
      <ClientRecommendations surface="home" widget="home_top_picks" k={8} />

      {/* Product Catalog */}
      <div>
        <h2 className="text-xl font-semibold mb-4">All Products</h2>
        <ul className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {products.map(
            (p: {
              id: string;
              name: string;
              brand?: string;
              category?: string;
              price: number;
              currency: string;
              imageUrl?: string;
            }) => (
              <li key={p.id}>
                <ProductCard
                  product={p}
                  surface="home"
                  widget="home_catalog"
                  showId={true}
                />
              </li>
            )
          )}
        </ul>
      </div>
    </main>
  );
}
