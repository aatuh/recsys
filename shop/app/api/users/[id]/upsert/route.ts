import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { upsertUsers } from "@/server/services/recsys";

export async function POST(
  _req: NextRequest,
  context: { params: Promise<{ id: string }> }
) {
  const { id } = await context.params;
  const user = await prisma.user.findUnique({ where: { id } });
  if (!user) return NextResponse.json({ error: "Not found" }, { status: 404 });
  let traits: unknown | undefined = undefined;
  if (user.traitsText) {
    try {
      traits = JSON.parse(user.traitsText);
    } catch {}
  }
  await upsertUsers([{ user_id: user.id, traits }]).catch(() => null);
  return NextResponse.json({ status: "ok" });
}
