/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_ManualOverrideCancelRequest } from '../models/types_ManualOverrideCancelRequest';
import type { types_ManualOverrideRequest } from '../models/types_ManualOverrideRequest';
import type { types_ManualOverrideResponse } from '../models/types_ManualOverrideResponse';
import type { types_RuleDryRunRequest } from '../models/types_RuleDryRunRequest';
import type { types_RuleDryRunResponse } from '../models/types_RuleDryRunResponse';
import type { types_RulePayload } from '../models/types_RulePayload';
import type { types_RuleResponse } from '../models/types_RuleResponse';
import type { types_RulesListResponse } from '../models/types_RulesListResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class AdminService {
    /**
     * List manual overrides
     * @param namespace Namespace
     * @param surface Surface
     * @param status Filter by status (active,cancelled,expired)
     * @param action Filter by action (boost,suppress)
     * @param includeExpired Include expired overrides
     * @returns types_ManualOverrideResponse OK
     * @throws ApiError
     */
    public static getV1AdminManualOverrides(
        namespace: string,
        surface?: string,
        status?: string,
        action?: string,
        includeExpired?: boolean,
    ): CancelablePromise<Array<types_ManualOverrideResponse>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/admin/manual_overrides',
            query: {
                'namespace': namespace,
                'surface': surface,
                'status': status,
                'action': action,
                'include_expired': includeExpired,
            },
            errors: {
                400: `Bad Request`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Create a manual boost/suppression override
     * @param payload Manual override payload
     * @returns types_ManualOverrideResponse Created
     * @throws ApiError
     */
    public static postV1AdminManualOverrides(
        payload: types_ManualOverrideRequest,
    ): CancelablePromise<types_ManualOverrideResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/admin/manual_overrides',
            body: payload,
            errors: {
                400: `Bad Request`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Cancel a manual override
     * @param overrideId Override ID
     * @param payload Optional cancellation metadata
     * @returns types_ManualOverrideResponse OK
     * @throws ApiError
     */
    public static postV1AdminManualOverridesCancel(
        overrideId: string,
        payload?: types_ManualOverrideCancelRequest,
    ): CancelablePromise<types_ManualOverrideResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/admin/manual_overrides/{override_id}/cancel',
            path: {
                'override_id': overrideId,
            },
            body: payload,
            errors: {
                400: `Bad Request`,
                404: `Not Found`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * List merchandising rules
     * List merchandising rules with optional filtering
     * @param namespace Namespace
     * @param surface Filter by surface
     * @param segmentId Filter by segment ID
     * @param enabled Filter by enabled status
     * @param activeNow Filter by active status
     * @param action Filter by action type
     * @param targetType Filter by target type
     * @returns types_RulesListResponse OK
     * @throws ApiError
     */
    public static getV1AdminRules(
        namespace: string,
        surface?: string,
        segmentId?: string,
        enabled?: boolean,
        activeNow?: boolean,
        action?: string,
        targetType?: string,
    ): CancelablePromise<types_RulesListResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/admin/rules',
            query: {
                'namespace': namespace,
                'surface': surface,
                'segment_id': segmentId,
                'enabled': enabled,
                'active_now': activeNow,
                'action': action,
                'target_type': targetType,
            },
            errors: {
                400: `Bad Request`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Create a merchandising rule
     * Create a new merchandising rule (BLOCK, PIN, BOOST)
     * @param payload Rule payload
     * @returns types_RuleResponse Created
     * @throws ApiError
     */
    public static postV1AdminRules(
        payload: types_RulePayload,
    ): CancelablePromise<types_RuleResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/admin/rules',
            body: payload,
            errors: {
                400: `Bad Request`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Preview rule effects
     * Preview matched rules and effects without mutating state
     * @param payload Dry run request
     * @returns types_RuleDryRunResponse OK
     * @throws ApiError
     */
    public static postV1AdminRulesDryRun(
        payload: types_RuleDryRunRequest,
    ): CancelablePromise<types_RuleDryRunResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/admin/rules/dry-run',
            body: payload,
            errors: {
                400: `Bad Request`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Update a merchandising rule
     * Update an existing merchandising rule
     * @param ruleId Rule ID
     * @param payload Rule payload
     * @returns types_RuleResponse OK
     * @throws ApiError
     */
    public static putV1AdminRules(
        ruleId: string,
        payload: types_RulePayload,
    ): CancelablePromise<types_RuleResponse> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/v1/admin/rules/{rule_id}',
            path: {
                'rule_id': ruleId,
            },
            body: payload,
            errors: {
                400: `Bad Request`,
                404: `Not Found`,
                500: `Internal Server Error`,
            },
        });
    }
}
