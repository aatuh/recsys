import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { upsertUsers } from "@/server/services/recsys";
import { buildUserContract } from "@/lib/contracts/user";

export async function POST(
  _req: NextRequest,
  context: { params: Promise<{ id: string }> }
) {
  const { id } = await context.params;
  const user = await prisma.user.findUnique({ where: { id } });
  if (!user) return NextResponse.json({ error: "Not found" }, { status: 404 });
  
  await upsertUsers([buildUserContract(user)]).catch(() => null);
  return NextResponse.json({ status: "ok" });
}
