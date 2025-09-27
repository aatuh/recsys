/**
 * Utilities for chunking and batching text processing.
 * Provides helpers for splitting long texts and batching requests.
 */

import type { ChunkingOptions, BatchingOptions } from "./types";
import { getLogger } from "../di";

export class ChunkingUtils {
  private static logger = getLogger().child({ component: "ChunkingUtils" });

  /**
   * Split a long text into smaller chunks for processing.
   */
  static chunkText(text: string, options: ChunkingOptions): string[] {
    if (text.length <= options.maxChunkSize) {
      return [text];
    }

    const chunks: string[] = [];
    let start = 0;

    while (start < text.length) {
      let end = Math.min(start + options.maxChunkSize, text.length);

      // Try to find a good break point
      if (end < text.length) {
        end = this.findBreakPoint(text, start, end, options);
      }

      const chunk = text.substring(start, end).trim();
      if (chunk.length > 0) {
        chunks.push(chunk);
      }

      // Move start position with overlap
      start = end - options.overlap;
      if (start >= text.length) break;
    }

    this.logger.debug("Text chunked", {
      originalLength: text.length,
      chunkCount: chunks.length,
      maxChunkSize: options.maxChunkSize,
      overlap: options.overlap,
    });

    return chunks;
  }

  /**
   * Find the best break point within the chunk boundaries.
   */
  private static findBreakPoint(
    text: string,
    start: number,
    end: number,
    options: ChunkingOptions
  ): number {
    const searchStart = Math.max(start, end - 100); // Search in last 100 chars

    switch (options.strategy) {
      case "sentence":
        return this.findSentenceBreak(text, searchStart, end);

      case "word":
        return this.findWordBreak(text, searchStart, end);

      case "character":
      default:
        return end;
    }
  }

  private static findSentenceBreak(
    text: string,
    start: number,
    end: number
  ): number {
    // Look for sentence endings
    const sentenceEndings = [".", "!", "?", "\n"];

    for (let i = end - 1; i >= start; i--) {
      if (sentenceEndings.includes(text[i] || "")) {
        return i + 1;
      }
    }

    return end;
  }

  private static findWordBreak(
    text: string,
    start: number,
    end: number
  ): number {
    // Look for word boundaries (whitespace)
    for (let i = end - 1; i >= start; i--) {
      if (/\s/.test(text[i] || "")) {
        return i;
      }
    }

    return end;
  }

  /**
   * Create batches from an array of items.
   */
  static createBatches<T>(items: T[], options: BatchingOptions): T[][] {
    const batches: T[][] = [];

    for (let i = 0; i < items.length; i += options.maxBatchSize) {
      batches.push(items.slice(i, i + options.maxBatchSize));
    }

    this.logger.debug("Items batched", {
      itemCount: items.length,
      batchCount: batches.length,
      maxBatchSize: options.maxBatchSize,
    });

    return batches;
  }

  /**
   * Process batches with concurrency control and error handling.
   */
  static async processBatches<T, R>(
    batches: T[][],
    processor: (batch: T[]) => Promise<R[]>,
    options: BatchingOptions
  ): Promise<R[]> {
    const results: R[] = [];
    const activePromises: Promise<void>[] = [];
    let batchIndex = 0;

    this.logger.debug("Starting batch processing", {
      batchCount: batches.length,
      maxConcurrentBatches: options.maxConcurrentBatches,
    });

    while (batchIndex < batches.length || activePromises.length > 0) {
      // Start new batches up to concurrency limit
      while (
        activePromises.length < options.maxConcurrentBatches &&
        batchIndex < batches.length
      ) {
        const batch = batches[batchIndex];
        const currentIndex = batchIndex;
        batchIndex++;

        if (batch) {
          const promise = this.processBatchWithRetry(
            batch,
            processor,
            options,
            currentIndex
          ).then((batchResults) => {
            if (batchResults) {
              results.push(...batchResults);
            }
          });

          activePromises.push(promise);
        }
      }

      // Wait for at least one batch to complete
      if (activePromises.length > 0) {
        await Promise.race(activePromises);

        // Remove completed promises
        for (let i = activePromises.length - 1; i >= 0; i--) {
          const promise = activePromises[i];
          if (promise && (await this.isPromiseSettled(promise))) {
            activePromises.splice(i, 1);
          }
        }
      }
    }

    this.logger.debug("Batch processing completed", {
      totalResults: results.length,
      processedBatches: batches.length,
    });

    return results;
  }

  /**
   * Process a single batch with retry logic.
   */
  private static async processBatchWithRetry<T, R>(
    batch: T[],
    processor: (batch: T[]) => Promise<R[]>,
    options: BatchingOptions,
    batchIndex: number
  ): Promise<R[]> {
    let lastError: Error | null = null;

    for (let attempt = 0; attempt <= options.retryAttempts; attempt++) {
      try {
        const timeoutPromise = new Promise<never>((_, reject) => {
          setTimeout(() => reject(new Error("Batch timeout")), options.timeout);
        });

        const result = await Promise.race([processor(batch), timeoutPromise]);

        this.logger.debug("Batch processed successfully", {
          batchIndex,
          attempt,
          batchSize: batch.length,
        });

        return result;
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error));

        this.logger.warn("Batch processing failed", {
          batchIndex,
          attempt,
          batchSize: batch.length,
          error: lastError.message,
        });

        if (attempt < options.retryAttempts) {
          // Wait before retry (exponential backoff)
          const delay = Math.min(1000 * Math.pow(2, attempt), 10000);
          await new Promise<void>((resolve) => setTimeout(resolve, delay));
        }
      }
    }

    throw lastError || new Error("Batch processing failed");
  }

  /**
   * Check if a promise is settled (resolved or rejected).
   */
  private static async isPromiseSettled(
    promise: Promise<any>
  ): Promise<boolean> {
    try {
      await Promise.race([promise, Promise.resolve()]);
      return true;
    } catch {
      return true;
    }
  }

  /**
   * Get recommended chunking options for different use cases.
   */
  static getRecommendedChunkingOptions(
    useCase: "local" | "remote" | "mixed"
  ): ChunkingOptions {
    switch (useCase) {
      case "local":
        return {
          maxChunkSize: 512,
          overlap: 50,
          strategy: "sentence",
          preserveWhitespace: true,
        };

      case "remote":
        return {
          maxChunkSize: 1000,
          overlap: 100,
          strategy: "sentence",
          preserveWhitespace: true,
        };

      case "mixed":
      default:
        return {
          maxChunkSize: 750,
          overlap: 75,
          strategy: "sentence",
          preserveWhitespace: true,
        };
    }
  }

  /**
   * Get recommended batching options for different providers.
   */
  static getRecommendedBatchingOptions(
    provider: "local" | "remote"
  ): BatchingOptions {
    switch (provider) {
      case "local":
        return {
          maxBatchSize: 16,
          maxConcurrentBatches: 2,
          timeout: 30000,
          retryAttempts: 2,
        };

      case "remote":
        return {
          maxBatchSize: 32,
          maxConcurrentBatches: 3,
          timeout: 15000,
          retryAttempts: 3,
        };

      default:
        return {
          maxBatchSize: 24,
          maxConcurrentBatches: 2,
          timeout: 20000,
          retryAttempts: 2,
        };
    }
  }
}
