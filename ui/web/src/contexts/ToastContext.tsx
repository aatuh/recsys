/**
 * Toast notification context for displaying non-blocking messages to users.
 */

import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  ReactNode,
} from "react";
import { getLogger } from "../di";

export type ToastType = "success" | "error" | "warning" | "info";

export interface Toast {
  id: string;
  type: ToastType;
  title: string;
  message: string;
  duration?: number; // Auto-dismiss duration in ms
  persistent?: boolean; // Don't auto-dismiss
  action?: {
    label: string;
    onClick: () => void;
  };
  onDismiss?: () => void;
}

export interface ToastContextType {
  toasts: Toast[];
  showToast: (toast: Omit<Toast, "id">) => string;
  dismissToast: (id: string) => void;
  dismissAllToasts: () => void;
  showSuccess: (
    title: string,
    message?: string,
    options?: Partial<Toast>
  ) => string;
  showError: (
    title: string,
    message?: string,
    options?: Partial<Toast>
  ) => string;
  showWarning: (
    title: string,
    message?: string,
    options?: Partial<Toast>
  ) => string;
  showInfo: (
    title: string,
    message?: string,
    options?: Partial<Toast>
  ) => string;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export interface ToastProviderProps {
  children: ReactNode;
  maxToasts?: number;
  defaultDuration?: number;
}

export function ToastProvider({
  children,
  maxToasts = 5,
  defaultDuration = 5000,
}: ToastProviderProps) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const logger = getLogger().child({ component: "ToastProvider" });

  const generateId = useCallback(() => {
    return `toast_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }, []);

  const showToast = useCallback(
    (toast: Omit<Toast, "id">): string => {
      const id = generateId();
      const newToast: Toast = {
        id,
        duration: defaultDuration,
        ...toast,
      };

      setToasts((prev) => {
        const updated = [...prev, newToast];

        // Limit number of toasts
        if (updated.length > maxToasts) {
          return updated.slice(-maxToasts);
        }

        return updated;
      });

      logger.debug("Toast shown", {
        id,
        type: toast.type,
        title: toast.title,
      });

      // Auto-dismiss if not persistent
      if (!toast.persistent && toast.duration !== 0) {
        setTimeout(() => {
          dismissToast(id);
        }, toast.duration || defaultDuration);
      }

      return id;
    },
    [generateId, defaultDuration, maxToasts, logger]
  );

  const dismissToast = useCallback(
    (id: string) => {
      setToasts((prev) => {
        const toast = prev.find((t) => t.id === id);
        if (toast?.onDismiss) {
          toast.onDismiss();
        }

        return prev.filter((t) => t.id !== id);
      });

      logger.debug("Toast dismissed", { id });
    },
    [logger]
  );

  const dismissAllToasts = useCallback(() => {
    setToasts((prev) => {
      prev.forEach((toast) => {
        if (toast.onDismiss) {
          toast.onDismiss();
        }
      });
      return [];
    });

    logger.debug("All toasts dismissed");
  }, [logger]);

  const showSuccess = useCallback(
    (title: string, message?: string, options?: Partial<Toast>) => {
      return showToast({
        type: "success",
        title,
        message: message || "",
        ...options,
      });
    },
    [showToast]
  );

  const showError = useCallback(
    (title: string, message?: string, options?: Partial<Toast>) => {
      return showToast({
        type: "error",
        title,
        message: message || "",
        persistent: true, // Errors are persistent by default
        ...options,
      });
    },
    [showToast]
  );

  const showWarning = useCallback(
    (title: string, message?: string, options?: Partial<Toast>) => {
      return showToast({
        type: "warning",
        title,
        message: message || "",
        ...options,
      });
    },
    [showToast]
  );

  const showInfo = useCallback(
    (title: string, message?: string, options?: Partial<Toast>) => {
      return showToast({
        type: "info",
        title,
        message: message || "",
        ...options,
      });
    },
    [showToast]
  );

  const contextValue: ToastContextType = {
    toasts,
    showToast,
    dismissToast,
    dismissAllToasts,
    showSuccess,
    showError,
    showWarning,
    showInfo,
  };

  return (
    <ToastContext.Provider value={contextValue}>
      {children}
    </ToastContext.Provider>
  );
}

export function useToast(): ToastContextType {
  const context = useContext(ToastContext);
  if (context === undefined) {
    throw new Error("useToast must be used within a ToastProvider");
  }
  return context;
}
