import { NextResponse } from "next/server";
import { upsertEventTypeConfig } from "@/server/services/recsys";

export async function POST() {
  try {
    const result = await upsertEventTypeConfig();
    return NextResponse.json({
      status: "success",
      message: "Event type configuration updated",
      result,
    });
  } catch (error) {
    console.error("Failed to upsert event type config:", error);
    return NextResponse.json(
      { error: "Failed to update event type configuration" },
      { status: 500 }
    );
  }
}
