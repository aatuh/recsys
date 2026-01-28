// UI-only types to avoid loose shapes and improve type safety

// ============================================================================
// Diff and Change Tracking Types
// ============================================================================

export type DiffValue = string | number | boolean | null | undefined | object;

export interface DiffItem {
  field: string;
  before: DiffValue;
  after: DiffValue;
  reason?: string;
}

export type ChangeType = "added" | "removed" | "modified";

export interface ChangeMetadata {
  timestamp: string;
  namespace: string;
  user_id?: string;
  changes: DiffItem[];
}

// ============================================================================
// Form and Input Types
// ============================================================================

export interface ValidationRule {
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  min?: number;
  max?: number;
  pattern?: RegExp;
  custom?: (value: any) => string | null;
}

export type ValidationRules<T> = {
  [K in keyof T]?: ValidationRule;
};

export type ValidationErrors = {
  [key: string]: string | undefined;
};

export type TouchedFields = {
  [key: string]: boolean;
};

// ============================================================================
// Table and Data Display Types
// ============================================================================

export interface TableColumn<T = any> {
  key: keyof T | string;
  label: string;
  sortable?: boolean;
  render?: (value: any, row: T) => React.ReactNode;
  width?: string | number;
  align?: "left" | "center" | "right";
}

export interface TableRow<T = any> {
  id: string | number;
  data: T;
  expanded?: boolean;
  selected?: boolean;
}

export interface PaginationState {
  page: number;
  pageSize: number;
  total: number;
  totalPages?: number;
}

export interface SortState {
  column: string;
  direction: "asc" | "desc";
}

// ============================================================================
// Modal and Dialog Types
// ============================================================================

export interface ModalProps {
  open: boolean;
  onClose: () => void;
  title?: string;
  size?: "sm" | "md" | "lg" | "xl";
  closable?: boolean;
}

export interface DrawerProps {
  open: boolean;
  onClose: () => void;
  title?: string;
  position?: "left" | "right" | "top" | "bottom";
  size?: string | number;
}

// ============================================================================
// Toast and Notification Types
// ============================================================================

export type ToastType = "success" | "error" | "warning" | "info";

export interface ToastMessage {
  id: string;
  type: ToastType;
  title?: string;
  message: string;
  duration?: number;
  action?: {
    label: string;
    onClick: () => void;
  };
}

// ============================================================================
// Event Handler Types
// ============================================================================

export type EventHandler<T = any> = (event: T) => void;

export interface AsyncState<T = any> {
  data: T | null;
  loading: boolean;
  error: Error | null;
}

export interface AsyncAction<T = any> {
  execute: (...args: any[]) => Promise<T>;
  reset: () => void;
}

// ============================================================================
// Component State Types
// ============================================================================

export interface ViewState {
  view: string;
  namespace: string;
  lastUpdated: string;
}

export interface UserSession {
  userId: string;
  events: UserEvent[];
  startTime: string;
  endTime?: string;
}

export interface UserEvent {
  id: string;
  type: string;
  itemId: string;
  timestamp: string;
  metadata?: Record<string, any>;
}

// ============================================================================
// Algorithm and Recommendation Types
// ============================================================================

export interface AlgorithmBlend extends Record<string, DiffValue> {
  pop: number;
  cooc: number;
  als: number;
}

export interface RecommendationRequest {
  userId: string;
  namespace: string;
  k: number;
  blend: AlgorithmBlend;
  overrides?: Record<string, any>;
}

export interface RecommendationResult {
  items: ScoredItem[];
  profileId?: string;
  segmentId?: string;
  metadata?: Record<string, any>;
}

export interface ScoredItem {
  item_id: string;
  score: number;
  explanation?: string;
  metadata?: Record<string, any>;
  boost_value?: number;
  pinned?: boolean;
}

// ============================================================================
// Rule and Policy Types
// ============================================================================

export type RuleAction = "BLOCK" | "PIN" | "BOOST";
export type RuleTargetType = "ITEM" | "TAG" | "BRAND" | "CATEGORY";

export interface RuleFormData {
  name: string;
  description: string;
  action: RuleAction;
  target_type: RuleTargetType;
  target_key: string;
  item_ids: string[];
  boost_value: number;
  max_pins: number;
  segment_id: string;
  priority: number;
  enabled: boolean;
  valid_from: string;
  valid_until: string;
}

export interface RuleDryRunRequest {
  surface: string;
  segment_id: string;
  items: string[];
}

export interface RuleDryRunResult {
  original_items: ScoredItem[];
  filtered_items: ScoredItem[];
  applied_rules: string[];
  metadata?: Record<string, any>;
}

// ============================================================================
// Segment and Profile Types
// ============================================================================

export interface SegmentProfile {
  id: string;
  name: string;
  description?: string;
  rules: SegmentRule[];
  metadata?: Record<string, any>;
}

export interface SegmentRule {
  id: string;
  condition: string;
  action: string;
  priority: number;
  enabled: boolean;
}

// ============================================================================
// Export and Sharing Types
// ============================================================================

export interface ExportData {
  timestamp: string;
  type: "recommendations" | "rules" | "segments" | "changes";
  namespace: string;
  data: any;
  metadata?: Record<string, any>;
}

export interface ShareableLink {
  url: string;
  title: string;
  description?: string;
  expiresAt?: string;
}

// ============================================================================
// Configuration Types
// ============================================================================

export interface UIConfig {
  theme: "light" | "dark" | "auto";
  language: string;
  timezone: string;
  dateFormat: string;
  numberFormat: string;
}

export interface FeatureFlags {
  enableExport: boolean;
  enableSharing: boolean;
  enableAnalytics: boolean;
  enableDebugMode: boolean;
}

// ============================================================================
// Error and Status Types
// ============================================================================

export interface UIError {
  code: string;
  message: string;
  details?: string;
  timestamp: string;
  context?: Record<string, any>;
}

export interface LoadingState {
  isLoading: boolean;
  progress?: number;
  message?: string;
}

export interface StatusMessage {
  type: "info" | "success" | "warning" | "error";
  message: string;
  duration?: number;
  dismissible?: boolean;
}
