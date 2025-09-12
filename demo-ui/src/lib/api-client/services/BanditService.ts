/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_Ack } from '../models/types_Ack';
import type { types_BanditDecideRequest } from '../models/types_BanditDecideRequest';
import type { types_BanditDecideResponse } from '../models/types_BanditDecideResponse';
import type { types_BanditPoliciesUpsertRequest } from '../models/types_BanditPoliciesUpsertRequest';
import type { types_BanditPolicy } from '../models/types_BanditPolicy';
import type { types_BanditRewardRequest } from '../models/types_BanditRewardRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class BanditService {
    /**
     * Decide best policy for this request context
     * @param payload Decision request
     * @returns types_BanditDecideResponse OK
     * @throws ApiError
     */
    public static postV1BanditDecide(
        payload: types_BanditDecideRequest,
    ): CancelablePromise<types_BanditDecideResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/bandit/decide',
            body: payload,
        });
    }
    /**
     * List active bandit policies
     * @param namespace Namespace
     * @returns types_BanditPolicy OK
     * @throws ApiError
     */
    public static getV1BanditPolicies(
        namespace: string,
    ): CancelablePromise<Array<types_BanditPolicy>> {
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
     * @param payload Policies
     * @returns types_Ack Accepted
     * @throws ApiError
     */
    public static upsertBanditPolicies(
        payload: types_BanditPoliciesUpsertRequest,
    ): CancelablePromise<types_Ack> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/bandit/policies:upsert',
            body: payload,
        });
    }
    /**
     * Report binary reward for a previous decision
     * @param payload Reward request
     * @returns types_Ack Accepted
     * @throws ApiError
     */
    public static postV1BanditReward(
        payload: types_BanditRewardRequest,
    ): CancelablePromise<types_Ack> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/bandit/reward',
            body: payload,
        });
    }
}
