// Design tokens for the demo UI. Now uses CSS custom properties for theme support.

export const color = {
  // Text colors - now using CSS custom properties
  text: "var(--color-text)",
  textMuted: "var(--color-text-muted)",
  textSubtle: "var(--color-text-subtle)",

  // Border colors
  border: "var(--color-border)",
  panelBorder: "var(--color-panel-border)",

  // Background colors
  background: "var(--color-background)",
  panelBg: "var(--color-panel-bg)",
  panelSubtle: "var(--color-panel-subtle)",

  // Primary colors
  primary: "var(--color-primary)",
  primaryTextOn: "var(--color-primary-text-on)",
  primaryHover: "var(--color-primary-hover)",

  // Button colors
  buttonBg: "var(--color-button-bg)",
  buttonBorder: "var(--color-button-border)",
  buttonHover: "var(--color-button-hover)",

  // Code colors
  codeBg: "var(--color-code-bg)",
  codeBorder: "var(--color-code-border)",

  // Status colors
  warningBg: "var(--color-warning-bg)",
  warning: "var(--color-warning)",
  dangerBg: "var(--color-danger-bg)",
  danger: "var(--color-danger)",
  successBg: "var(--color-success-bg)",
  success: "var(--color-success)",

  // Focus and interaction colors
  focus: "var(--color-focus)",
  focusRing: "var(--color-focus-ring)",
};

export const spacing = {
  xxs: 2,
  xs: 4,
  sm: 6,
  md: 8,
  lg: 12,
  xl: 16,
  xxl: 24,
};

export const radius = {
  sm: 4,
  md: 6,
  lg: 8,
  pill: 999,
};

export const text = {
  xs: 10,
  sm: 12,
  md: 14,
  lg: 16,
  xl: 24,
};

export const layout = {
  maxWidth: 1200,
};
