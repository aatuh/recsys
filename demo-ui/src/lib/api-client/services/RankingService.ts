/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_RecommendRequest } from '../models/types_RecommendRequest';
import type { types_RecommendResponse } from '../models/types_RecommendResponse';
import type { types_ScoredItem } from '../models/types_ScoredItem';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class RankingService {
    /**
     * Get similar items
     * @param itemId Item ID
     * @param namespace Namespace
     * @param k Top-K
     * @returns types_ScoredItem OK
     * @throws ApiError
     */
    public static getV1ItemsSimilar(
        itemId: string,
        namespace: string = 'default',
        k: number = 20,
    ): CancelablePromise<Array<types_ScoredItem>> {
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
     * @param payload Recommend request
     * @returns types_RecommendResponse OK
     * @throws ApiError
     */
    public static postV1Recommendations(
        payload: types_RecommendRequest,
    ): CancelablePromise<types_RecommendResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/recommendations',
            body: payload,
            errors: {
                400: `Bad Request`,
            },
        });
    }
}
