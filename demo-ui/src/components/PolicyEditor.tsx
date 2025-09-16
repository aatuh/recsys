import React, { useState } from "react";
import { Section, Row, Label, Button } from "./UIComponents";
import type { types_BanditPolicy } from "../lib/api-client";

interface PolicyEditorProps {
  policy: types_BanditPolicy;
  onSave: (policy: types_BanditPolicy) => void;
  onCancel: () => void;
  onCreateCustom: (policy: types_BanditPolicy) => void;
}

export function PolicyEditor({
  policy,
  onSave,
  onCancel,
  onCreateCustom,
}: PolicyEditorProps) {
  const [editedPolicy, setEditedPolicy] = useState<types_BanditPolicy>(policy);

  const handleSave = () => {
    onSave(editedPolicy);
  };

  const handleCreateCustom = () => {
    onCreateCustom(editedPolicy);
  };

  const handleFieldChange = (
    field: keyof types_BanditPolicy,
    value: string | number | boolean
  ) => {
    setEditedPolicy((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleClearField = (field: keyof types_BanditPolicy) => {
    const defaultValue = getDefaultValue(field);
    setEditedPolicy((prev) => ({
      ...prev,
      [field]: defaultValue,
    }));
  };

  const getDefaultValue = (field: keyof types_BanditPolicy) => {
    switch (field) {
      case "policy_id":
        return "";
      case "name":
        return "";
      case "active":
        return true;
      case "blend_alpha":
      case "blend_beta":
      case "blend_gamma":
      case "mmr_lambda":
      case "profile_boost":
      case "half_life_days":
        return 0.0;
      case "brand_cap":
      case "category_cap":
      case "co_vis_window_days":
      case "popularity_fanout":
        return 0;
      case "rule_exclude_purchased":
        return false;
      case "notes":
        return "";
      default:
        return "";
    }
  };

  const formatValue = (value: any) => {
    if (typeof value === "number") {
      return value.toString();
    }
    return value || "";
  };

  return (
    <Section title="Edit Bandit Policy">
      <div style={{ marginBottom: 16 }}>
        <p style={{ color: "#666", fontSize: 14, marginBottom: 16 }}>
          Configure the bandit policy parameters. These settings control how the
          recommendation algorithm behaves when this policy is selected.
        </p>

        {/* Basic Information */}
        <div style={{ marginBottom: 20 }}>
          <Label text="Basic Information">
            <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
              <Row>
                <div style={{ flex: 1 }}>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontSize: 12,
                      fontWeight: "bold",
                    }}
                  >
                    Policy ID
                  </label>
                  <input
                    type="text"
                    value={editedPolicy.policy_id}
                    onChange={(e) =>
                      handleFieldChange("policy_id", e.target.value)
                    }
                    style={{
                      width: "100%",
                      padding: "6px 8px",
                      border: "1px solid #ddd",
                      borderRadius: 4,
                      fontSize: 14,
                    }}
                    placeholder="e.g., policy-001"
                  />
                </div>
                <div style={{ flex: 1 }}>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontSize: 12,
                      fontWeight: "bold",
                    }}
                  >
                    Name
                  </label>
                  <input
                    type="text"
                    value={editedPolicy.name}
                    onChange={(e) => handleFieldChange("name", e.target.value)}
                    style={{
                      width: "100%",
                      padding: "6px 8px",
                      border: "1px solid #ddd",
                      borderRadius: 4,
                      fontSize: 14,
                    }}
                    placeholder="e.g., High Diversity Policy"
                  />
                </div>
                <div
                  style={{
                    flex: 0,
                    display: "flex",
                    alignItems: "center",
                    marginTop: 20,
                  }}
                >
                  <label
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 8,
                      fontSize: 14,
                    }}
                  >
                    <input
                      type="checkbox"
                      checked={editedPolicy.active}
                      onChange={(e) =>
                        handleFieldChange("active", e.target.checked)
                      }
                      style={{ margin: 0 }}
                    />
                    Active
                  </label>
                </div>
              </Row>
            </div>
          </Label>
        </div>

        {/* Algorithm Parameters */}
        <div style={{ marginBottom: 20 }}>
          <Label text="Algorithm Parameters">
            <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
              <Row>
                <div style={{ flex: 1 }}>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontSize: 12,
                      fontWeight: "bold",
                    }}
                  >
                    Blend Alpha (Popularity Weight)
                  </label>
                  <div style={{ display: "flex", gap: 4 }}>
                    <input
                      type="number"
                      step="0.1"
                      min="0"
                      max="1"
                      value={formatValue(editedPolicy.blend_alpha)}
                      onChange={(e) =>
                        handleFieldChange(
                          "blend_alpha",
                          parseFloat(e.target.value) || 0
                        )
                      }
                      style={{
                        flex: 1,
                        padding: "6px 8px",
                        border: "1px solid #ddd",
                        borderRadius: 4,
                        fontSize: 14,
                      }}
                      placeholder="0.0"
                    />
                    <Button
                      onClick={() => handleClearField("blend_alpha")}
                      style={{
                        padding: "6px 8px",
                        backgroundColor: "#6c757d",
                        color: "white",
                        border: "none",
                        borderRadius: 4,
                        fontSize: 12,
                        cursor: "pointer",
                      }}
                      title="Clear field"
                    >
                      Clear
                    </Button>
                  </div>
                  <div style={{ fontSize: 11, color: "#666", marginTop: 2 }}>
                    Weight for popularity-based recommendations (0.0-1.0)
                  </div>
                </div>
                <div style={{ flex: 1 }}>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontSize: 12,
                      fontWeight: "bold",
                    }}
                  >
                    Blend Beta (Co-occurrence Weight)
                  </label>
                  <div style={{ display: "flex", gap: 4 }}>
                    <input
                      type="number"
                      step="0.1"
                      min="0"
                      max="1"
                      value={formatValue(editedPolicy.blend_beta)}
                      onChange={(e) =>
                        handleFieldChange(
                          "blend_beta",
                          parseFloat(e.target.value) || 0
                        )
                      }
                      style={{
                        flex: 1,
                        padding: "6px 8px",
                        border: "1px solid #ddd",
                        borderRadius: 4,
                        fontSize: 14,
                      }}
                      placeholder="0.0"
                    />
                    <Button
                      onClick={() => handleClearField("blend_beta")}
                      style={{
                        padding: "6px 8px",
                        backgroundColor: "#6c757d",
                        color: "white",
                        border: "none",
                        borderRadius: 4,
                        fontSize: 12,
                        cursor: "pointer",
                      }}
                      title="Clear field"
                    >
                      Clear
                    </Button>
                  </div>
                  <div style={{ fontSize: 11, color: "#666", marginTop: 2 }}>
                    Weight for co-occurrence-based recommendations (0.0-1.0)
                  </div>
                </div>
              </Row>

              <Row>
                <div style={{ flex: 1 }}>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontSize: 12,
                      fontWeight: "bold",
                    }}
                  >
                    Blend Gamma (ALS Weight)
                  </label>
                  <div style={{ display: "flex", gap: 4 }}>
                    <input
                      type="number"
                      step="0.1"
                      min="0"
                      max="1"
                      value={formatValue(editedPolicy.blend_gamma)}
                      onChange={(e) =>
                        handleFieldChange(
                          "blend_gamma",
                          parseFloat(e.target.value) || 0
                        )
                      }
                      style={{
                        flex: 1,
                        padding: "6px 8px",
                        border: "1px solid #ddd",
                        borderRadius: 4,
                        fontSize: 14,
                      }}
                      placeholder="0.0"
                    />
                    <Button
                      onClick={() => handleClearField("blend_gamma")}
                      style={{
                        padding: "6px 8px",
                        backgroundColor: "#6c757d",
                        color: "white",
                        border: "none",
                        borderRadius: 4,
                        fontSize: 12,
                        cursor: "pointer",
                      }}
                      title="Clear field"
                    >
                      Clear
                    </Button>
                  </div>
                  <div style={{ fontSize: 11, color: "#666", marginTop: 2 }}>
                    Weight for ALS (Alternating Least Squares) recommendations
                    (0.0-1.0)
                  </div>
                </div>
                <div style={{ flex: 1 }}>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontSize: 12,
                      fontWeight: "bold",
                    }}
                  >
                    MMR Lambda (Diversity Parameter)
                  </label>
                  <div style={{ display: "flex", gap: 4 }}>
                    <input
                      type="number"
                      step="0.1"
                      min="0"
                      max="1"
                      value={formatValue(editedPolicy.mmr_lambda)}
                      onChange={(e) =>
                        handleFieldChange(
                          "mmr_lambda",
                          parseFloat(e.target.value) || 0
                        )
                      }
                      style={{
                        flex: 1,
                        padding: "6px 8px",
                        border: "1px solid #ddd",
                        borderRadius: 4,
                        fontSize: 14,
                      }}
                      placeholder="0.0"
                    />
                    <Button
                      onClick={() => handleClearField("mmr_lambda")}
                      style={{
                        padding: "6px 8px",
                        backgroundColor: "#6c757d",
                        color: "white",
                        border: "none",
                        borderRadius: 4,
                        fontSize: 12,
                        cursor: "pointer",
                      }}
                      title="Clear field"
                    >
                      Clear
                    </Button>
                  </div>
                  <div style={{ fontSize: 11, color: "#666", marginTop: 2 }}>
                    Diversity parameter for Maximal Marginal Relevance (0.0-1.0)
                  </div>
                </div>
              </Row>
            </div>
          </Label>
        </div>

        {/* Caps */}
        <div style={{ marginBottom: 20 }}>
          <Label text="Recommendation Caps">
            <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
              <Row>
                <div style={{ flex: 1 }}>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontSize: 12,
                      fontWeight: "bold",
                    }}
                  >
                    Brand Cap
                  </label>
                  <div style={{ display: "flex", gap: 4 }}>
                    <input
                      type="number"
                      min="0"
                      value={formatValue(editedPolicy.brand_cap)}
                      onChange={(e) =>
                        handleFieldChange(
                          "brand_cap",
                          parseInt(e.target.value) || 0
                        )
                      }
                      style={{
                        flex: 1,
                        padding: "6px 8px",
                        border: "1px solid #ddd",
                        borderRadius: 4,
                        fontSize: 14,
                      }}
                      placeholder="0"
                    />
                    <Button
                      onClick={() => handleClearField("brand_cap")}
                      style={{
                        padding: "6px 8px",
                        backgroundColor: "#6c757d",
                        color: "white",
                        border: "none",
                        borderRadius: 4,
                        fontSize: 12,
                        cursor: "pointer",
                      }}
                      title="Clear field"
                    >
                      Clear
                    </Button>
                  </div>
                  <div style={{ fontSize: 11, color: "#666", marginTop: 2 }}>
                    Maximum number of items from the same brand (0 = no limit)
                  </div>
                </div>
                <div style={{ flex: 1 }}>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontSize: 12,
                      fontWeight: "bold",
                    }}
                  >
                    Category Cap
                  </label>
                  <div style={{ display: "flex", gap: 4 }}>
                    <input
                      type="number"
                      min="0"
                      value={formatValue(editedPolicy.category_cap)}
                      onChange={(e) =>
                        handleFieldChange(
                          "category_cap",
                          parseInt(e.target.value) || 0
                        )
                      }
                      style={{
                        flex: 1,
                        padding: "6px 8px",
                        border: "1px solid #ddd",
                        borderRadius: 4,
                        fontSize: 14,
                      }}
                      placeholder="0"
                    />
                    <Button
                      onClick={() => handleClearField("category_cap")}
                      style={{
                        padding: "6px 8px",
                        backgroundColor: "#6c757d",
                        color: "white",
                        border: "none",
                        borderRadius: 4,
                        fontSize: 12,
                        cursor: "pointer",
                      }}
                      title="Clear field"
                    >
                      Clear
                    </Button>
                  </div>
                  <div style={{ fontSize: 11, color: "#666", marginTop: 2 }}>
                    Maximum number of items from the same category (0 = no
                    limit)
                  </div>
                </div>
              </Row>
            </div>
          </Label>
        </div>

        {/* Personalization and Filtering Parameters */}
        <div style={{ marginBottom: 20 }}>
          <Label text="Personalization and Filtering Parameters">
            <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
              <Row>
                <div style={{ flex: 1 }}>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontSize: 12,
                      fontWeight: "bold",
                    }}
                  >
                    Profile Boost
                  </label>
                  <div style={{ display: "flex", gap: 4 }}>
                    <input
                      type="number"
                      step="0.1"
                      min="0"
                      value={formatValue(editedPolicy.profile_boost)}
                      onChange={(e) =>
                        handleFieldChange(
                          "profile_boost",
                          parseFloat(e.target.value) || 0
                        )
                      }
                      style={{
                        flex: 1,
                        padding: "6px 8px",
                        border: "1px solid #ddd",
                        borderRadius: 4,
                        fontSize: 14,
                      }}
                      placeholder="0.0"
                    />
                    <Button
                      onClick={() => handleClearField("profile_boost")}
                      style={{
                        padding: "6px 8px",
                        backgroundColor: "#6c757d",
                        color: "white",
                        border: "none",
                        borderRadius: 4,
                        fontSize: 12,
                        cursor: "pointer",
                      }}
                      title="Clear field"
                    >
                      Clear
                    </Button>
                  </div>
                  <div style={{ fontSize: 11, color: "#666", marginTop: 2 }}>
                    Multiplier for personalization profile boost (≥0.0)
                  </div>
                </div>
                <div style={{ flex: 1 }}>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontSize: 12,
                      fontWeight: "bold",
                    }}
                  >
                    Half Life Days
                  </label>
                  <div style={{ display: "flex", gap: 4 }}>
                    <input
                      type="number"
                      step="0.1"
                      min="0"
                      value={formatValue(editedPolicy.half_life_days)}
                      onChange={(e) =>
                        handleFieldChange(
                          "half_life_days",
                          parseFloat(e.target.value) || 0
                        )
                      }
                      style={{
                        flex: 1,
                        padding: "6px 8px",
                        border: "1px solid #ddd",
                        borderRadius: 4,
                        fontSize: 14,
                      }}
                      placeholder="0.0"
                    />
                    <Button
                      onClick={() => handleClearField("half_life_days")}
                      style={{
                        padding: "6px 8px",
                        backgroundColor: "#6c757d",
                        color: "white",
                        border: "none",
                        borderRadius: 4,
                        fontSize: 12,
                        cursor: "pointer",
                      }}
                      title="Clear field"
                    >
                      Clear
                    </Button>
                  </div>
                  <div style={{ fontSize: 11, color: "#666", marginTop: 2 }}>
                    Popularity half-life in days (≥0.0)
                  </div>
                </div>
              </Row>

              <Row>
                <div style={{ flex: 1 }}>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontSize: 12,
                      fontWeight: "bold",
                    }}
                  >
                    Co-vis Window Days
                  </label>
                  <div style={{ display: "flex", gap: 4 }}>
                    <input
                      type="number"
                      min="0"
                      value={formatValue(editedPolicy.co_vis_window_days)}
                      onChange={(e) =>
                        handleFieldChange(
                          "co_vis_window_days",
                          parseInt(e.target.value) || 0
                        )
                      }
                      style={{
                        flex: 1,
                        padding: "6px 8px",
                        border: "1px solid #ddd",
                        borderRadius: 4,
                        fontSize: 14,
                      }}
                      placeholder="0"
                    />
                    <Button
                      onClick={() => handleClearField("co_vis_window_days")}
                      style={{
                        padding: "6px 8px",
                        backgroundColor: "#6c757d",
                        color: "white",
                        border: "none",
                        borderRadius: 4,
                        fontSize: 12,
                        cursor: "pointer",
                      }}
                      title="Clear field"
                    >
                      Clear
                    </Button>
                  </div>
                  <div style={{ fontSize: 11, color: "#666", marginTop: 2 }}>
                    Co-visitation window in days (≥0)
                  </div>
                </div>
                <div style={{ flex: 1 }}>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontSize: 12,
                      fontWeight: "bold",
                    }}
                  >
                    Popularity Fanout
                  </label>
                  <div style={{ display: "flex", gap: 4 }}>
                    <input
                      type="number"
                      min="0"
                      value={formatValue(editedPolicy.popularity_fanout)}
                      onChange={(e) =>
                        handleFieldChange(
                          "popularity_fanout",
                          parseInt(e.target.value) || 0
                        )
                      }
                      style={{
                        flex: 1,
                        padding: "6px 8px",
                        border: "1px solid #ddd",
                        borderRadius: 4,
                        fontSize: 14,
                      }}
                      placeholder="0"
                    />
                    <Button
                      onClick={() => handleClearField("popularity_fanout")}
                      style={{
                        padding: "6px 8px",
                        backgroundColor: "#6c757d",
                        color: "white",
                        border: "none",
                        borderRadius: 4,
                        fontSize: 12,
                        cursor: "pointer",
                      }}
                      title="Clear field"
                    >
                      Clear
                    </Button>
                  </div>
                  <div style={{ fontSize: 11, color: "#666", marginTop: 2 }}>
                    Fanout for popularity candidates (≥0)
                  </div>
                </div>
              </Row>

              <Row>
                <div style={{ flex: 1 }}>
                  <label
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 8,
                      fontSize: 14,
                      marginTop: 20,
                    }}
                  >
                    <input
                      type="checkbox"
                      checked={editedPolicy.rule_exclude_purchased || false}
                      onChange={(e) =>
                        handleFieldChange(
                          "rule_exclude_purchased",
                          e.target.checked
                        )
                      }
                      style={{ margin: 0 }}
                    />
                    Exclude Purchased Rule
                  </label>
                  <div style={{ fontSize: 11, color: "#666", marginTop: 2 }}>
                    Whether to exclude items the user has already purchased
                  </div>
                </div>
              </Row>
            </div>
          </Label>
        </div>

        {/* Notes */}
        <div style={{ marginBottom: 20 }}>
          <Label text="Notes">
            <textarea
              value={editedPolicy.notes || ""}
              onChange={(e) => handleFieldChange("notes", e.target.value)}
              style={{
                width: "100%",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: 4,
                fontSize: 14,
                minHeight: 80,
                resize: "vertical",
              }}
              placeholder="Optional notes about this policy..."
            />
          </Label>
        </div>

        {/* Action Buttons */}
        <div style={{ display: "flex", gap: 8, justifyContent: "flex-end" }}>
          <Button
            onClick={onCancel}
            style={{
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              padding: "8px 16px",
              borderRadius: 4,
              cursor: "pointer",
              fontSize: 14,
            }}
          >
            Cancel
          </Button>
          <Button
            onClick={handleCreateCustom}
            style={{
              backgroundColor: "#ffc107",
              color: "#212529",
              border: "none",
              padding: "8px 16px",
              borderRadius: 4,
              cursor: "pointer",
              fontSize: 14,
            }}
          >
            Save as Template
          </Button>
          <Button
            onClick={handleSave}
            style={{
              backgroundColor: "#28a745",
              color: "white",
              border: "none",
              padding: "8px 16px",
              borderRadius: 4,
              cursor: "pointer",
              fontSize: 14,
            }}
          >
            Save Policy
          </Button>
        </div>
      </div>
    </Section>
  );
}
