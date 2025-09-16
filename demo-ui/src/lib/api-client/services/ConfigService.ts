/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { definitions_types_EventTypeConfigUpsertRequest } from '../models/definitions_types_EventTypeConfigUpsertRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ConfigService {
    /**
     * List effective event-type config
     * @param namespace Namespace
     * @returns any OK
     * @throws ApiError
     */
    public static getV1EventTypes(
        namespace: any,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/event-types',
            query: {
                'namespace': namespace,
            },
        });
    }
    /**
     * Upsert tenant event-type config
     * @param requestBody Event types
     * @returns any Accepted
     * @throws ApiError
     */
    public static upsertEventTypes(
        requestBody: definitions_types_EventTypeConfigUpsertRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/event-types:upsert',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
}
