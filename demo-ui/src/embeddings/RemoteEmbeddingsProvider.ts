/**
 * Remote embeddings provider stub that calls an API endpoint.
 * Validates shape and error mapping; kept behind useRemoteEmbeddings flag.
 */

import type {
  EmbeddingsProvider,
  EmbeddingsConfig,
  EmbeddingsError,
} from "./types";
import { getLogger, getHttpClient } from "../di";

export class RemoteEmbeddingsProvider implements EmbeddingsProvider {
  private endpoint: string;
  private dimension: number;
  private modelName: string;
  private timeout: number;
  private logger = getLogger().child({ component: "RemoteEmbeddingsProvider" });
  private httpClient = getHttpClient();

  constructor(config: EmbeddingsConfig = {}) {
    this.endpoint = config.model || "/api/embeddings";
    this.dimension = config.dimension || 384;
    this.modelName = config.model || "remote-model";
    this.timeout = config.timeout || 30000;
  }

  async isReady(): Promise<boolean> {
    try {
      // Test the endpoint with a simple health check
      const response = await this.httpClient.get(`${this.endpoint}/health`, {
        timeout: 5000,
      });

      return response.status === 200;
    } catch (error) {
      this.logger.debug("Remote embeddings provider not ready", {
        endpoint: this.endpoint,
        error: error instanceof Error ? error.message : String(error),
      });
      return false;
    }
  }

  async embed(texts: string[]): Promise<number[][]> {
    const startTime = Date.now();

    try {
      // Validate inputs
      this.validateInputs(texts);

      // Prepare request payload
      const payload = {
        texts,
        model: this.modelName,
        normalize: true,
      };

      this.logger.debug("Sending embedding request", {
        textCount: texts.length,
        endpoint: this.endpoint,
      });

      // Make API request
      const response = await this.httpClient.post<{
        embeddings: number[][];
        model: string;
        dimension: number;
      }>(this.endpoint, payload, {
        timeout: this.timeout,
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!response.data) {
        throw this.createError(
          "VALIDATION_ERROR",
          "Empty response from server"
        );
      }

      // Validate response shape
      this.validateResponse(response.data, texts.length);

      const processingTime = Date.now() - startTime;
      this.logger.debug("Remote embeddings received", {
        textCount: texts.length,
        dimension: response.data.dimension,
        processingTime,
      });

      return response.data.embeddings;
    } catch (error) {
      const processingTime = Date.now() - startTime;
      this.logger.error("Remote embedding request failed", {
        textCount: texts.length,
        processingTime,
        error: error instanceof Error ? error.message : String(error),
      });

      throw this.mapError(error);
    }
  }

  getDimension(): number {
    return this.dimension;
  }

  getModelName(): string {
    return this.modelName;
  }

  async dispose(): Promise<void> {
    // No resources to clean up for remote provider
    this.logger.debug("Remote embeddings provider disposed");
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

    if (texts.length > 100) {
      // Reasonable limit for remote API
      throw this.createError(
        "VALIDATION_ERROR",
        "Too many texts. Maximum: 100"
      );
    }

    for (let i = 0; i < texts.length; i++) {
      if (typeof texts[i] !== "string") {
        throw this.createError(
          "VALIDATION_ERROR",
          `Text at index ${i} must be a string`
        );
      }

      if ((texts[i] || "").length > 10000) {
        // Reasonable limit for remote API
        throw this.createError(
          "VALIDATION_ERROR",
          `Text at index ${i} is too long. Maximum: 10000 characters`
        );
      }
    }
  }

  private validateResponse(data: any, expectedCount: number): void {
    if (!data || typeof data !== "object") {
      throw this.createError("VALIDATION_ERROR", "Invalid response format");
    }

    if (!Array.isArray(data.embeddings)) {
      throw this.createError(
        "VALIDATION_ERROR",
        "Response missing embeddings array"
      );
    }

    if (data.embeddings.length !== expectedCount) {
      throw this.createError(
        "VALIDATION_ERROR",
        `Expected ${expectedCount} embeddings, got ${data.embeddings.length}`
      );
    }

    if (typeof data.dimension !== "number" || data.dimension <= 0) {
      throw this.createError(
        "VALIDATION_ERROR",
        "Invalid dimension in response"
      );
    }

    // Validate each embedding vector
    for (let i = 0; i < data.embeddings.length; i++) {
      const embedding = data.embeddings[i];

      if (!Array.isArray(embedding)) {
        throw this.createError(
          "VALIDATION_ERROR",
          `Embedding at index ${i} is not an array`
        );
      }

      if (embedding.length !== data.dimension) {
        throw this.createError(
          "VALIDATION_ERROR",
          `Embedding at index ${i} has wrong dimension. Expected: ${data.dimension}, got: ${embedding.length}`
        );
      }

      // Check if all elements are numbers
      for (let j = 0; j < embedding.length; j++) {
        if (typeof embedding[j] !== "number" || !isFinite(embedding[j])) {
          throw this.createError(
            "VALIDATION_ERROR",
            `Embedding at index ${i}, position ${j} is not a valid number`
          );
        }
      }
    }
  }

  private mapError(error: unknown): EmbeddingsError {
    if (error instanceof Error && "status" in error) {
      const status = (error as any).status;

      if (status === 401 || status === 403) {
        return this.createError("NETWORK_ERROR", "Authentication failed");
      }

      if (status === 429) {
        return this.createError("NETWORK_ERROR", "Rate limit exceeded");
      }

      if (status >= 500) {
        return this.createError("NETWORK_ERROR", "Server error");
      }

      if (status >= 400) {
        return this.createError("VALIDATION_ERROR", "Invalid request");
      }
    }

    if (error instanceof Error && error.message.includes("timeout")) {
      return this.createError("TIMEOUT_ERROR", "Request timeout");
    }

    if (error instanceof Error && error.message.includes("network")) {
      return this.createError("NETWORK_ERROR", "Network error");
    }

    return this.createError(
      "EMBEDDING_ERROR",
      error instanceof Error ? error.message : String(error)
    );
  }

  private createError(
    code: EmbeddingsError["code"],
    message: string
  ): EmbeddingsError {
    const error = new Error(message) as EmbeddingsError;
    error.code = code;
    error.provider = "RemoteEmbeddingsProvider";
    error.model = this.modelName;
    error.retryable = code === "NETWORK_ERROR" || code === "TIMEOUT_ERROR";
    return error;
  }
}
