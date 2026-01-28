/**
 * Web Worker wrapper for local embeddings processing.
 * Provides the same interface as LocalEmbeddingsProvider but runs in a worker.
 */

import type {
  EmbeddingsProvider,
  EmbeddingsConfig,
  EmbeddingsError,
} from "./types";
import { getLogger } from "../di";
import type {
  WorkerMessage,
  EmbedRequest,
  EmbedResponse,
} from "./EmbeddingsWorker";

export class WorkerEmbeddingsProvider implements EmbeddingsProvider {
  private worker: any = null;
  private modelName: string;
  private dimension: number;
  private config: EmbeddingsConfig;
  private logger = getLogger().child({ component: "WorkerEmbeddingsProvider" });
  private pendingRequests = new Map<
    string,
    {
      resolve: (value: any) => void;
      reject: (error: any) => void;
      timeout: ReturnType<typeof setTimeout>;
    }
  >();
  private isInitialized = false;

  constructor(config: EmbeddingsConfig = {}) {
    this.config = {
      model: "Xenova/all-MiniLM-L6-v2",
      dimension: 384,
      maxBatchSize: 16,
      maxTextLength: 512,
      timeout: 30000,
      ...config,
    };
    this.modelName = this.config.model!;
    this.dimension = this.config.dimension!;
  }

  async isReady(): Promise<boolean> {
    if (!this.worker) {
      return false;
    }

    if (this.isInitialized) {
      return true;
    }

    try {
      await this.initializeWorker();
      return this.isInitialized;
    } catch (error) {
      this.logger.error("Worker initialization failed", { error });
      return false;
    }
  }

  async embed(texts: string[]): Promise<number[][]> {
    const startTime = Date.now();

    try {
      // Validate inputs
      this.validateInputs(texts);

      // Ensure worker is ready
      await this.ensureWorkerReady();

      if (!this.worker) {
        throw this.createError("MODEL_LOAD_ERROR", "Worker not available");
      }

      // Send embedding request to worker
      const request: EmbedRequest = {
        texts,
        model: this.modelName,
        maxTextLength: this.config.maxTextLength!,
        maxBatchSize: this.config.maxBatchSize!,
      };

      const response = await this.sendWorkerMessage<EmbedResponse>(
        "EMBED",
        request
      );

      const processingTime = Date.now() - startTime;
      this.logger.debug("Worker embeddings generated", {
        textCount: texts.length,
        dimension: this.dimension,
        processingTime,
        batchCount: response.batchCount,
      });

      return response.embeddings;
    } catch (error) {
      const processingTime = Date.now() - startTime;
      this.logger.error("Worker embedding generation failed", {
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
    if (this.worker) {
      try {
        await this.sendWorkerMessage("DISPOSE");
        this.worker.terminate();
        this.worker = null;
        this.isInitialized = false;
        this.logger.debug("Worker disposed");
      } catch (error) {
        this.logger.warn("Error disposing worker", { error });
      }
    }
  }

  private async ensureWorkerReady(): Promise<void> {
    if (this.isInitialized) {
      return;
    }

    if (!this.worker) {
      await this.initializeWorker();
    }
  }

  private async initializeWorker(): Promise<void> {
    if (this.worker && this.isInitialized) {
      return;
    }

    try {
      // Create worker
      this.worker = new (globalThis as any).Worker(
        new URL("./EmbeddingsWorker.ts", import.meta.url),
        { type: "module" }
      );

      // Set up message handler
      this.worker.onmessage = this.handleWorkerMessage.bind(this);
      this.worker.onerror = this.handleWorkerError.bind(this);

      // Initialize worker with model
      await this.sendWorkerMessage("INIT", { model: this.modelName });

      this.isInitialized = true;
      this.logger.info("Worker initialized successfully", {
        model: this.modelName,
      });
    } catch (error) {
      this.logger.error("Failed to initialize worker", {
        model: this.modelName,
        error: error instanceof Error ? error.message : String(error),
      });
      throw this.createError(
        "MODEL_LOAD_ERROR",
        `Failed to initialize worker: ${error}`
      );
    }
  }

  private handleWorkerMessage(event: any): void {
    const { id, payload, error } = event.data;

    const pending = this.pendingRequests.get(id);
    if (!pending) {
      return;
    }

    // Clear timeout
    clearTimeout(pending.timeout);
    this.pendingRequests.delete(id);

    if (error) {
      pending.reject(new Error(error));
    } else {
      pending.resolve(payload);
    }
  }

  private handleWorkerError(error: any): void {
    this.logger.error("Worker error", {
      message: error.message,
      filename: error.filename,
      lineno: error.lineno,
      colno: error.colno,
    });

    // Reject all pending requests
    for (const [, pending] of this.pendingRequests) {
      clearTimeout(pending.timeout);
      pending.reject(new Error(`Worker error: ${error.message}`));
    }
    this.pendingRequests.clear();
  }

  private async sendWorkerMessage<T = any>(
    type: WorkerMessage["type"],
    payload?: any
  ): Promise<T> {
    if (!this.worker) {
      throw this.createError("MODEL_LOAD_ERROR", "Worker not available");
    }

    const id = `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

    return new Promise<T>((resolve, reject) => {
      // Set up timeout
      const timeout = setTimeout(() => {
        this.pendingRequests.delete(id);
        reject(this.createError("TIMEOUT_ERROR", "Worker request timeout"));
      }, this.config.timeout!);

      // Store pending request
      this.pendingRequests.set(id, { resolve, reject, timeout });

      // Send message to worker
      const message: WorkerMessage = { id, type, payload };
      this.worker!.postMessage(message);
    });
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

  private createError(
    code: EmbeddingsError["code"],
    message: string
  ): EmbeddingsError {
    const error = new Error(message) as EmbeddingsError;
    error.code = code;
    error.provider = "WorkerEmbeddingsProvider";
    error.model = this.modelName;
    error.retryable = code === "EMBEDDING_ERROR" || code === "TIMEOUT_ERROR";
    return error;
  }
}
