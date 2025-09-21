// Minimal client-side logger and analytics shim

type LogLevel = "debug" | "info" | "warn" | "error";

interface LogFields {
  [key: string]: unknown;
}

function emit(level: LogLevel, message: string, fields?: LogFields) {
  const payload = { level, message, ...fields };

  console[level]("[recsys-ui]", payload);

  // Optional analytics shim; ignored if not present
  try {
    const anyWindow = window as any;
    if (anyWindow?.analytics?.track) {
      anyWindow.analytics.track("recsys_ui_event", payload);
    }
    if (anyWindow?.gtag) {
      anyWindow.gtag("event", message, payload);
    }
  } catch {
    // no-op
  }
}

export const logger = {
  debug: (message: string, fields?: LogFields) =>
    emit("debug", message, fields),
  info: (message: string, fields?: LogFields) => emit("info", message, fields),
  warn: (message: string, fields?: LogFields) => emit("warn", message, fields),
  error: (message: string, fields?: LogFields) =>
    emit("error", message, fields),
};
