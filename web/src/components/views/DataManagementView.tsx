import React from "react";
import { DataManagementSection } from "../sections/DataManagementSection";
import { spacing, text } from "../../ui/tokens";

interface DataManagementViewProps {
  namespace: string;
}

export function DataManagementView({ namespace }: DataManagementViewProps) {
  return (
    <div style={{ padding: spacing.xl, fontFamily: "system-ui, sans-serif" }}>
      <p style={{ color: "#444", marginBottom: spacing.xl, fontSize: text.md }}>
        Manage your data with bulk operations. View, update, and delete users,
        items, and events. Monitor system health and performance metrics.
      </p>

      <DataManagementSection namespace={namespace} />
    </div>
  );
}
