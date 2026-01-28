/**
 * Application-level error boundary with fallback UI and logging integration.
 */

import React, { Component, ReactNode } from "react";
import { getLogger } from "../di";

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
  errorInfo?: React.ErrorInfo;
  errorId?: string;
}

interface AppErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: React.ErrorInfo) => void;
}

export class AppErrorBoundary extends Component<
  AppErrorBoundaryProps,
  ErrorBoundaryState
> {
  private logger = getLogger().child({ component: "AppErrorBoundary" });

  constructor(props: AppErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    // Update state so the next render will show the fallback UI
    return {
      hasError: true,
      error,
      errorId: `error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
    };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    const { errorId } = this.state;

    // Log the error with structured data
    this.logger.error("Application error caught by boundary", {
      errorId,
      error: {
        name: error.name,
        message: error.message,
        stack: error.stack,
      },
      errorInfo: {
        componentStack: errorInfo.componentStack,
      },
      timestamp: new Date().toISOString(),
      userAgent: (globalThis as any).navigator?.userAgent || "unknown",
      url: window.location.href,
    });

    // Call custom error handler if provided
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }

    // Update state with error info
    this.setState({
      error,
      errorInfo,
    });
  }

  private handleReload = () => {
    // Clear error state and reload the page
    this.setState({ hasError: false, error: undefined, errorInfo: undefined });
    window.location.reload();
  };

  private handleRetry = () => {
    // Clear error state and try to continue
    this.setState({ hasError: false, error: undefined, errorInfo: undefined });
  };

  private handleReportError = () => {
    const { error, errorInfo, errorId } = this.state;

    // Copy error details to clipboard
    const errorDetails = {
      errorId,
      timestamp: new Date().toISOString(),
      error: {
        name: error?.name,
        message: error?.message,
        stack: error?.stack,
      },
      errorInfo: {
        componentStack: errorInfo?.componentStack,
      },
      userAgent: (globalThis as any).navigator?.userAgent || "unknown",
      url: window.location.href,
    };

    (globalThis as any).navigator?.clipboard
      ?.writeText(JSON.stringify(errorDetails, null, 2))
      .then(() => {
        this.logger.info("Error details copied to clipboard", { errorId });
        alert(
          "Error details copied to clipboard. Please share this with the development team."
        );
      })
      .catch(() => {
        this.logger.warn("Failed to copy error details to clipboard", {
          errorId,
        });
        alert(
          "Failed to copy error details. Please check the console for error information."
        );
      });
  };

  render() {
    if (this.state.hasError) {
      // Custom fallback UI
      if (this.props.fallback) {
        return this.props.fallback;
      }

      // Default fallback UI
      return (
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            minHeight: "100vh",
            padding: "20px",
            backgroundColor: "#f8f9fa",
            fontFamily: "system-ui, -apple-system, sans-serif",
          }}
        >
          <div
            style={{
              maxWidth: "600px",
              textAlign: "center",
              backgroundColor: "white",
              padding: "40px",
              borderRadius: "8px",
              boxShadow: "0 4px 6px rgba(0, 0, 0, 0.1)",
            }}
          >
            <div
              style={{
                fontSize: "48px",
                marginBottom: "20px",
                color: "#dc3545",
              }}
            >
              ⚠️
            </div>

            <h1
              style={{
                fontSize: "24px",
                fontWeight: "bold",
                marginBottom: "16px",
                color: "#212529",
              }}
            >
              Something went wrong
            </h1>

            <p
              style={{
                fontSize: "16px",
                color: "#6c757d",
                marginBottom: "24px",
                lineHeight: "1.5",
              }}
            >
              We're sorry, but something unexpected happened. This error has
              been logged and our team will investigate.
            </p>

            {this.state.errorId && (
              <div
                style={{
                  backgroundColor: "#f8f9fa",
                  padding: "12px",
                  borderRadius: "4px",
                  marginBottom: "24px",
                  fontFamily: "monospace",
                  fontSize: "12px",
                  color: "#6c757d",
                }}
              >
                Error ID: {this.state.errorId}
              </div>
            )}

            <div
              style={{
                display: "flex",
                gap: "12px",
                justifyContent: "center",
                flexWrap: "wrap",
              }}
            >
              <button
                onClick={this.handleRetry}
                style={{
                  padding: "12px 24px",
                  backgroundColor: "#007bff",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                  cursor: "pointer",
                  fontSize: "14px",
                  fontWeight: "500",
                }}
              >
                Try Again
              </button>

              <button
                onClick={this.handleReload}
                style={{
                  padding: "12px 24px",
                  backgroundColor: "#28a745",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                  cursor: "pointer",
                  fontSize: "14px",
                  fontWeight: "500",
                }}
              >
                Reload Page
              </button>

              <button
                onClick={this.handleReportError}
                style={{
                  padding: "12px 24px",
                  backgroundColor: "#6c757d",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                  cursor: "pointer",
                  fontSize: "14px",
                  fontWeight: "500",
                }}
              >
                Report Error
              </button>
            </div>

            {(globalThis as any).process?.env?.NODE_ENV === "development" &&
              this.state.error && (
                <details
                  style={{
                    marginTop: "24px",
                    textAlign: "left",
                    backgroundColor: "#f8f9fa",
                    padding: "16px",
                    borderRadius: "4px",
                    fontSize: "12px",
                  }}
                >
                  <summary
                    style={{
                      cursor: "pointer",
                      fontWeight: "bold",
                      marginBottom: "8px",
                    }}
                  >
                    Development Error Details
                  </summary>
                  <pre
                    style={{
                      whiteSpace: "pre-wrap",
                      wordBreak: "break-word",
                      color: "#dc3545",
                      margin: 0,
                    }}
                  >
                    {this.state.error.toString()}
                    {this.state.errorInfo?.componentStack}
                  </pre>
                </details>
              )}
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
