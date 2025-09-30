import {
  BanditService,
  ConfigService,
  IngestionService,
} from "../lib/api-client";
import { ensureApiBase } from "../lib/api";

export async function seedMinimal(
  apiBase: string,
  namespace: string
): Promise<void> {
  ensureApiBase(apiBase);

  const eventTypes = [
    { name: "view", type: 1, weight: 1, half_life_days: 30, is_active: true },
    { name: "click", type: 2, weight: 3, half_life_days: 21, is_active: true },
    {
      name: "purchase",
      type: 3,
      weight: 10,
      half_life_days: 60,
      is_active: true,
    },
  ];

  await ConfigService.upsertEventTypes({ namespace, types: eventTypes });

  const items = Array.from({ length: 24 }).map((_, i) => ({
    item_id: `item-${i + 1}`,
    price: 10 + (i % 8) * 7,
    available: true,
    tags: [
      `brand:${["NOVA", "ELMO", "ALFA"][i % 3]}`,
      `cat:${["sneaker", "boot", "runner"][i % 3]}`,
    ],
    props: { title: `Item ${i + 1}` },
  }));

  try {
    const titleLookup: Record<string, string> = {};
    const metaLookup: Record<
      string,
      {
        title?: string;
        brand?: string;
        price?: number;
        tags?: string[];
        available?: boolean;
      }
    > = {};
    for (const it of items) {
      const title = it.props?.title ?? it.item_id;
      if (!it.item_id) continue;
      titleLookup[it.item_id] = title;
      const brandTag = (it.tags || []).find((tag) => tag.startsWith("brand:"));
      metaLookup[it.item_id] = {
        title,
        brand: brandTag ? brandTag.split(":")[1] : undefined,
        price: it.price,
        tags: it.tags,
        available: it.available,
      };
    }
    localStorage.setItem(
      `demo:items:${namespace}`,
      JSON.stringify(titleLookup)
    );
    localStorage.setItem(
      `demo:itemMeta:${namespace}`,
      JSON.stringify(metaLookup)
    );
  } catch {
    // ignore storage errors in demo context
  }

  await IngestionService.upsertItems({ namespace, items });

  const users = Array.from({ length: 5 }).map((_, i) => ({
    user_id: `user-${i + 1}`,
    traits: { plan: ["free", "plus", "pro"][i % 3] },
  }));

  await IngestionService.upsertUsers({ namespace, users });

  const events = [] as Array<{
    user_id: string;
    item_id: string;
    type: number;
    ts: string;
    value?: number;
  }>;
  const now = Date.now();
  for (let u = 1; u <= 5; u++) {
    for (let e = 0; e < 30; e++) {
      const itemId = `item-${1 + Math.floor(Math.random() * 20)}`;
      const typ = [1, 1, 1, 2, 2, 3][Math.floor(Math.random() * 6)];
      const ageDays = Math.floor(Math.random() * 21);
      const ts = new Date(now - ageDays * 86400_000).toISOString();
      events.push({
        user_id: `user-${u}`,
        item_id: itemId,
        type: typ,
        ts,
        value: 1,
      });
    }
  }

  for (let i = 0; i < events.length; i += 500) {
    const batch = events.slice(i, i + 500);
    await IngestionService.batchEvents({ namespace, events: batch });
  }

  // Seed bandit policies for EDS-05 demo (Baseline vs Diverse)
  try {
    await BanditService.upsertBanditPolicies({
      namespace,
      policies: [
        {
          policy_id: "baseline",
          name: "Baseline",
          notes: "Balanced relevance with light diversity. Good default.",
          active: true,
          mmr_lambda: 0.2,
          brand_cap: 3,
          profile_boost: 1.0,
          popularity_fanout: 200,
        },
        {
          policy_id: "diverse",
          name: "Diverse",
          notes: "Emphasize exploration and brand variety.",
          active: true,
          mmr_lambda: 0.6,
          brand_cap: 2,
          profile_boost: 0.9,
          popularity_fanout: 200,
        },
      ],
    });
    // Ensure both policies are returned as candidates by the UI on first load
    // by marking them active and persisting minimal local state for the
    // current namespace. This is purely for demo UX resilience.
    try {
      localStorage.setItem(
        `demo:bandit:candidates:${namespace}`,
        JSON.stringify(["baseline", "diverse"])
      );
    } catch {
      // ignore localStorage failures
    }
  } catch {
    // Best-effort: keep demo resilient if bandit service is unavailable
  }
}
