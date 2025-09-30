/**
 * Local embeddings provider using @xenova/transformers.
 * Provides lazy loading, caching, and serialized processing to avoid concurrent loads.
 */

import { pipeline, FeatureExtractionPipeline } from "@xenova/transformers";
import type {
  EmbeddingsProvider,
  EmbeddingsConfig,
  EmbeddingsError,
} from "./types";
import { getLogger } from "../di";

export class LocalEmbeddingsProvider implements EmbeddingsProvider {
  private model: FeatureExtractionPipeline | null = null;
  private modelName: string;
  private dimension: number;
  private isModelLoading = false;
  private loadingPromise: Promise<void> | null = null;
  private logger = getLogger().child({ component: "LocalEmbeddingsProvider" });
  private config: EmbeddingsConfig;

  constructor(config: EmbeddingsConfig = {}) {
    this.config = {
      model: "Xenova/all-MiniLM-L6-v2",
      dimension: 384,
      maxBatchSize: 32,
      maxTextLength: 512,
      timeout: 30000,
      ...config,
    };
    this.modelName = this.config.model!;
    this.dimension = this.config.dimension!;
  }

  async isReady(): Promise<boolean> {
    if (this.model) {
      return true;
    }

    if (this.isModelLoading) {
      return false;
    }

    try {
      await this.loadModel();
      return this.model !== null;
    } catch (error) {
      this.logger.error("Failed to load model", { error });
      return false;
    }
  }

  async embed(texts: string[]): Promise<number[][]> {
    const startTime = Date.now();

    try {
      // Validate inputs
      this.validateInputs(texts);

      // Ensure model is loaded
      await this.ensureModelLoaded();

      if (!this.model) {
        throw this.createError("MODEL_LOAD_ERROR", "Model failed to load");
      }

      // Process texts in batches to avoid memory issues
      const batches = this.createBatches(texts);
      const allEmbeddings: number[][] = [];

      for (const batch of batches) {
        const batchEmbeddings = await this.processBatch(batch);
        allEmbeddings.push(...batchEmbeddings);
      }

      const processingTime = Date.now() - startTime;
      this.logger.debug("Embeddings generated", {
        textCount: texts.length,
        dimension: this.dimension,
        processingTime,
        batchCount: batches.length,
      });

      return allEmbeddings;
    } catch (error) {
      const processingTime = Date.now() - startTime;
      this.logger.error("Embedding generation failed", {
        textCount: texts.length,
        processingTime,
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  getDimension(): number {
    return this.dimension;
  }

  getModelName(): string {
    return this.modelName;
  }

  async dispose(): Promise<void> {
    if (this.model) {
      try {
        // Clean up model resources
        this.model = null;
        this.logger.debug("Model disposed");
      } catch (error) {
        this.logger.warn("Error disposing model", { error });
      }
    }
  }

  private async ensureModelLoaded(): Promise<void> {
    if (this.model) {
      return;
    }

    if (this.isModelLoading && this.loadingPromise) {
      return this.loadingPromise;
    }

    this.loadingPromise = this.loadModel();
    return this.loadingPromise;
  }

  private async loadModel(): Promise<void> {
    if (this.model || this.isModelLoading) {
      return;
    }

    this.isModelLoading = true;
    this.logger.info("Loading embeddings model", { model: this.modelName });

    try {
      // Load the feature extraction pipeline
      this.model = await pipeline("feature-extraction", this.modelName, {
        quantized: true, // Use quantized model for better performance
        progress_callback: (progress: any) => {
          this.logger.debug("Model loading progress", {
            progress: Math.round(progress.progress * 100),
          });
        },
      });

      this.logger.info("Model loaded successfully", { model: this.modelName });
    } catch (error) {
      this.logger.error("Failed to load model", {
        model: this.modelName,
        error: error instanceof Error ? error.message : String(error),
      });
      throw this.createError(
        "MODEL_LOAD_ERROR",
        `Failed to load model: ${error}`
      );
    } finally {
      this.isModelLoading = false;
    }
  }

  private async processBatch(texts: string[]): Promise<number[][]> {
    if (!this.model) {
      throw this.createError("MODEL_LOAD_ERROR", "Model not loaded");
    }

    try {
      // Truncate texts if they exceed max length
      const truncatedTexts = texts.map((text) =>
        text.length > this.config.maxTextLength!
          ? text.substring(0, this.config.maxTextLength!)
          : text
      );

      // Generate embeddings
      const result = await this.model(truncatedTexts, {
        pooling: "mean", // Use mean pooling for sentence embeddings
        normalize: true, // Normalize embeddings
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
    } catch (error) {
      this.logger.error("Batch processing failed", {
        batchSize: texts.length,
        error: error instanceof Error ? error.message : String(error),
      });
      throw this.createError(
        "EMBEDDING_ERROR",
        `Batch processing failed: ${error}`
      );
    }
  }

  private validateInputs(texts: string[]): void {
    if (!Array.isArray(texts)) {
      throw this.createError(
        "VALIDATION_ERROR",
        "Input must be an array of strings"
      );
    }

    if (texts.length === 0) {
      throw this.createError("VALIDATION_ERROR", "Input array cannot be empty");
    }

    if (texts.length > this.config.maxBatchSize! * 10) {
      // Allow some flexibility
      throw this.createError(
        "VALIDATION_ERROR",
        `Too many texts. Maximum: ${this.config.maxBatchSize! * 10}`
      );
    }

    for (let i = 0; i < texts.length; i++) {
      if (typeof texts[i] !== "string") {
        throw this.createError(
          "VALIDATION_ERROR",
          `Text at index ${i} must be a string`
        );
      }
    }
  }

  private createBatches(texts: string[]): string[][] {
    const maxBatchSize = this.config.maxBatchSize!;
    const batches: string[][] = [];

    for (let i = 0; i < texts.length; i += maxBatchSize) {
      batches.push(texts.slice(i, i + maxBatchSize));
    }

    return batches;
  }

  private createError(
    code: EmbeddingsError["code"],
    message: string
  ): EmbeddingsError {
    const error = new Error(message) as EmbeddingsError;
    error.code = code;
    error.provider = "LocalEmbeddingsProvider";
    error.model = this.modelName;
    error.retryable = code === "EMBEDDING_ERROR" || code === "TIMEOUT_ERROR";
    return error;
  }
}
