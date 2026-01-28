import { useState, useEffect } from "react";
import { Button } from "../primitives/UIComponents";
import { SafeMarkdown } from "../SafeMarkdown";
import { color, radius, spacing, text } from "../../ui/tokens";

interface ReadmeContent {
  title: string;
  content: string;
  source: string;
}

export function DocumentationView() {
  const [activeTab, setActiveTab] = useState<"root" | "web" | "api">("root");
  const [readmeContent, setReadmeContent] = useState<ReadmeContent | null>(
    null
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchReadme = async (source: "root" | "web" | "api") => {
    setLoading(true);
    setError(null);

    try {
      // eslint-disable-next-line no-restricted-globals
      const response = await fetch(`/readme-${source}.md`);
      if (!response.ok) {
        throw new Error(
          `Failed to fetch ${source} README: ${response.statusText}`
        );
      }
      const content = await response.text();

      const titles = {
        root: "Project Overview",
        web: "Web UI Documentation",
        api: "API Documentation",
      };

      setReadmeContent({
        title: titles[source],
        content,
        source,
      });
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to load documentation"
      );
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchReadme(activeTab);
  }, [activeTab]);

  // Removed renderMarkdown function - now using SafeMarkdown component

  return (
    <div style={{ padding: spacing.xl, fontFamily: "system-ui, sans-serif" }}>
      <div
        style={{
          border: `1px solid ${color.panelBorder}`,
          borderRadius: radius.lg,
          overflow: "hidden",
          backgroundColor: color.panelSubtle,
        }}
      >
        {/* Header */}
        <div
          style={{
            padding: `${spacing.lg}px ${spacing.xl}px`,
            backgroundColor: color.panelSubtle,
            borderBottom: `1px solid ${color.border}`,
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <h2 style={{ margin: 0, fontSize: text.lg, fontWeight: 600 }}>
            Documentation
          </h2>
          <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
            <div style={{ display: "flex", gap: 4 }}>
              <Button
                type="button"
                onClick={() => setActiveTab("root")}
                style={{
                  padding: `${spacing.sm}px ${spacing.xl}px`,
                  fontSize: text.md,
                  backgroundColor:
                    activeTab === "root" ? color.primary : "#fff",
                  color: activeTab === "root" ? "#fff" : color.textMuted,
                  border: `1px solid ${color.border}`,
                  borderRadius: radius.md,
                  cursor: "pointer",
                  fontWeight: activeTab === "root" ? 600 : 400,
                }}
              >
                Overview
              </Button>
              <Button
                type="button"
                onClick={() => setActiveTab("web")}
                style={{
                  padding: `${spacing.sm}px ${spacing.xl}px`,
                  fontSize: text.md,
                  backgroundColor: activeTab === "web" ? color.primary : "#fff",
                  color: activeTab === "web" ? "#fff" : color.textMuted,
                  border: `1px solid ${color.border}`,
                  borderRadius: radius.md,
                  cursor: "pointer",
                  fontWeight: activeTab === "web" ? 600 : 400,
                }}
              >
                Web UI
              </Button>
              <Button
                type="button"
                onClick={() => setActiveTab("api")}
                style={{
                  padding: `${spacing.sm}px ${spacing.xl}px`,
                  fontSize: text.md,
                  backgroundColor: activeTab === "api" ? color.primary : "#fff",
                  color: activeTab === "api" ? "#fff" : color.textMuted,
                  border: `1px solid ${color.border}`,
                  borderRadius: radius.md,
                  cursor: "pointer",
                  fontWeight: activeTab === "api" ? 600 : 400,
                }}
              >
                API
              </Button>
            </div>
            <div
              style={{
                marginLeft: spacing.md,
                paddingLeft: spacing.md,
                borderLeft: `1px solid ${color.border}`,
              }}
            >
              <a
                href={`/readme-${activeTab}.md`}
                target="_blank"
                rel="noopener noreferrer"
                style={{
                  display: "inline-flex",
                  alignItems: "center",
                  gap: 4,
                  padding: `${spacing.sm - 2}px ${spacing.lg}px`,
                  fontSize: text.sm,
                  color: color.primary,
                  textDecoration: "none",
                  border: `1px solid ${color.primary}`,
                  borderRadius: radius.sm,
                  backgroundColor: "#fff",
                  transition: "all 0.2s ease",
                }}
                onMouseOver={(e) => {
                  e.currentTarget.style.backgroundColor = "#f5f5f5";
                }}
                onMouseOut={(e) => {
                  e.currentTarget.style.backgroundColor = "#fff";
                }}
              >
                📄 View Raw
              </a>
            </div>
          </div>
        </div>

        {/* Content */}
        <div
          style={{
            padding: spacing.xl,
            maxHeight: "calc(100vh - 200px)",
            overflowY: "auto",
          }}
        >
          {loading && (
            <div style={{ textAlign: "center", color: "#666", padding: 40 }}>
              <div style={{ fontSize: 16, marginBottom: 8 }}>⏳</div>
              Loading documentation...
            </div>
          )}

          {error && (
            <div
              style={{
                color: "#d32f2f",
                backgroundColor: "#ffebee",
                padding: spacing.lg,
                borderRadius: radius.md,
                border: "1px solid #ffcdd2",
                marginBottom: spacing.lg,
              }}
            >
              <strong>Error:</strong> {error}
            </div>
          )}

          {readmeContent && !loading && !error && (
            <div>
              <div
                style={{
                  marginBottom: spacing.xl,
                  paddingBottom: spacing.md,
                  borderBottom: `1px solid ${color.panelBorder}`,
                }}
              >
                <h3 style={{ margin: 0, color: color.primary, fontSize: 18 }}>
                  {readmeContent.title}
                </h3>
                <small style={{ color: "#666", fontSize: 12 }}>
                  Source: {readmeContent.source}/README.md
                </small>
              </div>
              <div
                style={{
                  lineHeight: 1.6,
                  color: "#333",
                  fontSize: 14,
                }}
              >
                <SafeMarkdown
                  content={readmeContent.content}
                  allowHtml={false}
                  allowLinks={true}
                  allowImages={true}
                  allowCode={true}
                />
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
