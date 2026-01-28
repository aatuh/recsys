/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_AuditDecisionDetail } from '../models/types_AuditDecisionDetail';
import type { types_AuditDecisionListResponse } from '../models/types_AuditDecisionListResponse';
import type { types_AuditDecisionsSearchRequest } from '../models/types_AuditDecisionsSearchRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class AuditService {
    /**
     * List decision traces with optional filters
     * @param namespace Namespace
     * @param from From timestamp (RFC3339)
     * @param to To timestamp (RFC3339)
     * @param userHash User hash
     * @param requestId Request ID
     * @param limit Limit
     * @returns types_AuditDecisionListResponse OK
     * @throws ApiError
     */
    public static getV1AuditDecisions(
        namespace: string,
        from?: string,
        to?: string,
        userHash?: string,
        requestId?: string,
        limit?: number,
    ): CancelablePromise<types_AuditDecisionListResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/audit/decisions',
            query: {
                'namespace': namespace,
                'from': from,
                'to': to,
                'user_hash': userHash,
                'request_id': requestId,
                'limit': limit,
            },
            errors: {
                400: `Bad Request`,
            },
        });
    }
    /**
     * Get full decision trace by ID
     * @param decisionId Decision ID
     * @returns types_AuditDecisionDetail OK
     * @throws ApiError
     */
    public static getV1AuditDecisions1(
        decisionId: string,
    ): CancelablePromise<types_AuditDecisionDetail> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/audit/decisions/{decision_id}',
            path: {
                'decision_id': decisionId,
            },
            errors: {
                400: `Bad Request`,
                404: `Not Found`,
            },
        });
    }
    /**
     * Search decision traces with advanced filters
     * @param payload Search request
     * @returns types_AuditDecisionListResponse OK
     * @throws ApiError
     */
    public static postV1AuditSearch(
        payload: types_AuditDecisionsSearchRequest,
    ): CancelablePromise<types_AuditDecisionListResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/audit/search',
            body: payload,
            errors: {
                400: `Bad Request`,
            },
        });
    }
}
