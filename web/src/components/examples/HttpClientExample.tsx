import React, { useState } from "react";
import { getHttpClient, getLogger } from "../../di";
import { ApiError } from "../../di/http";
import { Button } from "../primitives/UIComponents";

/**
 * Example component demonstrating the enhanced HTTP client features.
 * Shows interceptors, retry logic, circuit breaker, and error handling.
 */
export function HttpClientExample() {
  const [logs, setLogs] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [_circuitState, _setCircuitState] = useState<string>("unknown");

  const httpClient = getHttpClient();
  const logger = getLogger();

  const addLog = (message: string) => {
    setLogs((prev) => [...prev, `${new Date().toISOString()}: ${message}`]);
  };

  const handleSuccessfulRequest = async () => {
    setLoading(true);
    addLog("Making successful request...");

    try {
      const response = await httpClient.get("/api/health", {
        headers: { "x-test": "success" },
      });

      addLog(
        `✅ Success: ${response.status} - ${JSON.stringify(response.data)}`
      );
      logger.info("HTTP request successful", {
        status: response.status,
        requestId: response.headers["x-request-id"],
      });
    } catch (error) {
      addLog(
        `❌ Error: ${error instanceof Error ? error.message : String(error)}`
      );
    } finally {
      setLoading(false);
    }
  };

  const handleRetryableRequest = async () => {
    setLoading(true);
    addLog("Making retryable request (will fail and retry)...");

    try {
      // This will likely fail and trigger retry logic
      const response = await httpClient.get("/api/nonexistent", {
        headers: { "x-test": "retry" },
      });

      addLog(`✅ Success: ${response.status}`);
    } catch (error) {
      if (error instanceof ApiError) {
        addLog(`❌ ApiError: ${error.message} (retryable: ${error.retryable})`);
        addLog(
          `   Status: ${error.status}, Auth: ${error.authError}, Server: ${error.serverError}`
        );
      } else {
        addLog(
          `❌ Error: ${error instanceof Error ? error.message : String(error)}`
        );
      }
    } finally {
      setLoading(false);
    }
  };

  const handleTimeoutRequest = async () => {
    setLoading(true);
    addLog("Making timeout request...");

    try {
      const response = await httpClient.get("/api/slow", {
        timeout: 1000, // 1 second timeout
      });

      addLog(`✅ Success: ${response.status}`);
    } catch (error) {
      if (error instanceof ApiError) {
        addLog(`❌ Timeout: ${error.message} (timeout: ${error.timeout})`);
      } else {
        addLog(
          `❌ Error: ${error instanceof Error ? error.message : String(error)}`
        );
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCircuitBreakerTest = async () => {
    setLoading(true);
    addLog("Testing circuit breaker with multiple failures...");

    try {
      // Make multiple requests that will fail to trigger circuit breaker
      for (let i = 0; i < 6; i++) {
        try {
          await httpClient.get("/api/circuit-breaker-test", {
            headers: { "x-test": `circuit-${i}` },
          });
          addLog(`Request ${i + 1}: ✅ Success`);
        } catch (error) {
          addLog(
            `Request ${i + 1}: ❌ Failed - ${
              error instanceof Error ? error.message : String(error)
            }`
          );
        }
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCancellationTest = async () => {
    setLoading(true);
    addLog("Testing request cancellation...");

    const token = httpClient.createCancellationToken();

    // Cancel after 2 seconds
    setTimeout(() => {
      token.cancel();
      addLog("🛑 Request cancelled");
    }, 2000);

    try {
      const response = await httpClient.get("/api/slow", {
        signal: token.cancelled
          ? undefined
          : new (globalThis as any).AbortController().signal,
      });

      addLog(`✅ Success: ${response.status}`);
    } catch (error) {
      if (error instanceof ApiError && error.timeout) {
        addLog(`❌ Cancelled: ${error.message}`);
      } else {
        addLog(
          `❌ Error: ${error instanceof Error ? error.message : String(error)}`
        );
      }
    } finally {
      setLoading(false);
    }
  };

  const handleInterceptorTest = async () => {
    setLoading(true);
    addLog("Testing interceptors...");

    // Add a custom interceptor
    const customInterceptor = {
      onRequest: (request: any) => {
        addLog(`🔧 Request interceptor: ${request.method} ${request.url}`);
        return {
          ...request,
          headers: {
            ...request.headers,
            "x-custom-header": "interceptor-test",
          },
        };
      },
      onResponse: (response: any) => {
        addLog(
          `🔧 Response interceptor: ${response.status} ${response.statusText}`
        );
        return response;
      },
    };

    httpClient.addRequestInterceptor(customInterceptor);
    httpClient.addResponseInterceptor(customInterceptor);

    try {
      const response = await httpClient.get("/api/test", {
        headers: { "x-test": "interceptor" },
      });

      addLog(`✅ Success: ${response.status}`);
    } catch (error) {
      addLog(
        `❌ Error: ${error instanceof Error ? error.message : String(error)}`
      );
    } finally {
      // Clean up interceptors
      httpClient.removeRequestInterceptor(customInterceptor);
      httpClient.removeResponseInterceptor(customInterceptor);
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: "20px", border: "1px solid #ccc", margin: "10px" }}>
      <h3>Enhanced HTTP Client Example</h3>

      <div style={{ marginBottom: "20px" }}>
        <h4>Request Tests</h4>
        <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
          <Button onClick={handleSuccessfulRequest} disabled={loading}>
            Successful Request
          </Button>
          <Button onClick={handleRetryableRequest} disabled={loading}>
            Retryable Request
          </Button>
          <Button onClick={handleTimeoutRequest} disabled={loading}>
            Timeout Request
          </Button>
          <Button onClick={handleCircuitBreakerTest} disabled={loading}>
            Circuit Breaker Test
          </Button>
          <Button onClick={handleCancellationTest} disabled={loading}>
            Cancellation Test
          </Button>
          <Button onClick={handleInterceptorTest} disabled={loading}>
            Interceptor Test
          </Button>
        </div>
      </div>

      <div>
        <h4>Request Logs</h4>
        <div
          style={{
            height: "300px",
            overflow: "auto",
            border: "1px solid #ddd",
            padding: "10px",
            backgroundColor: "#f9f9f9",
            fontFamily: "monospace",
            fontSize: "12px",
          }}
        >
          {logs.map((log, index) => (
            <div key={index} style={{ marginBottom: "2px" }}>
              {log}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
