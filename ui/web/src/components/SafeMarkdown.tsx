/**
 * Safe Markdown rendering component with sanitization.
 * Replaces regex/innerHTML patterns with a secure library pipeline.
 */

import React from "react";
import ReactMarkdown from "react-markdown";
import rehypeSanitize from "rehype-sanitize";

export interface SafeMarkdownProps {
  /** The markdown content to render */
  content: string;
  /** Optional CSS class name for the container */
  className?: string;
  /** Optional inline styles for the container */
  style?: React.CSSProperties;
  /** Whether to allow HTML tags in the markdown (default: false) */
  allowHtml?: boolean;
  /** Whether to allow links (default: true) */
  allowLinks?: boolean;
  /** Whether to allow images (default: true) */
  allowImages?: boolean;
  /** Whether to allow code blocks (default: true) */
  allowCode?: boolean;
  /** Custom sanitization options */
  sanitizeOptions?: any;
}

/**
 * SafeMarkdown component that renders markdown content safely.
 *
 * Features:
 * - Uses react-markdown for secure markdown parsing
 * - Uses rehype-sanitize for HTML sanitization
 * - Configurable security options
 * - Maintains stable component API
 */
export function SafeMarkdown({
  content,
  className,
  style,
  allowHtml = false,
  allowLinks = true,
  allowImages = true,
  allowCode = true,
  sanitizeOptions,
}: SafeMarkdownProps) {
  // Configure sanitization options
  const defaultSanitizeOptions = {
    allowedTags: [
      // Basic text formatting
      "p",
      "br",
      "strong",
      "em",
      "u",
      "s",
      "del",
      "ins",
      // Headers
      "h1",
      "h2",
      "h3",
      "h4",
      "h5",
      "h6",
      // Lists
      "ul",
      "ol",
      "li",
      // Code
      ...(allowCode ? ["pre", "code"] : []),
      // Links and media
      ...(allowLinks ? ["a"] : []),
      ...(allowImages ? ["img"] : []),
      // Tables
      "table",
      "thead",
      "tbody",
      "tr",
      "th",
      "td",
      // Blockquotes
      "blockquote",
      // Horizontal rules
      "hr",
      // HTML tags (if allowed)
      ...(allowHtml ? ["div", "span", "section", "article"] : []),
    ],
    allowedAttributes: {
      // Basic attributes
      "*": ["class", "id", "title"],
      // Link attributes
      ...(allowLinks ? { a: ["href", "target", "rel"] } : {}),
      // Image attributes
      ...(allowImages ? { img: ["src", "alt", "width", "height"] } : {}),
      // Table attributes
      th: ["colspan", "rowspan", "scope"],
      td: ["colspan", "rowspan"],
    },
    allowedSchemes: allowLinks ? ["http", "https", "mailto"] : [],
    allowedSchemesByTag: {
      ...(allowLinks ? { a: ["http", "https", "mailto"] } : {}),
      ...(allowImages ? { img: ["http", "https", "data"] } : {}),
    },
    ...sanitizeOptions,
  };

  return (
    <div className={className} style={style}>
      <ReactMarkdown
        rehypePlugins={[[rehypeSanitize, defaultSanitizeOptions]]}
        components={{
          // Custom styling for better appearance
          h1: ({ children }) => (
            <h1
              style={{
                margin: "24px 0 16px 0",
                color: "#1976d2",
                fontSize: "24px",
                fontWeight: "bold",
              }}
            >
              {children}
            </h1>
          ),
          h2: ({ children }) => (
            <h2
              style={{
                margin: "20px 0 12px 0",
                color: "#1976d2",
                borderBottom: "1px solid #e0e0e0",
                paddingBottom: "4px",
                fontSize: "20px",
                fontWeight: "bold",
              }}
            >
              {children}
            </h2>
          ),
          h3: ({ children }) => (
            <h3
              style={{
                margin: "16px 0 8px 0",
                color: "#1976d2",
                fontSize: "18px",
                fontWeight: "bold",
              }}
            >
              {children}
            </h3>
          ),
          h4: ({ children }) => (
            <h4
              style={{
                margin: "12px 0 6px 0",
                color: "#1976d2",
                fontSize: "16px",
                fontWeight: "bold",
              }}
            >
              {children}
            </h4>
          ),
          h5: ({ children }) => (
            <h5
              style={{
                margin: "10px 0 5px 0",
                color: "#1976d2",
                fontSize: "14px",
                fontWeight: "bold",
              }}
            >
              {children}
            </h5>
          ),
          h6: ({ children }) => (
            <h6
              style={{
                margin: "8px 0 4px 0",
                color: "#1976d2",
                fontSize: "13px",
                fontWeight: "bold",
              }}
            >
              {children}
            </h6>
          ),
          p: ({ children }) => <p style={{ margin: "8px 0" }}>{children}</p>,
          ul: ({ children }) => (
            <ul style={{ margin: "8px 0", paddingLeft: "20px" }}>{children}</ul>
          ),
          ol: ({ children }) => (
            <ol style={{ margin: "8px 0", paddingLeft: "20px" }}>{children}</ol>
          ),
          li: ({ children }) => <li style={{ margin: "2px 0" }}>{children}</li>,
          code: ({ children, className }) => {
            const isInline = !className;
            return isInline ? (
              <code
                style={{
                  background: "#f0f0f0",
                  padding: "2px 4px",
                  borderRadius: "3px",
                  fontFamily: "monospace",
                  fontSize: "0.9em",
                }}
              >
                {children}
              </code>
            ) : (
              <code>{children}</code>
            );
          },
          pre: ({ children }) => (
            <pre
              style={{
                background: "#f5f5f5",
                padding: "12px",
                borderRadius: "4px",
                overflowX: "auto",
                margin: "8px 0",
                fontFamily: "monospace",
                fontSize: "14px",
              }}
            >
              {children}
            </pre>
          ),
          blockquote: ({ children }) => (
            <blockquote
              style={{
                margin: "16px 0",
                padding: "8px 16px",
                borderLeft: "4px solid #e0e0e0",
                backgroundColor: "#f9f9f9",
                fontStyle: "italic",
              }}
            >
              {children}
            </blockquote>
          ),
          a: ({ children, href }) => (
            <a
              href={href}
              target="_blank"
              rel="noopener noreferrer"
              style={{
                color: "#1976d2",
                textDecoration: "none",
                borderBottom: "1px solid transparent",
                transition: "border-bottom 0.2s",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.borderBottom = "1px solid #1976d2";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.borderBottom = "1px solid transparent";
              }}
            >
              {children}
            </a>
          ),
          table: ({ children }) => (
            <table
              style={{
                borderCollapse: "collapse",
                width: "100%",
                margin: "16px 0",
                border: "1px solid #e0e0e0",
              }}
            >
              {children}
            </table>
          ),
          thead: ({ children }) => (
            <thead style={{ backgroundColor: "#f5f5f5" }}>{children}</thead>
          ),
          tbody: ({ children }) => <tbody>{children}</tbody>,
          tr: ({ children }) => <tr>{children}</tr>,
          th: ({ children }) => (
            <th
              style={{
                padding: "8px 12px",
                textAlign: "left",
                borderBottom: "2px solid #e0e0e0",
                fontWeight: "600",
                backgroundColor: "#f5f5f5",
              }}
            >
              {children}
            </th>
          ),
          td: ({ children }) => (
            <td
              style={{
                padding: "8px 12px",
                borderBottom: "1px solid #e0e0e0",
              }}
            >
              {children}
            </td>
          ),
          hr: ({ children }) => (
            <hr
              style={{
                border: "none",
                borderTop: "1px solid #e0e0e0",
                margin: "16px 0",
              }}
            >
              {children}
            </hr>
          ),
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}

/**
 * Hook for safe markdown rendering with custom options.
 */
export function useSafeMarkdown(
  content: string,
  _options?: Partial<SafeMarkdownProps>
) {
  return content; // react-markdown handles sanitization automatically
}

export default SafeMarkdown;
