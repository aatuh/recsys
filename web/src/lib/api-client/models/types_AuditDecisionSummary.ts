/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_AuditTraceFinalItem } from './types_AuditTraceFinalItem';
export type types_AuditDecisionSummary = {
    decision_id?: string;
    extras?: Record<string, any>;
    final_items?: Array<types_AuditTraceFinalItem>;
    'k'?: number;
    namespace?: string;
    request_id?: string;
    surface?: string;
    ts?: string;
    user_hash?: string;
};

