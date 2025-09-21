import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { ToastProvider } from "./ui/Toast";
import { logger } from "./utils/logger";

const el = document.getElementById("root")!;
// Initialize basic page view analytics
logger.info("app.start");
createRoot(el).render(
  <ToastProvider>
    <App />
  </ToastProvider>
);
