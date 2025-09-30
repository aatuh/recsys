import React from "react";
import { color, spacing, radius, text } from "../../ui/tokens";
import type { DiffItem, DiffValue, ChangeType } from "../../types/ui";

interface DiffBlockProps {
  title?: string;
  diffs: DiffItem[];
  showReasons?: boolean;
  compact?: boolean;
  showExport?: boolean;
  onExport?: (diffs: DiffItem[]) => void;
}

export function DiffBlock({
  title = "Changes Applied",
  diffs,
  showReasons = true,
  compact = false,
  showExport = false,
  onExport,
}: DiffBlockProps) {
  if (diffs.length === 0) return null;

  const formatValue = (value: DiffValue): string => {
    if (value === null || value === undefined) return "—";
    if (typeof value === "number") return value.toFixed(2);
    if (typeof value === "boolean") return value ? "Yes" : "No";
    if (typeof value === "object") return JSON.stringify(value);
    return String(value);
  };

  const getChangeType = (before: DiffValue, after: DiffValue): ChangeType => {
    if (before === null || before === undefined) return "added";
    if (after === null || after === undefined) return "removed";
    return "modified";
  };

  const getChangeColor = (type: ChangeType) => {
    switch (type) {
      case "added":
        return color.success;
      case "removed":
        return color.danger;
      case "modified":
        return color.warning;
    }
  };

  const getChangeIcon = (type: ChangeType) => {
    switch (type) {
      case "added":
        return "➕";
      case "removed":
        return "➖";
      case "modified":
        return "🔄";
    }
  };

  return (
    <div
      style={{
        backgroundColor: color.panelSubtle,
        border: `1px solid ${color.panelBorder}`,
        borderRadius: radius.md,
        padding: compact ? spacing.sm : spacing.md,
        margin: `${spacing.sm}px 0`,
        fontSize: compact ? text.sm : text.md,
      }}
    >
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: spacing.sm,
        }}
      >
        <strong
          style={{ color: color.text, fontSize: compact ? text.sm : text.md }}
        >
          {title}
        </strong>
        <div style={{ display: "flex", alignItems: "center", gap: spacing.sm }}>
          {showExport && onExport && (
            <button
              onClick={() => onExport(diffs)}
              style={{
                backgroundColor: color.buttonBg,
                color: color.text,
                border: `1px solid ${color.buttonBorder}`,
                borderRadius: radius.sm,
                padding: "4px 8px",
                fontSize: text.xs,
                cursor: "pointer",
                fontWeight: 500,
              }}
              title="Export changes"
            >
              📤 Export
            </button>
          )}
          <span
            style={{
              backgroundColor: color.primary,
              color: color.primaryTextOn,
              padding: "2px 6px",
              borderRadius: radius.sm,
              fontSize: text.xs,
              fontWeight: 500,
            }}
          >
            {diffs.length} change{diffs.length !== 1 ? "s" : ""}
          </span>
        </div>
      </div>

      <div
        style={{ display: "flex", flexDirection: "column", gap: spacing.xs }}
      >
        {diffs.map((diff, index) => {
          const changeType = getChangeType(diff.before, diff.after);
          const changeColor = getChangeColor(changeType);
          const changeIcon = getChangeIcon(changeType);

          return (
            <div
              key={index}
              style={{
                display: "flex",
                alignItems: "center",
                gap: spacing.sm,
                padding: spacing.xs,
                backgroundColor: color.panelBg,
                borderRadius: radius.sm,
                border: `1px solid ${color.border}`,
              }}
            >
              <span style={{ fontSize: text.sm }}>{changeIcon}</span>

              <div style={{ flex: 1, minWidth: 0 }}>
                <div
                  style={{
                    fontWeight: 500,
                    color: color.text,
                    marginBottom: 2,
                  }}
                >
                  {diff.field}
                </div>

                <div
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: spacing.xs,
                  }}
                >
                  {changeType !== "added" && (
                    <span
                      style={{
                        color: color.textMuted,
                        fontSize: text.xs,
                        textDecoration: "line-through",
                        backgroundColor: color.dangerBg,
                        padding: "1px 4px",
                        borderRadius: 2,
                      }}
                    >
                      {formatValue(diff.before)}
                    </span>
                  )}

                  {changeType !== "removed" && (
                    <span
                      style={{
                        color: changeColor,
                        fontSize: text.xs,
                        fontWeight: 500,
                        backgroundColor:
                          changeType === "added"
                            ? color.successBg
                            : color.warningBg,
                        padding: "1px 4px",
                        borderRadius: 2,
                      }}
                    >
                      {formatValue(diff.after)}
                    </span>
                  )}
                </div>

                {showReasons && diff.reason && (
                  <div
                    style={{
                      color: color.textMuted,
                      fontSize: text.xs,
                      marginTop: 2,
                      fontStyle: "italic",
                    }}
                  >
                    {diff.reason}
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// Hook to generate diffs from before/after states
export function useDiffGenerator() {
  const generateDiffs = (
    before: Record<string, DiffValue>,
    after: Record<string, DiffValue>,
    fieldLabels?: Record<string, string>
  ): DiffItem[] => {
    const diffs: DiffItem[] = [];

    // Get all unique keys from both objects
    const allKeys = new Set([...Object.keys(before), ...Object.keys(after)]);

    for (const key of allKeys) {
      const beforeValue = before[key];
      const afterValue = after[key];

      // Skip if values are the same
      if (JSON.stringify(beforeValue) === JSON.stringify(afterValue)) {
        continue;
      }

      const fieldLabel = fieldLabels?.[key] || key;
      diffs.push({
        field: fieldLabel,
        before: beforeValue,
        after: afterValue,
      });
    }

    return diffs;
  };

  const generateRuleDiffs = (
    before: Record<string, DiffValue>,
    after: Record<string, DiffValue>,
    ruleName?: string
  ): DiffItem[] => {
    const diffs = generateDiffs(before, after);

    // Add rule context to reasons
    if (ruleName) {
      diffs.forEach((diff) => {
        diff.reason = `Applied by rule: ${ruleName}`;
      });
    }

    return diffs;
  };

  const generateOverrideDiffs = (
    before: Record<string, DiffValue>,
    after: Record<string, DiffValue>,
    overrideName?: string
  ): DiffItem[] => {
    const diffs = generateDiffs(before, after);

    // Add override context to reasons
    if (overrideName) {
      diffs.forEach((diff) => {
        diff.reason = `Override: ${overrideName}`;
      });
    }

    return diffs;
  };

  return {
    generateDiffs,
    generateRuleDiffs,
    generateOverrideDiffs,
  };
}
