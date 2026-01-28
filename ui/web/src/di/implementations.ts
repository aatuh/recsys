import type { Embeddings } from "./interfaces";
import { embedText, itemToText } from "../embeddings/browserEmbedder";

// BrowserStorage has been replaced by EnhancedStorage in the storage module

/**
 * Default browser embeddings implementation.
 */
export class BrowserEmbeddings implements Embeddings {
  async embedText(text: string): Promise<number[]> {
    return embedText(text);
  }

  itemToText(item: {
    item_id: string;
    tags?: string[];
    price?: number;
    props?: Record<string, any>;
  }): string {
    return itemToText(item);
  }
}
