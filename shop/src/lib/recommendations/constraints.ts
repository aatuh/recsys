import { getRecommendations } from "@/server/services/recsys";

export interface RecommendationConstraints {
  price_between?: [number, number];
  include_tags_any?: string[];
  exclude_tags_any?: string[];
  brand_cap?: number;
  category_cap?: number;
}

export interface RecommendationParams {
  userId: string;
  k?: number;
  includeReasons?: boolean;
  constraints?: RecommendationConstraints;
}

export async function getConstrainedRecommendations(
  params: RecommendationParams
) {
  const { userId, k = 8, includeReasons = false, constraints } = params;

  try {
    // For now, we'll use the basic recommendations endpoint
    // In a full implementation, we'd pass constraints to the Recsys API
    const response = await getRecommendations({
      userId,
      k,
      includeReasons,
    });

    // Apply client-side filtering if constraints are provided
    if (constraints && response.items) {
      let filteredItems = response.items;

      // Filter by price range
      if (constraints.price_between) {
        const [,] = constraints.price_between;
        filteredItems = filteredItems.filter(() => {
          // Note: This assumes items have price info in their metadata
          // In a real implementation, you'd need to fetch product details
          return true; // Placeholder - would need product price data
        });
      }

      // Filter by included tags
      if (
        constraints.include_tags_any &&
        constraints.include_tags_any.length > 0
      ) {
        filteredItems = filteredItems.filter(() => {
          // Note: This assumes items have tag info in their metadata
          // In a real implementation, you'd need to fetch product tags
          return true; // Placeholder - would need product tag data
        });
      }

      // Apply brand/category caps
      if (constraints.brand_cap || constraints.category_cap) {
        filteredItems = filteredItems.filter(() => {
          // Note: This assumes items have brand/category info in their metadata
          // In a real implementation, you'd need to fetch product details
          return true; // Placeholder - would need product brand/category data
        });
      }

      response.items = filteredItems.slice(0, k);
    }

    return response;
  } catch (error) {
    console.error("Failed to get constrained recommendations:", error);
    throw error;
  }
}

// Helper function to build tag constraints from UI filters
export function buildTagConstraints(filters: {
  brands?: string[];
  categories?: string[];
  priceRange?: [number, number];
}): RecommendationConstraints {
  const constraints: RecommendationConstraints = {};

  if (filters.brands && filters.brands.length > 0) {
    constraints.include_tags_any = filters.brands.map(
      (brand) => `brand:${brand}`
    );
  }

  if (filters.categories && filters.categories.length > 0) {
    const categoryTags = filters.categories.map((cat) => `category:${cat}`);
    constraints.include_tags_any = [
      ...(constraints.include_tags_any || []),
      ...categoryTags,
    ];
  }

  if (filters.priceRange) {
    constraints.price_between = filters.priceRange;
  }

  return constraints;
}
