import React, { useState } from "react";
import {
  Section,
  Row,
  Label,
  TextInput,
  NumberInput,
  Button,
} from "../primitives/UIComponents";
import { color, spacing } from "../../ui/tokens";
import { useToast } from "../../ui/Toast";

export interface TraitValue {
  value: string;
  probability: number; // 0-1, probability of this value being selected
}

export interface TraitConfig {
  key: string;
  probability: number; // 0-1, probability of this trait being included
  values: TraitValue[];
}

export interface UserTraitsEditorProps {
  traitConfigs: TraitConfig[];
  setTraitConfigs: (value: TraitConfig[]) => void;
  generatedUsers: string[];
  onUpdateUser: (
    _userId: string,
    _traits: Record<string, any>
  ) => Promise<void>;
}

export function UserTraitsEditor({
  traitConfigs,
  setTraitConfigs,
  generatedUsers,
  onUpdateUser,
}: UserTraitsEditorProps) {
  const [selectedUser, setSelectedUser] = useState("");
  const [userTraits, setUserTraits] = useState<Record<string, any>>({});
  const [loading, setLoading] = useState(false);
  const toast = useToast();

  // Add new trait config
  const addTraitConfig = () => {
    const newConfig: TraitConfig = {
      key: `trait_${traitConfigs.length + 1}`,
      probability: 0.5,
      values: [
        { value: "value1", probability: 0.5 },
        { value: "value2", probability: 0.5 },
      ],
    };
    setTraitConfigs([...traitConfigs, newConfig]);
  };

  // Remove trait config
  const removeTraitConfig = (index: number) => {
    setTraitConfigs(traitConfigs.filter((_, i) => i !== index));
  };

  // Update trait config
  const updateTraitConfig = (
    index: number,
    field: keyof TraitConfig,
    value: any
  ) => {
    const updated = [...traitConfigs];
    if (updated[index]) {
      updated[index] = { ...updated[index], [field]: value };
      setTraitConfigs(updated);
    }
  };

  // Add value to trait config
  const addTraitValue = (traitIndex: number) => {
    const updated = [...traitConfigs];
    if (updated[traitIndex]) {
      const newValue: TraitValue = {
        value: `value_${updated[traitIndex].values.length + 1}`,
        probability: 0.5,
      };
      updated[traitIndex].values.push(newValue);
      setTraitConfigs(updated);
    }
  };

  // Remove value from trait config
  const removeTraitValue = (traitIndex: number, valueIndex: number) => {
    const updated = [...traitConfigs];
    if (updated[traitIndex]) {
      updated[traitIndex].values = updated[traitIndex].values.filter(
        (_, i) => i !== valueIndex
      );
      setTraitConfigs(updated);
    }
  };

  // Update trait value
  const updateTraitValue = (
    traitIndex: number,
    valueIndex: number,
    field: keyof TraitValue,
    value: any
  ) => {
    const updated = [...traitConfigs];
    if (updated[traitIndex] && updated[traitIndex].values[valueIndex]) {
      updated[traitIndex].values[valueIndex] = {
        ...updated[traitIndex].values[valueIndex],
        [field]: value,
      };
      setTraitConfigs(updated);
    }
  };

  // Load user traits (mock implementation - in real app, this would fetch from API)
  const loadUserTraits = async (userId: string) => {
    setLoading(true);
    try {
      // Mock: generate random traits based on current config
      const traits: Record<string, any> = {};
      traitConfigs.forEach((config) => {
        if (Math.random() < config.probability) {
          const selectedValue = selectWeightedValue(config.values);
          traits[config.key] = selectedValue;
        }
      });
      setUserTraits(traits);
      toast.success(`Loaded traits for ${userId}`);
    } catch (error: any) {
      toast.error(`Error loading traits: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };

  // Select weighted random value
  const selectWeightedValue = (values: TraitValue[]): string => {
    const totalWeight = values.reduce((sum, v) => sum + v.probability, 0);
    let random = Math.random() * totalWeight;

    for (const value of values) {
      random -= value.probability;
      if (random <= 0) {
        return value.value;
      }
    }
    return values[values.length - 1]?.value || "";
  };

  // Update user traits
  const updateUserTraits = async () => {
    if (!selectedUser) return;

    setLoading(true);
    try {
      await onUpdateUser(selectedUser, userTraits);
      toast.success(`Updated traits for ${selectedUser}`);
    } catch (error: any) {
      toast.error(`Error updating traits: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };

  // Add new trait to current user
  const addTraitToUser = () => {
    const newKey = prompt("Enter trait key:");
    const newValue = prompt("Enter trait value:");
    if (newKey && newValue) {
      setUserTraits({ ...userTraits, [newKey]: newValue });
    }
  };

  // Remove trait from current user
  const removeTraitFromUser = (key: string) => {
    const updated = { ...userTraits };
    delete updated[key];
    setUserTraits(updated);
  };

  return (
    <Section title="User Traits Editor">
      <div style={{ marginBottom: spacing.lg }}>
        <h3 style={{ marginTop: 0, marginBottom: 8 }}>Trait Configuration</h3>
        <p
          style={{
            color: color.textMuted,
            fontSize: 14,
            marginBottom: spacing.md,
          }}
        >
          Configure dynamic traits that will be randomly assigned to users
          during seeding.
        </p>

        {traitConfigs.map((config, traitIndex) => (
          <div
            key={traitIndex}
            style={{
              border: `1px solid ${color.border}`,
              borderRadius: 6,
              padding: spacing.md,
              marginBottom: spacing.md,
              backgroundColor: color.panelBg,
            }}
          >
            <Row>
              <Label text="Trait Key" width={150}>
                <TextInput
                  value={config.key}
                  onChange={(e) =>
                    updateTraitConfig(traitIndex, "key", e.target.value)
                  }
                />
              </Label>
              <Label text="Include Probability" width={150}>
                <NumberInput
                  min={0}
                  max={1}
                  step={0.1}
                  value={config.probability}
                  onChange={(e) =>
                    updateTraitConfig(
                      traitIndex,
                      "probability",
                      Number(e.target.value)
                    )
                  }
                />
              </Label>
              <Button
                onClick={() => removeTraitConfig(traitIndex)}
                aria-label={`Remove trait config ${
                  config.key || traitIndex + 1
                }`}
                style={{
                  backgroundColor: "#ffebee",
                  color: "#c62828",
                }}
              >
                Remove
              </Button>
            </Row>

            <div style={{ marginTop: 8 }}>
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: spacing.sm,
                  marginBottom: spacing.sm,
                }}
              >
                <span style={{ fontSize: 12, color: color.textMuted }}>
                  Values:
                </span>
                <Button
                  onClick={() => addTraitValue(traitIndex)}
                  aria-label={`Add value to ${
                    config.key || `trait ${traitIndex + 1}`
                  }`}
                  style={{ fontSize: 12, padding: "4px 8px" }}
                >
                  Add Value
                </Button>
              </div>

              {config.values.map((value, valueIndex) => (
                <div key={valueIndex} style={{ marginBottom: 4 }}>
                  <Row>
                    <Label text="Value" width={120}>
                      <TextInput
                        value={value.value}
                        onChange={(e) =>
                          updateTraitValue(
                            traitIndex,
                            valueIndex,
                            "value",
                            e.target.value
                          )
                        }
                      />
                    </Label>
                    <Label text="Probability" width={120}>
                      <NumberInput
                        min={0}
                        max={1}
                        step={0.1}
                        value={value.probability}
                        onChange={(e) =>
                          updateTraitValue(
                            traitIndex,
                            valueIndex,
                            "probability",
                            Number(e.target.value)
                          )
                        }
                      />
                    </Label>
                    <Button
                      onClick={() => removeTraitValue(traitIndex, valueIndex)}
                      aria-label={`Remove value ${valueIndex + 1} from ${
                        config.key || `trait ${traitIndex + 1}`
                      }`}
                      style={{
                        backgroundColor: "#ffebee",
                        color: "#c62828",
                        fontSize: 12,
                        padding: "4px 8px",
                      }}
                    >
                      Remove
                    </Button>
                  </Row>
                </div>
              ))}
            </div>
          </div>
        ))}

        <Button
          onClick={addTraitConfig}
          style={{ backgroundColor: "#e8f5e8", color: "#2e7d32" }}
        >
          Add Trait Configuration
        </Button>
      </div>

      <div
        style={{
          borderTop: `1px solid ${color.border}`,
          paddingTop: spacing.md,
        }}
      >
        <h3 style={{ marginTop: 0, marginBottom: 8 }}>Edit User Traits</h3>

        <Row>
          <Label text="Select User" width={200}>
            <select
              value={selectedUser}
              onChange={(e) => setSelectedUser(e.target.value)}
              aria-label="Select user to edit"
              style={{
                width: "100%",
                padding: "8px 10px",
                border: `1px solid ${color.border}`,
                borderRadius: 6,
                fontFamily: "monospace",
              }}
            >
              <option value="">Select a user...</option>
              {generatedUsers.map((userId) => (
                <option key={userId} value={userId}>
                  {userId}
                </option>
              ))}
            </select>
          </Label>
          <Button
            onClick={() => loadUserTraits(selectedUser)}
            disabled={!selectedUser || loading}
            aria-label="Load user traits"
          >
            Load Traits
          </Button>
        </Row>

        {selectedUser && (
          <div style={{ marginTop: spacing.md }}>
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: spacing.sm,
                marginBottom: spacing.sm,
              }}
            >
              <span style={{ fontSize: 12, color: color.textMuted }}>
                Current Traits:
              </span>
              <Button
                onClick={addTraitToUser}
                aria-label="Add trait to current user"
                style={{ fontSize: 12, padding: "4px 8px" }}
              >
                Add Trait
              </Button>
            </div>

            {Object.entries(userTraits).map(([key, value]) => (
              <div key={key} style={{ marginBottom: 4 }}>
                <Row>
                  <Label text="Key" width={120}>
                    <TextInput
                      value={key}
                      onChange={(e) => {
                        const newTraits = { ...userTraits };
                        delete newTraits[key];
                        newTraits[e.target.value] = value;
                        setUserTraits(newTraits);
                      }}
                    />
                  </Label>
                  <Label text="Value" width={120}>
                    <TextInput
                      value={value}
                      onChange={(e) =>
                        setUserTraits({ ...userTraits, [key]: e.target.value })
                      }
                    />
                  </Label>
                  <Button
                    onClick={() => removeTraitFromUser(key)}
                    aria-label={`Remove trait ${key}`}
                    style={{
                      backgroundColor: "#ffebee",
                      color: "#c62828",
                      fontSize: 12,
                      padding: "4px 8px",
                    }}
                  >
                    Remove
                  </Button>
                </Row>
              </div>
            ))}

            <div style={{ marginTop: spacing.sm }}>
              <Button
                onClick={updateUserTraits}
                disabled={loading}
                aria-label="Update user traits"
                style={{ backgroundColor: "#e3f2fd", color: "#1565c0" }}
              >
                {loading ? "Updating..." : "Update User Traits"}
              </Button>
            </div>
          </div>
        )}
      </div>
    </Section>
  );
}
