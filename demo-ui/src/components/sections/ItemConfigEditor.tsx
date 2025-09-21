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

export interface ItemValue {
  value: string;
  probability: number; // 0-1, probability of this value being selected
}

export interface ItemConfig {
  key: string;
  probability: number; // 0-1, probability of this config being included
  values: ItemValue[];
}

export interface PriceRange {
  min: number;
  max: number;
  probability: number; // 0-1, probability of this price range being selected
}

export interface ItemConfigEditorProps {
  itemConfigs: ItemConfig[];
  setItemConfigs: (configs: ItemConfig[]) => void;
  priceRanges: PriceRange[];
  setPriceRanges: (ranges: PriceRange[]) => void;
  generatedItems: string[];
  onUpdateItem: (itemId: string, updates: Record<string, any>) => Promise<void>;
}

export function ItemConfigEditor({
  itemConfigs,
  setItemConfigs,
  priceRanges,
  setPriceRanges,
  generatedItems,
  onUpdateItem,
}: ItemConfigEditorProps) {
  const [selectedItem, setSelectedItem] = useState("");
  const [itemUpdates, setItemUpdates] = useState<Record<string, any>>({});
  const [loading, setLoading] = useState(false);
  const toast = useToast();

  // Add new item config
  const addItemConfig = () => {
    const newConfig: ItemConfig = {
      key: `prop_${itemConfigs.length + 1}`,
      probability: 0.5,
      values: [
        { value: "value1", probability: 0.5 },
        { value: "value2", probability: 0.5 },
      ],
    };
    setItemConfigs([...itemConfigs, newConfig]);
  };

  // Remove item config
  const removeItemConfig = (index: number) => {
    setItemConfigs(itemConfigs.filter((_, i) => i !== index));
  };

  // Update item config
  const updateItemConfig = (
    index: number,
    field: keyof ItemConfig,
    value: any
  ) => {
    const updated = [...itemConfigs];
    if (updated[index]) {
      updated[index] = { ...updated[index], [field]: value };
      setItemConfigs(updated);
    }
  };

  // Add value to item config
  const addItemValue = (configIndex: number) => {
    const updated = [...itemConfigs];
    if (updated[configIndex]) {
      const newValue: ItemValue = {
        value: `value_${updated[configIndex].values.length + 1}`,
        probability: 0.5,
      };
      updated[configIndex].values.push(newValue);
      setItemConfigs(updated);
    }
  };

  // Remove value from item config
  const removeItemValue = (configIndex: number, valueIndex: number) => {
    const updated = [...itemConfigs];
    if (updated[configIndex]) {
      updated[configIndex].values = updated[configIndex].values.filter(
        (_, i) => i !== valueIndex
      );
      setItemConfigs(updated);
    }
  };

  // Update item value
  const updateItemValue = (
    configIndex: number,
    valueIndex: number,
    field: keyof ItemValue,
    value: any
  ) => {
    const updated = [...itemConfigs];
    if (updated[configIndex] && updated[configIndex].values[valueIndex]) {
      updated[configIndex].values[valueIndex] = {
        ...updated[configIndex].values[valueIndex],
        [field]: value,
      };
      setItemConfigs(updated);
    }
  };

  // Add new price range
  const addPriceRange = () => {
    const newRange: PriceRange = {
      min: 10,
      max: 50,
      probability: 0.5,
    };
    setPriceRanges([...priceRanges, newRange]);
  };

  // Remove price range
  const removePriceRange = (index: number) => {
    setPriceRanges(priceRanges.filter((_, i) => i !== index));
  };

  // Update price range
  const updatePriceRange = (
    index: number,
    field: keyof PriceRange,
    value: any
  ) => {
    const updated = [...priceRanges];
    if (updated[index]) {
      updated[index] = { ...updated[index], [field]: value };
      setPriceRanges(updated);
    }
  };

  // Load item data (mock implementation - in real app, this would fetch from API)
  const loadItemData = async (itemId: string) => {
    setLoading(true);
    try {
      // Mock: generate random data based on current config
      const updates: Record<string, any> = {};

      // Generate price from price ranges
      if (priceRanges.length > 0) {
        const selectedRange = selectWeightedPriceRange(priceRanges);
        updates.price = Math.floor(
          Math.random() * (selectedRange.max - selectedRange.min + 1) +
            selectedRange.min
        );
      }

      // Generate properties based on configs
      const props: Record<string, any> = {};
      itemConfigs.forEach((config) => {
        if (Math.random() < config.probability) {
          const selectedValue = selectWeightedValue(config.values);
          props[config.key] = selectedValue;
        }
      });
      if (Object.keys(props).length > 0) {
        updates.props = props;
      }

      setItemUpdates(updates);
      toast.success(`Loaded configuration for ${itemId}`);
    } catch (error: any) {
      toast.error(`Error loading item data: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };

  // Select weighted random value
  const selectWeightedValue = (values: ItemValue[]): string => {
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

  // Select weighted price range
  const selectWeightedPriceRange = (ranges: PriceRange[]): PriceRange => {
    const totalWeight = ranges.reduce((sum, r) => sum + r.probability, 0);
    let random = Math.random() * totalWeight;

    for (const range of ranges) {
      random -= range.probability;
      if (random <= 0) {
        return range;
      }
    }
    return ranges[ranges.length - 1] || { min: 10, max: 50, probability: 1.0 };
  };

  // Update item data
  const updateItemData = async () => {
    if (!selectedItem) return;

    setLoading(true);
    try {
      await onUpdateItem(selectedItem, itemUpdates);
      toast.success(`Updated data for ${selectedItem}`);
    } catch (error: any) {
      toast.error(`Error updating item: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };

  // Add new property to current item
  const addPropertyToItem = () => {
    const newKey = prompt("Enter property key:");
    const newValue = prompt("Enter property value:");
    if (newKey && newValue) {
      setItemUpdates({ ...itemUpdates, [newKey]: newValue });
    }
  };

  // Remove property from current item
  const removePropertyFromItem = (key: string) => {
    const updated = { ...itemUpdates };
    delete updated[key];
    setItemUpdates(updated);
  };

  return (
    <Section title="Item Configuration Editor">
      <div style={{ marginBottom: spacing.lg }}>
        <h3 style={{ marginTop: 0, marginBottom: 8 }}>
          Price Range Configuration
        </h3>
        <p
          style={{
            color: color.textMuted,
            fontSize: 14,
            marginBottom: spacing.md,
          }}
        >
          Configure price ranges that will be randomly assigned to items during
          seeding.
        </p>

        {priceRanges.map((range, rangeIndex) => (
          <div
            key={rangeIndex}
            style={{
              border: `1px solid ${color.border}`,
              borderRadius: 6,
              padding: spacing.md,
              marginBottom: spacing.md,
              backgroundColor: color.panelBg,
            }}
          >
            <Row>
              <Label text="Min Price" width={120}>
                <NumberInput
                  min={0}
                  value={range.min}
                  onChange={(e) =>
                    updatePriceRange(rangeIndex, "min", Number(e.target.value))
                  }
                />
              </Label>
              <Label text="Max Price" width={120}>
                <NumberInput
                  min={range.min}
                  value={range.max}
                  onChange={(e) =>
                    updatePriceRange(rangeIndex, "max", Number(e.target.value))
                  }
                />
              </Label>
              <Label text="Probability" width={120}>
                <NumberInput
                  min={0}
                  max={1}
                  step={0.1}
                  value={range.probability}
                  onChange={(e) =>
                    updatePriceRange(
                      rangeIndex,
                      "probability",
                      Number(e.target.value)
                    )
                  }
                />
              </Label>
              <Button
                onClick={() => removePriceRange(rangeIndex)}
                aria-label={`Remove price range ${rangeIndex + 1}`}
                style={{
                  backgroundColor: "#ffebee",
                  color: "#c62828",
                }}
              >
                Remove
              </Button>
            </Row>
          </div>
        ))}

        <Button
          onClick={addPriceRange}
          style={{ backgroundColor: "#e8f5e8", color: "#2e7d32" }}
        >
          Add Price Range
        </Button>
      </div>

      <div style={{ marginBottom: spacing.lg }}>
        <h3 style={{ marginTop: 0, marginBottom: 8 }}>
          Property Configuration
        </h3>
        <p
          style={{
            color: color.textMuted,
            fontSize: 14,
            marginBottom: spacing.md,
          }}
        >
          Configure dynamic properties that will be randomly assigned to items
          during seeding.
        </p>

        {itemConfigs.map((config, configIndex) => (
          <div
            key={configIndex}
            style={{
              border: `1px solid ${color.border}`,
              borderRadius: 6,
              padding: spacing.md,
              marginBottom: spacing.md,
              backgroundColor: color.panelBg,
            }}
          >
            <Row>
              <Label text="Property Key" width={150}>
                <TextInput
                  value={config.key}
                  onChange={(e) =>
                    updateItemConfig(configIndex, "key", e.target.value)
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
                    updateItemConfig(
                      configIndex,
                      "probability",
                      Number(e.target.value)
                    )
                  }
                />
              </Label>
              <Button
                onClick={() => removeItemConfig(configIndex)}
                aria-label={`Remove property config ${
                  config.key || configIndex + 1
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
                  onClick={() => addItemValue(configIndex)}
                  aria-label={`Add value to ${
                    config.key || `config ${configIndex + 1}`
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
                          updateItemValue(
                            configIndex,
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
                          updateItemValue(
                            configIndex,
                            valueIndex,
                            "probability",
                            Number(e.target.value)
                          )
                        }
                      />
                    </Label>
                    <Button
                      onClick={() => removeItemValue(configIndex, valueIndex)}
                      aria-label={`Remove value ${valueIndex + 1} from ${
                        config.key || `config ${configIndex + 1}`
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
          onClick={addItemConfig}
          style={{ backgroundColor: "#e8f5e8", color: "#2e7d32" }}
        >
          Add Property Configuration
        </Button>
      </div>

      <div
        style={{
          borderTop: `1px solid ${color.border}`,
          paddingTop: spacing.md,
        }}
      >
        <h3 style={{ marginTop: 0, marginBottom: 8 }}>Edit Item Data</h3>

        <Row>
          <Label text="Select Item" width={200}>
            <select
              value={selectedItem}
              onChange={(e) => setSelectedItem(e.target.value)}
              aria-label="Select item to edit"
              style={{
                width: "100%",
                padding: "8px 10px",
                border: `1px solid ${color.border}`,
                borderRadius: 6,
                fontFamily: "monospace",
              }}
            >
              <option value="">Select an item...</option>
              {generatedItems.map((itemId) => (
                <option key={itemId} value={itemId}>
                  {itemId}
                </option>
              ))}
            </select>
          </Label>
          <Button
            onClick={() => loadItemData(selectedItem)}
            disabled={!selectedItem || loading}
            aria-label="Load item configuration"
          >
            Load Configuration
          </Button>
        </Row>

        {selectedItem && (
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
                Current Data:
              </span>
              <Button
                onClick={addPropertyToItem}
                aria-label="Add property to current item"
                style={{ fontSize: 12, padding: "4px 8px" }}
              >
                Add Property
              </Button>
            </div>

            {Object.entries(itemUpdates).map(([key, value]) => (
              <div key={key} style={{ marginBottom: 4 }}>
                <Row>
                  <Label text="Key" width={120}>
                    <TextInput
                      value={key}
                      onChange={(e) => {
                        const newUpdates = { ...itemUpdates };
                        delete newUpdates[key];
                        newUpdates[e.target.value] = value;
                        setItemUpdates(newUpdates);
                      }}
                    />
                  </Label>
                  <Label text="Value" width={120}>
                    <TextInput
                      value={value}
                      onChange={(e) =>
                        setItemUpdates({
                          ...itemUpdates,
                          [key]: e.target.value,
                        })
                      }
                    />
                  </Label>
                  <Button
                    onClick={() => removePropertyFromItem(key)}
                    aria-label={`Remove property ${key}`}
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
                onClick={updateItemData}
                disabled={loading}
                aria-label="Update item data"
                style={{ backgroundColor: "#e3f2fd", color: "#1565c0" }}
              >
                {loading ? "Updating..." : "Update Item Data"}
              </Button>
            </div>
          </div>
        )}
      </div>
    </Section>
  );
}
