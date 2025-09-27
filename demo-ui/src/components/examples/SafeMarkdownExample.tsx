/**
 * Example component demonstrating safe markdown rendering and URL validation.
 */

import React, { useState } from "react";
import { SafeMarkdown } from "../SafeMarkdown";
import {
  useSafeQueryParam,
  useSafeQueryParams,
} from "../../hooks/useSafeQuerySync";
import {
  AppQuerySchemas,
  parseSearchParams,
  sanitizeQueryParam,
} from "../../utils/urlValidation";
import { Button } from "../primitives/UIComponents";
import { z } from "zod";

export function SafeMarkdownExample() {
  const [logs, setLogs] = useState<string[]>([]);
  const [markdownContent, setMarkdownContent] =
    useState(`# Safe Markdown Example

This demonstrates **safe markdown rendering** with sanitization.

## Features

- **Bold text** and *italic text*
- \`inline code\` and code blocks
- [Links](https://example.com) with security
- Lists and tables
- HTML sanitization

### Code Block Example

\`\`\`typescript
const safeMarkdown = <SafeMarkdown content={content} />;
\`\`\`

### Security Features

- XSS protection via DOMPurify
- Configurable HTML tags
- Link validation
- Script removal

> This is a blockquote with security considerations.

---

**Note**: All content is sanitized before rendering.`);

  const [allowHtml, setAllowHtml] = useState(false);
  const [allowLinks, setAllowLinks] = useState(true);
  const [allowImages, setAllowImages] = useState(true);
  const [allowCode, setAllowCode] = useState(true);

  // Example of safe query parameter usage
  const [view, setView] = useSafeQueryParam(
    "demo-view",
    AppQuerySchemas.view,
    "recommendations" as const,
    { storageKey: "demo-view", persist: true }
  );

  const [page, setPage] = useSafeQueryParam(
    "demo-page",
    AppQuerySchemas.page,
    1,
    { storageKey: "demo-page", persist: true }
  );

  // Example of multiple parameter validation
  const [filters, setFilters] = useSafeQueryParams(
    z.object({
      search: AppQuerySchemas.search,
      userId: AppQuerySchemas.userId,
      enabled: AppQuerySchemas.enabled,
    }),
    { search: undefined, userId: undefined, enabled: undefined },
    { storageKey: "demo-filters", persist: true }
  );

  const addLog = (message: string) => {
    setLogs((prev) => [...prev, `${new Date().toISOString()}: ${message}`]);
  };

  const handleTestUrlValidation = () => {
    // Test URL validation
    const testUrl =
      "https://example.com?view=recommendations&page=2&search=test&userId=123&enabled=true";

    try {
      const result = parseSearchParams(
        testUrl,
        z.object({
          view: AppQuerySchemas.view,
          page: AppQuerySchemas.page,
          search: AppQuerySchemas.search,
          userId: AppQuerySchemas.userId,
          enabled: AppQuerySchemas.enabled,
        })
      );

      if (result.success) {
        addLog(`✅ URL validation successful: ${JSON.stringify(result.data)}`);
      } else {
        addLog(`❌ URL validation failed: ${JSON.stringify(result.errors)}`);
      }
    } catch (error) {
      addLog(
        `❌ URL validation error: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  };

  const handleTestSanitization = () => {
    const maliciousInput = `<script>alert('XSS')</script><img src="x" onerror="alert('XSS')"><a href="javascript:alert('XSS')">Click me</a>`;
    const sanitized = sanitizeQueryParam(maliciousInput);

    addLog(`🧹 Sanitized input: "${sanitized}"`);
  };

  const handleTestMarkdownSecurity = () => {
    const maliciousMarkdown = `# Malicious Content

<script>alert('XSS')</script>

[Click me](javascript:alert('XSS'))

<img src="x" onerror="alert('XSS')">

\`\`\`html
<script>alert('XSS')</script>
\`\`\``;

    setMarkdownContent(maliciousMarkdown);
    addLog("🔒 Testing markdown security with malicious content");
  };

  const handleResetContent = () => {
    setMarkdownContent(`# Safe Markdown Example

This demonstrates **safe markdown rendering** with sanitization.

## Features

- **Bold text** and *italic text*
- \`inline code\` and code blocks
- [Links](https://example.com) with security
- Lists and tables
- HTML sanitization

### Code Block Example

\`\`\`typescript
const safeMarkdown = <SafeMarkdown content={content} />;
\`\`\`

### Security Features

- XSS protection via DOMPurify
- Configurable HTML tags
- Link validation
- Script removal

> This is a blockquote with security considerations.

---

**Note**: All content is sanitized before rendering.`);
    addLog("🔄 Reset to safe content");
  };

  return (
    <div style={{ padding: "20px", border: "1px solid #ccc", margin: "10px" }}>
      <h3>Safe Markdown & URL Validation Example</h3>
      <p style={{ fontSize: "14px", color: "#666", marginBottom: "20px" }}>
        Demonstrates secure markdown rendering and URL parameter validation.
      </p>

      <div style={{ marginBottom: "20px" }}>
        <h4>Markdown Security Controls</h4>
        <div
          style={{
            display: "flex",
            gap: "10px",
            marginBottom: "10px",
            flexWrap: "wrap",
          }}
        >
          <label>
            <input
              type="checkbox"
              checked={allowHtml}
              onChange={(e) => setAllowHtml(e.target.checked)}
            />
            Allow HTML
          </label>
          <label>
            <input
              type="checkbox"
              checked={allowLinks}
              onChange={(e) => setAllowLinks(e.target.checked)}
            />
            Allow Links
          </label>
          <label>
            <input
              type="checkbox"
              checked={allowImages}
              onChange={(e) => setAllowImages(e.target.checked)}
            />
            Allow Images
          </label>
          <label>
            <input
              type="checkbox"
              checked={allowCode}
              onChange={(e) => setAllowCode(e.target.checked)}
            />
            Allow Code
          </label>
        </div>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Safe Markdown Rendering</h4>
        <div
          style={{
            border: "1px solid #ddd",
            padding: "15px",
            borderRadius: "4px",
            backgroundColor: "#f9f9f9",
          }}
        >
          <SafeMarkdown
            content={markdownContent}
            allowHtml={allowHtml}
            allowLinks={allowLinks}
            allowImages={allowImages}
            allowCode={allowCode}
          />
        </div>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>URL Parameter Validation</h4>
        <div style={{ display: "flex", gap: "10px", marginBottom: "10px" }}>
          <Button onClick={handleTestUrlValidation}>Test URL Validation</Button>
          <Button onClick={handleTestSanitization}>Test Sanitization</Button>
          <Button onClick={handleTestMarkdownSecurity}>
            Test Markdown Security
          </Button>
          <Button onClick={handleResetContent}>Reset Content</Button>
        </div>

        <div style={{ fontSize: "12px", color: "#666", marginBottom: "10px" }}>
          Current view: {view} | Page: {page} | Search:{" "}
          {filters.search || "none"} | User ID: {filters.userId || "none"} |
          Enabled: {filters.enabled ? "true" : "false"}
        </div>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Query Parameter Controls</h4>
        <div
          style={{
            display: "flex",
            gap: "10px",
            marginBottom: "10px",
            flexWrap: "wrap",
          }}
        >
          <select
            value={view}
            onChange={(e) => setView(e.target.value as any)}
            style={{ padding: "5px" }}
          >
            <option value="recommendations">Recommendations</option>
            <option value="bandit">Bandit</option>
            <option value="data-management">Data Management</option>
            <option value="rules">Rules</option>
            <option value="documentation">Documentation</option>
            <option value="explain-llm">Explain LLM</option>
            <option value="privacy-policy">Privacy Policy</option>
          </select>

          <input
            type="number"
            value={page}
            onChange={(e) => setPage(parseInt(e.target.value) || 1)}
            placeholder="Page"
            style={{ padding: "5px", width: "80px" }}
          />

          <input
            type="text"
            value={filters.search || ""}
            onChange={(e) =>
              setFilters({ ...filters, search: e.target.value || undefined })
            }
            placeholder="Search"
            style={{ padding: "5px", width: "120px" }}
          />

          <input
            type="text"
            value={filters.userId || ""}
            onChange={(e) =>
              setFilters({ ...filters, userId: e.target.value || undefined })
            }
            placeholder="User ID"
            style={{ padding: "5px", width: "120px" }}
          />

          <label>
            <input
              type="checkbox"
              checked={filters.enabled || false}
              onChange={(e) =>
                setFilters({ ...filters, enabled: e.target.checked })
              }
            />
            Enabled
          </label>
        </div>
      </div>

      <div>
        <h4>Activity Logs</h4>
        <div style={{ display: "flex", gap: "10px", marginBottom: "10px" }}>
          <Button
            onClick={() => setLogs([])}
            style={{ backgroundColor: "#6c757d", color: "white" }}
          >
            Clear Logs
          </Button>
        </div>

        <div
          style={{
            maxHeight: "200px",
            overflowY: "auto",
            border: "1px solid #ddd",
            padding: "10px",
            backgroundColor: "#f8f9fa",
          }}
        >
          {logs.length === 0 ? (
            <div style={{ color: "#666", fontStyle: "italic" }}>
              No logs yet
            </div>
          ) : (
            logs.map((log, index) => (
              <div
                key={index}
                style={{
                  fontSize: "12px",
                  marginBottom: "4px",
                  fontFamily: "monospace",
                }}
              >
                {log}
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
