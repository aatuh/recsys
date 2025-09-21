import { useState, useEffect } from "react";
import { Button } from "../primitives/UIComponents";
import { color, radius, spacing, text } from "../../ui/tokens";

interface ReadmeContent {
  title: string;
  content: string;
  source: string;
}

export function DocumentationView() {
  const [activeTab, setActiveTab] = useState<"root" | "demo-ui" | "api">(
    "root"
  );
  const [readmeContent, setReadmeContent] = useState<ReadmeContent | null>(
    null
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchReadme = async (source: "root" | "demo-ui" | "api") => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(`/readme-${source}.md`);
      if (!response.ok) {
        throw new Error(
          `Failed to fetch ${source} README: ${response.statusText}`
        );
      }
      const content = await response.text();

      const titles = {
        root: "Project Overview",
        "demo-ui": "Demo UI Documentation",
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

  const renderMarkdown = (content: string) => {
    // Enhanced markdown rendering - convert markdown to HTML
    const html = content
      // Horizontal rules (---) - must be processed first
      .replace(
        /^---$/gm,
        '<hr style="border: none; border-top: 1px solid #e0e0e0; margin: 16px 0;">'
      )
      // Code blocks first (to avoid conflicts with other patterns)
      .replace(
        /```([\s\S]*?)```/g,
        '<pre style="background: #f5f5f5; padding: 12px; border-radius: 4px; overflow-x: auto; margin: 8px 0;"><code>$1</code></pre>'
      )
      // Headers - handle all levels including #### and deeper
      .replace(
        /^#### (.*$)/gim,
        '<h4 style="margin: 12px 0 6px 0; color: #1976d2; font-size: 16px;">$1</h4>'
      )
      .replace(
        /^### (.*$)/gim,
        '<h3 style="margin: 16px 0 8px 0; color: #1976d2; font-size: 18px;">$1</h3>'
      )
      .replace(
        /^## (.*$)/gim,
        '<h2 style="margin: 20px 0 12px 0; color: #1976d2; border-bottom: 1px solid #e0e0e0; padding-bottom: 4px; font-size: 20px;">$1</h2>'
      )
      .replace(
        /^# (.*$)/gim,
        '<h1 style="margin: 24px 0 16px 0; color: #1976d2; font-size: 24px;">$1</h1>'
      )
      // Handle deeper header levels (####, #####, ######)
      .replace(
        /^##### (.*$)/gim,
        '<h5 style="margin: 10px 0 5px 0; color: #1976d2; font-size: 14px;">$1</h5>'
      )
      .replace(
        /^###### (.*$)/gim,
        '<h6 style="margin: 8px 0 4px 0; color: #1976d2; font-size: 13px;">$1</h6>'
      )
      // Lists - improved handling to group consecutive list items
      .replace(/^- (.*$)/gim, '<li style="margin: 2px 0;">$1</li>')
      .replace(
        /(<li[^>]*>.*<\/li>)(\s*<li[^>]*>.*<\/li>)*/gs,
        '<ul style="margin: 8px 0; padding-left: 20px;">$&</ul>'
      )
      // Bold
      .replace(
        /\*\*(.*?)\*\*/g,
        '<strong style="font-weight: 600;">$1</strong>'
      )
      // Italic
      .replace(/\*(.*?)\*/g, "<em>$1</em>")
      // Inline code
      .replace(
        /`(.*?)`/g,
        '<code style="background: #f0f0f0; padding: 2px 4px; border-radius: 3px; font-family: monospace;">$1</code>'
      )
      // Tables - must be processed before line breaks
      .replace(
        /^(\|.*\|)\n(\|[\s|-]+\|)\n((?:\|.*\|\n?)*)/gm,
        (match, header, separator, rows) => {
          const headerCells = header
            .split("|")
            .slice(1, -1)
            .map(
              (cell: string) =>
                `<th style="padding: 8px 12px; text-align: left; border-bottom: 2px solid #e0e0e0; font-weight: 600;">${cell.trim()}</th>`
            )
            .join("");

          const rowLines = rows
            .trim()
            .split("\n")
            .filter((line: string) => line.trim());
          const tableRows = rowLines
            .map((row: string) => {
              const cells = row
                .split("|")
                .slice(1, -1)
                .map(
                  (cell: string) =>
                    `<td style="padding: 8px 12px; border-bottom: 1px solid #f0f0f0;">${cell.trim()}</td>`
                )
                .join("");
              return `<tr>${cells}</tr>`;
            })
            .join("");

          return `<table style="border-collapse: collapse; width: 100%; margin: 16px 0; font-size: 14px;">
            <thead><tr>${headerCells}</tr></thead>
            <tbody>${tableRows}</tbody>
          </table>`;
        }
      )
      // Links
      .replace(
        /\[([^\]]+)\]\(([^)]+)\)/g,
        '<a href="$2" target="_blank" rel="noopener noreferrer" style="color: #1976d2; text-decoration: none;">$1</a>'
      )
      // Line breaks - handle multiple consecutive line breaks better
      .replace(/\n\n+/g, '</p><p style="margin: 8px 0;">')
      .replace(/\n/g, "<br>")
      // Wrap in paragraph
      .replace(/^(.*)$/, '<p style="margin: 8px 0;">$1</p>');

    return html;
  };

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
                onClick={() => setActiveTab("demo-ui")}
                style={{
                  padding: `${spacing.sm}px ${spacing.xl}px`,
                  fontSize: text.md,
                  backgroundColor:
                    activeTab === "demo-ui" ? color.primary : "#fff",
                  color: activeTab === "demo-ui" ? "#fff" : color.textMuted,
                  border: `1px solid ${color.border}`,
                  borderRadius: radius.md,
                  cursor: "pointer",
                  fontWeight: activeTab === "demo-ui" ? 600 : 400,
                }}
              >
                Demo UI
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
                dangerouslySetInnerHTML={{
                  __html: renderMarkdown(readmeContent.content),
                }}
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
