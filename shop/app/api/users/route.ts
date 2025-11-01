import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { upsertUsers } from "@/server/services/recsys";
import { buildUserContract } from "@/lib/contracts/user";

export async function GET(req: NextRequest) {
  const { searchParams } = new URL(req.url);
  const limit = Number(searchParams.get("limit") ?? "20");
  const offset = Number(searchParams.get("offset") ?? "0");
  const q = searchParams.get("q") ?? undefined;

  const where = q ? { displayName: { contains: q, mode: "insensitive" } } : {};

  const [items, total] = await Promise.all([
    prisma.user.findMany({
      where,
      orderBy: { createdAt: "desc" },
      skip: offset,
      take: limit,
    }),
    prisma.user.count({ where }),
  ]);

  return NextResponse.json({ items, total, limit, offset });
}

export async function POST(req: NextRequest) {
  const body = await req.json();
  const item = await prisma.user.create({ data: body });
  void upsertUsers([buildUserContract(item)]).catch(() => null);
  return NextResponse.json(item, { status: 201 });
}
