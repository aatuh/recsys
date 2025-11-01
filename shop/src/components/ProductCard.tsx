import { AddToCartButton } from "@/components/AddToCartButton";
import ProductPlaceholder from "@/components/ProductPlaceholder";
import Image from "next/image";

interface ProductCardProps {
  product: {
    id: string;
    name: string;
    brand?: string;
    category?: string;
    price: number;
    currency: string;
    imageUrl?: string;
  };
  surface?: "home" | "pdp" | "cart" | "checkout" | "products";
  widget?: string;
  unitPrice?: number;
  // Recommendation-specific props
  recommended?: boolean;
  rank?: number;
  score?: number;
  showId?: boolean;
}

export function ProductCard({
  product,
  surface = "home",
  widget = "catalog",
  unitPrice,
  recommended = false,
  rank,
  score,
  showId = false,
}: ProductCardProps) {
  const displayPrice = unitPrice ?? product.price;

  return (
    <div className="border rounded p-3">
      {product.imageUrl ? (
        <Image
          src={product.imageUrl}
          alt={product.name}
          width={144}
          height={144}
          className="w-full h-36 object-cover border rounded mb-2"
        />
      ) : (
        <ProductPlaceholder
          seed={product.id}
          label={product.name}
          className="w-full h-36 border rounded mb-2"
        />
      )}

      <a
        href={`/products/${product.id}`}
        data-product-id={product.id}
        data-recommended={recommended ? "true" : undefined}
        data-widget={widget}
        data-rank={rank}
        className="text-sm font-medium block mb-1"
      >
        {product.name}
        {showId && (
          <span className="text-xs text-gray-500 font-mono ml-1">
            ({product.id})
          </span>
        )}
      </a>

      {product.brand && (
        <div className="text-xs text-gray-600">{product.brand}</div>
      )}

      {product.category && (
        <div className="text-xs text-gray-500">{product.category}</div>
      )}

      {/* Show score for recommended items */}
      {recommended && score !== undefined && (
        <div className="text-xs text-blue-600 font-medium">
          Score: {score.toFixed(3)}
        </div>
      )}

      <div className="mt-1 text-sm">
        ${displayPrice.toFixed(2)} {product.currency}
      </div>

      <div className="mt-2">
        <AddToCartButton
          productId={product.id}
          surface={surface}
          widget={widget}
          unitPrice={displayPrice}
          currency={product.currency}
          recommended={recommended}
          rank={rank}
        />
      </div>
    </div>
  );
}
