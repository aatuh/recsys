import React from "react";
import { DataManagementSection } from "./";

interface DataManagementViewProps {
  namespace: string;
}

export function DataManagementView({ namespace }: DataManagementViewProps) {
  return (
    <div style={{ padding: 16, fontFamily: "system-ui, sans-serif" }}>
      <p style={{ color: "#444", marginBottom: 24 }}>
        Manage your data with bulk operations. View, update, and delete users,
        items, and events. Monitor system health and performance metrics.
      </p>

      <DataManagementSection namespace={namespace} />
    </div>
  );
}
