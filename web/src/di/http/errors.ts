/**
 * Unified API error handling with classification for retry and auth behaviors.
 */

export interface ApiErrorDetails {
  status?: number;
  code?: string;
  message: string;
  retryable: boolean;
  authError: boolean;
  serverError: boolean;
  clientError: boolean;
  networkError: boolean;
  timeout: boolean;
}

export class ApiError extends Error {
  public readonly status?: number;
  public readonly code?: string;
  public readonly retryable: boolean;
  public readonly authError: boolean;
  public readonly serverError: boolean;
  public readonly clientError: boolean;
  public readonly networkError: boolean;
  public readonly timeout: boolean;
  public readonly originalError?: Error;

  constructor(
    message: string,
    details: Partial<ApiErrorDetails> = {},
    originalError?: Error
  ) {
    super(message);
    this.name = "ApiError";

    this.status = details.status;
    this.code = details.code;
    this.retryable = details.retryable ?? false;
    this.authError = details.authError ?? false;
    this.serverError = details.serverError ?? false;
    this.clientError = details.clientError ?? false;
    this.networkError = details.networkError ?? false;
    this.timeout = details.timeout ?? false;
    this.originalError = originalError;
  }

  /**
   * Classify an error based on status code and error type.
   */
  static fromHttpError(error: Error, status?: number): ApiError {
    const isNetworkError =
      !status &&
      (error.name === "TypeError" ||
        error.name === "NetworkError" ||
        error.message.includes("fetch"));

    const isTimeout =
      error.name === "AbortError" || error.message.includes("timeout");

    const isAuthError = status === 401 || status === 403;
    const isServerError = status ? status >= 500 : false;
    const isClientError = status ? status >= 400 && status < 500 : false;

    const retryable = isServerError || isNetworkError || isTimeout;

    return new ApiError(
      error.message,
      {
        status,
        retryable,
        authError: isAuthError,
        serverError: isServerError,
        clientError: isClientError,
        networkError: isNetworkError,
        timeout: isTimeout,
      },
      error
    );
  }

  /**
   * Create a timeout error.
   */
  static timeout(message: string = "Request timeout"): ApiError {
    return new ApiError(message, {
      timeout: true,
      retryable: true,
    });
  }

  /**
   * Create a network error.
   */
  static network(message: string = "Network error"): ApiError {
    return new ApiError(message, {
      networkError: true,
      retryable: true,
    });
  }

  /**
   * Create an authentication error.
   */
  static auth(
    message: string = "Authentication failed",
    status: number = 401
  ): ApiError {
    return new ApiError(message, {
      status,
      authError: true,
      retryable: false,
    });
  }

  /**
   * Create a server error.
   */
  static server(
    message: string = "Server error",
    status: number = 500
  ): ApiError {
    return new ApiError(message, {
      status,
      serverError: true,
      retryable: true,
    });
  }

  /**
   * Create a client error.
   */
  static client(
    message: string = "Client error",
    status: number = 400
  ): ApiError {
    return new ApiError(message, {
      status,
      clientError: true,
      retryable: false,
    });
  }
}
