/**
 * Enhanced HTTP client with interceptors, retry logic, circuit breaker, and request tracing.
 */

// Use crypto.randomUUID() instead of uuid library for browser compatibility
import type {
  HttpRequest,
  HttpResponse,
  HttpError,
  HttpClient,
  HttpClientConfig,
  RequestInterceptor,
  ResponseInterceptor,
  CancellationToken,
} from "./types";
import { ApiError } from "./errors";
import { CircuitBreaker } from "./circuit-breaker";
import { RetryManager } from "./retry";

export class EnhancedHttpClient implements HttpClient {
  private requestInterceptors: RequestInterceptor[] = [];
  private responseInterceptors: ResponseInterceptor[] = [];
  private circuitBreaker: CircuitBreaker;
  private retryManager: RetryManager;

  constructor(private config: HttpClientConfig = {}) {
    this.circuitBreaker = new CircuitBreaker({
      enabled: config.circuitBreaker?.enabled ?? false,
      failureThreshold: config.circuitBreaker?.failureThreshold ?? 5,
      timeout: config.circuitBreaker?.timeout ?? 5000,
      resetTimeout: config.circuitBreaker?.resetTimeout ?? 30000,
    });

    this.retryManager = new RetryManager({
      retries: config.retries ?? 3,
      retryDelay: config.retryDelay ?? 1000,
      retryBackoff: config.retryBackoff ?? true,
      jitter: config.jitter ?? true,
    });
  }

  async get<T>(
    url: string,
    options: Partial<HttpRequest> = {}
  ): Promise<HttpResponse<T>> {
    return this.request<T>({ ...options, url, method: "GET" });
  }

  async post<T>(
    url: string,
    data?: unknown,
    options: Partial<HttpRequest> = {}
  ): Promise<HttpResponse<T>> {
    return this.request<T>({ ...options, url, method: "POST", body: data });
  }

  async put<T>(
    url: string,
    data?: unknown,
    options: Partial<HttpRequest> = {}
  ): Promise<HttpResponse<T>> {
    return this.request<T>({ ...options, url, method: "PUT", body: data });
  }

  async delete<T>(
    url: string,
    options: Partial<HttpRequest> = {}
  ): Promise<HttpResponse<T>> {
    return this.request<T>({ ...options, url, method: "DELETE" });
  }

  private async request<T>(request: HttpRequest): Promise<HttpResponse<T>> {
    // Check circuit breaker
    if (!this.circuitBreaker.canExecute()) {
      throw ApiError.server("Circuit breaker is open");
    }

    // Add base URL if not present
    const fullUrl = request.url.startsWith("http")
      ? request.url
      : `${this.config.baseUrl || ""}${request.url}`;

    // Generate request ID
    const requestId = (globalThis as any).crypto.randomUUID();
    const headers = {
      "x-request-id": requestId,
      "Content-Type": "application/json",
      ...request.headers,
    };

    // Apply request interceptors
    let processedRequest: HttpRequest = {
      ...request,
      url: fullUrl,
      headers,
      timeout: request.timeout ?? this.config.timeout ?? 30000,
    };

    for (const interceptor of this.requestInterceptors) {
      if (interceptor.onRequest) {
        processedRequest = await interceptor.onRequest(processedRequest);
      }
    }

    // Execute request with retry logic
    let lastError: Error | null = null;

    for (let attempt = 0; attempt <= this.retryManager.retries; attempt++) {
      try {
        const response = await this.executeRequest<T>(processedRequest);

        // Record success in circuit breaker
        this.circuitBreaker.onSuccess();

        // Apply response interceptors
        let processedResponse = response;
        for (const interceptor of this.responseInterceptors) {
          if (interceptor.onResponse) {
            processedResponse = await interceptor.onResponse(processedResponse);
          }
        }

        return processedResponse;
      } catch (error) {
        lastError = error as Error;

        // Record failure in circuit breaker
        this.circuitBreaker.onFailure();

        // Check if we should retry
        if (
          attempt < this.retryManager.retries &&
          this.retryManager.shouldRetry(lastError, attempt)
        ) {
          await this.retryManager.waitForRetry(attempt);
          continue;
        }

        // Apply error interceptors
        for (const interceptor of this.responseInterceptors) {
          if (interceptor.onError) {
            lastError = await interceptor.onError(lastError as HttpError);
          }
        }

        throw lastError;
      }
    }

    throw lastError || new Error("Request failed");
  }

  private async executeRequest<T>(
    request: HttpRequest
  ): Promise<HttpResponse<T>> {
    const controller = new (globalThis as any).AbortController();
    const timeoutId = setTimeout(
      () => controller.abort(),
      request.timeout || 30000
    );

    if (request.signal) {
      request.signal.addEventListener("abort", () => controller.abort());
    }

    try {
      const response = await fetch(request.url, {
        method: request.method,
        headers: request.headers,
        body: request.body ? JSON.stringify(request.body) : undefined,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        const errorText = await response.text().catch(() => "Unknown error");
        throw ApiError.fromHttpError(
          new Error(`HTTP ${response.status}: ${errorText}`),
          response.status
        );
      }

      const data = await response.json().catch(() => null);
      const responseHeaders: Record<string, string> = {};
      response.headers.forEach((value, key) => {
        responseHeaders[key] = value;
      });

      return {
        data,
        status: response.status,
        statusText: response.statusText,
        headers: responseHeaders,
      };
    } catch (error) {
      clearTimeout(timeoutId);

      if (error instanceof ApiError) {
        throw error;
      }

      if (error instanceof Error) {
        if (error.name === "AbortError") {
          throw ApiError.timeout("Request was aborted");
        }
        throw ApiError.fromHttpError(error);
      }

      throw ApiError.network("Unknown network error");
    }
  }

  addRequestInterceptor(interceptor: RequestInterceptor): void {
    this.requestInterceptors.push(interceptor);
  }

  addResponseInterceptor(interceptor: ResponseInterceptor): void {
    this.responseInterceptors.push(interceptor);
  }

  removeRequestInterceptor(interceptor: RequestInterceptor): void {
    const index = this.requestInterceptors.indexOf(interceptor);
    if (index > -1) {
      this.requestInterceptors.splice(index, 1);
    }
  }

  removeResponseInterceptor(interceptor: ResponseInterceptor): void {
    const index = this.responseInterceptors.indexOf(interceptor);
    if (index > -1) {
      this.responseInterceptors.splice(index, 1);
    }
  }

  createCancellationToken(): CancellationToken {
    let cancelled = false;
    const callbacks: (() => void)[] = [];

    return {
      get cancelled() {
        return cancelled;
      },
      cancel() {
        cancelled = true;
        callbacks.forEach((callback) => callback());
      },
      onCancel(callback: () => void) {
        callbacks.push(callback);
      },
    };
  }
}
