import type {
  Container,
  HttpClient,
  Storage,
  Logger,
  Embeddings,
} from "./interfaces";
import { BrowserEmbeddings } from "./implementations";
import { createStorage } from "./storage";
import { StructuredLogger } from "./logger";
import {
  EnhancedHttpClient,
  RequestIdInterceptor,
  AuthInterceptor,
  ErrorInterceptor,
  ResponseLoggingInterceptor,
} from "./http";
import { config } from "../config";

/**
 * Default dependency injection container for the demo UI application.
 * Provides browser implementations of all required interfaces.
 */
export class DefaultContainer implements Container {
  private httpClient: EnhancedHttpClient;
  private storage: Storage;
  private logger: Logger;
  private embeddings: Embeddings;

  constructor() {
    // Initialize enhanced HTTP client with configuration
    this.httpClient = new EnhancedHttpClient({
      baseUrl: config.api.baseUrl,
      timeout: 30000,
      retries: 3,
      retryDelay: 1000,
      retryBackoff: true,
      jitter: true,
      circuitBreaker: {
        enabled: true,
        failureThreshold: 5,
        timeout: 5000,
        resetTimeout: 30000,
      },
    });

    // Add interceptors
    this.setupHttpInterceptors();

    // Initialize enhanced storage with hybrid backend
    this.storage = createStorage("hybrid", {
      prefix: "recsys_",
      defaultTTL: 24 * 60 * 60 * 1000, // 24 hours
      maxSize: 1000,
    });

    // Initialize logger based on configuration
    this.logger = new StructuredLogger(
      config.logging.level,
      config.logging.enableAnalytics,
      { app: "recsys-demo-ui" }
    );

    // Initialize embeddings service
    this.embeddings = new BrowserEmbeddings();
  }

  private setupHttpInterceptors(): void {
    // Add request ID to all requests
    this.httpClient.addRequestInterceptor({
      onRequest: (request) => RequestIdInterceptor.addRequestId(request),
    });

    // Add auth headers (stub for now)
    this.httpClient.addRequestInterceptor({
      onRequest: (request) => AuthInterceptor.attachAuthHeader(request),
    });

    // Handle errors
    this.httpClient.addResponseInterceptor({
      onError: (error) => {
        const apiError = ErrorInterceptor.processError(error);
        ErrorInterceptor.logError(
          apiError,
          error.request || { url: "", method: "GET" }
        );
        return apiError;
      },
    });

    // Log responses
    this.httpClient.addResponseInterceptor({
      onResponse: (response) =>
        ResponseLoggingInterceptor.logResponse(response),
    });
  }

  getHttpClient(): HttpClient {
    return this.httpClient;
  }

  getStorage(): Storage {
    return this.storage;
  }

  getLogger(): Logger {
    return this.logger;
  }

  getEmbeddings(): Embeddings {
    return this.embeddings;
  }
}

/**
 * Global container instance.
 * Can be replaced for testing or different environments.
 */
let container: Container = new DefaultContainer();

/**
 * Get the current container instance.
 */
export function getContainer(): Container {
  return container;
}

/**
 * Set a custom container instance.
 * Useful for testing or different environments.
 */
export function setContainer(newContainer: Container): void {
  container = newContainer;
}

/**
 * Reset to the default container.
 */
export function resetContainer(): void {
  container = new DefaultContainer();
}

// Export convenience functions for common dependencies
export const getHttpClient = () => getContainer().getHttpClient();
export const getStorage = () => getContainer().getStorage();
export const getLogger = () => getContainer().getLogger();
export const getEmbeddings = () => getContainer().getEmbeddings();
