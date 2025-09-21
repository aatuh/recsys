import React, { useState, useEffect } from "react";
import { Button, Section, Th, Td } from "../primitives/UIComponents";
import { RuleFormModal } from "../rules/RuleFormModal";
import { RuleDryRunModal } from "../rules/RuleDryRunModal";

// Placeholder types until API client is regenerated
type specs_types_RuleResponse = any;
type specs_types_RulePayload = any;
type specs_types_RuleDryRunRequest = any;
type specs_types_RuleDryRunResponse = any;

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
      // TODO: Uncomment when API client is regenerated
      // const response = await AdminService.getV1AdminRules(namespace);
      // setRules(response.rules || []);
      setRules([]); // Placeholder
    } catch (err) {
      console.error("Failed to load rules:", err);
      setError(err instanceof Error ? err.message : "Failed to load rules");
    } finally {
      setLoading(false);
    }
  };

  const handleCreateRule = async () => {
    try {
      const _payload: specs_types_RulePayload = {
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

      // TODO: Uncomment when API client is regenerated
      // await AdminService.rulesCreate(payload);
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
      const _payload: specs_types_RulePayload = {
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

      // TODO: Uncomment when API client is regenerated
      // await AdminService.rulesUpdate(editingRule.rule_id, payload);
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
      const _payload: specs_types_RuleDryRunRequest = {
        namespace,
        surface: dryRunFormData.surface,
        segment_id: dryRunFormData.segment_id || undefined,
        items: dryRunFormData.items,
      };

      // TODO: Uncomment when API client is regenerated
      // const response = await AdminService.rulesDryRun(payload);
      const response = null; // Placeholder
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
                <Td>{formatTTL(rule.valid_from, rule.valid_until)}</Td>
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
      <RuleFormModal
        open={showCreateForm || Boolean(editingRule)}
        editing={Boolean(editingRule)}
        formData={formData}
        setFormData={setFormData}
        onCancel={() => {
          setShowCreateForm(false);
          setEditingRule(null);
          resetForm();
        }}
        onSubmit={editingRule ? handleUpdateRule : handleCreateRule}
      />

      {/* Dry Run Modal */}
      <RuleDryRunModal
        open={showDryRun}
        loading={dryRunLoading}
        data={dryRunResult}
        form={dryRunFormData}
        setForm={setDryRunFormData}
        onRun={handleDryRun}
        onClose={() => setShowDryRun(false)}
      />
    </Section>
  );
}
