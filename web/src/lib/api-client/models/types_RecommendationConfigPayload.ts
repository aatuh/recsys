/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_BanditExperimentConfig } from './types_BanditExperimentConfig';
export type types_RecommendationConfigPayload = {
    bandit_experiment?: types_BanditExperimentConfig;
    blend_alpha?: number;
    blend_beta?: number;
    blend_gamma?: number;
    brand_cap?: number;
    brand_tag_prefixes?: Array<string>;
    category_cap?: number;
    category_tag_prefixes?: Array<string>;
    co_vis_window_days?: number;
    coverage_cache_ttl_seconds?: number;
    coverage_long_tail_hint_threshold?: number;
    exclude_event_types?: Array<number>;
    half_life_days?: number;
    mmr_lambda?: number;
    mmr_presets?: Record<string, number>;
    new_user_blend_alpha?: number;
    new_user_blend_beta?: number;
    new_user_blend_gamma?: number;
    new_user_mmr_lambda?: number;
    new_user_pop_fanout?: number;
    popularity_fanout?: number;
    profile_boost?: number;
    profile_cold_start_multiplier?: number;
    profile_min_events_for_boost?: number;
    profile_starter_blend_weight?: number;
    profile_top_n_tags?: number;
    profile_window_days?: number;
    purchased_window_days?: number;
    rule_exclude_events?: boolean;
    rules_enabled?: boolean;
};

