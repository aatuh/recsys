/**
 * Retry logic with exponential backoff and jitter.
 */

export interface RetryConfig {
  retries: number;
  retryDelay: number;
  retryBackoff: boolean;
  jitter: boolean;
}

export class RetryManager {
  public readonly retries: number;

  constructor(private config: RetryConfig) {
    this.retries = config.retries;
  }

  /**
   * Calculate the delay for the next retry attempt.
   */
  calculateDelay(attempt: number): number {
    if (attempt <= 0) {
      return 0;
    }

    let delay = this.config.retryDelay;

    // Apply exponential backoff if enabled
    if (this.config.retryBackoff) {
      delay = delay * Math.pow(2, attempt - 1);
    }

    // Apply jitter if enabled
    if (this.config.jitter) {
      const jitterRange = delay * 0.1; // 10% jitter
      const jitter = (Math.random() - 0.5) * 2 * jitterRange;
      delay = Math.max(0, delay + jitter);
    }

    return Math.min(delay, 30000); // Cap at 30 seconds
  }

  /**
   * Check if a request should be retried based on the error.
   */
  shouldRetry(error: Error, attempt: number): boolean {
    if (attempt >= this.config.retries) {
      return false;
    }

    // Don't retry on client errors (4xx) except for specific cases
    if (error.name === "ApiError") {
      const apiError = error as any;
      if (apiError.clientError && !apiError.timeout) {
        return false;
      }
    }

    // Don't retry on auth errors
    if (error.name === "ApiError") {
      const apiError = error as any;
      if (apiError.authError) {
        return false;
      }
    }

    // Retry on network errors, timeouts, and server errors
    return true;
  }

  /**
   * Wait for the calculated delay.
   */
  async waitForRetry(attempt: number): Promise<void> {
    const delay = this.calculateDelay(attempt);
    if (delay > 0) {
      await new Promise((resolve) => setTimeout(resolve, delay));
    }
  }
}
