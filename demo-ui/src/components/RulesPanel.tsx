import React, { useState, useEffect } from "react";
import {
  AdminService,
  type specs_types_RuleResponse,
  type specs_types_RulePayload,
  type specs_types_RuleDryRunRequest,
  type specs_types_RuleDryRunResponse,
  type specs_types_RuleMatchResponse,
  type specs_types_RuleItemEffect,
} from "../lib/api-client";
import { Button, Section, Th, Td } from "./UIComponents";

interface RulesPanelProps {
  namespace: string;
}

interface RuleFormData {
  name: string;
  description: string;
  action: "BLOCK" | "PIN" | "BOOST";
  target_type: "ITEM" | "TAG" | "BRAND" | "CATEGORY";
  target_key: string;
  item_ids: string[];
  boost_value: number;
  max_pins: number;
  segment_id: string;
  priority: number;
  enabled: boolean;
  valid_from: string;
  valid_until: string;
}

interface DryRunFormData {
  surface: string;
  segment_id: string;
  items: string[];
}

export function RulesPanel({ namespace }: RulesPanelProps) {
  const [rules, setRules] = useState<specs_types_RuleResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingRule, setEditingRule] =
    useState<specs_types_RuleResponse | null>(null);
  const [showDryRun, setShowDryRun] = useState(false);
  const [dryRunResult, setDryRunResult] =
    useState<specs_types_RuleDryRunResponse | null>(null);
  const [dryRunLoading, setDryRunLoading] = useState(false);

  const [formData, setFormData] = useState<RuleFormData>({
    name: "",
    description: "",
    action: "BOOST",
    target_type: "ITEM",
    target_key: "",
    item_ids: [],
    boost_value: 0.1,
    max_pins: 3,
    segment_id: "",
    priority: 0,
    enabled: true,
    valid_from: "",
    valid_until: "",
  });

  const [dryRunFormData, setDryRunFormData] = useState<DryRunFormData>({
    surface: "home",
    segment_id: "",
    items: [],
  });

  useEffect(() => {
    loadRules();
  }, [namespace]);

  const loadRules = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await AdminService.getV1AdminRules(namespace);
      setRules(response.rules || []);
    } catch (err) {
      console.error("Failed to load rules:", err);
      setError(err instanceof Error ? err.message : "Failed to load rules");
    } finally {
      setLoading(false);
    }
  };

  const handleCreateRule = async () => {
    try {
      const payload: specs_types_RulePayload = {
        namespace,
        surface: "home", // Default surface for now
        name: formData.name,
        description: formData.description,
        action: formData.action,
        target_type: formData.target_type,
        target_key: formData.target_key || undefined,
        item_ids: formData.item_ids.length > 0 ? formData.item_ids : undefined,
        boost_value:
          formData.action === "BOOST" ? formData.boost_value : undefined,
        max_pins: formData.action === "PIN" ? formData.max_pins : undefined,
        segment_id: formData.segment_id || undefined,
        priority: formData.priority,
        enabled: formData.enabled,
        valid_from: formData.valid_from || undefined,
        valid_until: formData.valid_until || undefined,
      };

      await AdminService.rulesCreate(payload);
      setShowCreateForm(false);
      resetForm();
      loadRules();
    } catch (err) {
      console.error("Failed to create rule:", err);
      setError(err instanceof Error ? err.message : "Failed to create rule");
    }
  };

  const handleUpdateRule = async () => {
    if (!editingRule) return;

    try {
      const payload: specs_types_RulePayload = {
        namespace,
        surface: editingRule.surface,
        name: formData.name,
        description: formData.description,
        action: formData.action,
        target_type: formData.target_type,
        target_key: formData.target_key || undefined,
        item_ids: formData.item_ids.length > 0 ? formData.item_ids : undefined,
        boost_value:
          formData.action === "BOOST" ? formData.boost_value : undefined,
        max_pins: formData.action === "PIN" ? formData.max_pins : undefined,
        segment_id: formData.segment_id || undefined,
        priority: formData.priority,
        enabled: formData.enabled,
        valid_from: formData.valid_from || undefined,
        valid_until: formData.valid_until || undefined,
      };

      await AdminService.rulesUpdate(editingRule.rule_id, payload);
      setEditingRule(null);
      resetForm();
      loadRules();
    } catch (err) {
      console.error("Failed to update rule:", err);
      setError(err instanceof Error ? err.message : "Failed to update rule");
    }
  };

  const handleDryRun = async () => {
    setDryRunLoading(true);
    setError(null);
    try {
      const payload: specs_types_RuleDryRunRequest = {
        namespace,
        surface: dryRunFormData.surface,
        segment_id: dryRunFormData.segment_id || undefined,
        items: dryRunFormData.items,
      };

      const response = await AdminService.rulesDryRun(payload);
      setDryRunResult(response);
    } catch (err) {
      console.error("Failed to run dry-run:", err);
      setError(err instanceof Error ? err.message : "Failed to run dry-run");
    } finally {
      setDryRunLoading(false);
    }
  };

  const resetForm = () => {
    setFormData({
      name: "",
      description: "",
      action: "BOOST",
      target_type: "ITEM",
      target_key: "",
      item_ids: [],
      boost_value: 0.1,
      max_pins: 3,
      segment_id: "",
      priority: 0,
      enabled: true,
      valid_from: "",
      valid_until: "",
    });
  };

  const startEdit = (rule: specs_types_RuleResponse) => {
    setEditingRule(rule);
    setFormData({
      name: rule.name,
      description: rule.description || "",
      action: rule.action as "BLOCK" | "PIN" | "BOOST",
      target_type: rule.target_type as "ITEM" | "TAG" | "BRAND" | "CATEGORY",
      target_key: rule.target_key || "",
      item_ids: rule.item_ids || [],
      boost_value: rule.boost_value || 0.1,
      max_pins: rule.max_pins || 3,
      segment_id: rule.segment_id || "",
      priority: rule.priority,
      enabled: rule.enabled,
      valid_from: rule.valid_from || "",
      valid_until: rule.valid_until || "",
    });
  };

  const formatTTL = (validFrom?: string, validUntil?: string) => {
    if (!validFrom && !validUntil) return "Always";
    if (validFrom && validUntil) {
      return `${new Date(validFrom).toLocaleDateString()} - ${new Date(
        validUntil
      ).toLocaleDateString()}`;
    }
    if (validFrom) return `From ${new Date(validFrom).toLocaleDateString()}`;
    if (validUntil) return `Until ${new Date(validUntil).toLocaleDateString()}`;
    return "Always";
  };

  const formatTarget = (rule: specs_types_RuleResponse) => {
    if (rule.target_type === "ITEM") {
      return rule.item_ids?.join(", ") || "No items";
    }
    return rule.target_key || "No key";
  };

  const getActionColor = (action: string) => {
    switch (action) {
      case "BLOCK":
        return "#dc3545";
      case "PIN":
        return "#ffc107";
      case "BOOST":
        return "#28a745";
      default:
        return "#6c757d";
    }
  };

  return (
    <Section title="Rule Engine v1">
      <p style={{ color: "#666", marginBottom: 16, fontSize: "14px" }}>
        Manage PIN/BLOCK/BOOST rules with TTL and precedence. Rules are
        evaluated at request time and can be scoped by namespace, surface, and
        segment.
      </p>

      {error && (
        <div
          style={{
            marginBottom: 16,
            padding: 12,
            backgroundColor: "#fee",
            border: "1px solid #fcc",
            borderRadius: 4,
            color: "#c33",
          }}
        >
          Error: {error}
        </div>
      )}

      <div style={{ display: "flex", gap: 8, marginBottom: 16 }}>
        <Button
          onClick={() => setShowCreateForm(true)}
          style={{ backgroundColor: "#28a745", color: "white" }}
        >
          Create Rule
        </Button>
        <Button
          onClick={() => setShowDryRun(true)}
          style={{ backgroundColor: "#17a2b8", color: "white" }}
        >
          Dry Run
        </Button>
        <Button onClick={loadRules} disabled={loading}>
          {loading ? "Loading..." : "Refresh"}
        </Button>
      </div>

      {/* Rules Table */}
      <div style={{ overflowX: "auto", marginBottom: 24 }}>
        <table
          style={{
            borderCollapse: "collapse",
            width: "100%",
            minWidth: 800,
          }}
        >
          <thead>
            <tr>
              <Th>Name</Th>
              <Th>Action</Th>
              <Th>Target</Th>
              <Th>Priority</Th>
              <Th>TTL</Th>
              <Th>Enabled</Th>
              <Th>Actions</Th>
            </tr>
          </thead>
          <tbody>
            {rules.map((rule) => (
              <tr key={rule.rule_id}>
                <Td>
                  <div>
                    <strong>{rule.name}</strong>
                    {rule.description && (
                      <div style={{ fontSize: "12px", color: "#666" }}>
                        {rule.description}
                      </div>
                    )}
                  </div>
                </Td>
                <Td>
                  <span
                    style={{
                      backgroundColor: getActionColor(rule.action),
                      color: "white",
                      padding: "2px 8px",
                      borderRadius: 4,
                      fontSize: "12px",
                      fontWeight: "bold",
                    }}
                  >
                    {rule.action}
                  </span>
                </Td>
                <Td>
                  <div style={{ fontSize: "12px" }}>
                    <strong>{rule.target_type}:</strong> {formatTarget(rule)}
                  </div>
                  {rule.boost_value && (
                    <div style={{ fontSize: "12px", color: "#28a745" }}>
                      +{rule.boost_value}
                    </div>
                  )}
                </Td>
                <Td>{rule.priority}</Td>
                <Td style={{ fontSize: "12px" }}>
                  {formatTTL(rule.valid_from, rule.valid_until)}
                </Td>
                <Td>
                  <span
                    style={{
                      color: rule.enabled ? "#28a745" : "#dc3545",
                      fontWeight: "bold",
                    }}
                  >
                    {rule.enabled ? "✓" : "✗"}
                  </span>
                </Td>
                <Td>
                  <Button
                    onClick={() => startEdit(rule)}
                    style={{ padding: "4px 8px", fontSize: "12px" }}
                  >
                    Edit
                  </Button>
                </Td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Create/Edit Form Modal */}
      {(showCreateForm || editingRule) && (
        <div
          style={{
            position: "fixed",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: "rgba(0,0,0,0.5)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 1000,
          }}
        >
          <div
            style={{
              backgroundColor: "white",
              padding: 24,
              borderRadius: 8,
              width: "90%",
              maxWidth: 600,
              maxHeight: "90vh",
              overflowY: "auto",
            }}
          >
            <h3 style={{ margin: "0 0 16px 0" }}>
              {editingRule ? "Edit Rule" : "Create Rule"}
            </h3>

            <div style={{ display: "grid", gap: 12, marginBottom: 16 }}>
              <div>
                <label
                  style={{
                    display: "block",
                    marginBottom: 4,
                    fontWeight: "bold",
                  }}
                >
                  Name *
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                  style={{
                    width: "100%",
                    padding: 8,
                    border: "1px solid #ddd",
                    borderRadius: 4,
                  }}
                />
              </div>

              <div>
                <label
                  style={{
                    display: "block",
                    marginBottom: 4,
                    fontWeight: "bold",
                  }}
                >
                  Description
                </label>
                <input
                  type="text"
                  value={formData.description}
                  onChange={(e) =>
                    setFormData({ ...formData, description: e.target.value })
                  }
                  style={{
                    width: "100%",
                    padding: 8,
                    border: "1px solid #ddd",
                    borderRadius: 4,
                  }}
                />
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: 12,
                }}
              >
                <div>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontWeight: "bold",
                    }}
                  >
                    Action *
                  </label>
                  <select
                    value={formData.action}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        action: e.target.value as any,
                      })
                    }
                    style={{
                      width: "100%",
                      padding: 8,
                      border: "1px solid #ddd",
                      borderRadius: 4,
                    }}
                  >
                    <option value="BLOCK">BLOCK</option>
                    <option value="PIN">PIN</option>
                    <option value="BOOST">BOOST</option>
                  </select>
                </div>

                <div>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontWeight: "bold",
                    }}
                  >
                    Target Type *
                  </label>
                  <select
                    value={formData.target_type}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        target_type: e.target.value as any,
                      })
                    }
                    style={{
                      width: "100%",
                      padding: 8,
                      border: "1px solid #ddd",
                      borderRadius: 4,
                    }}
                  >
                    <option value="ITEM">ITEM</option>
                    <option value="TAG">TAG</option>
                    <option value="BRAND">BRAND</option>
                    <option value="CATEGORY">CATEGORY</option>
                  </select>
                </div>
              </div>

              {formData.target_type === "ITEM" ? (
                <div>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontWeight: "bold",
                    }}
                  >
                    Item IDs (comma-separated)
                  </label>
                  <input
                    type="text"
                    value={formData.item_ids.join(", ")}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        item_ids: e.target.value
                          .split(",")
                          .map((s) => s.trim())
                          .filter((s) => s),
                      })
                    }
                    style={{
                      width: "100%",
                      padding: 8,
                      border: "1px solid #ddd",
                      borderRadius: 4,
                    }}
                    placeholder="item1, item2, item3"
                  />
                </div>
              ) : (
                <div>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontWeight: "bold",
                    }}
                  >
                    Target Key
                  </label>
                  <input
                    type="text"
                    value={formData.target_key}
                    onChange={(e) =>
                      setFormData({ ...formData, target_key: e.target.value })
                    }
                    style={{
                      width: "100%",
                      padding: 8,
                      border: "1px solid #ddd",
                      borderRadius: 4,
                    }}
                    placeholder="e.g., 'new', 'premium', 'action'"
                  />
                </div>
              )}

              {formData.action === "BOOST" && (
                <div>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontWeight: "bold",
                    }}
                  >
                    Boost Value
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.boost_value}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        boost_value: parseFloat(e.target.value),
                      })
                    }
                    style={{
                      width: "100%",
                      padding: 8,
                      border: "1px solid #ddd",
                      borderRadius: 4,
                    }}
                  />
                </div>
              )}

              {formData.action === "PIN" && (
                <div>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontWeight: "bold",
                    }}
                  >
                    Max Pins
                  </label>
                  <input
                    type="number"
                    value={formData.max_pins}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        max_pins: parseInt(e.target.value),
                      })
                    }
                    style={{
                      width: "100%",
                      padding: 8,
                      border: "1px solid #ddd",
                      borderRadius: 4,
                    }}
                  />
                </div>
              )}

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: 12,
                }}
              >
                <div>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontWeight: "bold",
                    }}
                  >
                    Priority
                  </label>
                  <input
                    type="number"
                    value={formData.priority}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        priority: parseInt(e.target.value),
                      })
                    }
                    style={{
                      width: "100%",
                      padding: 8,
                      border: "1px solid #ddd",
                      borderRadius: 4,
                    }}
                  />
                </div>

                <div>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontWeight: "bold",
                    }}
                  >
                    Segment ID
                  </label>
                  <input
                    type="text"
                    value={formData.segment_id}
                    onChange={(e) =>
                      setFormData({ ...formData, segment_id: e.target.value })
                    }
                    style={{
                      width: "100%",
                      padding: 8,
                      border: "1px solid #ddd",
                      borderRadius: 4,
                    }}
                    placeholder="Optional"
                  />
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: 12,
                }}
              >
                <div>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontWeight: "bold",
                    }}
                  >
                    Valid From
                  </label>
                  <input
                    type="datetime-local"
                    value={formData.valid_from}
                    onChange={(e) =>
                      setFormData({ ...formData, valid_from: e.target.value })
                    }
                    style={{
                      width: "100%",
                      padding: 8,
                      border: "1px solid #ddd",
                      borderRadius: 4,
                    }}
                  />
                </div>

                <div>
                  <label
                    style={{
                      display: "block",
                      marginBottom: 4,
                      fontWeight: "bold",
                    }}
                  >
                    Valid Until
                  </label>
                  <input
                    type="datetime-local"
                    value={formData.valid_until}
                    onChange={(e) =>
                      setFormData({ ...formData, valid_until: e.target.value })
                    }
                    style={{
                      width: "100%",
                      padding: 8,
                      border: "1px solid #ddd",
                      borderRadius: 4,
                    }}
                  />
                </div>
              </div>

              <div>
                <label
                  style={{ display: "flex", alignItems: "center", gap: 8 }}
                >
                  <input
                    type="checkbox"
                    checked={formData.enabled}
                    onChange={(e) =>
                      setFormData({ ...formData, enabled: e.target.checked })
                    }
                  />
                  <span style={{ fontWeight: "bold" }}>Enabled</span>
                </label>
              </div>
            </div>

            <div
              style={{ display: "flex", gap: 8, justifyContent: "flex-end" }}
            >
              <Button
                onClick={() => {
                  setShowCreateForm(false);
                  setEditingRule(null);
                  resetForm();
                }}
                style={{ backgroundColor: "#6c757d", color: "white" }}
              >
                Cancel
              </Button>
              <Button
                onClick={editingRule ? handleUpdateRule : handleCreateRule}
                style={{ backgroundColor: "#28a745", color: "white" }}
                disabled={!formData.name}
              >
                {editingRule ? "Update" : "Create"}
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Dry Run Modal */}
      {showDryRun && (
        <div
          style={{
            position: "fixed",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: "rgba(0,0,0,0.5)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 1000,
          }}
        >
          <div
            style={{
              backgroundColor: "white",
              padding: 24,
              borderRadius: 8,
              width: "90%",
              maxWidth: 800,
              maxHeight: "90vh",
              overflowY: "auto",
            }}
          >
            <h3 style={{ margin: "0 0 16px 0" }}>Rule Dry Run</h3>
            <p style={{ color: "#666", marginBottom: 16, fontSize: "14px" }}>
              Test which rules would fire for a given context and see the
              resulting effects.
            </p>

            <div style={{ display: "grid", gap: 12, marginBottom: 16 }}>
              <div>
                <label
                  style={{
                    display: "block",
                    marginBottom: 4,
                    fontWeight: "bold",
                  }}
                >
                  Surface *
                </label>
                <input
                  type="text"
                  value={dryRunFormData.surface}
                  onChange={(e) =>
                    setDryRunFormData({
                      ...dryRunFormData,
                      surface: e.target.value,
                    })
                  }
                  style={{
                    width: "100%",
                    padding: 8,
                    border: "1px solid #ddd",
                    borderRadius: 4,
                  }}
                  placeholder="e.g., home, gamepage"
                />
              </div>

              <div>
                <label
                  style={{
                    display: "block",
                    marginBottom: 4,
                    fontWeight: "bold",
                  }}
                >
                  Segment ID
                </label>
                <input
                  type="text"
                  value={dryRunFormData.segment_id}
                  onChange={(e) =>
                    setDryRunFormData({
                      ...dryRunFormData,
                      segment_id: e.target.value,
                    })
                  }
                  style={{
                    width: "100%",
                    padding: 8,
                    border: "1px solid #ddd",
                    borderRadius: 4,
                  }}
                  placeholder="Optional"
                />
              </div>

              <div>
                <label
                  style={{
                    display: "block",
                    marginBottom: 4,
                    fontWeight: "bold",
                  }}
                >
                  Candidate Item IDs *
                </label>
                <input
                  type="text"
                  value={dryRunFormData.items.join(", ")}
                  onChange={(e) =>
                    setDryRunFormData({
                      ...dryRunFormData,
                      items: e.target.value
                        .split(",")
                        .map((s) => s.trim())
                        .filter((s) => s),
                    })
                  }
                  style={{
                    width: "100%",
                    padding: 8,
                    border: "1px solid #ddd",
                    borderRadius: 4,
                  }}
                  placeholder="item1, item2, item3"
                />
              </div>
            </div>

            <div style={{ display: "flex", gap: 8, marginBottom: 16 }}>
              <Button
                onClick={handleDryRun}
                disabled={
                  dryRunLoading ||
                  !dryRunFormData.surface ||
                  dryRunFormData.items.length === 0
                }
                style={{ backgroundColor: "#17a2b8", color: "white" }}
              >
                {dryRunLoading ? "Running..." : "Run Dry Run"}
              </Button>
              <Button
                onClick={() => setShowDryRun(false)}
                style={{ backgroundColor: "#6c757d", color: "white" }}
              >
                Close
              </Button>
            </div>

            {/* Dry Run Results */}
            {dryRunResult && (
              <div>
                <h4 style={{ margin: "0 0 12px 0" }}>Dry Run Results</h4>

                {/* Matched Rules */}
                {dryRunResult.matched_rules &&
                  dryRunResult.matched_rules.length > 0 && (
                    <div style={{ marginBottom: 16 }}>
                      <h5 style={{ margin: "0 0 8px 0" }}>
                        Matched Rules ({dryRunResult.matched_rules.length})
                      </h5>
                      <div style={{ overflowX: "auto" }}>
                        <table
                          style={{
                            borderCollapse: "collapse",
                            width: "100%",
                            fontSize: "12px",
                          }}
                        >
                          <thead>
                            <tr>
                              <Th>Rule</Th>
                              <Th>Action</Th>
                              <Th>Target</Th>
                              <Th>Priority</Th>
                              <Th>Affected Items</Th>
                            </tr>
                          </thead>
                          <tbody>
                            {dryRunResult.matched_rules.map((rule, index) => (
                              <tr key={index}>
                                <Td>
                                  <div>
                                    <strong>{rule.name}</strong>
                                    <div
                                      style={{
                                        fontSize: "11px",
                                        color: "#666",
                                      }}
                                    >
                                      {rule.rule_id}
                                    </div>
                                  </div>
                                </Td>
                                <Td>
                                  <span
                                    style={{
                                      backgroundColor: getActionColor(
                                        rule.action
                                      ),
                                      color: "white",
                                      padding: "2px 6px",
                                      borderRadius: 3,
                                      fontSize: "11px",
                                      fontWeight: "bold",
                                    }}
                                  >
                                    {rule.action}
                                  </span>
                                </Td>
                                <Td>
                                  <div style={{ fontSize: "11px" }}>
                                    <strong>{rule.target_type}:</strong>{" "}
                                    {rule.target_key ||
                                      rule.item_ids?.join(", ")}
                                  </div>
                                  {rule.boost_value && (
                                    <div
                                      style={{
                                        fontSize: "11px",
                                        color: "#28a745",
                                      }}
                                    >
                                      +{rule.boost_value}
                                    </div>
                                  )}
                                </Td>
                                <Td>{rule.priority}</Td>
                                <Td style={{ fontSize: "11px" }}>
                                  {rule.affected_item_ids?.join(", ") || "None"}
                                </Td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}

                {/* Item Effects */}
                {dryRunResult.item_effects &&
                  Object.keys(dryRunResult.item_effects).length > 0 && (
                    <div>
                      <h5 style={{ margin: "0 0 8px 0" }}>Item Effects</h5>
                      <div style={{ overflowX: "auto" }}>
                        <table
                          style={{
                            borderCollapse: "collapse",
                            width: "100%",
                            fontSize: "12px",
                          }}
                        >
                          <thead>
                            <tr>
                              <Th>Item ID</Th>
                              <Th>Blocked</Th>
                              <Th>Pinned</Th>
                              <Th>Boost Delta</Th>
                              <Th>Matched Rules</Th>
                            </tr>
                          </thead>
                          <tbody>
                            {Object.entries(dryRunResult.item_effects).map(
                              ([itemId, effect]) => (
                                <tr key={itemId}>
                                  <Td mono>{itemId}</Td>
                                  <Td>
                                    <span
                                      style={{
                                        color: effect.blocked
                                          ? "#dc3545"
                                          : "#28a745",
                                      }}
                                    >
                                      {effect.blocked ? "✓" : "✗"}
                                    </span>
                                  </Td>
                                  <Td>
                                    <span
                                      style={{
                                        color: effect.pinned
                                          ? "#ffc107"
                                          : "#6c757d",
                                      }}
                                    >
                                      {effect.pinned ? "✓" : "✗"}
                                    </span>
                                  </Td>
                                  <Td>
                                    {effect.boost_delta !== 0 ? (
                                      <span
                                        style={{
                                          color:
                                            effect.boost_delta > 0
                                              ? "#28a745"
                                              : "#dc3545",
                                        }}
                                      >
                                        {effect.boost_delta > 0 ? "+" : ""}
                                        {effect.boost_delta}
                                      </span>
                                    ) : (
                                      "0"
                                    )}
                                  </Td>
                                  <Td style={{ fontSize: "11px" }}>
                                    {effect.matched_rules?.join(", ") || "None"}
                                  </Td>
                                </tr>
                              )
                            )}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}

                {(!dryRunResult.matched_rules ||
                  dryRunResult.matched_rules.length === 0) &&
                  (!dryRunResult.item_effects ||
                    Object.keys(dryRunResult.item_effects).length === 0) && (
                    <div
                      style={{
                        padding: 16,
                        backgroundColor: "#f8f9fa",
                        border: "1px solid #e9ecef",
                        borderRadius: 4,
                        textAlign: "center",
                        color: "#666",
                      }}
                    >
                      No rules matched for the given context.
                    </div>
                  )}
              </div>
            )}
          </div>
        </div>
      )}
    </Section>
  );
}


