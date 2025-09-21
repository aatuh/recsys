import React from "react";
import { color, layout, spacing } from "./tokens";

export function AppShell(props: {
  header: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <div
      style={{ minHeight: "100vh", display: "flex", flexDirection: "column" }}
    >
      {props.header}
      <main id="main-content" style={{ flex: 1 }}>
        <div
          style={{
            margin: "0 auto",
            maxWidth: layout.maxWidth,
            padding: `${spacing.lg}px ${spacing.md}px`,
          }}
        >
          {props.children}
        </div>
      </main>
    </div>
  );
}

export class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  { hasError: boolean; error?: any }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props);
    this.state = { hasError: false };
  }
  static getDerivedStateFromError(error: any) {
    return { hasError: true, error };
  }
  componentDidCatch() {}
  render() {
    if (this.state.hasError) {
      return (
        <div
          style={{
            margin: "0 auto",
            maxWidth: layout.maxWidth,
            padding: spacing.xl,
            color: color.text,
          }}
        >
          <h2>Something went wrong.</h2>
          <pre
            style={{
              background: "#f6f8fa",
              border: `1px solid ${color.border}`,
              padding: spacing.lg,
              borderRadius: 6,
              overflowX: "auto",
            }}
          >
            {String(this.state.error)}
          </pre>
        </div>
      );
    }
    return this.props.children as any;
  }
}
