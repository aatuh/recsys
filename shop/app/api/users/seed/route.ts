import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/server/db/client";
import { upsertUsers } from "@/server/services/recsys";
import { buildUserContract } from "@/lib/contracts/user";

export async function POST(req: NextRequest) {
  const { count = 20 } = (await req.json().catch(() => ({}))) as {
    count?: number;
  };
  const firsts = [
    "Alex",
    "Avery",
    "Blake",
    "Casey",
    "Dakota",
    "Drew",
    "Elliot",
    "Emerson",
    "Emery",
    "Finley",
    "Harlow",
    "Harper",
    "Hayden",
    "Indigo",
    "Jamie",
    "Jordan",
    "Jules",
    "Kai",
    "Lennon",
    "Logan",
    "Lux",
    "Morgan",
    "Oakley",
    "Parker",
    "Peyton",
    "Phoenix",
    "Quinn",
    "Reese",
    "Remy",
    "Riley",
    "Rowan",
    "Sage",
    "Sam",
    "Sawyer",
    "Skylar",
    "Skyler",
    "Sloane",
    "Sydney",
    "Tatum",
    "Taylor",
    "Wren",
  ];
  const lasts = [
    "Adams",
    "Bailey",
    "Bell",
    "Brooks",
    "Brown",
    "Campbell",
    "Carter",
    "Collins",
    "Cook",
    "Cooper",
    "Cox",
    "Edwards",
    "Evans",
    "Garcia",
    "Gray",
    "Hall",
    "Hill",
    "Howard",
    "James",
    "Johnson",
    "Kelly",
    "King",
    "Lee",
    "Martinez",
    "Mitchell",
    "Morgan",
    "Morris",
    "Murphy",
    "Nelson",
    "Parker",
    "Perez",
    "Peterson",
    "Phillips",
    "Price",
    "Ramirez",
    "Reed",
    "Richardson",
    "Rivera",
    "Roberts",
    "Rogers",
    "Sanchez",
    "Sanders",
    "Smith",
    "Stewart",
    "Torres",
    "Turner",
    "Walker",
    "Ward",
    "Watson",
    "Young",
  ];
  const usersData = Array.from({ length: count }).map((_, i) => ({
    displayName: `${firsts[i % firsts.length]} ${
      lasts[(i * 7) % lasts.length]
    }`,
  }));
  
  await prisma.user.createMany({
    data: usersData,
  });
  
  // Upsert to recsys
  void upsertUsers(usersData.map((user, index) => buildUserContract({
    ...user,
    id: `temp-${Date.now()}-${index}`, // Generate temporary ID for contract
  }))).catch(() => null);
  
  return NextResponse.json({ inserted: count });
}
