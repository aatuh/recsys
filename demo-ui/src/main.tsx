import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { ToastProvider } from "./contexts/ToastContext";
import { logger } from "./utils/logger";
import { initializeApiClient } from "./lib/api-client-config";

// Initialize API client configuration
initializeApiClient();

const el = document.getElementById("root")!;
// Initialize basic page view analytics
logger.info("app.start");
createRoot(el).render(
  <ToastProvider>
    <App />
  </ToastProvider>
);
