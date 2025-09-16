/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { definitions_types_BanditDecideRequest } from '../models/definitions_types_BanditDecideRequest';
import type { definitions_types_BanditPoliciesUpsertRequest } from '../models/definitions_types_BanditPoliciesUpsertRequest';
import type { definitions_types_BanditRewardRequest } from '../models/definitions_types_BanditRewardRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class BanditService {
    /**
     * Decide best policy for this request context
     * @param requestBody Decision request
     * @returns any OK
     * @throws ApiError
     */
    public static postV1BanditDecide(
        requestBody: definitions_types_BanditDecideRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/bandit/decide',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * List all bandit policies (active and inactive)
     * @param namespace Namespace
     * @returns any OK
     * @throws ApiError
     */
    public static getV1BanditPolicies(
        namespace: any,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/bandit/policies',
            query: {
                'namespace': namespace,
            },
        });
    }
    /**
     * Upsert bandit policies
     * @param requestBody Policies
     * @returns any Accepted
     * @throws ApiError
     */
    public static upsertBanditPolicies(
        requestBody: definitions_types_BanditPoliciesUpsertRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/bandit/policies:upsert',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * Report binary reward for a previous decision
     * @param requestBody Reward request
     * @returns any Accepted
     * @throws ApiError
     */
    public static postV1BanditReward(
        requestBody: definitions_types_BanditRewardRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/bandit/reward',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
}
