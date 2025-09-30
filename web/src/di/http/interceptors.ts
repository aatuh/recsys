/**
 * HTTP interceptors for authentication and error handling.
 * These are pluggable stubs that can be extended later.
 */

import type { HttpRequest, HttpResponse, HttpError } from "./types";
import { ApiError } from "./errors";

/**
 * Authentication interceptor stub.
 * Can be extended to attach auth headers, refresh tokens, etc.
 */
export class AuthInterceptor {
  /**
   * Attach authentication header to requests.
   * Currently a stub - can be extended to add JWT tokens, API keys, etc.
   */
  static attachAuthHeader(request: HttpRequest): HttpRequest {
    // TODO: Implement auth header attachment
    // Example: request.headers['Authorization'] = `Bearer ${token}`;
    return request;
  }

  /**
   * Handle 401 authentication errors.
   * Currently a stub - can be extended to refresh tokens, redirect to login, etc.
   */
  static async handle401(error: HttpError): Promise<HttpError> {
    // TODO: Implement 401 handling
    // Example: refresh token, redirect to login, etc.
    console.warn("Authentication error detected:", error);
    return error;
  }
}

/**
 * Error handling interceptor.
 * Provides centralized error processing and logging.
 */
export class ErrorInterceptor {
  /**
   * Process HTTP errors and convert them to ApiError instances.
   */
  static processError(error: Error): ApiError {
    if (error instanceof ApiError) {
      return error;
    }

    // Convert generic errors to ApiError
    return ApiError.fromHttpError(error);
  }

  /**
   * Log errors for debugging and monitoring.
   */
  static logError(error: ApiError, request: HttpRequest): void {
    const logData = {
      url: request.url,
      method: request.method,
      status: error.status,
      retryable: error.retryable,
      authError: error.authError,
      serverError: error.serverError,
      networkError: error.networkError,
      timeout: error.timeout,
    };

    if (error.serverError || error.networkError) {
      console.error("HTTP Error:", error.message, logData);
    } else if (error.authError) {
      console.warn("Auth Error:", error.message, logData);
    } else {
      console.info("HTTP Error:", error.message, logData);
    }
  }
}

/**
 * Request ID interceptor.
 * Ensures all requests have a unique request ID for tracing.
 */
export class RequestIdInterceptor {
  /**
   * Add request ID to all outgoing requests.
   */
  static addRequestId(request: HttpRequest): HttpRequest {
    if (!request.headers?.["x-request-id"]) {
      return {
        ...request,
        headers: {
          ...request.headers,
          "x-request-id": (globalThis as any).crypto.randomUUID(),
        },
      };
    }
    return request;
  }
}

/**
 * Response logging interceptor.
 * Logs successful responses for debugging.
 */
export class ResponseLoggingInterceptor {
  /**
   * Log successful responses.
   */
  static logResponse<T>(response: HttpResponse<T>): HttpResponse<T> {
    const logData = {
      status: response.status,
      statusText: response.statusText,
      requestId: response.headers["x-request-id"],
    };

    if (response.status >= 400) {
      console.warn("HTTP Response Error:", logData);
    } else {
      console.debug("HTTP Response:", logData);
    }

    return response;
  }
}
