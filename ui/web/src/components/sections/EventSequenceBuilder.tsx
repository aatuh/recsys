import React, { useState, useCallback } from "react";
import { Row, Label, TextInput, NumberInput, Button } from "../primitives/UIComponents";

export interface EventStep {
  id: string;
  type: number;
  typeName: string;
  delayMs: number;
  itemSelection: "random" | "recommended" | "similar" | "specific";
  specificItemId?: string;
  value?: number;
  description?: string;
}

export interface EventSequence {
  id: string;
  name: string;
  description: string;
  events: EventStep[];
}

interface EventSequenceBuilderProps {
  eventTypes: Array<{
    id: string;
    title: string;
    index: number;
    weight: number;
    halfLifeDays: number;
  }>;
  generatedItems: string[];
  onSaveSequence: (sequence: EventSequence) => void;
  onClose: () => void;
}

export function EventSequenceBuilder({
  eventTypes,
  generatedItems,
  onSaveSequence,
  onClose,
}: EventSequenceBuilderProps) {
  const [sequenceName, setSequenceName] = useState("");
  const [sequenceDescription, setSequenceDescription] = useState("");
  const [events, setEvents] = useState<EventStep[]>([]);

  const addEvent = useCallback(() => {
    const newEvent: EventStep = {
      id: `event_${Date.now()}`,
      type: eventTypes[0]?.index || 0,
      typeName: eventTypes[0]?.title || "View",
      delayMs: 1000,
      itemSelection: "random",
      value: 1,
      description: "",
    };
    setEvents((prev) => [...prev, newEvent]);
  }, [eventTypes]);

  const updateEvent = useCallback(
    (eventId: string, updates: Partial<EventStep>) => {
      setEvents((prev) =>
        prev.map((event) =>
          event.id === eventId
            ? {
                ...event,
                ...updates,
                typeName:
                  eventTypes.find((et) => et.index === updates.type)?.title ||
                  event.typeName,
              }
            : event
        )
      );
    },
    [eventTypes]
  );

  const removeEvent = useCallback((eventId: string) => {
    setEvents((prev) => prev.filter((event) => event.id !== eventId));
  }, []);

  const moveEvent = useCallback((eventId: string, direction: "up" | "down") => {
    setEvents((prev) => {
      const index = prev.findIndex((event) => event.id === eventId);
      if (index === -1) return prev;

      const newEvents: EventStep[] = [...prev];
      const targetIndex = direction === "up" ? index - 1 : index + 1;

      if (targetIndex >= 0 && targetIndex < newEvents.length) {
        const current = newEvents[index] as EventStep;
        const target = newEvents[targetIndex] as EventStep;
        newEvents[targetIndex] = current;
        newEvents[index] = target;
      }

      return newEvents;
    });
  }, []);

  const saveSequence = useCallback(() => {
    if (!sequenceName.trim()) {
      alert("Please enter a sequence name");
      return;
    }

    if (events.length === 0) {
      alert("Please add at least one event to the sequence");
      return;
    }

    const sequence: EventSequence = {
      id: `custom_${Date.now()}`,
      name: sequenceName.trim(),
      description: sequenceDescription.trim(),
      events,
    };

    onSaveSequence(sequence);
    onClose();
  }, [sequenceName, sequenceDescription, events, onSaveSequence, onClose]);

  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(0,0,0,0.5)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: 16,
        zIndex: 1000,
      }}
      onClick={onClose}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          width: "90vw",
          maxWidth: 800,
          maxHeight: "90vh",
          background: "white",
          borderRadius: 8,
          border: "1px solid #ddd",
          boxShadow: "0 8px 30px rgba(0,0,0,0.3)",
          padding: 16,
          overflow: "auto",
        }}
      >
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: 16,
          }}
        >
          <h2 style={{ margin: 0 }}>Build Event Sequence</h2>
          <button
            onClick={onClose}
            style={{
              padding: "4px 8px",
              borderRadius: 4,
              border: "1px solid #aaa",
              background: "#f5f5f5",
              cursor: "pointer",
            }}
          >
            Close
          </button>
        </div>

        {/* Sequence Info */}
        <div
          style={{
            border: "1px solid #e0e0e0",
            borderRadius: 6,
            padding: 12,
            marginBottom: 16,
            backgroundColor: "#fafafa",
          }}
        >
          <h3 style={{ marginTop: 0, marginBottom: 8, fontSize: 14 }}>
            Sequence Information
          </h3>
          <Row>
            <Label text="Name">
              <TextInput
                placeholder="e.g., High-Value Customer Journey"
                value={sequenceName}
                onChange={(e) => setSequenceName(e.target.value)}
                style={{ minWidth: 300 }}
              />
            </Label>
          </Row>
          <div style={{ marginTop: 8 }}>
            <Label text="Description">
              <TextInput
                placeholder="Describe what this sequence represents..."
                value={sequenceDescription}
                onChange={(e) => setSequenceDescription(e.target.value)}
                style={{ minWidth: 400 }}
              />
            </Label>
          </div>
        </div>

        {/* Events List */}
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
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: 12,
            }}
          >
            <h3 style={{ marginTop: 0, marginBottom: 0, fontSize: 14 }}>
              Events ({events.length})
            </h3>
            <Button onClick={addEvent} style={{ backgroundColor: "#4caf50" }}>
              Add Event
            </Button>
          </div>

          {events.length === 0 ? (
            <p style={{ color: "#666", fontSize: 12, textAlign: "center" }}>
              No events added yet. Click "Add Event" to start building your
              sequence.
            </p>
          ) : (
            <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
              {events.map((event, index) => (
                <div
                  key={event.id}
                  style={{
                    border: "1px solid #ddd",
                    borderRadius: 4,
                    padding: 12,
                    backgroundColor: "white",
                  }}
                >
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                      marginBottom: 8,
                    }}
                  >
                    <span style={{ fontWeight: 600, fontSize: 12 }}>
                      Step {index + 1}
                    </span>
                    <div style={{ display: "flex", gap: 4 }}>
                      <button
                        onClick={() => moveEvent(event.id, "up")}
                        disabled={index === 0}
                        style={{
                          padding: "2px 6px",
                          fontSize: 10,
                          border: "1px solid #ccc",
                          background: index === 0 ? "#f5f5f5" : "white",
                          cursor: index === 0 ? "not-allowed" : "pointer",
                        }}
                      >
                        ↑
                      </button>
                      <button
                        onClick={() => moveEvent(event.id, "down")}
                        disabled={index === events.length - 1}
                        style={{
                          padding: "2px 6px",
                          fontSize: 10,
                          border: "1px solid #ccc",
                          background:
                            index === events.length - 1 ? "#f5f5f5" : "white",
                          cursor:
                            index === events.length - 1
                              ? "not-allowed"
                              : "pointer",
                        }}
                      >
                        ↓
                      </button>
                      <button
                        onClick={() => removeEvent(event.id)}
                        style={{
                          padding: "2px 6px",
                          fontSize: 10,
                          border: "1px solid #f44336",
                          background: "white",
                          color: "#f44336",
                          cursor: "pointer",
                        }}
                      >
                        Remove
                      </button>
                    </div>
                  </div>

                  <div
                    style={{
                      display: "grid",
                      gridTemplateColumns: "1fr 1fr 1fr",
                      gap: 8,
                    }}
                  >
                    <div>
                      <label
                        style={{
                          fontSize: 11,
                          display: "block",
                          marginBottom: 2,
                        }}
                      >
                        Event Type
                      </label>
                      <select
                        value={event.type}
                        onChange={(e) =>
                          updateEvent(event.id, {
                            type: Number(e.target.value),
                          })
                        }
                        style={{
                          width: "100%",
                          padding: "4px 6px",
                          border: "1px solid #ccc",
                          borderRadius: 3,
                          fontSize: 11,
                        }}
                      >
                        {eventTypes.map((et) => (
                          <option key={et.id} value={et.index}>
                            {et.title}
                          </option>
                        ))}
                      </select>
                    </div>

                    <div>
                      <label
                        style={{
                          fontSize: 11,
                          display: "block",
                          marginBottom: 2,
                        }}
                      >
                        Delay (ms)
                      </label>
                      <NumberInput
                        min={0}
                        value={event.delayMs}
                        onChange={(e) =>
                          updateEvent(event.id, {
                            delayMs: Number(e.target.value),
                          })
                        }
                        style={{ fontSize: 11, padding: "4px 6px" }}
                      />
                    </div>

                    <div>
                      <label
                        style={{
                          fontSize: 11,
                          display: "block",
                          marginBottom: 2,
                        }}
                      >
                        Value
                      </label>
                      <NumberInput
                        min={1}
                        value={event.value || 1}
                        onChange={(e) =>
                          updateEvent(event.id, {
                            value: Number(e.target.value),
                          })
                        }
                        style={{ fontSize: 11, padding: "4px 6px" }}
                      />
                    </div>
                  </div>

                  <div style={{ marginTop: 8 }}>
                    <label
                      style={{
                        fontSize: 11,
                        display: "block",
                        marginBottom: 2,
                      }}
                    >
                      Item Selection
                    </label>
                    <select
                      value={event.itemSelection}
                      onChange={(e) =>
                        updateEvent(event.id, {
                          itemSelection: e.target.value as any,
                        })
                      }
                      style={{
                        width: "100%",
                        padding: "4px 6px",
                        border: "1px solid #ccc",
                        borderRadius: 3,
                        fontSize: 11,
                      }}
                    >
                      <option value="random">Random Item</option>
                      <option value="recommended">Recommended Item</option>
                      <option value="similar">Similar Item</option>
                      <option value="specific">Specific Item</option>
                    </select>
                  </div>

                  {event.itemSelection === "specific" && (
                    <div style={{ marginTop: 8 }}>
                      <label
                        style={{
                          fontSize: 11,
                          display: "block",
                          marginBottom: 2,
                        }}
                      >
                        Specific Item ID
                      </label>
                      <select
                        value={event.specificItemId || ""}
                        onChange={(e) =>
                          updateEvent(event.id, {
                            specificItemId: e.target.value,
                          })
                        }
                        style={{
                          width: "100%",
                          padding: "4px 6px",
                          border: "1px solid #ccc",
                          borderRadius: 3,
                          fontSize: 11,
                        }}
                      >
                        <option value="">Select an item...</option>
                        {generatedItems.map((itemId) => (
                          <option key={itemId} value={itemId}>
                            {itemId}
                          </option>
                        ))}
                      </select>
                    </div>
                  )}

                  <div style={{ marginTop: 8 }}>
                    <label
                      style={{
                        fontSize: 11,
                        display: "block",
                        marginBottom: 2,
                      }}
                    >
                      Description (optional)
                    </label>
                    <TextInput
                      placeholder="e.g., User clicks on recommended item"
                      value={event.description || ""}
                      onChange={(e) =>
                        updateEvent(event.id, { description: e.target.value })
                      }
                      style={{ fontSize: 11, padding: "4px 6px" }}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Actions */}
        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8 }}>
          <Button onClick={onClose} style={{ backgroundColor: "#666" }}>
            Cancel
          </Button>
          <Button onClick={saveSequence} style={{ backgroundColor: "#4caf50" }}>
            Save Sequence
          </Button>
        </div>
      </div>
    </div>
  );
}
