import React from "react";
import { color, radius, spacing, text } from "./tokens";

export type ToastKind = "info" | "success" | "error";

export interface ToastMessage {
  id: string;
  kind: ToastKind;
  title?: string;
  message: string;
  durationMs?: number;
}

interface ToastContextValue {
  show: (t: Omit<ToastMessage, "id">) => void;
  info: (message: string, title?: string) => void;
  success: (message: string, title?: string) => void;
  error: (message: string, title?: string) => void;
}

const ToastContext = React.createContext<ToastContextValue | undefined>(
  undefined
);

export function useToast() {
  const ctx = React.useContext(ToastContext);
  if (!ctx) throw new Error("useToast must be used within ToastProvider");
  return ctx;
}

export function ToastProvider(props: { children: React.ReactNode }) {
  const [items, setItems] = React.useState<ToastMessage[]>([]);

  const remove = React.useCallback((id: string) => {
    setItems((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const show = React.useCallback(
    (t: Omit<ToastMessage, "id">) => {
      const id = `${Date.now()}-${Math.random().toString(36).slice(2)}`;
      const toast: ToastMessage = {
        id,
        durationMs: 3500,
        ...t,
      };
      setItems((prev) => [...prev, toast]);
      if (toast.durationMs && toast.durationMs > 0) {
        window.setTimeout(() => remove(id), toast.durationMs);
      }
    },
    [remove]
  );

  const value: ToastContextValue = React.useMemo(
    () => ({
      show,
      info: (message, title) => show({ kind: "info", message, title }),
      success: (message, title) => show({ kind: "success", message, title }),
      error: (message, title) => show({ kind: "error", message, title }),
    }),
    [show]
  );

  return (
    <ToastContext.Provider value={value}>
      {props.children}
      <div
        style={{
          position: "fixed",
          top: spacing.xl,
          right: spacing.xl,
          display: "flex",
          flexDirection: "column",
          gap: spacing.md,
          zIndex: 1000,
        }}
      >
        {items.map((t) => (
          <ToastItem key={t.id} toast={t} onClose={() => remove(t.id)} />
        ))}
      </div>
    </ToastContext.Provider>
  );
}

function ToastItem(props: { toast: ToastMessage; onClose: () => void }) {
  const { toast } = props;

  let border = color.border;
  let bg = color.panelSubtle;
  let fg = color.text;
  if (toast.kind === "success") {
    border = color.success;
    bg = color.successBg;
    fg = color.text;
  } else if (toast.kind === "error") {
    border = color.danger;
    bg = color.dangerBg;
    fg = color.text;
  }

  return (
    <div
      role="status"
      aria-live="polite"
      style={{
        minWidth: 260,
        maxWidth: 360,
        border: `1px solid ${border}`,
        borderLeftWidth: 4,
        background: bg,
        color: fg,
        borderRadius: radius.lg,
        boxShadow: "0 2px 6px rgba(0,0,0,0.08)",
      }}
    >
      <div style={{ padding: spacing.lg }}>
        {toast.title && (
          <div style={{ fontWeight: 600, marginBottom: spacing.xs }}>
            {toast.title}
          </div>
        )}
        <div style={{ fontSize: text.md }}>{toast.message}</div>
        <div style={{ display: "flex", justifyContent: "flex-end" }}>
          <button
            type="button"
            onClick={props.onClose}
            style={{
              marginTop: spacing.md,
              fontSize: text.sm,
              background: "transparent",
              border: `1px solid ${color.border}`,
              borderRadius: radius.sm,
              padding: `${spacing.xs}px ${spacing.sm}px`,
              cursor: "pointer",
            }}
          >
            Dismiss
          </button>
        </div>
      </div>
    </div>
  );
}
