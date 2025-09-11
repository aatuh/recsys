import { embedText, itemToText } from "../embeddings/browserEmbedder";
import { upsertItems } from "../services/apiService";

/**
 * Compute an embedding for a single item in the browser and upsert it.
 */
export async function updateItemEmbedding(
  namespace: string,
  item: {
    item_id: string;
    tags?: string[];
    price?: number;
    props?: Record<string, any>;
  },
  log?: (s: string) => void
) {
  const text = itemToText(item);
  log?.(`Embedding "${item.item_id}" from text: "${text}"`);
  const vec = await embedText(text);

  await upsertItems(
    namespace,
    [{ item_id: item.item_id, embedding: vec }],
    (s) => log?.(s)
  );
  log?.(`✔ embedding upserted for ${item.item_id}`);
}
