import { useState, useEffect } from "react";
import { Button } from "./UIComponents";

interface DocumentationSectionProps {
  className?: string;
}

interface ReadmeContent {
  title: string;
  content: string;
  source: string;
}

export function DocumentationSection({ className }: DocumentationSectionProps) {
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
      // Code blocks first (to avoid conflicts with other patterns)
      .replace(
        /```([\s\S]*?)```/g,
        '<pre style="background: #f5f5f5; padding: 12px; border-radius: 4px; overflow-x: auto; margin: 8px 0;"><code>$1</code></pre>'
      )
      // Headers
      .replace(
        /^### (.*$)/gim,
        '<h3 style="margin: 16px 0 8px 0; color: #1976d2;">$1</h3>'
      )
      .replace(
        /^## (.*$)/gim,
        '<h2 style="margin: 20px 0 12px 0; color: #1976d2; border-bottom: 1px solid #e0e0e0; padding-bottom: 4px;">$1</h2>'
      )
      .replace(
        /^# (.*$)/gim,
        '<h1 style="margin: 24px 0 16px 0; color: #1976d2;">$1</h1>'
      )
      // Lists
      .replace(/^- (.*$)/gim, '<li style="margin: 4px 0;">$1</li>')
      .replace(
        /(<li.*<\/li>)/s,
        '<ul style="margin: 8px 0; padding-left: 20px;">$1</ul>'
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
      // Links
      .replace(
        /\[([^\]]+)\]\(([^)]+)\)/g,
        '<a href="$2" target="_blank" rel="noopener noreferrer" style="color: #1976d2; text-decoration: none;">$1</a>'
      )
      // Line breaks
      .replace(/\n\n/g, '</p><p style="margin: 8px 0;">')
      .replace(/\n/g, "<br>")
      // Wrap in paragraph
      .replace(/^(.*)$/, '<p style="margin: 8px 0;">$1</p>');

    return html;
  };

  return (
    <div className={className} style={{ marginTop: 24 }}>
      <div
        style={{
          border: "1px solid #e0e0e0",
          borderRadius: 8,
          overflow: "hidden",
          backgroundColor: "#fafafa",
        }}
      >
        {/* Header */}
        <div
          style={{
            padding: "12px 16px",
            backgroundColor: "#f5f5f5",
            borderBottom: "1px solid #e0e0e0",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <h2 style={{ margin: 0, fontSize: 18, fontWeight: 600 }}>
            Documentation
          </h2>
          <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
            <div style={{ display: "flex", gap: 4 }}>
              <Button
                type="button"
                onClick={() => setActiveTab("root")}
                style={{
                  padding: "4px 12px",
                  fontSize: 12,
                  backgroundColor: activeTab === "root" ? "#1976d2" : "#fff",
                  color: activeTab === "root" ? "#fff" : "#666",
                  border: "1px solid #ddd",
                  borderRadius: 4,
                  cursor: "pointer",
                }}
              >
                Overview
              </Button>
              <Button
                type="button"
                onClick={() => setActiveTab("demo-ui")}
                style={{
                  padding: "4px 12px",
                  fontSize: 12,
                  backgroundColor: activeTab === "demo-ui" ? "#1976d2" : "#fff",
                  color: activeTab === "demo-ui" ? "#fff" : "#666",
                  border: "1px solid #ddd",
                  borderRadius: 4,
                  cursor: "pointer",
                }}
              >
                Demo UI
              </Button>
              <Button
                type="button"
                onClick={() => setActiveTab("api")}
                style={{
                  padding: "4px 12px",
                  fontSize: 12,
                  backgroundColor: activeTab === "api" ? "#1976d2" : "#fff",
                  color: activeTab === "api" ? "#fff" : "#666",
                  border: "1px solid #ddd",
                  borderRadius: 4,
                  cursor: "pointer",
                }}
              >
                API
              </Button>
            </div>
            <div
              style={{
                marginLeft: 8,
                paddingLeft: 8,
                borderLeft: "1px solid #ddd",
              }}
            >
              <a
                href={`/readme-${activeTab}.md`}
                target="_blank"
                rel="noopener noreferrer"
                style={{
                  display: "inline-flex",
                  alignItems: "center",
                  gap: 2,
                  padding: "4px 8px",
                  fontSize: 10,
                  color: "#1976d2",
                  textDecoration: "none",
                  border: "1px solid #1976d2",
                  borderRadius: 3,
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
                📄 Raw
              </a>
            </div>
          </div>
        </div>

        {/* Content */}
        <div style={{ padding: 16, maxHeight: 600, overflowY: "auto" }}>
          {loading && (
            <div style={{ textAlign: "center", color: "#666", padding: 20 }}>
              Loading documentation...
            </div>
          )}

          {error && (
            <div
              style={{
                color: "#d32f2f",
                backgroundColor: "#ffebee",
                padding: 12,
                borderRadius: 4,
                border: "1px solid #ffcdd2",
              }}
            >
              <strong>Error:</strong> {error}
            </div>
          )}

          {readmeContent && !loading && !error && (
            <div>
              <div
                style={{
                  marginBottom: 16,
                  paddingBottom: 8,
                  borderBottom: "1px solid #e0e0e0",
                }}
              >
                <h3 style={{ margin: 0, color: "#1976d2" }}>
                  {readmeContent.title}
                </h3>
                <small style={{ color: "#666" }}>
                  Source: {readmeContent.source}/README.md
                </small>
              </div>
              <div
                style={{
                  lineHeight: 1.6,
                  color: "#333",
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
