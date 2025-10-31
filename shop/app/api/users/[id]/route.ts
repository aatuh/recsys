import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";

export async function GET(
  _req: NextRequest,
  context: { params: Promise<{ id: string }> }
) {
  const { id } = await context.params;
  const item = await prisma.user.findUnique({ where: { id } });
  if (!item) return NextResponse.json({ error: "Not found" }, { status: 404 });
  return NextResponse.json(item);
}

export async function PATCH(
  req: NextRequest,
  context: { params: Promise<{ id: string }> }
) {
  const { id } = await context.params;
  const body = await req.json();
  const item = await prisma.user.update({ where: { id }, data: body });
  return NextResponse.json(item);
}

export async function DELETE(
  _req: NextRequest,
  context: { params: Promise<{ id: string }> }
) {
  const { id } = await context.params;
  await prisma.user.delete({ where: { id } });
  return NextResponse.json({ status: "deleted" });
}
