import React, { useState, useEffect, useCallback, useMemo } from "react";
import {
  Section,
  Row,
  Label,
  TextInput,
  NumberInput,
  Button,
  Code,
} from "./UIComponents";
import { DataTable, Column } from "./DataTable";
import { EventSequenceBuilder } from "./EventSequenceBuilder";
import { UserJourneyVisualization } from "./UserJourneyVisualization";
import { batchEvents, recommend } from "../services/apiService";
import type { types_ScoredItem } from "../lib/api-client";
import { randChoice, randInt, iso, daysAgo } from "../utils/helpers";

interface UserSessionSimulatorProps {
  namespace: string;
  generatedUsers: string[];
  generatedItems: string[];
  eventTypes: Array<{
    id: string;
    title: string;
    index: number;
    weight: number;
    halfLifeDays: number;
  }>;
  blend: { pop: number; cooc: number; als: number };
  k: number;
}

interface EventSequence {
  id: string;
  name: string;
  description: string;
  events: Array<{
    type: number;
    delayMs: number;
    itemSelection: "random" | "recommended" | "similar" | "specific";
    specificItemId?: string;
    value?: number;
  }>;
}

interface UserEvent {
  id: string;
  user_id: string;
  item_id: string;
  type: number;
  typeName: string;
  ts: string;
  value: number;
  addedAt: Date;
}

const PREDEFINED_SEQUENCES: EventSequence[] = [
  {
    id: "browsing_session",
    name: "Browsing Session",
    description: "User browses items, clicks a few, adds one to cart",
    events: [
      { type: 0, delayMs: 0, itemSelection: "random" }, // view
      { type: 0, delayMs: 2000, itemSelection: "random" }, // view
      { type: 1, delayMs: 3000, itemSelection: "random" }, // click
      { type: 0, delayMs: 1000, itemSelection: "random" }, // view
      { type: 2, delayMs: 2000, itemSelection: "random" }, // add to cart
    ],
  },
  {
    id: "purchase_journey",
    name: "Purchase Journey",
    description: "Complete purchase funnel from view to purchase",
    events: [
      { type: 0, delayMs: 0, itemSelection: "random" }, // view
      { type: 1, delayMs: 1500, itemSelection: "recommended" }, // click recommended
      { type: 0, delayMs: 1000, itemSelection: "similar" }, // view similar
      { type: 2, delayMs: 2000, itemSelection: "recommended" }, // add to cart
      { type: 3, delayMs: 5000, itemSelection: "recommended", value: 1 }, // purchase
    ],
  },
  {
    id: "power_user",
    name: "Power User",
    description: "High-engagement user with many interactions",
    events: [
      { type: 0, delayMs: 0, itemSelection: "random" }, // view
      { type: 1, delayMs: 1000, itemSelection: "random" }, // click
      { type: 0, delayMs: 500, itemSelection: "recommended" }, // view recommended
      { type: 1, delayMs: 800, itemSelection: "recommended" }, // click recommended
      { type: 0, delayMs: 1200, itemSelection: "similar" }, // view similar
      { type: 2, delayMs: 1500, itemSelection: "recommended" }, // add to cart
      { type: 0, delayMs: 2000, itemSelection: "random" }, // view
      { type: 3, delayMs: 3000, itemSelection: "recommended", value: 2 }, // purchase
    ],
  },
  {
    id: "window_shopper",
    name: "Window Shopper",
    description: "User who views many items but doesn't convert",
    events: [
      { type: 0, delayMs: 0, itemSelection: "random" }, // view
      { type: 0, delayMs: 3000, itemSelection: "random" }, // view
      { type: 0, delayMs: 2000, itemSelection: "random" }, // view
      { type: 0, delayMs: 4000, itemSelection: "random" }, // view
      { type: 0, delayMs: 1500, itemSelection: "random" }, // view
      { type: 1, delayMs: 2000, itemSelection: "random" }, // click
      { type: 0, delayMs: 1000, itemSelection: "random" }, // view
    ],
  },
];

