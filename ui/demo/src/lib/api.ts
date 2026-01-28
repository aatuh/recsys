import { OpenAPI } from "./api-client";

let configuredBase: string | null = null;

export function ensureApiBase(baseUrl: string): void {
  if (!baseUrl) return;
  if (configuredBase === baseUrl) return;
  OpenAPI.BASE = baseUrl;
  OpenAPI.WITH_CREDENTIALS = false;
  configuredBase = baseUrl;
}
