import React from "react";
import { UserSessionSimulator } from "../sections/UserSessionSimulator";
import { useViewState } from "../../contexts/ViewStateContext";
import type { EventTypeConfig } from "../../types";
import { spacing, text } from "../../ui/tokens";

interface UserSessionViewProps {
  namespace: string;
  generatedUsers: string[];
  setGeneratedUsers: (value: string[]) => void;
  generatedItems: string[];
  setGeneratedItems: (value: string[]) => void;
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
    <div style={{ padding: spacing.xl, fontFamily: "system-ui, sans-serif" }}>
      <p style={{ color: "#444", marginBottom: spacing.xl, fontSize: text.md }}>
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
