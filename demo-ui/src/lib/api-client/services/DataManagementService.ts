/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { definitions_types_DeleteRequest } from '../models/definitions_types_DeleteRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class DataManagementService {
    /**
     * List events with pagination and filtering
     * Get a paginated list of events with optional filtering by user_id, item_id, event_type, date range, etc.
     * @param namespace Namespace
     * @param limit Limit (default: 100, max: 1000)
     * @param offset Offset (default: 0)
     * @param userId Filter by user ID
     * @param itemId Filter by item ID
     * @param eventType Filter by event type
     * @param createdAfter Filter by creation date (ISO8601)
     * @param createdBefore Filter by creation date (ISO8601)
     * @returns any OK
     * @throws ApiError
     */
    public static listEvents(
        namespace: any,
        limit?: any,
        offset?: any,
        userId?: any,
        itemId?: any,
        eventType?: any,
        createdAfter?: any,
        createdBefore?: any,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/events',
            query: {
                'namespace': namespace,
                'limit': limit,
                'offset': offset,
                'user_id': userId,
                'item_id': itemId,
                'event_type': eventType,
                'created_after': createdAfter,
                'created_before': createdBefore,
            },
            errors: {
                400: `Bad Request`,
            },
        });
    }
    /**
     * Delete events with optional filtering
     * Delete events based on filters. If no filters provided, deletes all events in namespace.
     * @param requestBody Delete request
     * @returns any OK
     * @throws ApiError
     */
    public static deleteEvents(
        requestBody: definitions_types_DeleteRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/events:delete',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad Request`,
            },
        });
    }
    /**
     * List items with pagination and filtering
     * Get a paginated list of items with optional filtering by item_id, date range, etc.
     * @param namespace Namespace
     * @param limit Limit (default: 100, max: 1000)
     * @param offset Offset (default: 0)
     * @param itemId Filter by item ID
     * @param createdAfter Filter by creation date (ISO8601)
     * @param createdBefore Filter by creation date (ISO8601)
     * @returns any OK
     * @throws ApiError
     */
    public static listItems(
        namespace: any,
        limit?: any,
        offset?: any,
        itemId?: any,
        createdAfter?: any,
        createdBefore?: any,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/items',
            query: {
                'namespace': namespace,
                'limit': limit,
                'offset': offset,
                'item_id': itemId,
                'created_after': createdAfter,
                'created_before': createdBefore,
            },
            errors: {
                400: `Bad Request`,
            },
        });
    }
    /**
     * Delete items with optional filtering
     * Delete items based on filters. If no filters provided, deletes all items in namespace.
     * @param requestBody Delete request
     * @returns any OK
     * @throws ApiError
     */
    public static deleteItems(
        requestBody: definitions_types_DeleteRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/items:delete',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad Request`,
            },
        });
    }
    /**
     * List users with pagination and filtering
     * Get a paginated list of users with optional filtering by user_id, date range, etc.
     * @param namespace Namespace
     * @param limit Limit (default: 100, max: 1000)
     * @param offset Offset (default: 0)
     * @param userId Filter by user ID
     * @param createdAfter Filter by creation date (ISO8601)
     * @param createdBefore Filter by creation date (ISO8601)
     * @returns any OK
     * @throws ApiError
     */
    public static listUsers(
        namespace: any,
        limit?: any,
        offset?: any,
        userId?: any,
        createdAfter?: any,
        createdBefore?: any,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/users',
            query: {
                'namespace': namespace,
                'limit': limit,
                'offset': offset,
                'user_id': userId,
                'created_after': createdAfter,
                'created_before': createdBefore,
            },
            errors: {
                400: `Bad Request`,
            },
        });
    }
    /**
     * Delete users with optional filtering
     * Delete users based on filters. If no filters provided, deletes all users in namespace.
     * @param requestBody Delete request
     * @returns any OK
     * @throws ApiError
     */
    public static deleteUsers(
        requestBody: definitions_types_DeleteRequest,
    ): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/users:delete',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad Request`,
            },
        });
    }
}
