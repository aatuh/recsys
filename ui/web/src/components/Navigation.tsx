import { Button } from "./primitives/UIComponents";
import { ThemeToggleCompact, ThemeToggleFull } from "./primitives/ThemeToggle";
import { color, radius, spacing, text } from "../ui/tokens";
import { useToast } from "../contexts/ToastContext";
import { useState } from "react";

export type ViewType =
  | "namespace-seed"
  | "recommendations-playground"
  | "bandit-playground"
  | "user-session"
  | "data-management"
  | "rules"
  | "documentation"
  | "explain-llm"
  | "privacy-policy";

interface NavigationProps {
  activeView: ViewType;
  onViewChange: (value: ViewType) => void;
  swaggerUrl: string;
  customChatGptUrl?: string;
  namespace: string;
}

const viewLabels: Record<ViewType, string> = {
  "namespace-seed": "Setup",
  "recommendations-playground": "Recommendations",
  "bandit-playground": "Bandit",
  "user-session": "User Session",
  "data-management": "Data",
  rules: "Rules",
  documentation: "Docs",
  "explain-llm": "Explain",
  "privacy-policy": "Privacy",
};

export function Navigation({
  activeView,
  onViewChange,
  swaggerUrl,
  customChatGptUrl,
  namespace,
}: NavigationProps) {
  const toast = useToast();
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  // Helper function to generate URL for a view
  const getViewUrl = (view: ViewType) => {
    const url = new URL(window.location.href);
    url.searchParams.set("view", view);
    url.searchParams.set("namespace", namespace);
    return url.toString();
  };

  // Handle middle-click to open in new tab
  const handleMouseDown = (view: ViewType, event: React.MouseEvent) => {
    if (event.button === 1) {
      // Middle mouse button
      event.preventDefault();
      window.open(getViewUrl(view), "_blank", "noopener,noreferrer");
    }
  };

  const copyLink = async () => {
    try {
      await window.navigator.clipboard.writeText(getViewUrl(activeView));
      toast.showSuccess("Shareable link", "Link copied");
    } catch {
      toast.showError("Failed to copy to clipboard");
    }
  };

  return (
    <nav
      style={{
        position: "sticky",
        top: 0,
        zIndex: 100,
        backgroundColor: "#ffffff",
        borderBottom: `1px solid ${color.border}`,
        marginBottom: spacing.lg,
      }}
    >
      {/* Mobile header */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          padding: `${spacing.md}px ${spacing.lg}px`,
          minHeight: 60,
        }}
      >
        <h1
          style={{
            margin: 0,
            fontSize: text.lg,
            fontWeight: 600,
            color: color.primary,
            cursor: "pointer",
          }}
          onClick={() => onViewChange("namespace-seed")}
          onMouseDown={(e) => handleMouseDown("namespace-seed", e)}
          title="RecSys"
        >
          RecSys
        </h1>

        {/* Mobile menu button - hidden on desktop */}
        <button
          type="button"
          onClick={() => setIsMenuOpen(!isMenuOpen)}
          style={{
            display: "flex",
            flexDirection: "column",
            gap: 4,
            padding: spacing.sm,
            background: "none",
            border: "none",
            cursor: "pointer",
          }}
          className="mobile-menu-button"
          aria-label="Toggle navigation menu"
        >
          <div
            style={{
              width: 20,
              height: 2,
              backgroundColor: color.text,
              transition: "transform 0.2s ease",
              transform: isMenuOpen
                ? "rotate(45deg) translate(6px, 6px)"
                : "none",
            }}
          />
          <div
            style={{
              width: 20,
              height: 2,
              backgroundColor: color.text,
              transition: "opacity 0.2s ease",
              opacity: isMenuOpen ? 0 : 1,
            }}
          />
          <div
            style={{
              width: 20,
              height: 2,
              backgroundColor: color.text,
              transition: "transform 0.2s ease",
              transform: isMenuOpen
                ? "rotate(-45deg) translate(6px, -6px)"
                : "none",
            }}
          />
        </button>
      </div>

      {/* Mobile menu */}
      {isMenuOpen && (
        <div
          style={{
            position: "absolute",
            top: "100%",
            left: 0,
            right: 0,
            backgroundColor: "#ffffff",
            borderBottom: `1px solid ${color.border}`,
            boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.1)",
            zIndex: 1000,
          }}
        >
          <div style={{ padding: spacing.md }}>
            {/* View selector */}
            <div style={{ marginBottom: spacing.lg }}>
              <label
                style={{
                  display: "block",
                  fontSize: text.sm,
                  fontWeight: 600,
                  color: color.textMuted,
                  marginBottom: spacing.sm,
                }}
              >
                Current View
              </label>
              <select
                value={activeView}
                onChange={(e) => {
                  onViewChange(e.target.value as ViewType);
                  setIsMenuOpen(false);
                }}
                style={{
                  width: "100%",
                  padding: `${spacing.md}px`,
                  border: `1px solid ${color.border}`,
                  borderRadius: radius.md,
                  fontSize: text.md,
                  backgroundColor: "#ffffff",
                }}
              >
                {Object.entries(viewLabels).map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </div>

            {/* Action buttons */}
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: spacing.sm,
              }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "center",
                  marginBottom: spacing.sm,
                }}
              >
                <ThemeToggleFull />
              </div>

              <Button
                type="button"
                onClick={copyLink}
                style={{
                  width: "100%",
                  padding: `${spacing.md}px`,
                  fontSize: text.md,
                  backgroundColor: color.panelSubtle,
                  color: color.text,
                  border: `1px solid ${color.border}`,
                  borderRadius: radius.md,
                }}
              >
                📋 Copy Link
              </Button>

              {customChatGptUrl && (
                <Button
                  type="button"
                  onClick={() => {
                    window.open(
                      customChatGptUrl,
                      "_blank",
                      "noopener,noreferrer"
                    );
                    setIsMenuOpen(false);
                  }}
                  style={{
                    width: "100%",
                    padding: `${spacing.md}px`,
                    fontSize: text.md,
                    backgroundColor: color.panelSubtle,
                    color: color.text,
                    border: `1px solid ${color.border}`,
                    borderRadius: radius.md,
                  }}
                >
                  🤖 Ask ChatGPT
                </Button>
              )}

              <Button
                type="button"
                onClick={() => {
                  window.open(swaggerUrl, "_blank", "noopener,noreferrer");
                  setIsMenuOpen(false);
                }}
                style={{
                  width: "100%",
                  padding: `${spacing.md}px`,
                  fontSize: text.md,
                  backgroundColor: color.panelSubtle,
                  color: color.text,
                  border: `1px solid ${color.border}`,
                  borderRadius: radius.md,
                }}
              >
                📚 Explore API
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Desktop navigation */}
      <div
        style={{
          display: "none",
          padding: `${spacing.lg}px`,
          borderTop: `1px solid ${color.border}`,
        }}
        className="desktop-nav"
      >
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            maxWidth: 1200,
            margin: "0 auto",
          }}
        >
          {/* Desktop view tabs */}
          <div
            style={{
              display: "flex",
              gap: spacing.sm,
              flexWrap: "wrap",
            }}
          >
            {Object.entries(viewLabels).map(([value, label]) => (
              <Button
                key={value}
                type="button"
                onClick={() => onViewChange(value as ViewType)}
                onMouseDown={(e) => handleMouseDown(value as ViewType, e)}
                style={{
                  padding: `${spacing.sm}px ${spacing.lg}px`,
                  fontSize: text.sm,
                  backgroundColor:
                    activeView === value ? color.primary : "transparent",
                  color:
                    activeView === value
                      ? color.primaryTextOn
                      : color.textMuted,
                  border: `1px solid ${
                    activeView === value ? color.primary : color.border
                  }`,
                  borderRadius: radius.md,
                  fontWeight: activeView === value ? 600 : 400,
                  transition: "all 0.2s ease",
                }}
                title={`${label} (middle-click to open in new tab)`}
              >
                {label}
              </Button>
            ))}
          </div>

          {/* Desktop action buttons */}
          <div
            style={{ display: "flex", gap: spacing.sm, alignItems: "center" }}
          >
            <ThemeToggleCompact />

            <Button
              type="button"
              onClick={copyLink}
              style={{
                padding: `${spacing.sm}px ${spacing.md}px`,
                fontSize: text.sm,
                backgroundColor: color.panelSubtle,
                color: color.textMuted,
                border: `1px solid ${color.border}`,
                borderRadius: radius.sm,
              }}
              title="Copy shareable link"
            >
              Copy link
            </Button>

            {customChatGptUrl && (
              <Button
                type="button"
                onClick={() => {
                  window.open(
                    customChatGptUrl,
                    "_blank",
                    "noopener,noreferrer"
                  );
                }}
                style={{
                  padding: `${spacing.sm}px ${spacing.md}px`,
                  fontSize: text.sm,
                  backgroundColor: color.panelSubtle,
                  color: color.textMuted,
                  border: `1px solid ${color.border}`,
                  borderRadius: radius.sm,
                }}
                title="Ask ChatGPT"
              >
                Ask ChatGPT
              </Button>
            )}

            <Button
              type="button"
              onClick={() => {
                window.open(swaggerUrl, "_blank", "noopener,noreferrer");
              }}
              style={{
                padding: `${spacing.sm}px ${spacing.md}px`,
                fontSize: text.sm,
                backgroundColor: color.panelSubtle,
                color: color.textMuted,
                border: `1px solid ${color.border}`,
                borderRadius: radius.sm,
              }}
              title="Explore API"
            >
              Explore API
            </Button>
          </div>
        </div>
      </div>
    </nav>
  );
}
