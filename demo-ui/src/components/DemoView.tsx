import { useMemo, useRef, useState } from "react";
import type {
  internal_http_types_ScoredItem,
  types_Overrides,
} from "../lib/api-client";
import {
  NamespaceSection,
  SeedDataSection,
  RecommendationsSection,
  SimilarItemsSection,
  DataManagementSection,
  UserSessionSimulator,
} from "./";
import type { TraitConfig } from "./UserTraitsEditor";
import type { ItemConfig, PriceRange } from "./ItemConfigEditor";
import type { EventTypeConfig } from "../types";

interface DemoViewProps {
  namespace: string;
  setNamespace: (value: string) => void;
  apiBase: string;
}

export function DemoView({ namespace, setNamespace, apiBase }: DemoViewProps) {
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
  const [overrides, _setOverrides] = useState<types_Overrides | null>(null);

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

  /* State for generated users and items to trigger re-renders */
  const [generatedUsers, setGeneratedUsers] = useState<string[]>([]);
  const [generatedItems, setGeneratedItems] = useState<string[]>([]);

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
  const [recOut, setRecOut] = useState<internal_http_types_ScoredItem[] | null>(
    null
  );
  const [recLoading, setRecLoading] = useState(false);

  const [simItemId, setSimItemId] = useState("");
  const [simOut, setSimOut] = useState<internal_http_types_ScoredItem[] | null>(
    null
  );
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

  return (
    <div style={{ padding: 16, fontFamily: "system-ui, sans-serif" }}>
      <p style={{ color: "#444", marginBottom: 24 }}>
        Namespace-aware demo. Seed synthetic users, items, and events, tune
        event-type weights and half-life, then explore top-K with reasons and
        similar-items.
      </p>

      <NamespaceSection
        namespace={namespace}
        setNamespace={setNamespace}
        apiBase={apiBase}
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
          setGeneratedUsers(users);
        }}
        setGeneratedItems={(items) => {
          generatedItemsRef.current = items;
          setGeneratedItems(items);
        }}
        traitConfigs={traitConfigs}
        setTraitConfigs={setTraitConfigs}
        itemConfigs={itemConfigs}
        setItemConfigs={setItemConfigs}
        priceRanges={priceRanges}
        setPriceRanges={setPriceRanges}
        generatedUsers={generatedUsers}
        generatedItems={generatedItems}
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
        overrides={overrides}
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
        generatedUsers={generatedUsers}
        setGeneratedUsers={setGeneratedUsers}
        generatedItems={generatedItems}
        setGeneratedItems={setGeneratedItems}
        eventTypes={eventTypes}
        blend={blend}
        k={k}
      />

      <DataManagementSection namespace={namespace} />
    </div>
  );
}
