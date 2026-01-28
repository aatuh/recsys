import type { Logger } from "./interfaces";

type LogLevel = "debug" | "info" | "warn" | "error";

interface LogEntry {
  level: LogLevel;
  message: string;
  fields?: Record<string, unknown>;
  timestamp: string;
}

/**
 * Enhanced logger implementation with structured logging, levels, and child loggers.
 * Supports no-op transport by default with hooks for external services.
 */
export class StructuredLogger implements Logger {
  private baseFields: Record<string, unknown>;
  private level: LogLevel;
  private enableAnalytics: boolean;

  constructor(
    level: LogLevel = "info",
    enableAnalytics: boolean = false,
    baseFields: Record<string, unknown> = {}
  ) {
    this.level = level;
    this.enableAnalytics = enableAnalytics;
    this.baseFields = baseFields;
  }

  private shouldLog(level: LogLevel): boolean {
    const levels = ["debug", "info", "warn", "error"];
    return levels.indexOf(level) >= levels.indexOf(this.level);
  }

  private createLogEntry(
    level: LogLevel,
    message: string,
    fields?: Record<string, unknown>
  ): LogEntry {
    return {
      level,
      message,
      fields: { ...this.baseFields, ...fields },
      timestamp: new Date().toISOString(),
    };
  }

  private emit(entry: LogEntry): void {
    if (!this.shouldLog(entry.level)) {
      return;
    }

    // Console output with structured format
    const consoleMethod = entry.level === "error" ? "error" : "log";
    console[consoleMethod](
      `[recsys-ui:${entry.level}]`,
      entry.message,
      entry.fields
    );

    // Analytics integration (optional)
    if (this.enableAnalytics) {
      this.sendToAnalytics(entry);
    }
  }

  private sendToAnalytics(entry: LogEntry): void {
    try {
      const anyWindow = window as any;

      // Segment.io integration
      if (anyWindow?.analytics?.track) {
        anyWindow.analytics.track("recsys_ui_log", {
          level: entry.level,
          message: entry.message,
          ...entry.fields,
        });
      }

      // Google Analytics integration
      if (anyWindow?.gtag) {
        anyWindow.gtag("event", "recsys_ui_log", {
          level: entry.level,
          message: entry.message,
          ...entry.fields,
        });
      }

      // Sentry integration (if available)
      if (anyWindow?.Sentry?.addBreadcrumb) {
        anyWindow.Sentry.addBreadcrumb({
          category: "log",
          level: entry.level,
          message: entry.message,
          data: entry.fields,
        });
      }
    } catch {
      // Silently fail if analytics services are not available
    }
  }

  debug(message: string, fields?: Record<string, unknown>): void {
    const entry = this.createLogEntry("debug", message, fields);
    this.emit(entry);
  }

  info(message: string, fields?: Record<string, unknown>): void {
    const entry = this.createLogEntry("info", message, fields);
    this.emit(entry);
  }

  warn(message: string, fields?: Record<string, unknown>): void {
    const entry = this.createLogEntry("warn", message, fields);
    this.emit(entry);
  }

  error(message: string, fields?: Record<string, unknown>): void {
    const entry = this.createLogEntry("error", message, fields);
    this.emit(entry);
  }

  child(fields: Record<string, unknown>): Logger {
    return new StructuredLogger(this.level, this.enableAnalytics, {
      ...this.baseFields,
      ...fields,
    });
  }
}

/**
 * No-op logger implementation for testing or when logging is disabled.
 */
export class NoOpLogger implements Logger {
  debug(): void {}
  info(): void {}
  warn(): void {}
  error(): void {}
  child(): Logger {
    return this;
  }
}
