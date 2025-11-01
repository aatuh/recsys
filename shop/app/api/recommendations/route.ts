import { NextRequest, NextResponse } from "next/server";
import { getRecommendations } from "@/server/services/recsys";
import { RecommendationConstraints } from "@/lib/recommendations/constraints";

export async function GET(req: NextRequest) {
  try {
    const { searchParams } = new URL(req.url);
    const userId = searchParams.get("userId");
    const k = Number(searchParams.get("k") ?? "8");
    const includeReasons = searchParams.get("includeReasons") === "true";

    if (!userId) {
      return NextResponse.json({ error: "userId required" }, { status: 400 });
    }

    // Parse constraints from query parameters
    const constraints: RecommendationConstraints = {};
    
    const minPrice = searchParams.get("minPrice");
    const maxPrice = searchParams.get("maxPrice");
    if (minPrice && maxPrice) {
      constraints.price_between = [Number(minPrice), Number(maxPrice)];
    }
    
    const includeTags = searchParams.get("includeTags");
    if (includeTags) {
      constraints.include_tags_any = includeTags.split(",").map(tag => tag.trim());
    }
    
    const excludeTags = searchParams.get("excludeTags");
    if (excludeTags) {
      constraints.exclude_tags_any = excludeTags.split(",").map(tag => tag.trim());
    }
    
    const brandCap = searchParams.get("brandCap");
    if (brandCap) {
      constraints.brand_cap = Number(brandCap);
    }
    
    const categoryCap = searchParams.get("categoryCap");
    if (categoryCap) {
      constraints.category_cap = Number(categoryCap);
    }

    const response = await getRecommendations({
      userId,
      k,
      includeReasons,
      constraints: Object.keys(constraints).length > 0 ? constraints : undefined,
    });

    return NextResponse.json(response);
  } catch (error) {
    console.error("Recommendation API error:", error);
    return NextResponse.json(
      { error: "Failed to get recommendations" },
      { status: 500 }
    );
  }
}
