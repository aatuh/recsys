import { NextRequest, NextResponse } from "next/server";
import {
  syncAllItemsToRecsys,
  syncItemAvailability,
  syncItemPrice,
  syncItemTags,
} from "@/server/services/itemSync";

export async function POST(req: NextRequest) {
  try {
    const { action = "sync-all", productId } = await req
      .json()
      .catch(() => ({}));

    switch (action) {
      case "sync-all":
        const result = await syncAllItemsToRecsys();
        return NextResponse.json({
          status: "success",
          synced: result.synced || "all",
        });

      case "sync-availability":
        if (!productId) {
          return NextResponse.json(
            { error: "productId required" },
            { status: 400 }
          );
        }
        await syncItemAvailability(productId);
        return NextResponse.json({ status: "success" });

      case "sync-price":
        if (!productId) {
          return NextResponse.json(
            { error: "productId required" },
            { status: 400 }
          );
        }
        await syncItemPrice(productId);
        return NextResponse.json({ status: "success" });

      case "sync-tags":
        if (!productId) {
          return NextResponse.json(
            { error: "productId required" },
            { status: 400 }
          );
        }
        await syncItemTags(productId);
        return NextResponse.json({ status: "success" });

      default:
        return NextResponse.json({ error: "Invalid action" }, { status: 400 });
    }
  } catch (error) {
    console.error("Item sync error:", error);
    return NextResponse.json(
      { error: "Sync operation failed" },
      { status: 500 }
    );
  }
}
