"use client";
import { useEffect, useState } from "react";
import { ProductCard } from "@/components/ProductCard";
import { RecommendationConstraints } from "@/lib/recommendations/constraints";
import { BanditMeta } from "@/lib/recommendations/bandit";

interface RecommendationWidgetProps {
  userId: string;
  surface: "home" | "pdp" | "cart" | "checkout";
  widget: string;
  k?: number;
  className?: string;
  constraints?: RecommendationConstraints;
  profileId?: string;
}

interface RecommendationItem {
  item_id: string;
  score: number;
  reasons?: string[];
  metadata?: {
    cold_start?: boolean;
  };
}

interface Product {
  id: string;
  name: string;
  brand?: string;
  category?: string;
  price: number;
  currency: string;
  imageUrl?: string;
}

export function RecommendationWidget({
  userId,
  surface,
  widget,
  k = 8,
  className = "",
  constraints,
  profileId,
}: RecommendationWidgetProps) {
  const [items, setItems] = useState<RecommendationItem[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [banditMeta, setBanditMeta] = useState<BanditMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!userId) {
      return;
    }

    let cancelled = false;
    const abortController = new AbortController();

    async function fetchRecommendations() {
      try {
        setLoading(true);
        setError(null);

        const params = new URLSearchParams({
          userId,
          k: k.toString(),
          includeReasons: "true",
          surface,
          widget,
        });

        if (profileId) {
          params.set("profileId", profileId);
        }

        if (constraints) {
          if (constraints.price_between) {
            params.set("minPrice", constraints.price_between[0].toString());
            params.set("maxPrice", constraints.price_between[1].toString());
          }
          if (constraints.include_tags_any) {
            params.set("includeTags", constraints.include_tags_any.join(","));
          }
          if (constraints.exclude_tags_any) {
            params.set("excludeTags", constraints.exclude_tags_any.join(","));
          }
          if (constraints.brand_cap) {
            params.set("brandCap", constraints.brand_cap.toString());
          }
          if (constraints.category_cap) {
            params.set("categoryCap", constraints.category_cap.toString());
          }
        }

        const response = await fetch(`/api/recommendations?${params}`, {
          signal: abortController.signal,
        });
        if (!response.ok) {
          throw new Error(
            `Failed to fetch recommendations: ${response.status}`
          );
        }

        const data = await response.json();
        if (cancelled) return;

        if (data.bandit) {
          setBanditMeta({
            policyId: data.bandit.chosen_policy_id ?? undefined,
            requestId: data.bandit.request_id ?? undefined,
            algorithm: data.bandit.algorithm ?? undefined,
            bucket:
              data.bandit.bandit_bucket ?? data.bandit.bucket ?? undefined,
            explore: data.bandit.explore ?? undefined,
            experiment:
              data.bandit.bandit_experiment ?? data.bandit.experiment ?? undefined,
            variant:
              data.bandit.bandit_variant ?? data.bandit.variant ?? undefined,
          });
        } else {
          setBanditMeta(null);
        }

        const recommendationItems: RecommendationItem[] = data.items || [];
        setItems(recommendationItems);

        if (recommendationItems.length > 0) {
          const productIds = recommendationItems.map(
            (item: RecommendationItem) => item.item_id
          );
          const productsResponse = await fetch(
            `/api/products?ids=${productIds.join(",")}`,
            { signal: abortController.signal }
          );
          if (!productsResponse.ok) {
            throw new Error(
              `Failed to load products: ${productsResponse.status}`
            );
          }
          const productsData = await productsResponse.json();
          if (cancelled) return;
          setProducts(productsData.items || []);
        } else {
          setProducts([]);
        }
      } catch (err) {
        if (err instanceof DOMException && err.name === "AbortError") {
          return;
        }
        console.error("Failed to fetch recommendations:", err);
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Unknown error");
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    fetchRecommendations();

    return () => {
      cancelled = true;
      abortController.abort();
    };
  }, [userId, k, constraints, surface, widget, profileId]);

  if (loading) {
    return (
      <div className={`space-y-3 ${className}`}>
        <h3 className="text-lg font-semibold">Recommended for you</h3>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {Array.from({ length: k }).map((_, i) => (
            <div key={i} className="border rounded p-3 animate-pulse">
              <div className="w-full h-36 bg-gray-200 rounded mb-2"></div>
              <div className="h-4 bg-gray-200 rounded mb-1"></div>
              <div className="h-3 bg-gray-200 rounded w-2/3"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`space-y-3 ${className}`}>
        <h3 className="text-lg font-semibold">Recommended for you</h3>
        <div className="text-sm text-red-600">
          Failed to load recommendations: {error}
        </div>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className={`space-y-3 ${className}`}>
        <h3 className="text-lg font-semibold">Recommended for you</h3>
        <div className="text-sm text-gray-500">
          No recommendations available yet. Try browsing some products!
        </div>
      </div>
    );
  }

  return (
    <div className={`space-y-3 ${className}`}>
      <h3 className="text-lg font-semibold">Recommended for you</h3>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {items.map((item, index) => {
          const product = products.find((p) => p.id === item.item_id);
          const isColdStart =
            item.metadata?.cold_start || item.reasons?.includes("cold_start");

          // If we have product data, use ProductCard, otherwise show fallback
          if (product) {
            return (
              <ProductCard
                key={item.item_id}
                product={product}
                surface={surface}
                widget={widget}
                recommended={true}
                rank={index + 1}
                score={item.score}
                coldStart={isColdStart}
                showId={true}
                banditMeta={banditMeta ?? undefined}
              />
            );
          }

          return (
            <ProductCard
              key={item.item_id}
              product={{
                id: item.item_id,
                name: `Product ${item.item_id}`,
                price: 0,
                currency: "USD",
              }}
              surface={surface}
              widget={widget}
              recommended={true}
              rank={index + 1}
              score={item.score}
              coldStart={isColdStart}
              showId={true}
              banditMeta={banditMeta ?? undefined}
            />
          );
        })}
      </div>
    </div>
  );
}
