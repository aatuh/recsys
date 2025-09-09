/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_Ack } from '../models/types_Ack';
import type { types_EventsBatchRequest } from '../models/types_EventsBatchRequest';
import type { types_ItemsUpsertRequest } from '../models/types_ItemsUpsertRequest';
import type { types_UsersUpsertRequest } from '../models/types_UsersUpsertRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class IngestionService {
    /**
     * Ingest events (batch)
     * @param payload Events batch
     * @returns types_Ack Accepted
     * @throws ApiError
     */
    public static postV1Events:batch(
        payload: types_EventsBatchRequest,
    ): CancelablePromise<types_Ack> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/events:batch',
            body: payload,
            errors: {
                400: `Bad Request`,
            },
        });
    }
    /**
     * Upsert items (batch)
     * Create or update items by opaque IDs.
     * @param payload Items upsert
     * @returns types_Ack Accepted
     * @throws ApiError
     */
    public static postV1Items:upsert(
        payload: types_ItemsUpsertRequest,
    ): CancelablePromise<types_Ack> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/items:upsert',
            body: payload,
            errors: {
                400: `Bad Request`,
            },
        });
    }
    /**
     * Upsert users (batch)
     * @param payload Users upsert
     * @returns types_Ack Accepted
     * @throws ApiError
     */
    public static postV1Users:upsert(
        payload: types_UsersUpsertRequest,
    ): CancelablePromise<types_Ack> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/users:upsert',
            body: payload,
            errors: {
                400: `Bad Request`,
            },
        });
    }
}
