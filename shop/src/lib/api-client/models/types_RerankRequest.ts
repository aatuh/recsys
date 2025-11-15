/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_Overrides } from './types_Overrides';
import type { types_RecommendBlend } from './types_RecommendBlend';
import type { types_RecommendConstraints } from './types_RecommendConstraints';
import type { types_RerankCandidate } from './types_RerankCandidate';
export type types_RerankRequest = {
    blend?: types_RecommendBlend;
    constraints?: types_RecommendConstraints;
    context?: Record<string, any>;
    explain_level?: types_RerankRequest.explain_level;
    include_reasons?: boolean;
    items?: Array<types_RerankCandidate>;
    'k'?: number;
    namespace?: string;
    overrides?: types_Overrides;
    user_id?: string;
};
export namespace types_RerankRequest {
    export enum explain_level {
        TAGS = 'tags',
        NUMERIC = 'numeric',
        FULL = 'full',
    }
}

