import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { upsertUsers } from "@/server/services/recsys";

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
  void upsertUsers([
    {
      user_id: item.id,
      traits: item.traitsText ? safeParseJson(item.traitsText) : undefined,
    },
  ]).catch(() => null);
  return NextResponse.json(item, { status: 201 });
}

function safeParseJson(s?: string | null): unknown | undefined {
  if (!s) return undefined;
  try {
    return JSON.parse(s);
  } catch {
    return undefined;
  }
}
