/**
 * HTTP client types and interfaces for advanced HTTP operations.
 */

export interface HttpRequest {
  url: string;
  method: string;
  headers?: Record<string, string>;
  body?: unknown;
  timeout?: number;
  signal?: any;
}

export interface HttpResponse<T = unknown> {
  data: T;
  status: number;
  statusText: string;
  headers: Record<string, string>;
}

export interface HttpError extends Error {
  status?: number;
  statusText?: string;
  response?: HttpResponse;
  request?: HttpRequest;
}

export interface RequestInterceptor {
  onRequest?: (request: HttpRequest) => HttpRequest | Promise<HttpRequest>;
}

export interface ResponseInterceptor {
  onResponse?: <T>(
    response: HttpResponse<T>
  ) => HttpResponse<T> | Promise<HttpResponse<T>>;
  onError?: (error: HttpError) => HttpError | Promise<HttpError>;
}

export interface HttpClientConfig {
  baseUrl?: string;
  timeout?: number;
  retries?: number;
  retryDelay?: number;
  retryBackoff?: boolean;
  circuitBreaker?: {
    enabled: boolean;
    failureThreshold: number;
    timeout: number;
    resetTimeout: number;
  };
  jitter?: boolean;
}

export interface CancellationToken {
  readonly cancelled: boolean;
  cancel(): void;
  onCancel(callback: () => void): void;
}

export interface HttpClient {
  get<T>(url: string, options?: Partial<HttpRequest>): Promise<HttpResponse<T>>;
  post<T>(
    url: string,
    data?: unknown,
    options?: Partial<HttpRequest>
  ): Promise<HttpResponse<T>>;
  put<T>(
    url: string,
    data?: unknown,
    options?: Partial<HttpRequest>
  ): Promise<HttpResponse<T>>;
  delete<T>(
    url: string,
    options?: Partial<HttpRequest>
  ): Promise<HttpResponse<T>>;

  addRequestInterceptor(interceptor: RequestInterceptor): void;
  addResponseInterceptor(interceptor: ResponseInterceptor): void;
  removeRequestInterceptor(interceptor: RequestInterceptor): void;
  removeResponseInterceptor(interceptor: ResponseInterceptor): void;

  createCancellationToken(): CancellationToken;
}
