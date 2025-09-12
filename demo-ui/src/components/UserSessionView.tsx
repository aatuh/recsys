import React from "react";
import { UserSessionSimulator } from "./";
import { useViewState } from "../contexts/ViewStateContext";
import type { EventTypeConfig } from "../types";

interface UserSessionViewProps {
  namespace: string;
  generatedUsers: string[];
  setGeneratedUsers: (users: string[]) => void;
  generatedItems: string[];
  setGeneratedItems: (items: string[]) => void;
  eventTypes: EventTypeConfig[];
  blend: { pop: number; cooc: number; als: number };
  k: number;
}

export function UserSessionView({
  namespace,
  generatedUsers,
  setGeneratedUsers,
  generatedItems,
  setGeneratedItems,
  eventTypes,
  blend,
  k,
}: UserSessionViewProps) {
  const { userSession, setUserSession } = useViewState();

  return (
    <div style={{ padding: 16, fontFamily: "system-ui, sans-serif" }}>
      <p style={{ color: "#444", marginBottom: 24 }}>
        Simulate realistic user sessions with event sequences. Create user
        journeys, generate events, and observe how the recommendation system
        responds to user behavior patterns.
      </p>

      <UserSessionSimulator
        namespace={namespace}
        generatedUsers={generatedUsers}
        setGeneratedUsers={setGeneratedUsers}
        generatedItems={generatedItems}
        setGeneratedItems={setGeneratedItems}
        eventTypes={eventTypes}
        blend={blend}
        k={k}
        userSession={userSession}
        setUserSession={setUserSession}
      />
    </div>
  );
}
