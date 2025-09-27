/**
 * Web Worker for local embeddings processing.
 * Keeps the UI responsive by offloading heavy computation to a background thread.
 */

import { pipeline, FeatureExtractionPipeline } from "@xenova/transformers";

// Worker message types
export interface WorkerMessage {
  id: string;
  type: "INIT" | "EMBED" | "DISPOSE";
  payload?: any;
}

export interface WorkerResponse {
  id: string;
  type:
    | "INIT_SUCCESS"
    | "INIT_ERROR"
    | "EMBED_SUCCESS"
    | "EMBED_ERROR"
    | "DISPOSE_SUCCESS";
  payload?: any;
  error?: string;
}

export interface EmbedRequest {
  texts: string[];
  model: string;
  maxTextLength: number;
  maxBatchSize: number;
}

export interface EmbedResponse {
  embeddings: number[][];
  processingTime: number;
  batchCount: number;
}

// Worker state
let model: FeatureExtractionPipeline | null = null;
let isModelLoading = false;
let loadingPromise: Promise<void> | null = null;

// Message handler
(globalThis as any).onmessage = async (event: any) => {
  const { id, type, payload } = event.data;

  try {
    switch (type) {
      case "INIT":
        await handleInit(id, payload);
        break;

      case "EMBED":
        await handleEmbed(id, payload);
        break;

      case "DISPOSE":
        await handleDispose(id);
        break;

      default:
        sendResponse(id, "EMBED_ERROR", null, `Unknown message type: ${type}`);
    }
  } catch (error) {
    sendResponse(
      id,
      "EMBED_ERROR",
      null,
      error instanceof Error ? error.message : String(error)
    );
  }
};

async function handleInit(
  id: string,
  payload: { model: string }
): Promise<void> {
  try {
    if (model) {
      sendResponse(id, "INIT_SUCCESS", {
        model: payload.model,
        alreadyLoaded: true,
      });
      return;
    }

    if (isModelLoading && loadingPromise) {
      await loadingPromise;
      sendResponse(id, "INIT_SUCCESS", {
        model: payload.model,
        alreadyLoaded: true,
      });
      return;
    }
  } catch (error) {
    sendResponse(
      id,
      "INIT_ERROR",
      null,
      error instanceof Error ? error.message : String(error)
    );
    return;
  }

  isModelLoading = true;
  loadingPromise = loadModel(payload.model);

  try {
    await loadingPromise;
    sendResponse(id, "INIT_SUCCESS", { model: payload.model });
  } catch (error) {
    sendResponse(
      id,
      "INIT_ERROR",
      null,
      error instanceof Error ? error.message : String(error)
    );
  } finally {
    isModelLoading = false;
    loadingPromise = null;
  }
}

async function handleEmbed(id: string, payload: EmbedRequest): Promise<void> {
  const startTime = Date.now();

  try {
    if (!model) {
      throw new Error("Model not initialized");
    }

    const { texts, maxTextLength, maxBatchSize } = payload;

    // Validate inputs
    if (!Array.isArray(texts) || texts.length === 0) {
      throw new Error("Invalid input: texts must be a non-empty array");
    }

    // Process in batches
    const batches = createBatches(texts, maxBatchSize);
    const allEmbeddings: number[][] = [];

    for (const batch of batches) {
      const batchEmbeddings = await processBatch(batch, maxTextLength);
      allEmbeddings.push(...batchEmbeddings);
    }

    const processingTime = Date.now() - startTime;

    sendResponse(id, "EMBED_SUCCESS", {
      embeddings: allEmbeddings,
      processingTime,
      batchCount: batches.length,
    });
  } catch (error) {
    const processingTime = Date.now() - startTime;
    sendResponse(
      id,
      "EMBED_ERROR",
      { processingTime },
      error instanceof Error ? error.message : String(error)
    );
  }
}

async function handleDispose(id: string): Promise<void> {
  try {
    if (model) {
      model = null;
    }
    sendResponse(id, "DISPOSE_SUCCESS");
  } catch (error) {
    sendResponse(
      id,
      "EMBED_ERROR",
      null,
      error instanceof Error ? error.message : String(error)
    );
  }
}

async function loadModel(modelName: string): Promise<void> {
  try {
    model = await pipeline("feature-extraction", modelName, {
      quantized: true,
      progress_callback: (progress: any) => {
        // Send progress updates to main thread
        (globalThis as any).postMessage({
          id: "progress",
          type: "PROGRESS",
          payload: { progress: Math.round(progress.progress * 100) },
        });
      },
    });
  } catch (error) {
    throw new Error(`Failed to load model: ${error}`);
  }
}

async function processBatch(
  texts: string[],
  maxTextLength: number
): Promise<number[][]> {
  if (!model) {
    throw new Error("Model not loaded");
  }

  // Truncate texts if needed
  const truncatedTexts = texts.map((text) =>
    text.length > maxTextLength ? text.substring(0, maxTextLength) : text
  );

  // Generate embeddings
  const result = await model(truncatedTexts, {
    pooling: "mean",
    normalize: true,
  });

  // Convert to number arrays
  const embeddings: number[][] = [];

  // Handle single tensor result
  if (result && result.data) {
    embeddings.push(Array.from(result.data));
  } else if (Array.isArray(result)) {
    // Handle array of tensors
    for (let i = 0; i < result.length; i++) {
      const tensor = result[i];
      if (tensor && tensor.data) {
        embeddings.push(Array.from(tensor.data));
      }
    }
  }

  return embeddings;
}

function createBatches(texts: string[], maxBatchSize: number): string[][] {
  const batches: string[][] = [];

  for (let i = 0; i < texts.length; i += maxBatchSize) {
    batches.push(texts.slice(i, i + maxBatchSize));
  }

  return batches;
}

function sendResponse(
  id: string,
  type: WorkerResponse["type"],
  payload?: any,
  error?: string
): void {
  const response: WorkerResponse = { id, type };

  if (payload) {
    response.payload = payload;
  }

  if (error) {
    response.error = error;
  }

  (globalThis as any).postMessage(response);
}
