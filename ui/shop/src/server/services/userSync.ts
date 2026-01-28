import { prisma } from "@/server/db/client";
import { upsertUsers } from "@/server/services/recsys";
import { buildUserContract } from "@/lib/contracts/user";

// Batch size for upsert operations
const BATCH_SIZE = 50;

export async function syncAllUsersToRecsys() {
  console.log("Starting full user sync to Recsys...");

  let offset = 0;
  let totalSynced = 0;

  while (true) {
    const users = await prisma.user.findMany({
      skip: offset,
      take: BATCH_SIZE,
      orderBy: { createdAt: "desc" },
    });

    if (users.length === 0) break;

    try {
      const userContracts = users.map(buildUserContract);
      await upsertUsers(userContracts);
      totalSynced += users.length;
      console.log(`Synced ${users.length} users (total: ${totalSynced})`);
    } catch (error) {
      console.error(
        `Failed to sync batch starting at offset ${offset}:`,
        error
      );
      // Continue with next batch instead of failing completely
    }

    offset += BATCH_SIZE;
  }

  console.log(`User sync completed. Total synced: ${totalSynced}`);
  return { synced: totalSynced };
}

export async function syncUserTraits(userId: string) {
  const user = await prisma.user.findUnique({ where: { id: userId } });
  if (!user) return;

  try {
    const userContract = buildUserContract(user);
    await upsertUsers([userContract]);
    console.log(`Synced traits for user ${userId}`);
  } catch (error) {
    console.error(`Failed to sync traits for user ${userId}:`, error);
  }
}

export async function updateUserLastSeen(userId: string) {
  const user = await prisma.user.findUnique({ where: { id: userId } });
  if (!user) return;

  try {
    // Update last_seen_ts in traits
    const traits = user.traitsText ? JSON.parse(user.traitsText) : {};
    traits.last_seen_ts = new Date().toISOString();

    await prisma.user.update({
      where: { id: userId },
      data: { traitsText: JSON.stringify(traits) },
    });

    // Sync to Recsys
    const userContract = buildUserContract({
      ...user,
      traitsText: JSON.stringify(traits),
    });
    await upsertUsers([userContract]);

    console.log(`Updated last seen for user ${userId}`);
  } catch (error) {
    console.error(`Failed to update last seen for user ${userId}:`, error);
  }
}

export async function enrichUserTraits(
  userId: string,
  traits: Record<string, unknown>
) {
  const user = await prisma.user.findUnique({ where: { id: userId } });
  if (!user) return;

  try {
    const existingTraits = user.traitsText ? JSON.parse(user.traitsText) : {};
    const enrichedTraits = { ...existingTraits, ...traits };

    await prisma.user.update({
      where: { id: userId },
      data: { traitsText: JSON.stringify(enrichedTraits) },
    });

    // Sync to Recsys
    const userContract = buildUserContract({
      ...user,
      traitsText: JSON.stringify(enrichedTraits),
    });
    await upsertUsers([userContract]);

    console.log(`Enriched traits for user ${userId}:`, enrichedTraits);
  } catch (error) {
    console.error(`Failed to enrich traits for user ${userId}:`, error);
  }
}
