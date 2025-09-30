import React from "react";
import { Button } from "../primitives/UIComponents";

interface RuleFormModalProps {
  open: boolean;
  editing: boolean;
  formData: any;
  setFormData: (v: any) => void;
  onCancel: () => void;
  onSubmit: () => void;
}

export function RuleFormModal(props: RuleFormModalProps) {
  if (!props.open) return null;
  const { formData, setFormData } = props;

  return (
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
        <h3 style={{ margin: "0 0 16px 0" }}>
          {props.editing ? "Edit Rule" : "Create Rule"}
        </h3>
        <div style={{ display: "grid", gap: 12, marginBottom: 16 }}>
          <div style={{ display: "grid", gap: 12 }}>
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
                placeholder="Rule name"
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
              <textarea
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
                rows={3}
                placeholder="Optional"
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
                    setFormData({ ...formData, action: e.target.value })
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
                    setFormData({ ...formData, target_type: e.target.value })
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
                  Item IDs (comma-separated)
                </label>
                <input
                  type="text"
                  value={(formData.item_ids || []).join(", ")}
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
                  placeholder="Optional"
                />
              </div>
            </div>

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
                  value={formData.boost_value}
                  step={0.05}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      boost_value: parseFloat(e.target.value || "0"),
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
                  min={1}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      max_pins: parseInt(e.target.value || "0", 10),
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
                      priority: parseInt(e.target.value || "0", 10),
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
              <label style={{ display: "flex", alignItems: "center", gap: 8 }}>
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

          <div style={{ display: "flex", gap: 8, justifyContent: "flex-end" }}>
            <Button
              onClick={props.onCancel}
              style={{ backgroundColor: "#6c757d", color: "white" }}
            >
              Cancel
            </Button>
            <Button
              onClick={props.onSubmit}
              style={{ backgroundColor: "#28a745", color: "white" }}
              disabled={!formData.name}
            >
              {props.editing ? "Update" : "Create"}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
