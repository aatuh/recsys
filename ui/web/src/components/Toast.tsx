/**
 * Toast notification component for displaying messages to users.
 */

import React, { useState, useEffect } from "react";
import { useToast, type Toast as ToastType } from "../contexts/ToastContext";

export interface ToastProps {
  toast: ToastType;
  onDismiss: (id: string) => void;
}

export function Toast({ toast, onDismiss }: ToastProps) {
  const [isVisible, setIsVisible] = useState(false);
  const [isLeaving, setIsLeaving] = useState(false);

  useEffect(() => {
    // Animate in
    const timer = setTimeout(() => setIsVisible(true), 10);
    return () => clearTimeout(timer);
  }, []);

  const handleDismiss = () => {
    setIsLeaving(true);
    setTimeout(() => onDismiss(toast.id), 300); // Match animation duration
  };

  const getToastStyles = () => {
    const baseStyles = {
      display: "flex",
      alignItems: "flex-start",
      padding: "16px",
      marginBottom: "8px",
      borderRadius: "8px",
      boxShadow: "0 4px 12px rgba(0, 0, 0, 0.15)",
      maxWidth: "400px",
      minWidth: "300px",
      transform: isVisible && !isLeaving ? "translateX(0)" : "translateX(100%)",
      opacity: isVisible && !isLeaving ? 1 : 0,
      transition: "all 0.3s ease-in-out",
      cursor: "pointer",
      position: "relative" as const,
    };

    switch (toast.type) {
      case "success":
        return {
          ...baseStyles,
          backgroundColor: "#d4edda",
          borderLeft: "4px solid #28a745",
          color: "#155724",
        };
      case "error":
        return {
          ...baseStyles,
          backgroundColor: "#f8d7da",
          borderLeft: "4px solid #dc3545",
          color: "#721c24",
        };
      case "warning":
        return {
          ...baseStyles,
          backgroundColor: "#fff3cd",
          borderLeft: "4px solid #ffc107",
          color: "#856404",
        };
      case "info":
      default:
        return {
          ...baseStyles,
          backgroundColor: "#d1ecf1",
          borderLeft: "4px solid #17a2b8",
          color: "#0c5460",
        };
    }
  };

  const getIcon = () => {
    switch (toast.type) {
      case "success":
        return "✅";
      case "error":
        return "❌";
      case "warning":
        return "⚠️";
      case "info":
      default:
        return "ℹ️";
    }
  };

  return (
    <div
      style={getToastStyles()}
      onClick={handleDismiss}
      role="alert"
      aria-live="polite"
    >
      <div style={{ marginRight: "12px", fontSize: "18px" }}>{getIcon()}</div>

      <div style={{ flex: 1 }}>
        <div
          style={{
            fontWeight: "bold",
            fontSize: "14px",
            marginBottom: toast.message ? "4px" : "0",
          }}
        >
          {toast.title}
        </div>

        {toast.message && (
          <div
            style={{
              fontSize: "13px",
              lineHeight: "1.4",
              opacity: 0.9,
            }}
          >
            {toast.message}
          </div>
        )}

        {toast.action && (
          <button
            onClick={(e) => {
              e.stopPropagation();
              toast.action!.onClick();
            }}
            style={{
              marginTop: "8px",
              padding: "4px 8px",
              backgroundColor: "transparent",
              border: "1px solid currentColor",
              borderRadius: "4px",
              fontSize: "12px",
              cursor: "pointer",
              color: "inherit",
            }}
          >
            {toast.action.label}
          </button>
        )}
      </div>

      <button
        onClick={(e) => {
          e.stopPropagation();
          handleDismiss();
        }}
        style={{
          marginLeft: "8px",
          background: "none",
          border: "none",
          fontSize: "16px",
          cursor: "pointer",
          color: "inherit",
          opacity: 0.7,
          padding: "0",
          width: "20px",
          height: "20px",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
        aria-label="Dismiss notification"
      >
        ×
      </button>
    </div>
  );
}

export function ToastContainer() {
  const { toasts, dismissToast } = useToast();

  if (toasts.length === 0) {
    return null;
  }

  return (
    <div
      style={{
        position: "fixed",
        top: "20px",
        right: "20px",
        zIndex: 9999,
        pointerEvents: "none",
      }}
    >
      {toasts.map((toast) => (
        <div
          key={toast.id}
          style={{
            pointerEvents: "auto",
            marginBottom: "8px",
          }}
        >
          <Toast toast={toast} onDismiss={dismissToast} />
        </div>
      ))}
    </div>
  );
}
