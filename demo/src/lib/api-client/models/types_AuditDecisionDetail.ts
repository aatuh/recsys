/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_AuditTraceBandit } from './types_AuditTraceBandit';
import type { types_AuditTraceCandidate } from './types_AuditTraceCandidate';
import type { types_AuditTraceCap } from './types_AuditTraceCap';
import type { types_AuditTraceConfig } from './types_AuditTraceConfig';
import type { types_AuditTraceConstraints } from './types_AuditTraceConstraints';
import type { types_AuditTraceFinalItem } from './types_AuditTraceFinalItem';
import type { types_AuditTraceMMR } from './types_AuditTraceMMR';
export type types_AuditDecisionDetail = {
    bandit?: types_AuditTraceBandit;
    candidates_pre?: Array<types_AuditTraceCandidate>;
    caps?: Record<string, types_AuditTraceCap>;
    constraints?: types_AuditTraceConstraints;
    decision_id?: string;
    effective_config?: types_AuditTraceConfig;
    extras?: Record<string, any>;
    final_items?: Array<types_AuditTraceFinalItem>;
    'k'?: number;
    mmr_info?: Array<types_AuditTraceMMR>;
    namespace?: string;
    org_id?: string;
    request_id?: string;
    surface?: string;
    ts?: string;
    user_hash?: string;
};

