import React, { useState } from "react";
import { Section, Row, Label, NumberInput, Button, Code } from "./UIComponents";
import type { EventTypeConfig } from "../types";

interface EventTypeEditorProps {
  eventTypes: EventTypeConfig[];
  setEventTypes: (eventTypes: EventTypeConfig[]) => void;
}

export function EventTypeEditor({
  eventTypes,
  setEventTypes,
}: EventTypeEditorProps) {
  const [isEditorOpen, setIsEditorOpen] = useState(false);
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [newEventType, setNewEventType] = useState<Partial<EventTypeConfig>>({
    id: "",
    title: "",
    index: 0,
    weight: 0.5,
    halfLifeDays: 30,
  });

  const addEventType = () => {
    if (!newEventType.id || !newEventType.title) {
      alert("Please fill in both ID and Title");
      return;
    }

    // Check for duplicate IDs or indexes
    if (eventTypes.some((et) => et.id === newEventType.id)) {
      alert("Event type ID already exists");
      return;
    }
    if (eventTypes.some((et) => et.index === newEventType.index)) {
      alert("Event type index already exists");
      return;
    }

    const eventType: EventTypeConfig = {
      id: newEventType.id!,
      title: newEventType.title!,
      index: newEventType.index!,
      weight: newEventType.weight!,
      halfLifeDays: newEventType.halfLifeDays!,
    };

    setEventTypes([...eventTypes, eventType].sort((a, b) => a.index - b.index));
    setNewEventType({
      id: "",
      title: "",
      index: Math.max(...eventTypes.map((et) => et.index), -1) + 1,
      weight: 0.5,
      halfLifeDays: 30,
    });
  };

  const updateEventType = (
    index: number,
    field: keyof EventTypeConfig,
    value: any
  ) => {
    const updated = [...eventTypes];
    const currentEventType = updated[index];
    if (currentEventType) {
      const newEventType = { ...currentEventType, [field]: value };

      // Validate for duplicates when changing ID or index
      if (field === "id" && value !== currentEventType.id) {
        if (updated.some((et, i) => i !== index && et.id === value)) {
          alert("Event type ID already exists");
          return;
        }
      }

      if (field === "index" && value !== currentEventType.index) {
        if (updated.some((et, i) => i !== index && et.index === value)) {
          alert("Event type index already exists");
          return;
        }
      }

      updated[index] = newEventType;
      setEventTypes(updated);
    }
  };

  const deleteEventType = (index: number) => {
    if (eventTypes.length <= 1) {
      alert("At least one event type is required");
      return;
    }
    setEventTypes(eventTypes.filter((_, i) => i !== index));
  };

  const resetToDefaults = () => {
    setEventTypes([
      { id: "view", title: "View", index: 0, weight: 0.2, halfLifeDays: 30 },
      { id: "click", title: "Click", index: 1, weight: 0.7, halfLifeDays: 30 },
      {
        id: "add",
        title: "Add to Cart",
        index: 2,
        weight: 0.8,
        halfLifeDays: 45,
      },
      {
        id: "purchase",
        title: "Purchase",
        index: 3,
        weight: 1.0,
        halfLifeDays: 60,
      },
    ]);
  };

  return (
    <div
      style={{
        border: "1px solid #e0e0e0",
        borderRadius: 6,
        padding: 12,
        marginBottom: 16,
        backgroundColor: "#fafafa",
      }}
    >
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: 8,
        }}
      >
        <h3
          style={{
            marginTop: 0,
            marginBottom: 0,
            fontSize: 14,
            color: "#333",
          }}
        >
          Event Type Configuration
        </h3>
        <div style={{ display: "flex", gap: 8 }}>
          <Button
            onClick={() => setIsEditorOpen(!isEditorOpen)}
            style={{
              padding: "4px 8px",
              fontSize: 12,
              backgroundColor: isEditorOpen ? "#e3f2fd" : "#f5f5f5",
              color: isEditorOpen ? "#1565c0" : "#666",
            }}
          >
            {isEditorOpen ? "▼ Hide" : "▶ Configure"}
          </Button>
          <Button
            onClick={resetToDefaults}
            style={{
              padding: "4px 8px",
              fontSize: 12,
              backgroundColor: "#f5f5f5",
              color: "#666",
            }}
          >
            Reset to Defaults
          </Button>
        </div>
      </div>

      <p style={{ color: "#666", fontSize: 12, marginBottom: 8 }}>
        Configure custom event types with their IDs, titles, indexes, weights,
        and half-life parameters. All fields are editable including existing
        event types.
      </p>

      {/* Current event types preview */}
      <div
        style={{
          backgroundColor: "#f0f0f0",
          border: "1px solid #ddd",
          borderRadius: 4,
          padding: 8,
          fontSize: 11,
          color: "#555",
          fontFamily: "monospace",
          marginBottom: 8,
        }}
      >
        <strong>Current Event Types:</strong>{" "}
        {eventTypes
          .sort((a, b) => a.index - b.index)
          .map(
            (et) =>
              `${et.title} (idx:${et.index}, w:${et.weight}, h:${et.halfLifeDays}d)`
          )
          .join(", ")}
      </div>

      {/* Editor content */}
      {isEditorOpen && (
        <div style={{ marginTop: 12 }}>
          {/* Add new event type */}
          <div
            style={{
              border: "1px solid #ddd",
              borderRadius: 4,
              padding: 12,
              marginBottom: 12,
              backgroundColor: "#fff",
            }}
          >
            <h4 style={{ marginTop: 0, marginBottom: 8, fontSize: 13 }}>
              Add New Event Type
            </h4>
            <Row>
              <Label text="ID">
                <input
                  type="text"
                  value={newEventType.id || ""}
                  onChange={(e) =>
                    setNewEventType({ ...newEventType, id: e.target.value })
                  }
                  placeholder="e.g., 'like'"
                  style={{
                    padding: "4px 8px",
                    border: "1px solid #ccc",
                    borderRadius: 4,
                    fontSize: 12,
                    width: 80,
                  }}
                />
              </Label>
              <Label text="Title">
                <input
                  type="text"
                  value={newEventType.title || ""}
                  onChange={(e) =>
                    setNewEventType({ ...newEventType, title: e.target.value })
                  }
                  placeholder="e.g., 'Like'"
                  style={{
                    padding: "4px 8px",
                    border: "1px solid #ccc",
                    borderRadius: 4,
                    fontSize: 12,
                    width: 120,
                  }}
                />
              </Label>
              <Label text="Index">
                <NumberInput
                  min={0}
                  value={newEventType.index || 0}
                  onChange={(e) =>
                    setNewEventType({
                      ...newEventType,
                      index: Number(e.target.value),
                    })
                  }
                  style={{ width: 60 }}
                />
              </Label>
              <Label text="Weight">
                <NumberInput
                  step="0.1"
                  min="0"
                  max="1"
                  value={newEventType.weight || 0.5}
                  onChange={(e) =>
                    setNewEventType({
                      ...newEventType,
                      weight: Number(e.target.value),
                    })
                  }
                  style={{ width: 60 }}
                />
              </Label>
              <Label text="Half-life (days)">
                <NumberInput
                  min="1"
                  value={newEventType.halfLifeDays || 30}
                  onChange={(e) =>
                    setNewEventType({
                      ...newEventType,
                      halfLifeDays: Number(e.target.value),
                    })
                  }
                  style={{ width: 60 }}
                />
              </Label>
              <Button
                onClick={addEventType}
                style={{
                  padding: "4px 8px",
                  fontSize: 12,
                  backgroundColor: "#4caf50",
                  color: "white",
                }}
              >
                Add
              </Button>
            </Row>
          </div>

          {/* Edit existing event types */}
          <div
            style={{
              border: "1px solid #ddd",
              borderRadius: 4,
              padding: 12,
              backgroundColor: "#fff",
            }}
          >
            <h4 style={{ marginTop: 0, marginBottom: 8, fontSize: 13 }}>
              Edit Existing Event Types
            </h4>
            <p
              style={{
                color: "#888",
                fontSize: 11,
                marginBottom: 8,
                marginTop: 0,
              }}
            >
              All fields are editable. Duplicate IDs or indexes will be
              prevented.
            </p>
            {eventTypes
              .sort((a, b) => a.index - b.index)
              .map((eventType, index) => (
                <div
                  key={eventType.id}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: 8,
                    marginBottom: 8,
                    padding: 8,
                    backgroundColor: "#f9f9f9",
                    borderRadius: 4,
                  }}
                >
                  <Label text="ID">
                    <input
                      type="text"
                      value={eventType.id}
                      onChange={(e) =>
                        updateEventType(
                          eventTypes.findIndex((et) => et.id === eventType.id),
                          "id",
                          e.target.value
                        )
                      }
                      style={{
                        padding: "2px 4px",
                        border: "1px solid #ccc",
                        borderRadius: 3,
                        fontSize: 11,
                        width: 60,
                      }}
                    />
                  </Label>
                  <Label text="Title">
                    <input
                      type="text"
                      value={eventType.title}
                      onChange={(e) =>
                        updateEventType(
                          eventTypes.findIndex((et) => et.id === eventType.id),
                          "title",
                          e.target.value
                        )
                      }
                      style={{
                        padding: "2px 4px",
                        border: "1px solid #ccc",
                        borderRadius: 3,
                        fontSize: 11,
                        width: 80,
                      }}
                    />
                  </Label>
                  <Label text="Index">
                    <NumberInput
                      min={0}
                      value={eventType.index}
                      onChange={(e) =>
                        updateEventType(
                          eventTypes.findIndex((et) => et.id === eventType.id),
                          "index",
                          Number(e.target.value)
                        )
                      }
                      style={{ width: 50 }}
                    />
                  </Label>
                  <Label text="Weight">
                    <NumberInput
                      step="0.1"
                      min="0"
                      max="1"
                      value={eventType.weight}
                      onChange={(e) =>
                        updateEventType(
                          eventTypes.findIndex((et) => et.id === eventType.id),
                          "weight",
                          Number(e.target.value)
                        )
                      }
                      style={{ width: 60 }}
                    />
                  </Label>
                  <Label text="Half-life">
                    <NumberInput
                      min="1"
                      value={eventType.halfLifeDays}
                      onChange={(e) =>
                        updateEventType(
                          eventTypes.findIndex((et) => et.id === eventType.id),
                          "halfLifeDays",
                          Number(e.target.value)
                        )
                      }
                      style={{ width: 60 }}
                    />
                  </Label>
                  <Button
                    onClick={() =>
                      deleteEventType(
                        eventTypes.findIndex((et) => et.id === eventType.id)
                      )
                    }
                    style={{
                      padding: "2px 6px",
                      fontSize: 10,
                      backgroundColor: "#f44336",
                      color: "white",
                    }}
                  >
                    Delete
                  </Button>
                </div>
              ))}
          </div>
        </div>
      )}
    </div>
  );
}
