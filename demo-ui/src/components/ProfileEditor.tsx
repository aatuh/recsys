import React, { useState, useEffect } from "react";
import {
  Section,
  Row,
  Label,
  Button,
  NumberInput,
  TextInput,
} from "./UIComponents";
import type { types_Overrides } from "../lib/api-client";

interface ProfileEditorProps {
  profileName: string;
  profileDescription: string;
  overrides: types_Overrides;
  onSave: (
    name: string,
    description: string,
    overrides: types_Overrides
  ) => void;
  onCancel: () => void;
  onClearField: (field: keyof types_Overrides) => void;
}

// Define all possible override fields with their types and labels
const OVERRIDE_FIELDS = [
  {
    key: "popularity_halflife_days" as const,
    label: "Popularity Half-Life (days)",
    type: "number" as const,
  },
  {
    key: "covis_window_days" as const,
    label: "Co-Vis Window (days)",
    type: "number" as const,
  },
  {
    key: "popularity_fanout" as const,
    label: "Popularity Fanout",
    type: "number" as const,
  },
  { key: "mmr_lambda" as const, label: "MMR Lambda", type: "number" as const },
  { key: "brand_cap" as const, label: "Brand Cap", type: "number" as const },
  {
    key: "category_cap" as const,
    label: "Category Cap",
    type: "number" as const,
  },
  {
    key: "rule_exclude_purchased" as const,
    label: "Exclude Purchased",
    type: "boolean" as const,
  },
  {
    key: "purchased_window_days" as const,
    label: "Purchased Window (days)",
    type: "number" as const,
  },
  {
    key: "profile_window_days" as const,
    label: "Profile Window (days)",
    type: "number" as const,
  },
  {
    key: "profile_boost" as const,
    label: "Profile Boost",
    type: "number" as const,
  },
  {
    key: "profile_top_n" as const,
    label: "Profile Top N",
    type: "number" as const,
  },
  {
    key: "blend_alpha" as const,
    label: "Blend Alpha",
    type: "number" as const,
  },
  { key: "blend_beta" as const, label: "Blend Beta", type: "number" as const },
  {
    key: "blend_gamma" as const,
    label: "Blend Gamma",
    type: "number" as const,
  },
];

export function ProfileEditor({
  profileName,
  profileDescription,
  overrides,
  onSave,
  onCancel,
  onClearField,
}: ProfileEditorProps) {
  const [name, setName] = useState(profileName);
  const [description, setDescription] = useState(profileDescription);
  const [editedOverrides, setEditedOverrides] = useState<types_Overrides>({
    ...overrides,
  });

  useEffect(() => {
    setName(profileName);
    setDescription(profileDescription);
    setEditedOverrides({ ...overrides });
  }, [profileName, profileDescription, overrides]);

  const handleSave = () => {
    onSave(name, description, editedOverrides);
  };

  const updateOverride = (key: keyof types_Overrides, value: string) => {
    if (value === "") {
      const newOverrides = { ...editedOverrides };
      delete newOverrides[key];
      setEditedOverrides(newOverrides);
    } else {
      const numValue = parseFloat(value);
      if (!isNaN(numValue)) {
        setEditedOverrides({ ...editedOverrides, [key]: numValue });
      }
    }
  };

  const updateBooleanOverride = (
    key: keyof types_Overrides,
    value: boolean
  ) => {
    setEditedOverrides({ ...editedOverrides, [key]: value });
  };

  const updateStringOverride = (key: keyof types_Overrides, value: string) => {
    if (value === "") {
      const newOverrides = { ...editedOverrides };
      delete newOverrides[key];
      setEditedOverrides(newOverrides);
    } else {
      setEditedOverrides({ ...editedOverrides, [key]: value });
    }
  };

  const clearField = (key: keyof types_Overrides) => {
    const newOverrides = { ...editedOverrides };
    delete newOverrides[key];
    setEditedOverrides(newOverrides);
    onClearField(key);
  };

  return (
    <Section title="Edit Profile">
      <div style={{ marginBottom: 16 }}>
        <Row>
          <Label text="Profile Name">
            <TextInput
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Enter profile name"
            />
          </Label>
        </Row>

        <Row>
          <Label text="Description">
            <TextInput
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Enter profile description"
            />
          </Label>
        </Row>

        <div style={{ marginTop: 20 }}>
          <h4 style={{ marginBottom: 16, color: "#333" }}>
            Override Parameters
          </h4>

          {OVERRIDE_FIELDS.map((field) => {
            const value = editedOverrides[field.key];
            const hasValue = value !== undefined && value !== null;

            return (
              <Row key={field.key} style={{ marginBottom: 12 }}>
                <Label text={field.label}>
                  <div
                    style={{ display: "flex", alignItems: "center", gap: 8 }}
                  >
                    {field.type === "number" ? (
                      <NumberInput
                        value={hasValue ? value.toString() : ""}
                        onChange={(e) =>
                          updateOverride(field.key, e.target.value)
                        }
                        placeholder="Use default"
                        style={{ flex: 1 }}
                      />
                    ) : (
                      <select
                        value={hasValue ? value.toString() : ""}
                        onChange={(e) =>
                          updateBooleanOverride(
                            field.key,
                            e.target.value === "true"
                          )
                        }
                        style={{
                          padding: "8px 12px",
                          border: "1px solid #ddd",
                          borderRadius: 4,
                          flex: 1,
                        }}
                      >
                        <option value="">Use default</option>
                        <option value="true">True</option>
                        <option value="false">False</option>
                      </select>
                    )}
                    {hasValue && (
                      <Button
                        onClick={() => clearField(field.key)}
                        style={{
                          backgroundColor: "#dc3545",
                          color: "white",
                          border: "none",
                          padding: "4px 8px",
                          borderRadius: 4,
                          fontSize: 12,
                          cursor: "pointer",
                        }}
                      >
                        Clear
                      </Button>
                    )}
                  </div>
                </Label>
              </Row>
            );
          })}
        </div>

        <Row style={{ marginTop: 20 }}>
          <Button
            onClick={handleSave}
            style={{
              backgroundColor: "#28a745",
              color: "white",
              border: "none",
              padding: "10px 20px",
              borderRadius: 4,
              cursor: "pointer",
              marginRight: 8,
            }}
          >
            Save Profile
          </Button>
          <Button
            onClick={onCancel}
            style={{
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              padding: "10px 20px",
              borderRadius: 4,
              cursor: "pointer",
            }}
          >
            Cancel
          </Button>
        </Row>
      </div>
    </Section>
  );
}
