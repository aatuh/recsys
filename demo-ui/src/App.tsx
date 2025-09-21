import { useState } from "react";
import {
  Navigation,
  NamespaceSeedView,
  RecommendationsPlaygroundView,
  BanditPlaygroundView,
  UserSessionView,
  DataManagementView,
  RulesView,
  DocumentationView,
  PrivacyPolicyView,
  ExplainLLMView,
} from "./components";
import { config } from "./config";
import { AppShell, ErrorBoundary } from "./ui/AppShell";
import "./ui/global.css";
import { useQuerySync } from "./hooks/useQuerySync";
import { ViewStateProvider } from "./contexts/ViewStateContext";
import { ThemeProvider } from "./contexts/ThemeContext";
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

// Centralized configuration is in ./config

/* --------------- App component --------------- */

export default function App() {
  const validViews: readonly ViewType[] = [
    "namespace-seed",
    "recommendations-playground",
    "bandit-playground",
    "user-session",
    "data-management",
    "rules",
    "documentation",
    "explain-llm",
    "privacy-policy",
  ];
  const [activeView, setActiveView] = useQuerySync<ViewType>(
    "view",
    "namespace-seed",
    validViews,
    { storageKey: "recsys-active-view" }
  );
  const [namespace, setNamespace] = useQuerySync<string>(
    "namespace",
    "default",
    undefined,
    { storageKey: "recsys-namespace" }
  );

  // Shared state for generated data that needs to be passed between views
  const [generatedUsers, setGeneratedUsers] = useState<string[]>([]);
  const [generatedItems, setGeneratedItems] = useState<string[]>([]);

  // Shared configuration state
  const [eventTypes, _setEventTypes] = useState([
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
  const [blend, _setBlend] = useState({ pop: 1.0, cooc: 0.5, als: 0.0 });
  const [k, _setK] = useState(20);

  return (
    <ThemeProvider>
      <ViewStateProvider>
        <ErrorBoundary>
          {/* Skip link for screen readers */}
          <a href="#main-content" className="skip-link">
            Skip to main content
          </a>
          <AppShell
            header={
              <Navigation
                activeView={activeView}
                onViewChange={setActiveView}
                swaggerUrl={config.swaggerUiUrl}
                customChatGptUrl={config.customChatGptUrl}
                namespace={namespace}
              />
            }
          >
            {activeView === "namespace-seed" && (
              <NamespaceSeedView
                namespace={namespace}
                setNamespace={setNamespace}
                apiBase={config.apiBase}
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

            {activeView === "rules" && <RulesView namespace={namespace} />}

            {activeView === "documentation" && <DocumentationView />}

            {activeView === "explain-llm" && (
              <ExplainLLMView namespace={namespace} />
            )}

            {activeView === "privacy-policy" && <PrivacyPolicyView />}
          </AppShell>
        </ErrorBoundary>
      </ViewStateProvider>
    </ThemeProvider>
  );
}
