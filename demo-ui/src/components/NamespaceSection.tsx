import React from "react";
import { Section, Row, Label, TextInput } from "./UIComponents";

interface NamespaceSectionProps {
  namespace: string;
  setNamespace: (namespace: string) => void;
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
    </Section>
  );
}
