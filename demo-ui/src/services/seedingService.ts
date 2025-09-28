import {
  randChoice,
  randInt,
  id,
  iso,
  daysAgo,
  randomBoolean,
  weightedChoice,
} from "../utils/helpers";
import { upsertEventTypes, upsertUsers, upsertItems, batchEvents } from "./api";
import type { TraitConfig } from "../components/sections/UserTraitsEditor";
import type {
  ItemConfig,
  PriceRange,
} from "../components/sections/ItemConfigEditor";
import type { EventTypeConfig } from "../types";

/**
 * Seeding service for generating and uploading synthetic data.
 */

export function buildUsers(
  n: number,
  startIndex: number = 1,
  traitConfigs: TraitConfig[] = []
): any[] {
  const users: any[] = [];
  for (let i = 0; i < n; i++) {
    const userIndex = startIndex + i;
    const traits: Record<string, any> = {};

    // Generate traits based on configuration
    traitConfigs.forEach((config) => {
      if (randomBoolean(config.probability)) {
        const values = config.values.map((v) => v.value);
        const weights = config.values.map((v) => v.probability);
        traits[config.key] = weightedChoice(values, weights);
      }
    });

    // Fallback to default plan trait if no configs provided
    if (Object.keys(traits).length === 0) {
      traits.plan = randChoice(["free", "plus", "pro"]);
    }

    users.push({
      user_id: id("user", userIndex),
      traits,
    });
  }
  return users;
}

export function buildItems(
  n: number,
  brands: string[],
  tags: string[],
  itemConfigs: ItemConfig[] = [],
  priceRanges: PriceRange[] = []
): any[] {
  const items: any[] = [];
  for (let i = 1; i <= n; i++) {
    const brand = randChoice(brands);
    const tagCount = randInt(1, 3);
    const itemTags = new Set<string>();
    for (let t = 0; t < tagCount; t++) {
      itemTags.add(randChoice(tags));
    }

    // Generate price from price ranges or fallback to random
    let price: number;
    if (priceRanges.length > 0) {
      const selectedRange = selectWeightedPriceRange(priceRanges);
      price = Math.floor(
        Math.random() * (selectedRange.max - selectedRange.min + 1) +
          selectedRange.min
      );
    } else {
      price = randInt(5, 200);
    }

    // Generate properties based on configuration
    const props: Record<string, any> = { brand };
    itemConfigs.forEach((config) => {
      if (randomBoolean(config.probability)) {
        const values = config.values.map((v) => v.value);
        const weights = config.values.map((v) => v.probability);
        props[config.key] = weightedChoice(values, weights);
      }
    });

    items.push({
      item_id: id("item", i),
      price,
      available: Math.random() > 0.05,
      tags: Array.from(itemTags),
      props,
    });
  }
  return items;
}

// Helper function to select weighted price range
function selectWeightedPriceRange(ranges: PriceRange[]): PriceRange {
  const totalWeight = ranges.reduce((sum, r) => sum + r.probability, 0);
  let random = Math.random() * totalWeight;

  for (const range of ranges) {
    random -= range.probability;
    if (random <= 0) {
      return range;
    }
  }
  return ranges[ranges.length - 1] || { min: 10, max: 50, probability: 1.0 };
}

export function buildEvents(
  users: any[],
  items: any[],
  minEventsPerUser: number,
  maxEventsPerUser: number,
  eventTypes: EventTypeConfig[]
): any[] {
  const res: any[] = [];
  const recentBiasDays = 21; // more recent events

  // Create weighted event type selection based on event types configuration
  const eventTypeIndexes = eventTypes.map((et) => et.index);
  const eventTypeWeights = eventTypes.map((et) => et.weight);

  // Generate events for each user with random count between min and max
  for (const user of users) {
    const eventsForThisUser = randInt(minEventsPerUser, maxEventsPerUser);

    for (let i = 0; i < eventsForThisUser; i++) {
      const it = randChoice(items).item_id!;
      const typ = weightedChoice(eventTypeIndexes, eventTypeWeights);
      const ageDays = Math.max(
        0,
        Math.round(Math.abs(randInt(-5, 60) + randInt(0, recentBiasDays) / 3))
      );
      const ts = iso(daysAgo(ageDays));
      res.push({
        user_id: user.user_id!,
        item_id: it,
        type: typ,
        ts,
        value: typ === 3 ? randInt(1, 3) : 1, // Keep purchase value logic for now
      });
    }
  }
  return res;
}

export async function handleSeed(
  namespace: string,
  userCount: number,
  userStartIndex: number,
  itemCount: number,
  minEventsPerUser: number,
  maxEventsPerUser: number,
  brands: string[],
  tags: string[],
  eventTypes: EventTypeConfig[],
  append: (value: string) => void,
  setGeneratedUsers: (value: string[]) => void,
  setGeneratedItems: (value: string[]) => void,
  traitConfigs: TraitConfig[] = [],
  itemConfigs: ItemConfig[] = [],
  priceRanges: PriceRange[] = []
) {
  setGeneratedUsers([]);
  setGeneratedItems([]);
  try {
    // Convert EventTypeConfig to types_EventTypeConfigUpsertRequest
    const eventTypeRequests = eventTypes.map((et) => ({
      name: et.title,
      type: et.index,
      weight: et.weight,
      half_life_days: et.halfLifeDays,
      is_active: true,
    }));
    await upsertEventTypes({ namespace, types: eventTypeRequests }, append);
    const users = buildUsers(userCount, userStartIndex, traitConfigs);
    const items = buildItems(itemCount, brands, tags, itemConfigs, priceRanges);
    setGeneratedUsers(users.map((u) => u.user_id!));
    setGeneratedItems(items.map((i) => i.item_id!));

    await upsertUsers({ namespace, users }, append);
    // Chunk to keep payload sizes modest
    const chunk = 200;
    for (let i = 0; i < items.length; i += chunk) {
      await upsertItems(
        { namespace, items: items.slice(i, i + chunk) },
        append
      );
    }
    const events = buildEvents(
      users,
      items,
      minEventsPerUser,
      maxEventsPerUser,
      eventTypes
    );
    for (let i = 0; i < events.length; i += 1000) {
      await batchEvents(
        { namespace, events: events.slice(i, i + 1000) },
        append
      );
    }
    append("✅ Seed complete");
  } catch (e: any) {
    append(`❌ Seed error: ${e.message}`);
  }
}
