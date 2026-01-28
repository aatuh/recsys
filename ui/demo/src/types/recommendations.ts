export type ScoredItem = {
  id: string;
  title?: string;
  score?: number;
  reasons?: string[];
  delta?: number;
  brand?: string;
  annotation?: string;
  muted?: boolean;
  price?: number;
  tags?: string[];
  available?: boolean;
  position?: number;
};
