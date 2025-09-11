import { useMemo, useRef, useState } from "react";
import { OpenAPI } from "./lib/api-client";
import type { types_ScoredItem } from "./lib/api-client";
import {
  NamespaceSection,
  SeedDataSection,
  RecommendationsSection,
  SimilarItemsSection,
  DataManagementSection,
  UserSessionSimulator,
  Button,
} from "./components";
import type { TraitConfig } from "./components/UserTraitsEditor";
import type { ItemConfig, PriceRange } from "./components/ItemConfigEditor";

export interface EventTypeConfig {
  id: string;
  title: string;
  index: number;
  weight: number;
  halfLifeDays: number;
}

/**
 * Demo UI for RecSys.
 *
 * Features:
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
  const [namespace, setNamespace] = useState("default");

  /* Seeding config */
  const [userCount, setUserCount] = useState(50);
  const [userStartIndex, setUserStartIndex] = useState(1);
  const [itemCount, setItemCount] = useState(100);
  const [minEventsPerUser, setMinEventsPerUser] = useState(10);
  const [maxEventsPerUser, setMaxEventsPerUser] = useState(30);
  const [brands] = useState([
    "alfa",
    "bravo",
    "charlie",
    "delta",
    "echo",
    "foxtrot",
  ]);
  const [tags] = useState([
    "action",
    "indie",
    "rpg",
    "strategy",
    "sim",
    "puzzle",
  ]);

  /* Event-type config */
  const [eventTypes, setEventTypes] = useState<EventTypeConfig[]>([
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

  /* Blend controls for recommendations */
  const [blend, setBlend] = useState({ pop: 1.0, cooc: 0.5, als: 0.0 });
  const [k, setK] = useState(20);

  /* User trait configuration */
  const [traitConfigs, setTraitConfigs] = useState<TraitConfig[]>([
    {
      key: "plan",
      probability: 1.0,
      values: [
        { value: "free", probability: 0.6 },
        { value: "plus", probability: 0.3 },
        { value: "pro", probability: 0.1 },
      ],
    },
    {
      key: "age_group",
      probability: 0.8,
      values: [
        { value: "18-24", probability: 0.2 },
        { value: "25-34", probability: 0.3 },
        { value: "35-44", probability: 0.25 },
        { value: "45-54", probability: 0.15 },
        { value: "55+", probability: 0.1 },
      ],
    },
    {
      key: "interests",
      probability: 0.7,
      values: [
        { value: "gaming", probability: 0.4 },
        { value: "music", probability: 0.3 },
        { value: "movies", probability: 0.2 },
        { value: "books", probability: 0.1 },
      ],
    },
  ]);

  /* Item configuration */
  const [itemConfigs, setItemConfigs] = useState<ItemConfig[]>([
    {
      key: "category",
      probability: 0.9,
      values: [
        { value: "electronics", probability: 0.3 },
        { value: "clothing", probability: 0.25 },
        { value: "books", probability: 0.2 },
        { value: "home", probability: 0.15 },
        { value: "sports", probability: 0.1 },
      ],
    },
    {
      key: "condition",
      probability: 0.7,
      values: [
        { value: "new", probability: 0.6 },
        { value: "used", probability: 0.3 },
        { value: "refurbished", probability: 0.1 },
      ],
    },
  ]);

  const [priceRanges, setPriceRanges] = useState<PriceRange[]>([
    { min: 5, max: 25, probability: 0.4 },
    { min: 25, max: 75, probability: 0.3 },
    { min: 75, max: 150, probability: 0.2 },
    { min: 150, max: 300, probability: 0.1 },
  ]);

  /* Local cache for generated ids to ease testing */
  const generatedUsersRef = useRef<string[]>([]);
  const generatedItemsRef = useRef<string[]>([]);

  /* Logs */
  const [log, setLog] = useState<string>("");

  const exampleItem = useMemo(() => {
    return generatedItemsRef.current[0] || "item-0001";
  }, [generatedItemsRef.current.length]);

  const exampleUser = useMemo(() => {
    return generatedUsersRef.current[0] || "user-0001";
  }, [generatedUsersRef.current.length]);

  /* --------------- UI state for playground --------------- */

  const [recUserId, setRecUserId] = useState("");
  const [recOut, setRecOut] = useState<types_ScoredItem[] | null>(null);
  const [recLoading, setRecLoading] = useState(false);

  const [simItemId, setSimItemId] = useState("");
  const [simOut, setSimOut] = useState<types_ScoredItem[] | null>(null);
  const [simLoading, setSimLoading] = useState(false);

  /* User traits update handler */
  const handleUpdateUser = async (
    userId: string,
    traits: Record<string, any>
  ) => {
    // In a real application, this would call the API to update user traits
    // For now, we'll just log the update
    console.log(`Updating user ${userId} with traits:`, traits);
    // TODO: Implement actual API call to update user traits
    return Promise.resolve();
  };

  /* Item update handler */
  const handleUpdateItem = async (
    itemId: string,
    updates: Record<string, any>
  ) => {
    // In a real application, this would call the API to update item data
    // For now, we'll just log the update
    console.log(`Updating item ${itemId} with data:`, updates);
    // TODO: Implement actual API call to update item data
    return Promise.resolve();
  };

  /* --------------- Render --------------- */

  return (
    <div style={{ padding: 16, fontFamily: "system-ui, sans-serif" }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 8,
        }}
      >
        <h1 style={{ marginTop: 0, marginBottom: 0 }}>RecSys Demo UI</h1>
        <Button
          type="button"
          style={{
            padding: "4px 8px",
            fontSize: 12,
            backgroundColor: "#fff3e0",
            color: "#e65100",
          }}
          onClick={() => {
            window.open(
              `${SWAGGER_UI_URL}/docs/`,
              "_blank",
              "noopener,noreferrer"
            );
          }}
        >
          🔍 Explore API
        </Button>
      </div>
      <p style={{ color: "#444" }}>
        Namespace-aware demo. Seed synthetic users, items, and events, tune
        event-type weights and half-life, then explore top-K with reasons and
        similar-items.
      </p>

      <NamespaceSection
        namespace={namespace}
        setNamespace={setNamespace}
        apiBase={API_BASE}
      />

      <SeedDataSection
        userCount={userCount}
        setUserCount={setUserCount}
        userStartIndex={userStartIndex}
        setUserStartIndex={setUserStartIndex}
        itemCount={itemCount}
        setItemCount={setItemCount}
        minEventsPerUser={minEventsPerUser}
        setMinEventsPerUser={setMinEventsPerUser}
        maxEventsPerUser={maxEventsPerUser}
        setMaxEventsPerUser={setMaxEventsPerUser}
        eventTypes={eventTypes}
        setEventTypes={setEventTypes}
        namespace={namespace}
        brands={brands}
        tags={tags}
        log={log}
        setLog={setLog}
        setGeneratedUsers={(users) => {
          generatedUsersRef.current = users;
        }}
        setGeneratedItems={(items) => {
          generatedItemsRef.current = items;
        }}
        traitConfigs={traitConfigs}
        setTraitConfigs={setTraitConfigs}
        itemConfigs={itemConfigs}
        setItemConfigs={setItemConfigs}
        priceRanges={priceRanges}
        setPriceRanges={setPriceRanges}
        generatedUsers={generatedUsersRef.current}
        generatedItems={generatedItemsRef.current}
        onUpdateUser={handleUpdateUser}
        onUpdateItem={handleUpdateItem}
      />

      <RecommendationsSection
        recUserId={recUserId}
        setRecUserId={setRecUserId}
        k={k}
        setK={setK}
        blend={blend}
        setBlend={setBlend}
        namespace={namespace}
        exampleUser={exampleUser}
        recOut={recOut}
        setRecOut={setRecOut}
        recLoading={recLoading}
        setRecLoading={setRecLoading}
      />

      <SimilarItemsSection
        simItemId={simItemId}
        setSimItemId={setSimItemId}
        k={k}
        setK={setK}
        namespace={namespace}
        exampleItem={exampleItem}
        simOut={simOut}
        setSimOut={setSimOut}
        simLoading={simLoading}
        setSimLoading={setSimLoading}
      />

      <UserSessionSimulator
        namespace={namespace}
        generatedUsers={generatedUsersRef.current}
        generatedItems={generatedItemsRef.current}
        eventTypes={eventTypes}
        blend={blend}
        k={k}
      />

      <DataManagementSection namespace={namespace} />

      <footer style={{ color: "#666", fontSize: 12 }}>
        Endpoints used: users:upsert, items:upsert, events:batch,
        event-types:upsert, recommendations, items/{"{id}"}/similar, users,
        items, events (list/delete).
      </footer>
    </div>
  );
}
