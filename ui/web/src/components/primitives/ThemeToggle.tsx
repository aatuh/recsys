import React from "react";
import { useTheme } from "../../contexts/ThemeContext";
import { color, spacing, radius, text } from "../../ui/tokens";

interface ThemeToggleProps {
  size?: "sm" | "md" | "lg";
  showLabel?: boolean;
  className?: string;
}

export function ThemeToggle({
  size = "md",
  showLabel = true,
  className,
}: ThemeToggleProps) {
  const { theme, toggleTheme, isDark } = useTheme();

  const getSizeStyles = () => {
    switch (size) {
      case "sm":
        return {
          button: { width: 32, height: 32, fontSize: text.sm },
          icon: { fontSize: 12 },
        };
      case "lg":
        return {
          button: { width: 48, height: 48, fontSize: text.lg },
          icon: { fontSize: 20 },
        };
      default:
        return {
          button: { width: 40, height: 40, fontSize: text.md },
          icon: { fontSize: 16 },
        };
    }
  };

  const getThemeIcon = () => {
    switch (theme) {
      case "light":
        return "☀️";
      case "dark":
        return "🌙";
      case "auto":
        return isDark ? "🌙" : "☀️";
      default:
        return "☀️";
    }
  };

  const getThemeLabel = () => {
    switch (theme) {
      case "light":
        return "Light";
      case "dark":
        return "Dark";
      case "auto":
        return "Auto";
      default:
        return "Light";
    }
  };

  const sizeStyles = getSizeStyles();

  return (
    <div
      className={className}
      style={{
        display: "flex",
        alignItems: "center",
        gap: spacing.sm,
      }}
    >
      <button
        onClick={toggleTheme}
        style={{
          ...sizeStyles.button,
          backgroundColor: color.buttonBg,
          border: `1px solid ${color.buttonBorder}`,
          borderRadius: radius.md,
          cursor: "pointer",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          transition: "all 0.2s ease",
          color: color.text,
          fontSize: sizeStyles.icon.fontSize,
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.backgroundColor = color.buttonHover;
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.backgroundColor = color.buttonBg;
        }}
        onFocus={(e) => {
          e.currentTarget.style.outline = `2px solid ${color.focus}`;
          e.currentTarget.style.outlineOffset = "2px";
        }}
        onBlur={(e) => {
          e.currentTarget.style.outline = "none";
        }}
        title={`Current theme: ${getThemeLabel()}. Click to cycle through themes.`}
        aria-label={`Toggle theme. Current: ${getThemeLabel()}`}
      >
        {getThemeIcon()}
      </button>

      {showLabel && (
        <span
          style={{
            fontSize: text.sm,
            color: color.textMuted,
            fontWeight: 500,
            minWidth: 40,
            textAlign: "left",
          }}
        >
          {getThemeLabel()}
        </span>
      )}
    </div>
  );
}

// Compact theme toggle for headers/navigation
export function ThemeToggleCompact() {
  return <ThemeToggle size="sm" showLabel={false} />;
}

// Full theme toggle with label
export function ThemeToggleFull() {
  return <ThemeToggle size="md" showLabel={true} />;
}
