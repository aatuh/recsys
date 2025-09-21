export interface EventTypeConfig {
  id: string;
  title: string;
  index: number;
  weight: number;
  halfLifeDays: number;
}

// Re-export UI types
export * from "./ui";
