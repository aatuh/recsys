/**
 * Central configuration for the OpenAPI-generated client.
 * Provides unified base URL, headers, and auth-ready setup.
 */

import { OpenAPI } from "./api-client";
import { config } from "../config";

/**
 * Generate a unique request ID for tracking requests.
 */
function generateRequestId(): string {
  return globalThis.crypto.randomUUID();
}

/**
 * Get authorization token (stubbed for now, ready for future auth).
 */
async function getAuthToken(): Promise<string> {
  // TODO: Implement actual auth token retrieval when auth is enabled
  // For now, return empty string to indicate no auth
  return "";
}

/**
 * Get request headers including request ID and auth.
 */
async function getRequestHeaders(): Promise<Record<string, string>> {
  const headers: Record<string, string> = {
    "x-request-id": generateRequestId(),
  };

  // Add auth header if token is available
  const token = await getAuthToken();
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  return headers;
}

/**
 * Remove a single trailing slash from a URL if present.
 */
function stripTrailingSlash(url: string): string {
  return url.endsWith("/") ? url.slice(0, -1) : url;
}

/**
 * Configure the OpenAPI client with centralized settings.
 */
export function configureApiClient(): void {
  // Set base URL from config
  // Ensure no double slashes by removing leading slash from baseUrl when constructing absolute URL
  const baseUrl = config.api.baseUrl.startsWith("/")
    ? config.api.baseUrl.substring(1)
    : config.api.baseUrl;

  const absoluteBase = config.api.host
    ? new URL(baseUrl, config.api.host).toString()
    : config.api.baseUrl;

  // Normalize to avoid trailing slash
  OpenAPI.BASE = stripTrailingSlash(absoluteBase);

  // Set headers resolver for request ID and auth
  OpenAPI.HEADERS = getRequestHeaders;

  // Set token resolver (stubbed for now)
  OpenAPI.TOKEN = getAuthToken;

  // Configure credentials for CORS
  OpenAPI.WITH_CREDENTIALS = false;
  OpenAPI.CREDENTIALS = "include";
}

/**
 * Re-export configuration values for app use.
 */
export const apiBase = (() => {
  const baseUrl = config.api.baseUrl.startsWith("/")
    ? config.api.baseUrl.substring(1)
    : config.api.baseUrl;

  const absoluteBase = config.api.host
    ? new URL(baseUrl, config.api.host).toString()
    : config.api.baseUrl;

  return stripTrailingSlash(absoluteBase);
})();

export const swaggerUiUrl = config.api.swaggerUiUrl;

/**
 * Initialize the API client configuration.
 * Call this once at app startup.
 */
export function initializeApiClient(): void {
  configureApiClient();
}
