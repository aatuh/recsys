import { useState, useEffect, useRef } from "react";
import { Section, Row, Label, Button } from "./UIComponents";
import { BanditService } from "../lib/api-client";
import { PolicyEditor } from "./PolicyEditor";
import { ALGORITHM_PROFILES } from "../types/algorithmProfiles";

// Define BanditPolicy type locally since it's not exported from api-client
interface types_BanditPolicy {
  policy_id?: string;
  name: string;
  notes?: string;
  active: boolean;
  blend_alpha?: number;
  blend_beta?: number;
  blend_gamma?: number;
  mmr_lambda?: number;
  brand_cap?: number;
  category_cap?: number;
  profile_boost?: number;
  rule_exclude_events?: boolean;
  half_life_days?: number;
  co_vis_window_days?: number;
  popularity_fanout?: number;
  [key: string]: any; // Allow additional properties
}

interface PolicyManagerProps {
  namespace: string;
  onPoliciesChange?: (value: types_BanditPolicy[]) => void;
}

interface CustomPolicy {
  id: string;
  name: string;
  description: string;
  policy: types_BanditPolicy;
}

export function PolicyManager({
  namespace,
  onPoliciesChange,
}: PolicyManagerProps) {
  const [policies, setPolicies] = useState<types_BanditPolicy[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [customPolicies, setCustomPolicies] = useState<CustomPolicy[]>([]);
  const [editingPolicy, setEditingPolicy] = useState<types_BanditPolicy | null>(
    null
  );
  const [isEditing, setIsEditing] = useState(false);
  const [highlightedPolicyId, setHighlightedPolicyId] = useState<string | null>(
    null
  );
  const prevCustomPoliciesRef = useRef<string>("");
  const hasLoadedFromStorageRef = useRef<boolean>(false);

  // Convert algorithm profile to bandit policy
  const convertProfileToPolicy = (
    profile: (typeof ALGORITHM_PROFILES)[0]
  ): types_BanditPolicy => {
    return {
      policy_id: `template-${profile.id}`,
      name: profile.name,
      active: true,
      blend_alpha: profile.overrides.blend_alpha || 0.0,
      blend_beta: profile.overrides.blend_beta || 0.0,
      blend_gamma: profile.overrides.blend_gamma || 0.0,
      mmr_lambda: profile.overrides.mmr_lambda || 0.0,
      brand_cap: profile.overrides.brand_cap || 0,
      category_cap: profile.overrides.category_cap || 0,
      profile_boost: profile.overrides.profile_boost || 0.0,
      rule_exclude_events: profile.overrides.rule_exclude_events || false,
      half_life_days: profile.overrides.popularity_halflife_days || 0.0,
      co_vis_window_days: profile.overrides.covis_window_days || 0,
      popularity_fanout: profile.overrides.popularity_fanout || 0,
      notes: profile.description,
    };
  };

  // Create predefined policy templates from algorithm profiles
  const createPredefinedTemplates = (): CustomPolicy[] => {
    return ALGORITHM_PROFILES.map((profile) => ({
      id: `predefined-${profile.id}`,
      name: profile.name,
      description: profile.description,
      policy: convertProfileToPolicy(profile),
    }));
  };

  // Load custom policies from localStorage on mount
  useEffect(() => {
    if (hasLoadedFromStorageRef.current) {
      return; // Only load once
    }

    const predefinedTemplates = createPredefinedTemplates();
    let userCustomPolicies: CustomPolicy[] = [];

    const saved = localStorage.getItem("recsys-custom-bandit-policies");
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        userCustomPolicies = parsed;
      } catch (e) {
        console.warn("Failed to load custom policies from localStorage:", e);
      }
    }

    // Merge predefined templates with user-created custom policies
    const allCustomPolicies = [...predefinedTemplates, ...userCustomPolicies];
    setCustomPolicies(allCustomPolicies);
    prevCustomPoliciesRef.current = JSON.stringify(userCustomPolicies); // Only save user-created ones

    hasLoadedFromStorageRef.current = true;
  }, []); // Empty dependency array - only run once on mount

  // Save custom policies to localStorage whenever they change
  useEffect(() => {
    // Only save user-created custom policies (filter out predefined ones)
    const userCustomPolicies = customPolicies.filter(
      (policy) => !policy.id.startsWith("predefined-")
    );
    const currentPoliciesString = JSON.stringify(userCustomPolicies);

    // Only save if the user-created policies have actually changed
    if (currentPoliciesString !== prevCustomPoliciesRef.current) {
      prevCustomPoliciesRef.current = currentPoliciesString;
      localStorage.setItem(
        "recsys-custom-bandit-policies",
        currentPoliciesString
      );
    }
  }, [customPolicies]);

  // Load policies from API
  const loadPolicies = async () => {
    setLoading(true);
    setError(null);
    try {
      console.log(`Loading policies for namespace: ${namespace}`);
      const response = await BanditService.getV1BanditPolicies(namespace);
      console.log("Policies response:", response);
      console.log("Response type:", typeof response);
      console.log("Is array:", Array.isArray(response));

      // Handle different response formats (array | json-string | wrapped)
      let policiesArray: types_BanditPolicy[] = [];

      if (Array.isArray(response)) {
        policiesArray = response as types_BanditPolicy[];
      } else if (typeof response === "string") {
        try {
          const parsed = JSON.parse(response);
          if (Array.isArray(parsed)) {
            policiesArray = parsed as types_BanditPolicy[];
          } else if (parsed && typeof parsed === "object") {
            if (Array.isArray((parsed as any).data)) {
              policiesArray = (parsed as any).data as types_BanditPolicy[];
            } else if (Array.isArray((parsed as any).policies)) {
              policiesArray = (parsed as any).policies as types_BanditPolicy[];
            } else if (Array.isArray((parsed as any).results)) {
              policiesArray = (parsed as any).results as types_BanditPolicy[];
            }
          }
        } catch (e) {
          console.warn("Failed to parse policies JSON string:", e);
        }
      } else if (response && typeof response === "object") {
        // Check if response has a data property or similar
        const anyResp = response as any;
        if (Array.isArray(anyResp.data)) {
          policiesArray = anyResp.data as types_BanditPolicy[];
        } else if (Array.isArray(anyResp.policies)) {
          policiesArray = anyResp.policies as types_BanditPolicy[];
        } else if (Array.isArray(anyResp.results)) {
          policiesArray = anyResp.results as types_BanditPolicy[];
        }
      }

      console.log(`Loaded ${policiesArray.length} policies`);
      console.log("Policies array:", policiesArray);
      setPolicies(policiesArray);

      // Notify parent component of policy changes
      if (onPoliciesChange) {
        onPoliciesChange(policiesArray);
      }
    } catch (err) {
      console.error("Failed to load bandit policies:", err);
      setError(
        err instanceof Error
          ? err.message
          : "Failed to load policies. Make sure the API server is running."
      );
      // Set empty array on error to prevent map errors
      setPolicies([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPolicies();
  }, [namespace]);

  // Debug: Log whenever policies state changes
  useEffect(() => {
    console.log("Policies state changed:", policies);
    console.log("Policies state length:", policies?.length);
  }, [policies]);

  const handleCreateNewPolicy = () => {
    const newPolicy: types_BanditPolicy = {
      policy_id: `policy-${Date.now()}`,
      name: "New Policy",
      active: true,
      blend_alpha: 0.0,
      blend_beta: 0.0,
      blend_gamma: 0.0,
      mmr_lambda: 0.0,
      brand_cap: 0,
      category_cap: 0,
      profile_boost: 0.0,
      rule_exclude_events: false,
      half_life_days: 0.0,
      co_vis_window_days: 0,
      popularity_fanout: 0,
      notes: "",
    };
    setEditingPolicy(newPolicy);
    setIsEditing(true);
  };

  const handleEditPolicy = (policy: types_BanditPolicy) => {
    setEditingPolicy(policy);
    setIsEditing(true);
  };

  const handleCopyPolicy = (policy: types_BanditPolicy) => {
    const copiedPolicy: types_BanditPolicy = {
      ...policy,
      policy_id: `policy-${Date.now()}`,
      name: `${policy.name} (Copy)`,
      notes: policy.notes
        ? `${policy.notes} (Copy)`
        : "Copy of existing policy",
    };
    setEditingPolicy(copiedPolicy);
    setIsEditing(true);
  };

  const handleSavePolicy = async (policy: types_BanditPolicy) => {
    try {
      console.log("Saving policy:", policy);
      const saveResponse = await BanditService.upsertBanditPolicies({
        namespace,
        policies: [policy],
      });
      console.log("Save response:", saveResponse);

      // Add a small delay to ensure database transaction is committed
      await new Promise((resolve) => setTimeout(resolve, 100));

      // Reload policies to get updated data
      await loadPolicies();

      // Highlight the saved policy
      setHighlightedPolicyId(policy.policy_id || null);
      setTimeout(() => {
        setHighlightedPolicyId(null);
      }, 2000);
    } catch (err) {
      console.error("Failed to save policy:", err);
      setError(err instanceof Error ? err.message : "Failed to save policy");
    }
    setIsEditing(false);
    setEditingPolicy(null);
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
    setEditingPolicy(null);
  };

  const handleCreateCustomPolicy = (policy: types_BanditPolicy) => {
    const customPolicy: CustomPolicy = {
      id: `custom-${Date.now()}`,
      name: policy.name || "Unnamed Policy",
      description: policy.notes || "Custom bandit policy",
      policy,
    };
    setCustomPolicies([...customPolicies, customPolicy]);

    // Highlight the newly created custom policy
    setHighlightedPolicyId(customPolicy.id);
    setTimeout(() => {
      setHighlightedPolicyId(null);
    }, 2000);
  };

  const handleDeleteCustomPolicy = (policyId: string) => {
    // Don't allow deletion of predefined templates
    if (policyId.startsWith("predefined-")) {
      return;
    }
    setCustomPolicies(customPolicies.filter((p) => p.id !== policyId));
  };

  const handleCopyCustomPolicy = (customPolicy: CustomPolicy) => {
    const copiedPolicy: types_BanditPolicy = {
      ...customPolicy.policy,
      policy_id: `policy-${Date.now()}`,
      name: `${customPolicy.policy.name} (Copy)`,
      notes: customPolicy.policy.notes
        ? `${customPolicy.policy.notes} (Copy)`
        : "Copy of custom policy",
    };
    setEditingPolicy(copiedPolicy);
    setIsEditing(true);
  };

  const handleDeletePolicy = async (policy: types_BanditPolicy) => {
    const confirmed = window.confirm(
      `Are you sure you want to delete the policy "${policy.name}"?\n\nThis will deactivate the policy and it will no longer be available for bandit decisions.`
    );

    if (!confirmed) {
      return;
    }

    try {
      console.log("Deleting policy:", policy);

      // Soft delete by setting active to false
      const deletedPolicy: types_BanditPolicy = {
        ...policy,
        active: false,
      };

      await BanditService.upsertBanditPolicies({
        namespace,
        policies: [deletedPolicy],
      });

      console.log("Policy deleted successfully, reloading policies...");

      // Add a small delay to ensure database transaction is committed
      await new Promise((resolve) => setTimeout(resolve, 100));

      // Reload policies to get updated data
      await loadPolicies();

      console.log("Policies reloaded after deletion");
    } catch (err) {
      console.error("Failed to delete policy:", err);
      setError(err instanceof Error ? err.message : "Failed to delete policy");
    }
  };

  if (isEditing && editingPolicy) {
    return (
      <PolicyEditor
        policy={editingPolicy}
        onSave={handleSavePolicy}
        onCancel={handleCancelEdit}
        onCreateCustom={handleCreateCustomPolicy}
      />
    );
  }

  return (
    <Section title="Bandit Policy Manager">
      <div style={{ marginBottom: 16 }}>
        <p style={{ color: "#666", fontSize: 14, marginBottom: 16 }}>
          Manage all bandit policies (active and inactive) and create custom
          policy templates. Predefined templates are based on algorithm profiles
          from the recommendations playground. Policies define algorithm
          parameters like blend weights, MMR lambda, and caps for recommendation
          optimization.
        </p>

        {/* Action Buttons */}
        <div style={{ marginBottom: 20 }}>
          <Row>
            <Button
              onClick={handleCreateNewPolicy}
              style={{
                backgroundColor: "#28a745",
                color: "white",
                border: "none",
                padding: "8px 16px",
                borderRadius: 4,
                cursor: "pointer",
                marginRight: 8,
                fontSize: 14,
              }}
            >
              + New Policy
            </Button>
            <Button
              onClick={loadPolicies}
              disabled={loading}
              style={{
                backgroundColor: "#007acc",
                color: "white",
                border: "none",
                padding: "8px 16px",
                borderRadius: 4,
                cursor: loading ? "not-allowed" : "pointer",
                opacity: loading ? 0.6 : 1,
                fontSize: 14,
              }}
            >
              {loading ? "Loading..." : "Refresh"}
            </Button>
          </Row>
        </div>

        {/* Loading Display */}
        {loading && (
          <div
            style={{
              backgroundColor: "#d1ecf1",
              color: "#0c5460",
              border: "1px solid #bee5eb",
              borderRadius: 4,
              padding: 12,
              marginBottom: 16,
              fontSize: 14,
            }}
          >
            Loading policies...
          </div>
        )}

        {/* Error Display */}
        {error && (
          <div
            style={{
              backgroundColor: "#f8d7da",
              color: "#721c24",
              border: "1px solid #f5c6cb",
              borderRadius: 4,
              padding: 12,
              marginBottom: 16,
              fontSize: 14,
            }}
          >
            Error: {error}
          </div>
        )}

        {/* All Policies */}
        <div style={{ marginBottom: 20 }}>
          <Label
            text={`All Policies (${
              Array.isArray(policies) ? policies.length : 0
            })`}
          >
            {(() => {
              return Array.isArray(policies) && policies.length > 0;
            })() ? (
              <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
                {policies.map((policy) => (
                  <div
                    key={policy.policy_id}
                    id={`policy-${policy.policy_id}`}
                    style={{ position: "relative" }}
                  >
                    <Button
                      onClick={() => handleEditPolicy(policy)}
                      style={{
                        backgroundColor:
                          highlightedPolicyId === policy.policy_id
                            ? "#28a745"
                            : policy.active
                            ? "#007acc"
                            : "#6c757d",
                        color: "white",
                        border: "none",
                        padding: "8px 12px",
                        borderRadius: 4,
                        fontSize: 12,
                        cursor: "pointer",
                        textAlign: "left",
                        minWidth: 250,
                        paddingRight: 116,
                        transition: "all 0.3s ease",
                        transform:
                          highlightedPolicyId === policy.policy_id
                            ? "scale(1.02)"
                            : "scale(1)",
                        boxShadow:
                          highlightedPolicyId === policy.policy_id
                            ? "0 4px 8px rgba(40, 167, 69, 0.3)"
                            : "none",
                      }}
                    >
                      <div
                        style={{
                          fontWeight: "bold",
                          marginBottom: 2,
                          display: "flex",
                          alignItems: "center",
                          gap: 4,
                        }}
                      >
                        {policy.name}
                        {!policy.active && (
                          <span
                            style={{
                              backgroundColor: "rgba(255, 255, 255, 0.9)",
                              color: "#6c757d",
                              fontSize: 8,
                              fontWeight: "bold",
                              padding: "1px 4px",
                              borderRadius: 2,
                              textTransform: "uppercase",
                            }}
                          >
                            INACTIVE
                          </span>
                        )}
                        {highlightedPolicyId === policy.policy_id && (
                          <span
                            style={{
                              backgroundColor: "rgba(255, 255, 255, 0.9)",
                              color: "#28a745",
                              fontSize: 8,
                              fontWeight: "bold",
                              padding: "1px 4px",
                              borderRadius: 2,
                              textTransform: "uppercase",
                            }}
                          >
                            UPDATED
                          </span>
                        )}
                      </div>
                      <div style={{ fontSize: 11, opacity: 0.8 }}>
                        ID: {policy.policy_id}
                      </div>
                      {policy.notes && (
                        <div
                          style={{ fontSize: 10, opacity: 0.7, marginTop: 2 }}
                        >
                          {policy.notes}
                        </div>
                      )}
                    </Button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleCopyPolicy(policy);
                      }}
                      style={{
                        position: "absolute",
                        top: 4,
                        right: 86,
                        backgroundColor: "#6c757d",
                        color: "white",
                        border: "none",
                        borderRadius: 2,
                        padding: "2px 4px",
                        fontSize: 10,
                        cursor: "pointer",
                        zIndex: 10,
                      }}
                      title="Copy policy"
                    >
                      📋
                    </button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleEditPolicy(policy);
                      }}
                      style={{
                        position: "absolute",
                        top: 4,
                        right: 60,
                        backgroundColor: "#007acc",
                        color: "white",
                        border: "none",
                        borderRadius: 2,
                        padding: "2px 4px",
                        fontSize: 10,
                        cursor: "pointer",
                        zIndex: 10,
                      }}
                      title="Edit policy"
                    >
                      ✏️
                    </button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeletePolicy(policy);
                      }}
                      style={{
                        position: "absolute",
                        top: 4,
                        right: 34,
                        backgroundColor: "#dc3545",
                        color: "white",
                        border: "none",
                        borderRadius: 2,
                        padding: "2px 4px",
                        fontSize: 10,
                        cursor: "pointer",
                        zIndex: 10,
                      }}
                      title="Delete policy"
                    >
                      🗑️
                    </button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleCreateCustomPolicy(policy);
                      }}
                      style={{
                        position: "absolute",
                        top: 4,
                        right: 4,
                        backgroundColor: "#28a745",
                        color: "white",
                        border: "none",
                        borderRadius: 2,
                        padding: "2px 4px",
                        fontSize: 10,
                        cursor: "pointer",
                        zIndex: 10,
                      }}
                      title="Save as custom template"
                    >
                      ⭐
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <div
                style={{
                  backgroundColor: "#f8f9fa",
                  border: "1px solid #e9ecef",
                  borderRadius: 4,
                  padding: 16,
                  textAlign: "center",
                  color: "#666",
                  fontSize: 14,
                }}
              >
                No policies found. Create a new policy to get started.
              </div>
            )}
          </Label>
        </div>

        {/* Custom Policy Templates */}
        <div style={{ marginBottom: 20 }}>
          <Label text={`Custom Policy Templates (${customPolicies.length})`}>
            {customPolicies.length > 0 ? (
              <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
                {customPolicies.map((customPolicy) => (
                  <div
                    key={customPolicy.id}
                    id={`custom-policy-${customPolicy.id}`}
                    style={{ position: "relative" }}
                  >
                    <Button
                      onClick={() => handleEditPolicy(customPolicy.policy)}
                      style={{
                        backgroundColor:
                          highlightedPolicyId === customPolicy.id
                            ? "#28a745"
                            : customPolicy.id.startsWith("predefined-")
                            ? "#f8f9fa"
                            : "#e9ecef",
                        color:
                          highlightedPolicyId === customPolicy.id
                            ? "white"
                            : customPolicy.id.startsWith("predefined-")
                            ? "#6c757d"
                            : "#333",
                        border:
                          highlightedPolicyId === customPolicy.id
                            ? "2px solid #28a745"
                            : customPolicy.id.startsWith("predefined-")
                            ? "2px solid #dee2e6"
                            : "1px solid #ddd",
                        padding: "8px 12px",
                        borderRadius: 4,
                        fontSize: 12,
                        cursor: "pointer",
                        textAlign: "left",
                        minWidth: 250,
                        paddingRight: customPolicy.id.startsWith("predefined-")
                          ? 56
                          : 116,
                        transition: "all 0.3s ease",
                        transform:
                          highlightedPolicyId === customPolicy.id
                            ? "scale(1.02)"
                            : "scale(1)",
                        boxShadow:
                          highlightedPolicyId === customPolicy.id
                            ? "0 4px 8px rgba(40, 167, 69, 0.3)"
                            : "none",
                      }}
                    >
                      <div
                        style={{
                          fontWeight: "bold",
                          marginBottom: 2,
                          display: "flex",
                          alignItems: "center",
                          gap: 4,
                        }}
                      >
                        {customPolicy.name}
                        {customPolicy.id.startsWith("predefined-") && (
                          <span
                            style={{
                              backgroundColor: "rgba(108, 117, 125, 0.9)",
                              color: "white",
                              fontSize: 8,
                              fontWeight: "bold",
                              padding: "1px 4px",
                              borderRadius: 2,
                              textTransform: "uppercase",
                            }}
                          >
                            TEMPLATE
                          </span>
                        )}
                        {highlightedPolicyId === customPolicy.id && (
                          <span
                            style={{
                              backgroundColor: "rgba(255, 255, 255, 0.9)",
                              color: "#28a745",
                              fontSize: 8,
                              fontWeight: "bold",
                              padding: "1px 4px",
                              borderRadius: 2,
                              textTransform: "uppercase",
                            }}
                          >
                            NEW
                          </span>
                        )}
                      </div>
                      <div style={{ fontSize: 11, opacity: 0.8 }}>
                        {customPolicy.description}
                      </div>
                    </Button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleCopyCustomPolicy(customPolicy);
                      }}
                      style={{
                        position: "absolute",
                        top: 4,
                        right: 56,
                        backgroundColor: "#6c757d",
                        color: "white",
                        border: "none",
                        borderRadius: 2,
                        padding: "2px 4px",
                        fontSize: 10,
                        cursor: "pointer",
                        zIndex: 10,
                      }}
                      title="Copy policy"
                    >
                      📋
                    </button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleEditPolicy(customPolicy.policy);
                      }}
                      style={{
                        position: "absolute",
                        top: 4,
                        right: 30,
                        backgroundColor: "#007acc",
                        color: "white",
                        border: "none",
                        borderRadius: 2,
                        padding: "2px 4px",
                        fontSize: 10,
                        cursor: "pointer",
                        zIndex: 10,
                      }}
                      title="Edit policy"
                    >
                      ✏️
                    </button>
                    {!customPolicy.id.startsWith("predefined-") && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDeleteCustomPolicy(customPolicy.id);
                        }}
                        style={{
                          position: "absolute",
                          top: 4,
                          right: 4,
                          backgroundColor: "#dc3545",
                          color: "white",
                          border: "none",
                          borderRadius: 2,
                          padding: "2px 4px",
                          fontSize: 10,
                          cursor: "pointer",
                          zIndex: 10,
                        }}
                        title="Delete custom template"
                      >
                        🗑️
                      </button>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <div
                style={{
                  backgroundColor: "#f8f9fa",
                  border: "1px solid #e9ecef",
                  borderRadius: 4,
                  padding: 16,
                  textAlign: "center",
                  color: "#666",
                  fontSize: 14,
                }}
              >
                No custom policy templates yet. Use the predefined templates
                above or create one by saving an existing policy as a template
                using the ⭐ button.
              </div>
            )}
          </Label>
        </div>
      </div>
    </Section>
  );
}
