import { prisma } from "@/server/db/client";
import { upsertItems } from "@/server/services/recsys";
import { buildItemContract } from "@/lib/contracts/item";

// Batch size for upsert operations
const BATCH_SIZE = 50;

export async function syncAllItemsToRecsys() {
  console.log("Starting full item sync to Recsys...");
  
  let offset = 0;
  let totalSynced = 0;
  
  while (true) {
    const products = await prisma.product.findMany({
      skip: offset,
      take: BATCH_SIZE,
      orderBy: { createdAt: "desc" },
    });
    
    if (products.length === 0) break;
    
    try {
      const itemContracts = products.map(buildItemContract);
      await upsertItems(itemContracts);
      totalSynced += products.length;
      console.log(`Synced ${products.length} items (total: ${totalSynced})`);
    } catch (error) {
      console.error(`Failed to sync batch starting at offset ${offset}:`, error);
      // Continue with next batch instead of failing completely
    }
    
    offset += BATCH_SIZE;
  }
  
  console.log(`Item sync completed. Total synced: ${totalSynced}`);
  return { synced: totalSynced };
}

export async function syncItemAvailability(productId: string) {
  const product = await prisma.product.findUnique({ where: { id: productId } });
  if (!product) return;
  
  try {
    const itemContract = buildItemContract(product);
    await upsertItems([itemContract]);
    console.log(`Synced availability for product ${productId}: ${itemContract.available}`);
  } catch (error) {
    console.error(`Failed to sync availability for product ${productId}:`, error);
  }
}

export async function syncItemPrice(productId: string) {
  const product = await prisma.product.findUnique({ where: { id: productId } });
  if (!product) return;
  
  try {
    const itemContract = buildItemContract(product);
    await upsertItems([itemContract]);
    console.log(`Synced price for product ${productId}: $${itemContract.price}`);
  } catch (error) {
    console.error(`Failed to sync price for product ${productId}:`, error);
  }
}

export async function syncItemTags(productId: string) {
  const product = await prisma.product.findUnique({ where: { id: productId } });
  if (!product) return;
  
  try {
    const itemContract = buildItemContract(product);
    await upsertItems([itemContract]);
    console.log(`Synced tags for product ${productId}:`, itemContract.tags);
  } catch (error) {
    console.error(`Failed to sync tags for product ${productId}:`, error);
  }
}
