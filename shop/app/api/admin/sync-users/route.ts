import { NextRequest, NextResponse } from "next/server";
import { syncAllUsersToRecsys, syncUserTraits, updateUserLastSeen, enrichUserTraits } from "@/server/services/userSync";

export async function POST(req: NextRequest) {
  try {
    const { action, userId, traits } = await req.json().catch(() => ({}));
    
    switch (action) {
      case "sync-all":
        const result = await syncAllUsersToRecsys();
        return NextResponse.json({ status: "success", ...result });
        
      case "sync-traits":
        if (!userId) {
          return NextResponse.json({ error: "userId required" }, { status: 400 });
        }
        await syncUserTraits(userId);
        return NextResponse.json({ status: "success" });
        
      case "update-last-seen":
        if (!userId) {
          return NextResponse.json({ error: "userId required" }, { status: 400 });
        }
        await updateUserLastSeen(userId);
        return NextResponse.json({ status: "success" });
        
      case "enrich-traits":
        if (!userId || !traits) {
          return NextResponse.json({ error: "userId and traits required" }, { status: 400 });
        }
        await enrichUserTraits(userId, traits);
        return NextResponse.json({ status: "success" });
        
      default:
        return NextResponse.json({ error: "Invalid action" }, { status: 400 });
    }
  } catch (error) {
    console.error("User sync error:", error);
    return NextResponse.json(
      { error: "Sync operation failed" },
      { status: 500 }
    );
  }
}
