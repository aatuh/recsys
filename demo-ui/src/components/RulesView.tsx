import React from "react";
import { RulesPanel } from "./RulesPanel";

interface RulesViewProps {
  namespace: string;
}

export function RulesView({ namespace }: RulesViewProps) {
  return (
    <div style={{ padding: 16, fontFamily: "system-ui, sans-serif" }}>
      <p style={{ color: "#444", marginBottom: 24 }}>
        Manage the Rule Engine v1 for PIN/BLOCK/BOOST operations with TTL and
        precedence. Create, edit, and test rules that are evaluated at request
        time to influence recommendation results. Use the dry-run feature to
        preview rule effects before applying them to live traffic.
      </p>

      <RulesPanel namespace={namespace} />
    </div>
  );
}


