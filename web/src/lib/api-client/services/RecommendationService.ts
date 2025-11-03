/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { handlers_RecommendationPresetsResponse } from '../models/handlers_RecommendationPresetsResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class RecommendationService {
    /**
     * List recommendation presets
     * @returns handlers_RecommendationPresetsResponse OK
     * @throws ApiError
     */
    public static getV1AdminRecommendationPresets(): CancelablePromise<handlers_RecommendationPresetsResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/admin/recommendation/presets',
            errors: {
                500: `Internal Server Error`,
            },
        });
    }
}
