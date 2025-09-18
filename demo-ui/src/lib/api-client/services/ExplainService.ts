/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_ExplainLLMRequest } from '../models/types_ExplainLLMRequest';
import type { types_ExplainLLMResponse } from '../models/types_ExplainLLMResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ExplainService {
    /**
     * Generate RCA explanation via LLM
     * @param payload Explain request
     * @returns types_ExplainLLMResponse OK
     * @throws ApiError
     */
    public static postV1ExplainLlm(
        payload: types_ExplainLLMRequest,
    ): CancelablePromise<types_ExplainLLMResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/explain/llm',
            body: payload,
            errors: {
                400: `Bad Request`,
                500: `Internal Server Error`,
            },
        });
    }
}
