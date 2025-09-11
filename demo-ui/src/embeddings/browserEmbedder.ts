import { pipeline, FeatureExtractionPipeline } from "@xenova/transformers";

const EXPECTED_EMBED_DIM = 384;

// Lazy-load and cache the pipeline so we only fetch the model once.
let extractorPromise: Promise<FeatureExtractionPipeline> | null = null;

async function getExtractor(): Promise<FeatureExtractionPipeline> {
  if (!extractorPromise) {
    extractorPromise = pipeline(
      "feature-extraction",
      "Xenova/all-MiniLM-L6-v2",
      { quantized: true } // smaller download, good enough for demo
    );
  }
  return extractorPromise;
}

/**
 * Compute a normalized 384-dim embedding for a piece of text.
 */
export async function embedText(text: string): Promise<number[]> {
  const extractor = await getExtractor();
  const output = await extractor(text, {
    pooling: "mean",
    normalize: true,
  });
  const data = Array.from(output.data as Float32Array);
  if (data.length !== EXPECTED_EMBED_DIM) {
    throw new Error(
      `Unexpected dim ${data.length}; expected ${EXPECTED_EMBED_DIM}`
    );
  }
  return data;
}

/**
 * An example of how you might build a stable, descriptive string
 * from your item for embedding.
 */
export function itemToText(item: {
  item_id: string;
  tags?: string[];
  price?: number;
  props?: Record<string, any>;
}): string {
  const brand = item.props?.brand ? `brand ${item.props.brand}` : "";
  const tags = (item.tags || []).join(" ");
  const price = item.price !== undefined ? `price ${item.price}` : "";
  return [brand, tags, price].filter(Boolean).join(" ");
}
