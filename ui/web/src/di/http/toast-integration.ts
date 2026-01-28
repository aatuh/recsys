/**
 * HTTP client integration with toast notifications for user-friendly error handling.
 */

import { ApiError } from "./errors";
import { useToast } from "../../contexts/ToastContext";
import { useTelemetry, TELEMETRY_EVENTS } from "../../hooks/useTelemetry";

/**
 * Enhanced error interceptor that shows toast notifications for network failures.
 */
export class ToastErrorInterceptor {
  private toast: ReturnType<typeof useToast>;
  private telemetry: ReturnType<typeof useTelemetry>;

  constructor(
    toast: ReturnType<typeof useToast>,
    telemetry: ReturnType<typeof useTelemetry>
  ) {
    this.toast = toast;
    this.telemetry = telemetry;
  }

  /**
   * Process HTTP errors and show appropriate toast notifications.
   */
  processError(error: ApiError): ApiError {
    // Track the error
    this.telemetry.track(TELEMETRY_EVENTS.API_ERROR, {
      status: error.status,
      code: error.code,
      message: error.message,
      isRetryable: error.retryable,
      isAuthError: error.authError,
      isNetworkError: error.networkError,
      isTimeout: error.timeout,
    });

    // Show appropriate toast based on error type
    if (error.authError) {
      this.toast.showError(
        "Authentication Required",
        "Please log in to continue.",
        {
          action: {
            label: "Log In",
            onClick: () => {
              // Future: Navigate to login page
              console.log("Navigate to login page");
            },
          },
        }
      );
    } else if (error.networkError) {
      this.toast.showError(
        "Network Error",
        "Unable to connect to the server. Please check your internet connection.",
        {
          action: {
            label: "Retry",
            onClick: () => {
              // Future: Retry the failed request
              console.log("Retry failed request");
            },
          },
        }
      );
    } else if (error.timeout) {
      this.toast.showWarning(
        "Request Timeout",
        "The request took too long to complete. Please try again.",
        {
          action: {
            label: "Retry",
            onClick: () => {
              // Future: Retry the failed request
              console.log("Retry timed out request");
            },
          },
        }
      );
    } else if (error.status && error.status >= 500) {
      this.toast.showError(
        "Server Error",
        "Something went wrong on our end. We're working to fix it.",
        {
          action: {
            label: "Report Issue",
            onClick: () => {
              // Future: Open issue reporting
              console.log("Open issue reporting");
            },
          },
        }
      );
    } else if (error.status && error.status >= 400) {
      this.toast.showWarning(
        "Request Failed",
        `The request failed with status ${error.status}.`,
        {
          action: {
            label: "Details",
            onClick: () => {
              // Show error details in console
              console.error("Request failed:", error);
            },
          },
        }
      );
    } else {
      // Generic error
      this.toast.showError(
        "Request Failed",
        "An unexpected error occurred. Please try again.",
        {
          action: {
            label: "Details",
            onClick: () => {
              console.error("Unexpected error:", error);
            },
          },
        }
      );
    }

    return error;
  }
}

/**
 * Success interceptor that shows success toasts for important operations.
 */
export class ToastSuccessInterceptor {
  private toast: ReturnType<typeof useToast>;
  private telemetry: ReturnType<typeof useTelemetry>;

  constructor(
    toast: ReturnType<typeof useToast>,
    telemetry: ReturnType<typeof useTelemetry>
  ) {
    this.toast = toast;
    this.telemetry = telemetry;
  }

  /**
   * Process successful responses and show appropriate success toasts.
   */
  processSuccess(response: any, request: any): any {
    // Track successful API calls
    this.telemetry.track(TELEMETRY_EVENTS.API_SUCCESS, {
      status: response.status,
      url: request.url,
      method: request.method,
      duration: Date.now() - (request.startTime || Date.now()),
    });

    // Show success toast for important operations
    if (this.shouldShowSuccessToast(request)) {
      this.toast.showSuccess("Success", this.getSuccessMessage(request), {
        duration: 3000, // Shorter duration for success messages
      });
    }

    return response;
  }

  private shouldShowSuccessToast(request: any): boolean {
    // Show success toast for important operations
    const importantOperations = [
      "POST",
      "PUT",
      "DELETE", // Write operations
    ];

    return (
      importantOperations.includes(request.method) ||
      request.url.includes("/recommendations") ||
      request.url.includes("/users") ||
      request.url.includes("/items")
    );
  }

  private getSuccessMessage(request: any): string {
    if (request.url.includes("/recommendations")) {
      return "Recommendations updated successfully";
    } else if (request.url.includes("/users")) {
      return "User data saved successfully";
    } else if (request.url.includes("/items")) {
      return "Item data saved successfully";
    } else if (request.method === "POST") {
      return "Data created successfully";
    } else if (request.method === "PUT") {
      return "Data updated successfully";
    } else if (request.method === "DELETE") {
      return "Data deleted successfully";
    } else {
      return "Operation completed successfully";
    }
  }
}

/**
 * Hook for creating toast-integrated HTTP interceptors.
 */
export function useToastHttpInterceptors() {
  const toast = useToast();
  const telemetry = useTelemetry();

  return {
    errorInterceptor: new ToastErrorInterceptor(toast, telemetry),
    successInterceptor: new ToastSuccessInterceptor(toast, telemetry),
  };
}
