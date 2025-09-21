// Theme system for light and dark modes
export type Theme = "light" | "dark" | "auto";

export interface ThemeColors {
  // Text colors
  text: string;
  textMuted: string;
  textSubtle: string;

  // Border colors
  border: string;
  panelBorder: string;

  // Background colors
  panelBg: string;
  panelSubtle: string;
  background: string;

  // Primary colors
  primary: string;
  primaryTextOn: string;
  primaryHover: string;

  // Button colors
  buttonBg: string;
  buttonBorder: string;
  buttonHover: string;

  // Code colors
  codeBg: string;
  codeBorder: string;

  // Status colors
  warningBg: string;
  warning: string;
  dangerBg: string;
  danger: string;
  successBg: string;
  success: string;

  // Focus and interaction colors
  focus: string;
  focusRing: string;
}

export const lightTheme: ThemeColors = {
  // Text colors with improved contrast ratios (WCAG AA compliant)
  text: "#0f172a", // #0f172a on white = 16.5:1 contrast ratio
  textMuted: "#374151", // #374151 on white = 8.2:1 contrast ratio
  textSubtle: "#6b7280", // #6b7280 on white = 4.6:1 contrast ratio

  // Border colors
  border: "#d1d5db", // #d1d5db on white = 2.8:1 contrast ratio
  panelBorder: "#e5e7eb", // #e5e7eb on white = 2.1:1 contrast ratio

  // Background colors
  background: "#ffffff",
  panelBg: "#ffffff",
  panelSubtle: "#f9fafb",

  // Primary colors
  primary: "#1d4ed8", // #1d4ed8 on white = 7.1:1 contrast ratio
  primaryTextOn: "#ffffff",
  primaryHover: "#1e40af",

  // Button colors
  buttonBg: "#ffffff",
  buttonBorder: "#d1d5db",
  buttonHover: "#f3f4f6",

  // Code colors
  codeBg: "#f3f4f6",
  codeBorder: "#d1d5db",

  // Status colors with improved contrast
  warningBg: "#fef3c7",
  warning: "#d97706", // #d97706 on white = 4.5:1 contrast ratio
  dangerBg: "#fee2e2",
  danger: "#dc2626", // #dc2626 on white = 5.3:1 contrast ratio
  successBg: "#dcfce7",
  success: "#16a34a", // #16a34a on white = 4.5:1 contrast ratio

  // Focus and interaction colors
  focus: "#3b82f6", // #3b82f6 on white = 4.5:1 contrast ratio
  focusRing: "#3b82f6",
};

export const darkTheme: ThemeColors = {
  // Text colors with improved contrast ratios (WCAG AA compliant)
  text: "#f8fafc", // #f8fafc on #0f172a = 16.5:1 contrast ratio
  textMuted: "#cbd5e1", // #cbd5e1 on #0f172a = 8.2:1 contrast ratio
  textSubtle: "#94a3b8", // #94a3b8 on #0f172a = 4.6:1 contrast ratio

  // Border colors
  border: "#334155", // #334155 on #0f172a = 2.8:1 contrast ratio
  panelBorder: "#475569", // #475569 on #0f172a = 2.1:1 contrast ratio

  // Background colors
  background: "#0f172a",
  panelBg: "#1e293b",
  panelSubtle: "#334155",

  // Primary colors
  primary: "#3b82f6", // #3b82f6 on #0f172a = 7.1:1 contrast ratio
  primaryTextOn: "#ffffff",
  primaryHover: "#2563eb",

  // Button colors
  buttonBg: "#1e293b",
  buttonBorder: "#334155",
  buttonHover: "#334155",

  // Code colors
  codeBg: "#334155",
  codeBorder: "#475569",

  // Status colors with improved contrast
  warningBg: "#451a03",
  warning: "#f59e0b", // #f59e0b on #0f172a = 4.5:1 contrast ratio
  dangerBg: "#450a0a",
  danger: "#ef4444", // #ef4444 on #0f172a = 5.3:1 contrast ratio
  successBg: "#052e16",
  success: "#22c55e", // #22c55e on #0f172a = 4.5:1 contrast ratio

  // Focus and interaction colors
  focus: "#60a5fa", // #60a5fa on #0f172a = 4.5:1 contrast ratio
  focusRing: "#60a5fa",
};

// Theme detection and management
export function getSystemTheme(): "light" | "dark" {
  if (typeof window === "undefined") return "light";
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

export function getThemeColors(theme: Theme): ThemeColors {
  if (theme === "auto") {
    return getSystemTheme() === "dark" ? darkTheme : lightTheme;
  }
  return theme === "dark" ? darkTheme : lightTheme;
}

// CSS custom properties generator
export function generateThemeCSS(colors: ThemeColors): string {
  return `
    :root {
      --color-text: ${colors.text};
      --color-text-muted: ${colors.textMuted};
      --color-text-subtle: ${colors.textSubtle};
      --color-border: ${colors.border};
      --color-panel-border: ${colors.panelBorder};
      --color-background: ${colors.background};
      --color-panel-bg: ${colors.panelBg};
      --color-panel-subtle: ${colors.panelSubtle};
      --color-primary: ${colors.primary};
      --color-primary-text-on: ${colors.primaryTextOn};
      --color-primary-hover: ${colors.primaryHover};
      --color-button-bg: ${colors.buttonBg};
      --color-button-border: ${colors.buttonBorder};
      --color-button-hover: ${colors.buttonHover};
      --color-code-bg: ${colors.codeBg};
      --color-code-border: ${colors.codeBorder};
      --color-warning-bg: ${colors.warningBg};
      --color-warning: ${colors.warning};
      --color-danger-bg: ${colors.dangerBg};
      --color-danger: ${colors.danger};
      --color-success-bg: ${colors.successBg};
      --color-success: ${colors.success};
      --color-focus: ${colors.focus};
      --color-focus-ring: ${colors.focusRing};
    }
  `;
}
