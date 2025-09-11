/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_Ack } from '../models/types_Ack';
import type { types_EventTypeConfigUpsertRequest } from '../models/types_EventTypeConfigUpsertRequest';
import type { types_EventTypeConfigUpsertResponse } from '../models/types_EventTypeConfigUpsertResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ConfigService {
    /**
     * List effective event-type config
     * @param namespace Namespace
     * @returns types_EventTypeConfigUpsertResponse OK
     * @throws ApiError
     */
    public static getV1EventTypes(
        namespace: string,
    ): CancelablePromise<Array<types_EventTypeConfigUpsertResponse>> {
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
     * @param payload Event types
     * @returns types_Ack Accepted
     * @throws ApiError
     */
    public static upsertEventTypes(
        payload: types_EventTypeConfigUpsertRequest,
    ): CancelablePromise<types_Ack> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/event-types:upsert',
            body: payload,
        });
    }
}
