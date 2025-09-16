/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { definitions_types_RecommendRequest } from '../models/definitions_types_RecommendRequest';
import type { definitions_types_RecommendWithBanditRequest } from '../models/definitions_types_RecommendWithBanditRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class RankingService {
    /**
     * Recommend with bandit-selected policy
     * @param requestBody Request
     * @returns any OK
     * @throws ApiError
     */
    public static postV1BanditRecommendations(
        requestBody: definitions_types_RecommendWithBanditRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/bandit/recommendations',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * Get similar items
     * @param itemId Item ID
     * @param namespace Namespace
     * @param k Top-K
     * @returns any OK
     * @throws ApiError
     */
    public static getV1ItemsSimilar(
        itemId: any,
        namespace?: any,
        k?: any,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/items/{item_id}/similar',
            path: {
                'item_id': itemId,
            },
            query: {
                'namespace': namespace,
                'k': k,
            },
            errors: {
                400: `Bad Request`,
            },
        });
    }
    /**
     * Get recommendations for a user
     * @param requestBody Recommend request
     * @returns any OK
     * @throws ApiError
     */
    public static postV1Recommendations(
        requestBody: definitions_types_RecommendRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/recommendations',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad Request`,
            },
        });
    }
}
