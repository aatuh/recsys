import { useState, useEffect } from "react";
import { OpenAPI } from "./lib/api-client";
import {
  Navigation,
  NamespaceSeedView,
  RecommendationsPlaygroundView,
  BanditPlaygroundView,
  UserSessionView,
  DataManagementView,
  DocumentationView,
  PrivacyPolicyView,
} from "./components";
import { ViewStateProvider } from "./contexts/ViewStateContext";
import type { ViewType } from "./components/Navigation";

/**
 * Demo UI for RecSys.
 *
 * Features:
 * - Navigation between Demo and Documentation views
 * - Namespace selector
 * - Seed generator: items, users, events
 * - Event-type config upsert (weights, half-life)
 * - Recommendation playground with reasons
 * - Similar-items lookup
 *
 * Notes:
 * - Uses generated API client with proper operation IDs.
 */

// Configure API base URL at runtime
const API_BASE =
  (import.meta as any).env?.VITE_API_BASE_URL?.toString() || "/api";
OpenAPI.BASE = API_BASE;

// Configure Swagger UI URL (direct to API server, not through proxy)
const SWAGGER_UI_URL =
  (import.meta as any).env?.VITE_SWAGGER_UI_URL?.toString() ||
  "http://localhost:8000";

/* --------------- App component --------------- */

export default function App() {
  const [activeView, setActiveView] = useState<ViewType>("namespace-seed");
  const [namespace, setNamespace] = useState("default");

  // Handle URL parameters for view state
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const viewParam = urlParams.get("view") as ViewType;
    const namespaceParam = urlParams.get("namespace");

    if (
      viewParam &&
      [
        "namespace-seed",
        "recommendations-playground",
        "bandit-playground",
        "user-session",
        "data-management",
        "documentation",
        "privacy-policy",
      ].includes(viewParam)
    ) {
      setActiveView(viewParam);
    }

    if (namespaceParam) {
      setNamespace(namespaceParam);
    }
  }, []);

  // Update URL when view or namespace changes
  useEffect(() => {
    const url = new URL(window.location.href);
    url.searchParams.set("view", activeView);
    url.searchParams.set("namespace", namespace);
    window.history.replaceState({}, "", url.toString());
  }, [activeView, namespace]);

  // Shared state for generated data that needs to be passed between views
  const [generatedUsers, setGeneratedUsers] = useState<string[]>([]);
  const [generatedItems, setGeneratedItems] = useState<string[]>([]);

  // Shared configuration state
  const [eventTypes, setEventTypes] = useState([
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
  const [blend, setBlend] = useState({ pop: 1.0, cooc: 0.5, als: 0.0 });
  const [k, setK] = useState(20);

  return (
    <ViewStateProvider>
      <div style={{ fontFamily: "system-ui, sans-serif" }}>
        <Navigation
          activeView={activeView}
          onViewChange={setActiveView}
          apiBase={API_BASE}
          swaggerUrl={SWAGGER_UI_URL}
          namespace={namespace}
        />

        {activeView === "namespace-seed" && (
          <NamespaceSeedView
            namespace={namespace}
            setNamespace={setNamespace}
            apiBase={API_BASE}
            setGeneratedUsers={setGeneratedUsers}
            setGeneratedItems={setGeneratedItems}
          />
        )}

        {activeView === "recommendations-playground" && (
          <RecommendationsPlaygroundView
            namespace={namespace}
            generatedUsers={generatedUsers}
            generatedItems={generatedItems}
          />
        )}

        {activeView === "bandit-playground" && (
          <BanditPlaygroundView
            namespace={namespace}
            generatedUsers={generatedUsers}
          />
        )}

        {activeView === "user-session" && (
          <UserSessionView
            namespace={namespace}
            generatedUsers={generatedUsers}
            setGeneratedUsers={setGeneratedUsers}
            generatedItems={generatedItems}
            setGeneratedItems={setGeneratedItems}
            eventTypes={eventTypes}
            blend={blend}
            k={k}
          />
        )}

        {activeView === "data-management" && (
          <DataManagementView namespace={namespace} />
        )}

        {activeView === "documentation" && <DocumentationView />}

        {activeView === "privacy-policy" && <PrivacyPolicyView />}
      </div>
    </ViewStateProvider>
  );
}
