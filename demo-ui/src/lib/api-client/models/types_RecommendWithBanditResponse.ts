/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { internal_http_types_ScoredItem } from './internal_http_types_ScoredItem';
export type types_RecommendWithBanditResponse = {
    algorithm?: string;
    bandit_bucket?: string;
    bandit_explain?: Record<string, string>;
    chosen_policy_id?: string;
    explore?: boolean;
    items?: Array<internal_http_types_ScoredItem>;
    model_version?: string;
};

