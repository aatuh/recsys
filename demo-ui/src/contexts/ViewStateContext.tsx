import React, { createContext, useContext, useState, ReactNode } from "react";
import type { EventTypeConfig } from "../types";
import type { TraitConfig } from "../components/UserTraitsEditor";
import type { ItemConfig, PriceRange } from "../components/ItemConfigEditor";
import type { internal_http_types_ScoredItem } from "../lib/api-client";

// UserSession specific types
export interface UserEvent {
  id: string;
  user_id: string;
  item_id: string;
  type: number;
  typeName: string;
  ts: string;
  value: number;
  addedAt: Date;
}

export interface EventSequence {
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

// Types for view-specific state
interface NamespaceSeedState {
  userCount: number;
  userStartIndex: number;
  itemCount: number;
  minEventsPerUser: number;
  maxEventsPerUser: number;
  eventTypes: EventTypeConfig[];
  traitConfigs: TraitConfig[];
  itemConfigs: ItemConfig[];
  priceRanges: PriceRange[];
  log: string;
}

interface RecommendationsPlaygroundState {
  blend: { pop: number; cooc: number; als: number };
  k: number;
  recUserId: string;
  recOut: internal_http_types_ScoredItem[] | null;
  recLoading: boolean;
  simItemId: string;
  simOut: internal_http_types_ScoredItem[] | null;
  simLoading: boolean;
}

interface DataManagementState {
  dataType: "users" | "items" | "events";
  selectedRows: Set<string>;
  filters: {
    user_id: string;
    item_id: string;
    event_type: string;
    created_after: string;
    created_before: string;
  };
  sortBy: string;
  sortDirection: "asc" | "desc";
  pagination: {
    page: number;
    pageSize: number;
    total: number;
  };
  embeddingsLoading: boolean;
  embeddingsProgress: {
    current: number;
    total: number;
    message: string;
  };
  exportLoading: boolean;
  exportProgress: {
    current: number;
    total: number;
    message: string;
  };
  selectedExportTables: string[];
}

interface UserSessionState {
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
}

interface ViewStateContextType {
  // Namespace Seed View State
  namespaceSeed: NamespaceSeedState;
  setNamespaceSeed: React.Dispatch<React.SetStateAction<NamespaceSeedState>>;

  // Recommendations Playground State
  recommendationsPlayground: RecommendationsPlaygroundState;
  setRecommendationsPlayground: React.Dispatch<
    React.SetStateAction<RecommendationsPlaygroundState>
  >;

  // Data Management State
  dataManagement: DataManagementState;
  setDataManagement: React.Dispatch<React.SetStateAction<DataManagementState>>;

  // User Session State
  userSession: UserSessionState;
  setUserSession: React.Dispatch<React.SetStateAction<UserSessionState>>;

  // Reset functions for each view
  resetNamespaceSeed: () => void;
  resetRecommendationsPlayground: () => void;
  resetDataManagement: () => void;
  resetUserSession: () => void;
}

// Initial state values
const initialNamespaceSeedState: NamespaceSeedState = {
  userCount: 50,
  userStartIndex: 1,
  itemCount: 100,
  minEventsPerUser: 10,
  maxEventsPerUser: 30,
  eventTypes: [
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
  ],
  traitConfigs: [
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
  ],
  itemConfigs: [
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
  ],
  priceRanges: [
    { min: 5, max: 25, probability: 0.4 },
    { min: 25, max: 75, probability: 0.3 },
    { min: 75, max: 150, probability: 0.2 },
    { min: 150, max: 300, probability: 0.1 },
  ],
  log: "",
};

const initialRecommendationsPlaygroundState: RecommendationsPlaygroundState = {
  blend: { pop: 1.0, cooc: 0.5, als: 0.0 },
  k: 20,
  recUserId: "",
  recOut: null,
  recLoading: false,
  simItemId: "",
  simOut: null,
  simLoading: false,
};

const initialDataManagementState: DataManagementState = {
  dataType: "users",
  selectedRows: new Set(),
  filters: {
    user_id: "",
    item_id: "",
    event_type: "",
    created_after: "",
    created_before: "",
  },
  sortBy: "",
  sortDirection: "desc",
  pagination: {
    page: 1,
    pageSize: 25,
    total: 0,
  },
  embeddingsLoading: false,
  embeddingsProgress: {
    current: 0,
    total: 0,
    message: "",
  },
  exportLoading: false,
  exportProgress: {
    current: 0,
    total: 0,
    message: "",
  },
  selectedExportTables: ["users", "items", "events"],
};

const initialUserSessionState: UserSessionState = {
  selectedUserId: "",
  userEvents: [],
  currentRecommendations: null,
  recommendationHistory: [],
  isSimulating: false,
  selectedSequence: "",
  customEventType: 0,
  customItemId: "",
  autoRefresh: true,
  simulationSpeed: 1,
  log: "",
  showSequenceBuilder: false,
  customSequences: [],
  showJourneyViz: false,
  hasAutoLoaded: false,
};

const ViewStateContext = createContext<ViewStateContextType | undefined>(
  undefined
);

interface ViewStateProviderProps {
  children: ReactNode;
}

export function ViewStateProvider({ children }: ViewStateProviderProps) {
  const [namespaceSeed, setNamespaceSeed] = useState<NamespaceSeedState>(
    initialNamespaceSeedState
  );
  const [recommendationsPlayground, setRecommendationsPlayground] =
    useState<RecommendationsPlaygroundState>(
      initialRecommendationsPlaygroundState
    );
  const [dataManagement, setDataManagement] = useState<DataManagementState>(
    initialDataManagementState
  );
  const [userSession, setUserSession] = useState<UserSessionState>(
    initialUserSessionState
  );

  const resetNamespaceSeed = () => {
    setNamespaceSeed(initialNamespaceSeedState);
  };

  const resetRecommendationsPlayground = () => {
    setRecommendationsPlayground(initialRecommendationsPlaygroundState);
  };

  const resetDataManagement = () => {
    setDataManagement(initialDataManagementState);
  };

  const resetUserSession = () => {
    setUserSession(initialUserSessionState);
  };

  const value: ViewStateContextType = {
    namespaceSeed,
    setNamespaceSeed,
    recommendationsPlayground,
    setRecommendationsPlayground,
    dataManagement,
    setDataManagement,
    userSession,
    setUserSession,
    resetNamespaceSeed,
    resetRecommendationsPlayground,
    resetDataManagement,
    resetUserSession,
  };

  return (
    <ViewStateContext.Provider value={value}>
      {children}
    </ViewStateContext.Provider>
  );
}

export function useViewState() {
  const context = useContext(ViewStateContext);
  if (context === undefined) {
    throw new Error("useViewState must be used within a ViewStateProvider");
  }
  return context;
}