export function UserSessionSimulator({
  namespace,
  generatedUsers,
  generatedItems,
  eventTypes,
  blend,
  k,
}: UserSessionSimulatorProps) {
  const [selectedUserId, setSelectedUserId] = useState("");
  const [userEvents, setUserEvents] = useState<UserEvent[]>([]);
  const [currentRecommendations, setCurrentRecommendations] = useState<
    types_ScoredItem[] | null
  >(null);
  const [recommendationHistory, setRecommendationHistory] = useState<
    Array<{
      timestamp: Date;
      events: UserEvent[];
      recommendations: types_ScoredItem[];
    }>
  >([]);
  const [isSimulating, setIsSimulating] = useState(false);
  const [selectedSequence, setSelectedSequence] = useState<string>("");
  const [customEventType, setCustomEventType] = useState<number>(0);
  const [customItemId, setCustomItemId] = useState("");
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [simulationSpeed, setSimulationSpeed] = useState(1); // 1x, 2x, 5x, 10x
  const [log, setLog] = useState<string>("");
  const [showSequenceBuilder, setShowSequenceBuilder] = useState(false);
  const [customSequences, setCustomSequences] = useState<EventSequence[]>([]);
  const [showJourneyViz, setShowJourneyViz] = useState(false);

  const appendLog = useCallback((message: string) => {
    const timestamp = new Date().toLocaleTimeString();
    setLog((prev) => `${prev}${prev ? "\n" : ""}[${timestamp}] ${message}`);
  }, []);

  // Auto-suggest first available user if none selected
  useEffect(() => {
    if (generatedUsers.length > 0 && !selectedUserId) {
      // Just show a hint in the placeholder, don't auto-select
      appendLog(`Tip: Try using ${generatedUsers[0]} or any other user ID`);
    }
  }, [generatedUsers, selectedUserId, appendLog]);

  const getEventTypeName = useCallback(
    (typeIndex: number) => {
      const eventType = eventTypes.find((et) => et.index === typeIndex);
      return eventType?.title || `Type ${typeIndex}`;
    },
    [eventTypes]
  );

  const getRecommendations = useCallback(
    async (userId: string) => {
      try {
        const response = await recommend(userId, namespace, k, blend);
        return response.items || [];
      } catch (error) {
        appendLog(`Error getting recommendations: ${error}`);
        return [];
      }
    },
    [namespace, k, blend, appendLog]
  );

  const refreshRecommendations = useCallback(async () => {
    if (!selectedUserId) return;

    const recommendations = await getRecommendations(selectedUserId);
    setCurrentRecommendations(recommendations);
    appendLog(`Updated recommendations (${recommendations.length} items)`);
  }, [selectedUserId, getRecommendations, appendLog]);

  const addEvent = useCallback(
    async (userId: string, itemId: string, type: number, value: number = 1) => {
      const event: UserEvent = {
        id: `event_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        user_id: userId,
        item_id: itemId,
        type,
        typeName: getEventTypeName(type),
        ts: iso(new Date()),
        value,
        addedAt: new Date(),
      };

      // Add to local state
      setUserEvents((prev) => [...prev, event]);

      // Send to API
      try {
        await batchEvents(namespace, [event], appendLog);
        appendLog(`Added ${event.typeName} event for ${userId} on ${itemId}`);

        // Refresh recommendations if auto-refresh is enabled
        if (autoRefresh) {
          setTimeout(refreshRecommendations, 500); // Small delay to ensure event is processed
        }
      } catch (error) {
        appendLog(`Error adding event: ${error}`);
      }
    },
    [
      namespace,
      getEventTypeName,
      appendLog,
      autoRefresh,
      refreshRecommendations,
    ]
  );

  const selectItemForEvent = useCallback(
    (
      selection: "random" | "recommended" | "similar" | "specific",
      specificItemId?: string
    ): string => {
      switch (selection) {
        case "random":
          return randChoice(generatedItems);
        case "recommended":
          if (currentRecommendations && currentRecommendations.length > 0) {
            const rec = randChoice(currentRecommendations);
            return rec.item_id || randChoice(generatedItems);
          }
          return randChoice(generatedItems); // fallback
        case "similar":
          // For now, just return a random item
          // In a real implementation, you'd call the similar items API
          return randChoice(generatedItems);
        case "specific":
          return specificItemId || randChoice(generatedItems);
        default:
          return randChoice(generatedItems);
      }
    },
    [generatedItems, currentRecommendations]
  );

  const runEventSequence = useCallback(
    async (sequence: EventSequence) => {
      if (!selectedUserId) {
        appendLog("No user selected");
        return;
      }

      setIsSimulating(true);
      appendLog(`Starting sequence: ${sequence.name}`);

      for (let i = 0; i < sequence.events.length; i++) {
        const eventConfig = sequence.events[i];
        if (!eventConfig) continue;

        const itemId = selectItemForEvent(
          eventConfig.itemSelection,
          eventConfig.specificItemId
        );

        await addEvent(
          selectedUserId,
          itemId,
          eventConfig.type,
          eventConfig.value || 1
        );

        // Wait for the specified delay (adjusted by simulation speed)
        if (i < sequence.events.length - 1) {
          const delay = eventConfig.delayMs / simulationSpeed;
          await new Promise((resolve) => setTimeout(resolve, delay));
        }
      }

      setIsSimulating(false);
      appendLog(`Completed sequence: ${sequence.name}`);

      // Save recommendation snapshot
      const recommendations = await getRecommendations(selectedUserId);
      setRecommendationHistory((prev) => [
        ...prev,
        {
          timestamp: new Date(),
          events: [...userEvents],
          recommendations: recommendations.map((rec) => ({
            item_id: rec.item_id || "",
            score: rec.score,
          })),
        },
      ]);
    },
    [
      selectedUserId,
      addEvent,
      selectItemForEvent,
      simulationSpeed,
      appendLog,
      getRecommendations,
      userEvents,
    ]
  );

  const saveCustomSequence = useCallback(
    (sequence: EventSequence) => {
      setCustomSequences((prev) => [...prev, sequence]);
      appendLog(`Saved custom sequence: ${sequence.name}`);
    },
    [appendLog]
  );

  const allSequences = useMemo(() => {
    return [...PREDEFINED_SEQUENCES, ...customSequences];
  }, [customSequences]);

  const addCustomEvent = useCallback(async () => {
    if (!selectedUserId) {
      appendLog("No user selected");
      return;
    }

    const itemId = customItemId || randChoice(generatedItems);
    await addEvent(selectedUserId, itemId, customEventType);
  }, [selectedUserId, customItemId, customEventType, generatedItems, addEvent]);

  const clearUserEvents = useCallback(() => {
    setUserEvents([]);
    setRecommendationHistory([]);
    setCurrentRecommendations(null);
    appendLog("Cleared user events and history");
  }, [appendLog]);

  const exportUserJourney = useCallback(() => {
    const journey = {
      userId: selectedUserId,
      events: userEvents,
      recommendationHistory,
      timestamp: new Date().toISOString(),
    };
    const blob = new Blob([JSON.stringify(journey, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `user-journey-${selectedUserId}-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
    appendLog("Exported user journey");
  }, [selectedUserId, userEvents, recommendationHistory, appendLog]);

  // Event table columns
  const eventColumns: Column<UserEvent>[] = [
    { key: "addedAt", title: "Added", width: "120px", sortable: true },
    { key: "typeName", title: "Type", width: "100px" },
    { key: "item_id", title: "Item", width: "120px" },
    { key: "value", title: "Value", width: "60px", align: "right" },
    { key: "ts", title: "Timestamp", width: "150px", sortable: true },
  ];

  // Recommendation history columns
  const historyColumns: Column<any>[] = [
    { key: "timestamp", title: "Time", width: "120px", sortable: true },
    { key: "eventCount", title: "Events", width: "80px", align: "right" },
    { key: "topRecommendation", title: "Top Rec", width: "120px" },
    { key: "topScore", title: "Score", width: "80px", align: "right" },
  ];

  const historyData = useMemo(() => {
    return recommendationHistory.map((entry, index) => ({
      timestamp: entry.timestamp.toLocaleTimeString(),
      eventCount: entry.events.length,
      topRecommendation: entry.recommendations[0]?.item_id || "None",
      topScore: entry.recommendations[0]?.score?.toFixed(3) || "0.000",
    }));
  }, [recommendationHistory]);

  return (
    <Section title="User Session Simulator">
      <div style={{ marginBottom: "16px" }}>
        <p style={{ color: "#666", fontSize: "14px", marginBottom: "16px" }}>
          Focus on a single user and simulate realistic event sequences to
          observe how recommendations evolve. Perfect for understanding
          recommendation system behavior.
        </p>

        {/* User Selection */}
        <div
          style={{
            border: "1px solid #e0e0e0",
            borderRadius: 6,
            padding: 12,
            marginBottom: 16,
            backgroundColor: "#fafafa",
          }}
        >
          <h3
            style={{
              marginTop: 0,
              marginBottom: 8,
              fontSize: 14,
              color: "#333",
            }}
          >
            User Selection
          </h3>
          <Row>
            <Label text="User ID">
              <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
                <TextInput
                  placeholder="e.g., user-0001 or any valid user ID"
                  value={selectedUserId}
                  onChange={(e) => {
                    setSelectedUserId(e.target.value);
                    setUserEvents([]);
                    setRecommendationHistory([]);
                    setCurrentRecommendations(null);
                  }}
                  style={{ minWidth: 200 }}
                />
                {generatedUsers.length > 0 && (
                  <Button
                    onClick={() => {
                      const randomUser = randChoice(generatedUsers);
                      setSelectedUserId(randomUser);
                      setUserEvents([]);
                      setRecommendationHistory([]);
                      setCurrentRecommendations(null);
                    }}
                    style={{
                      padding: "4px 8px",
                      fontSize: 11,
                      backgroundColor: "#e3f2fd",
                      color: "#1565c0",
                    }}
                  >
                    Pick Random
                  </Button>
                )}
              </div>
            </Label>
            <Label text="Auto-refresh recommendations">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                style={{ marginLeft: 8 }}
              />
            </Label>
          </Row>
          {generatedUsers.length > 0 && (
            <div style={{ marginTop: 8 }}>
              <p style={{ fontSize: 12, color: "#666", margin: 0 }}>
                Available users: {generatedUsers.slice(0, 5).join(", ")}
                {generatedUsers.length > 5 &&
                  ` (+${generatedUsers.length - 5} more)`}
              </p>
            </div>
          )}
          {selectedUserId && (
            <div style={{ marginTop: 8 }}>
              <p style={{ fontSize: 12, color: "#666", margin: 0 }}>
                Events: {userEvents.length} | Recommendations:{" "}
                {currentRecommendations?.length || 0} | History:{" "}
                {recommendationHistory.length}
              </p>
              {generatedUsers.length > 0 &&
                !generatedUsers.includes(selectedUserId) && (
                  <p
                    style={{
                      fontSize: 11,
                      color: "#ff9800",
                      margin: "4px 0 0 0",
                    }}
                  >
                    ⚠️ This user ID is not in the generated users list. You can
                    still use it, but make sure it exists in your system.
                  </p>
                )}
            </div>
          )}
        </div>

        {/* Event Sequences */}
        <div
          style={{
            border: "1px solid #e0e0e0",
            borderRadius: 6,
            padding: 12,
            marginBottom: 16,
            backgroundColor: "#fafafa",
          }}
        >
          <h3
            style={{
              marginTop: 0,
              marginBottom: 8,
              fontSize: 14,
              color: "#333",
            }}
          >
            Predefined Event Sequences
          </h3>
          <p style={{ color: "#666", fontSize: 12, marginBottom: 12 }}>
            Run realistic user behavior patterns to see how recommendations
            change.
          </p>
          <Row>
            <Label text="Sequence">
              <select
                value={selectedSequence}
                onChange={(e) => setSelectedSequence(e.target.value)}
                style={{
                  padding: "6px 8px",
                  border: "1px solid #ccc",
                  borderRadius: 4,
                  fontSize: 12,
                  minWidth: 200,
                }}
              >
                <option value="">Select a sequence...</option>
                {allSequences.map((seq) => (
                  <option key={seq.id} value={seq.id}>
                    {seq.name} - {seq.description}
                  </option>
                ))}
              </select>
            </Label>
            <Label text="Speed">
              <select
                value={simulationSpeed}
                onChange={(e) => setSimulationSpeed(Number(e.target.value))}
                style={{
                  padding: "6px 8px",
                  border: "1px solid #ccc",
                  borderRadius: 4,
                  fontSize: 12,
                }}
              >
                <option value={1}>1x (Normal)</option>
                <option value={2}>2x (Fast)</option>
                <option value={5}>5x (Very Fast)</option>
                <option value={10}>10x (Instant)</option>
              </select>
            </Label>
          </Row>
          <div style={{ marginTop: 8 }}>
            <Button
              onClick={() => {
                const sequence = allSequences.find(
                  (s) => s.id === selectedSequence
                );
                if (sequence) runEventSequence(sequence);
              }}
              disabled={!selectedSequence || !selectedUserId || isSimulating}
              style={{ backgroundColor: "#4caf50" }}
            >
              {isSimulating ? "Running..." : "Run Sequence"}
            </Button>
            <Button
              onClick={() => setShowSequenceBuilder(true)}
              style={{ backgroundColor: "#9c27b0" }}
            >
              Build Custom
            </Button>
          </div>
        </div>

        {/* Custom Event */}
        <div
          style={{
            border: "1px solid #e0e0e0",
            borderRadius: 6,
            padding: 12,
            marginBottom: 16,
            backgroundColor: "#fafafa",
          }}
        >
          <h3
            style={{
              marginTop: 0,
              marginBottom: 8,
              fontSize: 14,
              color: "#333",
            }}
          >
            Add Custom Event
          </h3>
          <Row>
            <Label text="Event Type">
              <select
                value={customEventType}
                onChange={(e) => setCustomEventType(Number(e.target.value))}
                style={{
                  padding: "6px 8px",
                  border: "1px solid #ccc",
                  borderRadius: 4,
                  fontSize: 12,
                }}
              >
                {eventTypes.map((et) => (
                  <option key={et.id} value={et.index}>
                    {et.title}
                  </option>
                ))}
              </select>
            </Label>
            <Label text="Item ID (leave blank for random)">
              <TextInput
                placeholder="item-0001 or leave blank"
                value={customItemId}
                onChange={(e) => setCustomItemId(e.target.value)}
                style={{ minWidth: 150 }}
              />
            </Label>
          </Row>
          <div style={{ marginTop: 8 }}>
            <Button
              onClick={addCustomEvent}
              disabled={!selectedUserId}
              style={{ backgroundColor: "#2196f3" }}
            >
              Add Event
            </Button>
          </div>
        </div>

        {/* Actions */}
        <div style={{ marginBottom: 16 }}>
          <Row>
            <Button
              onClick={refreshRecommendations}
              disabled={!selectedUserId}
              style={{ backgroundColor: "#ff9800" }}
            >
              Refresh Recommendations
            </Button>
            <Button
              onClick={clearUserEvents}
              disabled={!selectedUserId}
              style={{ backgroundColor: "#f44336" }}
            >
              Clear Events
            </Button>
            <Button
              onClick={exportUserJourney}
              disabled={!selectedUserId || userEvents.length === 0}
              style={{ backgroundColor: "#9c27b0" }}
            >
              Export Journey
            </Button>
            <Button
              onClick={() => setShowJourneyViz(!showJourneyViz)}
              disabled={!selectedUserId}
              style={{ backgroundColor: "#ff5722" }}
            >
              {showJourneyViz ? "Hide" : "Show"} Timeline
            </Button>
          </Row>
        </div>

        {/* Current Recommendations */}
        {currentRecommendations && currentRecommendations.length > 0 && (
          <div
            style={{
              border: "1px solid #e0e0e0",
              borderRadius: 6,
              padding: 12,
              marginBottom: 16,
              backgroundColor: "#f8f9fa",
            }}
          >
            <h3
              style={{
                marginTop: 0,
                marginBottom: 8,
                fontSize: 14,
                color: "#333",
              }}
            >
              Current Recommendations
            </h3>
            <div style={{ display: "flex", flexWrap: "wrap", gap: "8px" }}>
              {currentRecommendations.slice(0, 10).map((rec, index) => (
                <span
                  key={rec.item_id}
                  style={{
                    backgroundColor: "#e3f2fd",
                    color: "#1565c0",
                    padding: "4px 8px",
                    borderRadius: "4px",
                    fontSize: "12px",
                    border: "1px solid #bbdefb",
                  }}
                >
                  #{index + 1} {rec.item_id} ({rec.score?.toFixed(3)})
                </span>
              ))}
            </div>
          </div>
        )}

        {/* Event History */}
        {userEvents.length > 0 && (
          <div style={{ marginBottom: 16 }}>
            <h3 style={{ fontSize: 14, color: "#333", marginBottom: 8 }}>
              Event History ({userEvents.length} events)
            </h3>
            <DataTable
              data={userEvents}
              columns={eventColumns}
              loading={false}
              selectable={false}
              pagination={{
                page: 1,
                pageSize: 10,
                total: userEvents.length,
                onPageChange: () => {},
                onPageSizeChange: () => {},
              }}
            />
          </div>
        )}

        {/* Recommendation History */}
        {recommendationHistory.length > 0 && (
          <div style={{ marginBottom: 16 }}>
            <h3 style={{ fontSize: 14, color: "#333", marginBottom: 8 }}>
              Recommendation Evolution ({recommendationHistory.length}{" "}
              snapshots)
            </h3>
            <DataTable
              data={historyData}
              columns={historyColumns}
              loading={false}
              selectable={false}
              pagination={{
                page: 1,
                pageSize: 10,
                total: historyData.length,
                onPageChange: () => {},
                onPageSizeChange: () => {},
              }}
            />
          </div>
        )}

        {/* User Journey Visualization */}
        {showJourneyViz && (
          <UserJourneyVisualization
            events={userEvents}
            recommendationHistory={recommendationHistory.map((entry) => ({
              ...entry,
              recommendations: entry.recommendations.map((rec) => ({
                item_id: rec.item_id || "",
                score: rec.score,
              })),
            }))}
            selectedUserId={selectedUserId}
          />
        )}

        {/* Log */}
        <Code>{log || "Ready to simulate user events."}</Code>
      </div>

      {/* Event Sequence Builder Modal */}
      {showSequenceBuilder && (
        <EventSequenceBuilder
          eventTypes={eventTypes}
          generatedItems={generatedItems}
          onSaveSequence={saveCustomSequence}
          onClose={() => setShowSequenceBuilder(false)}
        />
      )}
    </Section>
  );
}
