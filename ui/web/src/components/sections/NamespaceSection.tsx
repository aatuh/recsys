import React from "react";
import { Section, Row, Label, TextInput } from "../primitives/UIComponents";
import { spacing, text } from "../../ui/tokens";

interface NamespaceSectionProps {
  namespace: string;
  setNamespace: (value: string) => void;
  apiBase: string;
}

export function NamespaceSection({
  namespace,
  setNamespace,
  apiBase,
}: NamespaceSectionProps) {
  return (
    <Section title="Namespace">
      <Row>
        <Label text="Namespace">
          <TextInput
            value={namespace}
            onChange={(e) => setNamespace(e.target.value)}
          />
        </Label>
        <Label text="API Base (read-only)">
          <TextInput value={apiBase} readOnly />
        </Label>
      </Row>
      <div style={{ color: "#666", fontSize: text.md, marginTop: spacing.sm }}>
        Use different namespaces to isolate demo datasets without affecting
        others.
      </div>
    </Section>
  );
}
