import React, { useEffect, useCallback, useMemo } from "react";
import {
  Section,
  Row,
  Label,
  TextInput,
  Button,
  Code,
} from "./UIComponents";
import { DataTable, Column } from "./DataTable";
import { EventSequenceBuilder } from "./EventSequenceBuilder";
import { UserJourneyVisualization } from "./UserJourneyVisualization";
import {
  batchEvents,
  recommend,
  listItems,
  listUsers,
} from "../services/apiService";
import type { internal_http_types_ScoredItem } from "../lib/api-client";
import { randChoice, iso } from "../utils/helpers";

// Import types from context
import type { UserEvent, EventSequence } from "../contexts/ViewStateContext";

interface UserSessionSimulatorProps {
  namespace: string;
  generatedUsers: string[];
  setGeneratedUsers: (value: string[]) => void;
  generatedItems: string[];
  setGeneratedItems: (value: string[]) => void;
  eventTypes: Array<{
    id: string;
    title: string;
    index: number;
    weight: number;
    halfLifeDays: number;
  }>;
  blend: { pop: number; cooc: number; als: number };
  k: number;
  userSession: {
    selectedUserId: string;
    userEvents: UserEvent[];
    currentRecommendations: internal_http_types_ScoredItem[] | null;
    recommendationHistory: Array<{
      timestamp: Date;
      events: UserEvent[];
      recommendations: internal_http_types_ScoredItem[];
    }>;
    isSimulating: boolean;
    selectedSequence: string;
    customEventType: number;
    customItemId: string;
    autoRefresh: boolean;
    simulationSpeed: number;
    log: string;
    showSequenceBuilder: boolean;
    customSequences: EventSequence[];
    showJourneyViz: boolean;
    hasAutoLoaded: boolean;
  };
  setUserSession: React.Dispatch<
    React.SetStateAction<{
      selectedUserId: string;
      userEvents: UserEvent[];
      currentRecommendations: internal_http_types_ScoredItem[] | null;
      recommendationHistory: Array<{
        timestamp: Date;
        events: UserEvent[];
        recommendations: internal_http_types_ScoredItem[];
      }>;
      isSimulating: boolean;
      selectedSequence: string;
      customEventType: number;
      customItemId: string;
      autoRefresh: boolean;
      simulationSpeed: number;
      log: string;
      showSequenceBuilder: boolean;
      customSequences: EventSequence[];
      showJourneyViz: boolean;
      hasAutoLoaded: boolean;
    }>
  >;
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
  setGeneratedUsers,
  generatedItems,
  setGeneratedItems,
  eventTypes,
  blend,
  k,
  userSession,
  setUserSession,
}: UserSessionSimulatorProps) {
  const appendLog = useCallback(
    (message: string) => {
      const timestamp = new Date().toLocaleTimeString();
      setUserSession((prev) => ({
        ...prev,
        log: `${prev.log}${prev.log ? "\n" : ""}[${timestamp}] ${message}`,
      }));
    },
    [setUserSession]
  );

  // Get the first available user for fallback
  const firstUser = useMemo(() => {
    return generatedUsers.length > 0 ? generatedUsers[0] : "user-0001";
  }, [generatedUsers]);

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
    const userId = userSession.selectedUserId || firstUser;
    if (!userId) return;

    const recommendations = await getRecommendations(userId);
    setUserSession((prev) => ({
      ...prev,
      currentRecommendations: recommendations,
    }));
    appendLog(`Updated recommendations (${recommendations.length} items)`);
  }, [
    userSession.selectedUserId,
    firstUser,
    getRecommendations,
    appendLog,
    setUserSession,
  ]);

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

      // Add to context state
      setUserSession((prev) => ({
        ...prev,
        userEvents: [...prev.userEvents, event],
      }));

      // Send to API
      try {
        await batchEvents(namespace, [event], appendLog);
        appendLog(`Added ${event.typeName} event for ${userId} on ${itemId}`);

        // Refresh recommendations if auto-refresh is enabled
        if (userSession.autoRefresh) {
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
      userSession.autoRefresh,
      refreshRecommendations,
      setUserSession,
    ]
  );

  const selectItemForEvent = useCallback(
    (
      selection: "random" | "recommended" | "similar" | "specific",
      specificItemId?: string
    ): string => {
      // Fallback item ID when no items are available
      const fallbackItemId = "item-0001";

      const getRandomItem = (): string => {
        if (generatedItems.length === 0) {
          appendLog(
            `⚠️ No generated items available, using fallback: ${fallbackItemId}`
          );
          return fallbackItemId;
        }
        return randChoice(generatedItems);
      };

      switch (selection) {
        case "random":
          return getRandomItem();
        case "recommended":
          if (
            userSession.currentRecommendations &&
            userSession.currentRecommendations.length > 0
          ) {
            const rec = randChoice(userSession.currentRecommendations);
            return rec.item_id || getRandomItem();
          }
          return getRandomItem(); // fallback
        case "similar":
          // For now, just return a random item
          // In a real implementation, you'd call the similar items API
          return getRandomItem();
        case "specific":
          return specificItemId || getRandomItem();
        default:
          return getRandomItem();
      }
    },
    [generatedItems, userSession.currentRecommendations, appendLog]
  );

  const runEventSequence = useCallback(
    async (sequence: EventSequence) => {
      const userId = userSession.selectedUserId || firstUser;
      if (!userId) {
        appendLog("No user available");
        return;
      }

      setUserSession((prev) => ({ ...prev, isSimulating: true }));
      appendLog(`Starting sequence: ${sequence.name}`);

      for (let i = 0; i < sequence.events.length; i++) {
        const eventConfig = sequence.events[i];
        if (!eventConfig) continue;

        const itemId = selectItemForEvent(
          eventConfig.itemSelection,
          eventConfig.specificItemId
        );

        await addEvent(
          userId,
          itemId,
          eventConfig.type,
          eventConfig.value || 1
        );

        // Wait for the specified delay (adjusted by simulation speed)
        if (i < sequence.events.length - 1) {
          const delay = eventConfig.delayMs / userSession.simulationSpeed;
          await new Promise((resolve) => setTimeout(resolve, delay));
        }
      }

      setUserSession((prev) => ({ ...prev, isSimulating: false }));
      appendLog(`Completed sequence: ${sequence.name}`);

      // Save recommendation snapshot
      const recommendations = await getRecommendations(userId);
      setUserSession((prev) => ({
        ...prev,
        recommendationHistory: [
          ...prev.recommendationHistory,
          {
            timestamp: new Date(),
            events: [...prev.userEvents],
            recommendations: recommendations.map((rec) => ({
              item_id: rec.item_id || "",
              score: rec.score,
            })),
          },
        ],
      }));
    },
    [
      userSession.selectedUserId,
      userSession.simulationSpeed,
      firstUser,
      addEvent,
      selectItemForEvent,
      appendLog,
      getRecommendations,
      setUserSession,
    ]
  );

  const saveCustomSequence = useCallback(
    (sequence: EventSequence) => {
      setUserSession((prev) => ({
        ...prev,
        customSequences: [...prev.customSequences, sequence],
      }));
      appendLog(`Saved custom sequence: ${sequence.name}`);
    },
    [appendLog, setUserSession]
  );

  const allSequences = useMemo(() => {
    return [...PREDEFINED_SEQUENCES, ...userSession.customSequences];
  }, [userSession.customSequences]);

  const addCustomEvent = useCallback(async () => {
    const userId = userSession.selectedUserId || firstUser;
    if (!userId) {
      appendLog("No user available");
      return;
    }

    const itemId =
      userSession.customItemId ||
      (generatedItems.length > 0 ? randChoice(generatedItems) : "item-0001");
    await addEvent(userId, itemId, userSession.customEventType);
  }, [
    userSession.selectedUserId,
    userSession.customItemId,
    userSession.customEventType,
    firstUser,
    generatedItems,
    addEvent,
  ]);

  const clearUserEvents = useCallback(() => {
    setUserSession((prev) => ({
      ...prev,
      userEvents: [],
      recommendationHistory: [],
      currentRecommendations: null,
    }));
    appendLog("Cleared user events and history");
  }, [appendLog, setUserSession]);

  const loadExistingItems = useCallback(async () => {
    if (!namespace) {
      appendLog("No namespace selected");
      return;
    }

    try {
      setUserSession((prev) => ({ ...prev, hasAutoLoaded: false })); // Reset auto-load flag
      appendLog("Loading existing items from database...");
      const response = await listItems({
        namespace,
        limit: 1000, // Load up to 1000 items
        offset: 0,
      });

      const itemIds = response.items
        .map((item: any) => item.item_id)
        .filter(Boolean);
      setGeneratedItems(itemIds);
      appendLog(`✅ Loaded ${itemIds.length} existing items from database`);
    } catch (error) {
      appendLog(`❌ Error loading items: ${error}`);
    }
  }, [namespace, setGeneratedItems, appendLog, setUserSession]);

  const loadExistingUsers = useCallback(async () => {
    if (!namespace) {
      appendLog("No namespace selected");
      return;
    }

    try {
      setUserSession((prev) => ({ ...prev, hasAutoLoaded: false })); // Reset auto-load flag
      appendLog("Loading existing users from database...");
      const response = await listUsers({
        namespace,
        limit: 1000, // Load up to 1000 users
        offset: 0,
      });

      const userIds = response.items
        .map((user: any) => user.user_id)
        .filter(Boolean);
      setGeneratedUsers(userIds);
      appendLog(`✅ Loaded ${userIds.length} existing users from database`);
    } catch (error) {
      appendLog(`❌ Error loading users: ${error}`);
    }
  }, [namespace, setGeneratedUsers, appendLog, setUserSession]);

  const autoLoadExistingData = useCallback(async () => {
    if (!namespace || userSession.hasAutoLoaded) {
      return;
    }

    try {
      setUserSession((prev) => ({ ...prev, hasAutoLoaded: true }));
      appendLog("🔄 Auto-loading existing data from database...");

      // Load both users and items in parallel
      const [usersResponse, itemsResponse] = await Promise.all([
        listUsers({ namespace, limit: 1000, offset: 0 }),
        listItems({ namespace, limit: 1000, offset: 0 }),
      ]);

      const userIds = usersResponse.items
        .map((user: any) => user.user_id)
        .filter(Boolean);
      const itemIds = itemsResponse.items
        .map((item: any) => item.item_id)
        .filter(Boolean);

      setGeneratedUsers(userIds);
      setGeneratedItems(itemIds);

      appendLog(
        `✅ Auto-loaded ${userIds.length} users and ${itemIds.length} items from database`
      );
    } catch (error) {
      appendLog(`❌ Error auto-loading data: ${error}`);
      setUserSession((prev) => ({ ...prev, hasAutoLoaded: false })); // Reset so it can try again
    }
  }, [
    namespace,
    userSession.hasAutoLoaded,
    setGeneratedUsers,
    setGeneratedItems,
    appendLog,
    setUserSession,
  ]);

  // Auto-load existing data when component mounts or namespace changes
  useEffect(() => {
    if (
      namespace &&
      generatedUsers.length === 0 &&
      generatedItems.length === 0
    ) {
      autoLoadExistingData();
    }
  }, [
    namespace,
    generatedUsers.length,
    generatedItems.length,
    autoLoadExistingData,
  ]);

  const refreshAllData = useCallback(async () => {
    if (!namespace) {
      appendLog("No namespace selected");
      return;
    }

    try {
      setUserSession((prev) => ({ ...prev, hasAutoLoaded: false })); // Reset auto-load flag
      appendLog("🔄 Refreshing all data from database...");

      // Load both users and items in parallel
      const [usersResponse, itemsResponse] = await Promise.all([
        listUsers({ namespace, limit: 1000, offset: 0 }),
        listItems({ namespace, limit: 1000, offset: 0 }),
      ]);

      const userIds = usersResponse.items
        .map((user: any) => user.user_id)
        .filter(Boolean);
      const itemIds = itemsResponse.items
        .map((item: any) => item.item_id)
        .filter(Boolean);

      setGeneratedUsers(userIds);
      setGeneratedItems(itemIds);

      appendLog(
        `✅ Refreshed ${userIds.length} users and ${itemIds.length} items from database`
      );
    } catch (error) {
      appendLog(`❌ Error refreshing data: ${error}`);
    }
  }, [
    namespace,
    setGeneratedUsers,
    setGeneratedItems,
    appendLog,
    setUserSession,
  ]);

  const exportUserJourney = useCallback(() => {
    const userId = userSession.selectedUserId || firstUser;
    const journey = {
      userId,
      events: userSession.userEvents,
      recommendationHistory: userSession.recommendationHistory,
      timestamp: new Date().toISOString(),
    };
    const blob = new Blob([JSON.stringify(journey, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `user-journey-${userId}-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
    appendLog("Exported user journey");
  }, [
    userSession.selectedUserId,
    userSession.userEvents,
    userSession.recommendationHistory,
    firstUser,
    appendLog,
  ]);

  // Event table columns
  const eventColumns: Column<UserEvent>[] = [
    { key: "addedAt", title: "Added", width: "120px", sortable: true },
    { key: "typeName", title: "Type", width: "100px" },
    { key: "item_id", title: "Item", width: "120px" },
    { key: "value", title: "Value", width: "60px", align: "right" },
    { key: "ts", title: "Timestamp", width: "150px", sortable: true },
  ];

  // Recommendation history data type
  interface HistoryDataEntry {
    timestamp: string;
    eventCount: number;
    topRecommendation: string;
    topScore: string;
  }

  // Recommendation history columns
  const historyColumns: Column<HistoryDataEntry>[] = [
    { key: "timestamp", title: "Time", width: "120px", sortable: true },
    { key: "eventCount", title: "Events", width: "80px", align: "right" },
    { key: "topRecommendation", title: "Top Rec", width: "120px" },
    { key: "topScore", title: "Score", width: "80px", align: "right" },
  ];

  const historyData = useMemo((): HistoryDataEntry[] => {
    return userSession.recommendationHistory.map((entry) => ({
      timestamp: entry.timestamp.toLocaleTimeString(),
      eventCount: entry.events.length,
      topRecommendation: entry.recommendations[0]?.item_id || "None",
      topScore: entry.recommendations[0]?.score?.toFixed(3) || "0.000",
    }));
  }, [userSession.recommendationHistory]);

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
            <Label text="User ID (leave blank to use first generated)">
              <TextInput
                placeholder={firstUser || "user-0001"}
                value={userSession.selectedUserId}
                onChange={(e) => {
                  setUserSession((prev) => ({
                    ...prev,
                    selectedUserId: e.target.value,
                    userEvents: [],
                    recommendationHistory: [],
                    currentRecommendations: null,
                  }));
                }}
                style={{ minWidth: 200 }}
              />
            </Label>
            <Label text="Auto-refresh recommendations">
              <input
                type="checkbox"
                checked={userSession.autoRefresh}
                onChange={(e) =>
                  setUserSession((prev) => ({
                    ...prev,
                    autoRefresh: e.target.checked,
                  }))
                }
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
          {generatedItems.length > 0 && (
            <div style={{ marginTop: 4 }}>
              <p style={{ fontSize: 12, color: "#666", margin: 0 }}>
                Available items: {generatedItems.slice(0, 5).join(", ")}
                {generatedItems.length > 5 &&
                  ` (+${generatedItems.length - 5} more)`}
              </p>
            </div>
          )}
          {generatedItems.length === 0 && (
            <div style={{ marginTop: 4 }}>
              <p style={{ fontSize: 12, color: "#ff9800", margin: 0 }}>
                ⚠️ No items available. Generate some items first or use specific
                item IDs.
              </p>
            </div>
          )}
          {(userSession.selectedUserId || firstUser) && (
            <div style={{ marginTop: 8 }}>
              <p style={{ fontSize: 12, color: "#666", margin: 0 }}>
                Using: {userSession.selectedUserId || firstUser} | Events:{" "}
                {userSession.userEvents.length} | Recommendations:{" "}
                {userSession.currentRecommendations?.length || 0} | History:{" "}
                {userSession.recommendationHistory.length}
              </p>
              {userSession.selectedUserId &&
                generatedUsers.length > 0 &&
                !generatedUsers.includes(userSession.selectedUserId) && (
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
                value={userSession.selectedSequence}
                onChange={(e) =>
                  setUserSession((prev) => ({
                    ...prev,
                    selectedSequence: e.target.value,
                  }))
                }
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
                value={userSession.simulationSpeed}
                onChange={(e) =>
                  setUserSession((prev) => ({
                    ...prev,
                    simulationSpeed: Number(e.target.value),
                  }))
                }
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
                  (s) => s.id === userSession.selectedSequence
                );
                if (sequence) runEventSequence(sequence);
              }}
              disabled={
                !userSession.selectedSequence || userSession.isSimulating
              }
            >
              {userSession.isSimulating ? "Running..." : "Run Sequence"}
            </Button>
            <Button
              onClick={() =>
                setUserSession((prev) => ({
                  ...prev,
                  showSequenceBuilder: true,
                }))
              }
            >
              Build Custom
            </Button>
            {!userSession.selectedSequence && (
              <span style={{ fontSize: 12, color: "#666", marginLeft: 8 }}>
                Select a sequence above to enable
              </span>
            )}
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
                value={userSession.customEventType}
                onChange={(e) =>
                  setUserSession((prev) => ({
                    ...prev,
                    customEventType: Number(e.target.value),
                  }))
                }
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
                value={userSession.customItemId}
                onChange={(e) =>
                  setUserSession((prev) => ({
                    ...prev,
                    customItemId: e.target.value,
                  }))
                }
                style={{ minWidth: 150 }}
              />
            </Label>
          </Row>
          <div style={{ marginTop: 8 }}>
            <Button onClick={addCustomEvent}>Add Event</Button>
          </div>
        </div>

        {/* Actions */}
        <div style={{ marginBottom: 16 }}>
          <Row>
            <Button onClick={refreshRecommendations}>
              Refresh Recommendations
            </Button>
            <Button onClick={clearUserEvents}>Clear Events</Button>
            <Button onClick={refreshAllData}>Refresh Data</Button>
            <Button onClick={loadExistingUsers}>Load Existing Users</Button>
            <Button onClick={loadExistingItems}>Load Existing Items</Button>
            <Button
              onClick={exportUserJourney}
              disabled={userSession.userEvents.length === 0}
            >
              Export Journey
            </Button>
            <Button
              onClick={() =>
                setUserSession((prev) => ({
                  ...prev,
                  showJourneyViz: !prev.showJourneyViz,
                }))
              }
            >
              {userSession.showJourneyViz ? "Hide" : "Show"} Timeline
            </Button>
          </Row>
        </div>

        {/* Current Recommendations */}
        {userSession.currentRecommendations &&
          userSession.currentRecommendations.length > 0 && (
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
                {userSession.currentRecommendations
                  .slice(0, 10)
                  .map((rec, index) => (
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
        {userSession.userEvents.length > 0 && (
          <div style={{ marginBottom: 16 }}>
            <h3 style={{ fontSize: 14, color: "#333", marginBottom: 8 }}>
              Event History ({userSession.userEvents.length} events)
            </h3>
            <DataTable
              data={userSession.userEvents}
              columns={eventColumns}
              loading={false}
              selectable={false}
              pagination={{
                page: 1,
                pageSize: 10,
                total: userSession.userEvents.length,
                onPageChange: () => {},
                onPageSizeChange: () => {},
              }}
            />
          </div>
        )}

        {/* Recommendation History */}
        {userSession.recommendationHistory.length > 0 && (
          <div style={{ marginBottom: 16 }}>
            <h3 style={{ fontSize: 14, color: "#333", marginBottom: 8 }}>
              Recommendation Evolution (
              {userSession.recommendationHistory.length} snapshots)
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
        {userSession.showJourneyViz && (
          <UserJourneyVisualization
            events={userSession.userEvents}
            recommendationHistory={userSession.recommendationHistory.map(
              (entry) => ({
                ...entry,
                recommendations: entry.recommendations.map((rec) => ({
                  item_id: rec.item_id || "",
                  score: rec.score,
                })),
              })
            )}
            selectedUserId={userSession.selectedUserId}
          />
        )}

        {/* Log */}
        <Code>{userSession.log || "Ready to simulate user events."}</Code>
      </div>

      {/* Event Sequence Builder Modal */}
      {userSession.showSequenceBuilder && (
        <EventSequenceBuilder
          eventTypes={eventTypes}
          generatedItems={generatedItems}
          onSaveSequence={saveCustomSequence}
          onClose={() =>
            setUserSession((prev) => ({ ...prev, showSequenceBuilder: false }))
          }
        />
      )}
    </Section>
  );
}
