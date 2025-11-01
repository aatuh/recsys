import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { deleteEvents } from "@/server/services/recsys";

export async function POST(req: NextRequest) {
  const body = (await req.json()) as {
    action: "delete" | "delete-pending" | "delete-from-recsys";
    ids?: string[];
  };

  if (body.action === "delete") {
    if (!body.ids?.length) {
      return NextResponse.json({ error: "invalid" }, { status: 400 });
    }
    const res = await prisma.event.deleteMany({
      where: { id: { in: body.ids } },
    });
    return NextResponse.json({ deleted: res.count });
  }

  if (body.action === "delete-pending") {
    const res = await prisma.event.deleteMany({
      where: { recsysStatus: "pending" },
    });
    return NextResponse.json({ deleted: res.count });
  }

  if (body.action === "delete-from-recsys") {
    if (!body.ids?.length) {
      return NextResponse.json({ error: "invalid" }, { status: 400 });
    }

    try {
      // Delete from Recsys
      await deleteEvents(body.ids);

      // Update local status to indicate deletion from Recsys
      const res = await prisma.event.updateMany({
        where: { id: { in: body.ids } },
        data: { recsysStatus: "deleted_from_recsys" },
      });

      return NextResponse.json({ deleted: res.count });
    } catch (error) {
      console.error("Failed to delete events from Recsys:", error);
      return NextResponse.json(
        { error: "recsys_deletion_failed" },
        { status: 502 }
      );
    }
  }

  return NextResponse.json({ error: "invalid action" }, { status: 400 });
}
