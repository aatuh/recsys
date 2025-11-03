import { NextRequest, NextResponse } from "next/server";
import { randomUUID } from "crypto";
import { prisma } from "@/server/db/client";
import { upsertUsers } from "@/server/services/recsys";
import { buildUserContract } from "@/lib/contracts/user";

export async function POST(req: NextRequest) {
  const {
    count = 20,
    locales,
    countries,
    loyaltyTiers,
    preferredCategories,
    devices,
    priceSensitivity,
    interests,
  } = (await req.json().catch(() => ({}))) as {
    count?: number;
    locales?: string[];
    countries?: string[];
    loyaltyTiers?: string[];
    preferredCategories?: string[];
    devices?: Array<"mobile" | "desktop">;
    priceSensitivity?: Array<"low" | "mid" | "high">;
    interests?: string[];
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

  const localePool =
    Array.isArray(locales) && locales.length > 0
      ? locales
      : ["en-US", "en-GB", "en-CA", "de-DE", "fr-FR"];
  const countryPool =
    Array.isArray(countries) && countries.length > 0
      ? countries
      : ["US", "GB", "CA", "DE", "FR", "AU"];
  const loyaltyPool =
    Array.isArray(loyaltyTiers) && loyaltyTiers.length > 0
      ? loyaltyTiers
      : ["bronze", "silver", "gold", "vip"];
  const categoryPool =
    Array.isArray(preferredCategories) && preferredCategories.length > 0
      ? preferredCategories
      : [
          "electronics",
          "audio",
          "fitness",
          "home",
          "outdoors",
          "wellness",
          "lifestyle",
        ];
  const devicePool =
    Array.isArray(devices) && devices.length > 0
      ? devices
      : (["mobile", "desktop"] as const);
  const pricePool =
    Array.isArray(priceSensitivity) && priceSensitivity.length > 0
      ? priceSensitivity
      : (["low", "mid", "high"] as const);
  const interestPool =
    Array.isArray(interests) && interests.length > 0
      ? interests
      : [
          "running",
          "cycling",
          "gaming",
          "photography",
          "cooking",
          "travel",
          "yoga",
          "wellness",
        ];

  const now = Date.now();
  const usersData = Array.from({ length: count }).map((_, index) => {
    const first = firsts[(index + now) % firsts.length];
    const last = lasts[(index * 7 + now) % lasts.length];
    const displayName = `${first} ${last}`;
    const preferred = shuffle(categoryPool)
      .slice(0, Math.max(2, Math.floor(Math.random() * 3) + 1))
      .map((category) => category.toLowerCase());
    const affinity: Record<string, number> = {};
    shuffle(categoryPool)
      .slice(0, 3)
      .forEach((category, idx) => {
        affinity[category] = Number(
          (0.4 + Math.random() * 0.6 + idx * 0.05).toFixed(2)
        );
      });
    const traitInterests = shuffle(interestPool).slice(0, 3);
    const locale = localePool[index % localePool.length];
    const country = countryPool[index % countryPool.length];
    const device = devicePool[index % devicePool.length];
    const loyalty =
      loyaltyPool[(index + Math.floor(Math.random() * loyaltyPool.length)) %
        loyaltyPool.length];
    const price = pricePool[index % pricePool.length];
    const lifetimeBuckets = ["L", "M", "H"] as const;
    const lifetimeValue =
      lifetimeBuckets[(index + Math.floor(Math.random() * 3)) % 3];

    const traits = {
      display_name: displayName,
      locale,
      country,
      device,
      signup_ts: new Date(now - Math.random() * 86400000 * 180).toISOString(),
      last_seen_ts: new Date().toISOString(),
      loyalty_tier: loyalty,
      newsletter: Math.random() < 0.45,
      preferred_categories: preferred,
      brand_affinity: affinity,
      interests: traitInterests,
      price_sensitivity: price,
      lifetime_value_bucket: lifetimeValue,
      active_subscription: Math.random() < 0.2,
      session_count_30d: Math.floor(Math.random() * 12),
      promotion_opt_in: Math.random() < 0.35,
    };

    return {
      id: randomUUID(),
      displayName,
      traitsText: JSON.stringify(traits),
    };
  });

  await prisma.user.createMany({
    data: usersData,
    skipDuplicates: true,
  });

  const createdUsers = await prisma.user.findMany({
    where: {
      id: { in: usersData.map((user) => user.id) },
    },
  });

  void upsertUsers(createdUsers.map((user) => buildUserContract(user))).catch(
    () => null
  );

  return NextResponse.json({ inserted: createdUsers.length });
}

function shuffle<T>(input: T[]): T[] {
  const copy = [...input];
  for (let i = copy.length - 1; i > 0; i -= 1) {
    const j = Math.floor(Math.random() * (i + 1));
    [copy[i], copy[j]] = [copy[j], copy[i]];
  }
  return copy;
}
