/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_Overrides } from './types_Overrides';
import type { types_RecommendBlend } from './types_RecommendBlend';
import type { types_RecommendConstraints } from './types_RecommendConstraints';
export type types_RecommendWithBanditRequest = {
    algorithm?: string;
    blend?: types_RecommendBlend;
    candidate_policy_ids?: Array<string>;
    constraints?: types_RecommendConstraints;
    context?: Record<string, string>;
    explain_level?: types_RecommendWithBanditRequest.explain_level;
    include_reasons?: boolean;
    'k'?: number;
    namespace?: string;
    overrides?: types_Overrides;
    surface?: string;
    user_id?: string;
};
export namespace types_RecommendWithBanditRequest {
    export enum explain_level {
        TAGS = 'tags',
        NUMERIC = 'numeric',
        FULL = 'full',
    }
}

