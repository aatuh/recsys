// Centralized runtime configuration for the demo UI.
// Reads once at startup and configures the generated API client.

import { OpenAPI } from "./lib/api-client";

export const apiBase: string =
  (import.meta as any).env?.VITE_API_BASE_URL?.toString() || "/api";

export const swaggerUiUrl: string =
  (import.meta as any).env?.VITE_SWAGGER_UI_URL?.toString() ||
  "http://localhost:8081";

export const customChatGptUrl: string | undefined = (
  import.meta as any
).env?.VITE_CUSTOM_CHATGPT_URL?.toString();

// Configure the generated API client base once.
OpenAPI.BASE = apiBase;

export const config = {
  apiBase,
  swaggerUiUrl,
  customChatGptUrl,
};
