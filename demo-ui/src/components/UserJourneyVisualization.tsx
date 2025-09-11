import React, { useMemo } from "react";
import { Section } from "./UIComponents";

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

interface RecommendationSnapshot {
  timestamp: Date;
  events: UserEvent[];
  recommendations: Array<{
    item_id: string;
    score?: number;
  }>;
}

interface UserJourneyVisualizationProps {
  events: UserEvent[];
  recommendationHistory: RecommendationSnapshot[];
  selectedUserId: string;
}

export function UserJourneyVisualization({
  events,
  recommendationHistory,
  selectedUserId,
}: UserJourneyVisualizationProps) {
  const timelineData = useMemo(() => {
    const timeline: Array<{
      timestamp: Date;
      type: "event" | "recommendation";
      data: any;
      eventType?: string;
      itemId?: string;
      score?: number;
    }> = [];

    // Add events to timeline
    events.forEach((event) => {
      timeline.push({
        timestamp: event.addedAt,
        type: "event",
        data: event,
        eventType: event.typeName,
        itemId: event.item_id,
      });
    });

    // Add recommendation snapshots to timeline
    recommendationHistory.forEach((snapshot) => {
      timeline.push({
        timestamp: snapshot.timestamp,
        type: "recommendation",
        data: snapshot,
        score: snapshot.recommendations[0]?.score,
      });
    });

    // Sort by timestamp
    return timeline.sort(
      (a, b) => a.timestamp.getTime() - b.timestamp.getTime()
    );
  }, [events, recommendationHistory]);

  const getEventTypeColor = (eventType: string) => {
    const colors: Record<string, string> = {
      View: "#9e9e9e",
      Click: "#2196f3",
      "Add to Cart": "#ff9800",
      Purchase: "#4caf50",
    };
    return colors[eventType] || "#666";
  };

  const getEventTypeIcon = (eventType: string) => {
    const icons: Record<string, string> = {
      View: "👁️",
      Click: "👆",
      "Add to Cart": "🛒",
      Purchase: "💰",
    };
    return icons[eventType] || "📝";
  };

  if (!selectedUserId || timelineData.length === 0) {
    return (
      <Section title="User Journey Timeline">
        <div
          style={{
            textAlign: "center",
            color: "#666",
            padding: "40px 20px",
            backgroundColor: "#fafafa",
            borderRadius: 6,
            border: "1px solid #e0e0e0",
          }}
        >
          <p style={{ margin: 0, fontSize: 14 }}>
            {!selectedUserId
              ? "Select a user to view their journey timeline"
              : "No events or recommendations to display"}
          </p>
        </div>
      </Section>
    );
  }

  return (
    <Section title="User Journey Timeline">
      <div style={{ marginBottom: 16 }}>
        <p style={{ color: "#666", fontSize: 12, margin: 0 }}>
          Timeline showing events and recommendation changes for{" "}
          <strong>{selectedUserId}</strong>
        </p>
      </div>

      <div
        style={{
          border: "1px solid #e0e0e0",
          borderRadius: 6,
          padding: 16,
          backgroundColor: "#fafafa",
          maxHeight: 400,
          overflow: "auto",
        }}
      >
        <div style={{ position: "relative" }}>
          {/* Timeline line */}
          <div
            style={{
              position: "absolute",
              left: 20,
              top: 0,
              bottom: 0,
              width: 2,
              backgroundColor: "#ddd",
            }}
          />

          {timelineData.map((item, index) => (
            <div
              key={`${item.type}-${index}`}
              style={{
                position: "relative",
                marginBottom: 16,
                paddingLeft: 50,
              }}
            >
              {/* Timeline dot */}
              <div
                style={{
                  position: "absolute",
                  left: 12,
                  top: 8,
                  width: 16,
                  height: 16,
                  borderRadius: "50%",
                  backgroundColor:
                    item.type === "event"
                      ? getEventTypeColor(item.eventType || "")
                      : "#9c27b0",
                  border: "2px solid white",
                  boxShadow: "0 0 0 2px #ddd",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  fontSize: 8,
                }}
              >
                {item.type === "event"
                  ? getEventTypeIcon(item.eventType || "")
                  : "🎯"}
              </div>

              {/* Content */}
              <div
                style={{
                  backgroundColor: "white",
                  border: "1px solid #e0e0e0",
                  borderRadius: 6,
                  padding: 12,
                  boxShadow: "0 1px 3px rgba(0,0,0,0.1)",
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
                  <div
                    style={{ display: "flex", alignItems: "center", gap: 8 }}
                  >
                    <span
                      style={{
                        fontSize: 12,
                        fontWeight: 600,
                        color: item.type === "event" ? "#333" : "#9c27b0",
                      }}
                    >
                      {item.type === "event"
                        ? item.eventType
                        : "Recommendations Updated"}
                    </span>
                    <span
                      style={{
                        fontSize: 10,
                        color: "#666",
                        backgroundColor: "#f0f0f0",
                        padding: "2px 6px",
                        borderRadius: 3,
                      }}
                    >
                      {item.timestamp.toLocaleTimeString()}
                    </span>
                  </div>
                </div>

                {item.type === "event" ? (
                  <div>
                    <p style={{ margin: 0, fontSize: 12, color: "#555" }}>
                      <strong>Item:</strong> {item.itemId}
                    </p>
                    <p
                      style={{
                        margin: "4px 0 0 0",
                        fontSize: 12,
                        color: "#555",
                      }}
                    >
                      <strong>Value:</strong> {item.data.value}
                    </p>
                  </div>
                ) : (
                  <div>
                    <p style={{ margin: 0, fontSize: 12, color: "#555" }}>
                      <strong>Top Recommendations:</strong>
                    </p>
                    <div
                      style={{
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 4,
                        marginTop: 4,
                      }}
                    >
                      {item.data.recommendations
                        .slice(0, 5)
                        .map((rec: any, i: number) => (
                          <span
                            key={rec.item_id}
                            style={{
                              backgroundColor: "#e3f2fd",
                              color: "#1565c0",
                              padding: "2px 6px",
                              borderRadius: 3,
                              fontSize: 10,
                              border: "1px solid #bbdefb",
                            }}
                          >
                            #{i + 1} {rec.item_id} ({rec.score?.toFixed(3)})
                          </span>
                        ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Summary Stats */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(150px, 1fr))",
          gap: 12,
          marginTop: 16,
        }}
      >
        <div
          style={{
            backgroundColor: "#f8f9fa",
            border: "1px solid #e0e0e0",
            borderRadius: 6,
            padding: 12,
            textAlign: "center",
          }}
        >
          <div style={{ fontSize: 20, fontWeight: 600, color: "#333" }}>
            {events.length}
          </div>
          <div style={{ fontSize: 12, color: "#666" }}>Total Events</div>
        </div>

        <div
          style={{
            backgroundColor: "#f8f9fa",
            border: "1px solid #e0e0e0",
            borderRadius: 6,
            padding: 12,
            textAlign: "center",
          }}
        >
          <div style={{ fontSize: 20, fontWeight: 600, color: "#333" }}>
            {recommendationHistory.length}
          </div>
          <div style={{ fontSize: 12, color: "#666" }}>Rec. Updates</div>
        </div>

        <div
          style={{
            backgroundColor: "#f8f9fa",
            border: "1px solid #e0e0e0",
            borderRadius: 6,
            padding: 12,
            textAlign: "center",
          }}
        >
          <div style={{ fontSize: 20, fontWeight: 600, color: "#333" }}>
            {events.filter((e) => e.typeName === "Purchase").length}
          </div>
          <div style={{ fontSize: 12, color: "#666" }}>Purchases</div>
        </div>

        <div
          style={{
            backgroundColor: "#f8f9fa",
            border: "1px solid #e0e0e0",
            borderRadius: 6,
            padding: 12,
            textAlign: "center",
          }}
        >
          <div style={{ fontSize: 20, fontWeight: 600, color: "#333" }}>
            {new Set(events.map((e) => e.item_id)).size}
          </div>
          <div style={{ fontSize: 12, color: "#666" }}>Unique Items</div>
        </div>
      </div>
    </Section>
  );
}
