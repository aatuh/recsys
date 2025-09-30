export interface RulePayload {
  namespace?: string;
  surface: string;
  name: string;
  description?: string;
  action: string;
  target_type: string;
  target_key?: string;
  item_ids?: string[];
  boost_value?: number;
  max_pins?: number;
  segment_id?: string;
  priority?: number;
  enabled?: boolean;
  valid_from?: string;
  valid_until?: string;
}

export interface RuleResponse {
  rule_id: string;
  namespace: string;
  surface: string;
  name: string;
  description?: string;
  action: string;
  target_type: string;
  target_key?: string;
  item_ids?: string[];
  boost_value?: number;
  max_pins?: number;
  segment_id?: string;
  priority: number;
  enabled: boolean;
  valid_from?: string;
  valid_until?: string;
  created_at: string;
  updated_at: string;
}

export interface RulesListResponse {
  rules: RuleResponse[];
}

export interface RuleDryRunRequest {
  namespace?: string;
  surface: string;
  segment_id?: string;
  items: string[];
}

export interface RuleDryRunResponse {
  rules_evaluated: string[];
  rules_matched: Array<{
    rule_id: string;
    action: string;
    target_type: string;
    item_ids: string[];
  }>;
  rule_effects_per_item: Record<
    string,
    {
      blocked: boolean;
      pinned: boolean;
      boost_delta: number;
    }
  >;
  reason_tags?: Record<string, string[]>;
  pinned_items?: Array<{
    item_id: string;
    rule_ids: string[];
    from_candidates: boolean;
  }>;
}
