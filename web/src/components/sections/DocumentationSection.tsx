import { useState, useEffect } from "react";
import { Button } from "../primitives/UIComponents";
import { SafeMarkdown } from "../SafeMarkdown";

interface DocumentationSectionProps {
  className?: string;
}

interface ReadmeContent {
  title: string;
  content: string;
  source: string;
}

export function DocumentationSection({ className }: DocumentationSectionProps) {
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
                onClick={() => setActiveTab("web")}
                style={{
                  padding: "4px 12px",
                  fontSize: 12,
                  backgroundColor: activeTab === "web" ? "#1976d2" : "#fff",
                  color: activeTab === "web" ? "#fff" : "#666",
                  border: "1px solid #ddd",
                  borderRadius: 4,
                  cursor: "pointer",
                }}
              >
                Web UI
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
