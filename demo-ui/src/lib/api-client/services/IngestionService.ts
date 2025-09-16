/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { definitions_types_EventsBatchRequest } from '../models/definitions_types_EventsBatchRequest';
import type { definitions_types_ItemsUpsertRequest } from '../models/definitions_types_ItemsUpsertRequest';
import type { definitions_types_UsersUpsertRequest } from '../models/definitions_types_UsersUpsertRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class IngestionService {
    /**
     * Ingest events (batch)
     * @param requestBody Events batch
     * @returns any Accepted
     * @throws ApiError
     */
    public static batchEvents(
        requestBody: definitions_types_EventsBatchRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/events:batch',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad Request`,
            },
        });
    }
    /**
     * Upsert items (batch)
     * Create or update items by opaque IDs.
     * @param requestBody Items upsert
     * @returns any Accepted
     * @throws ApiError
     */
    public static upsertItems(
        requestBody: definitions_types_ItemsUpsertRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/items:upsert',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad Request`,
            },
        });
    }
    /**
     * Upsert users (batch)
     * @param requestBody Users upsert
     * @returns any Accepted
     * @throws ApiError
     */
    public static upsertUsers(
        requestBody: definitions_types_UsersUpsertRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/users:upsert',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad Request`,
            },
        });
    }
}
