/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_Ack } from '../models/types_Ack';
import type { types_EventTypeConfigUpsertRequest } from '../models/types_EventTypeConfigUpsertRequest';
import type { types_EventTypeConfigUpsertResponse } from '../models/types_EventTypeConfigUpsertResponse';
import type { types_IDListRequest } from '../models/types_IDListRequest';
import type { types_SegmentDryRunRequest } from '../models/types_SegmentDryRunRequest';
import type { types_SegmentDryRunResponse } from '../models/types_SegmentDryRunResponse';
import type { types_SegmentProfilesListResponse } from '../models/types_SegmentProfilesListResponse';
import type { types_SegmentProfilesUpsertRequest } from '../models/types_SegmentProfilesUpsertRequest';
import type { types_SegmentsListResponse } from '../models/types_SegmentsListResponse';
import type { types_SegmentsUpsertRequest } from '../models/types_SegmentsUpsertRequest';
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
    /**
     * List segment profiles
     * @param namespace Namespace
     * @returns types_SegmentProfilesListResponse OK
     * @throws ApiError
     */
    public static getV1SegmentProfiles(
        namespace: string = 'default',
    ): CancelablePromise<types_SegmentProfilesListResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/segment-profiles',
            query: {
                'namespace': namespace,
            },
            errors: {
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Delete segment profiles
     * @param payload IDs
     * @returns types_Ack OK
     * @throws ApiError
     */
    public static segmentProfilesDelete(
        payload: types_IDListRequest,
    ): CancelablePromise<types_Ack> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/segment-profiles:delete',
            body: payload,
            errors: {
                400: `Bad Request`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Upsert segment profiles
     * @param payload Profiles
     * @returns types_Ack OK
     * @throws ApiError
     */
    public static segmentProfilesUpsert(
        payload: types_SegmentProfilesUpsertRequest,
    ): CancelablePromise<types_Ack> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/segment-profiles:upsert',
            body: payload,
            errors: {
                400: `Bad Request`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * List segments with rules
     * @param namespace Namespace
     * @returns types_SegmentsListResponse OK
     * @throws ApiError
     */
    public static getV1Segments(
        namespace: string = 'default',
    ): CancelablePromise<types_SegmentsListResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/v1/segments',
            query: {
                'namespace': namespace,
            },
            errors: {
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Delete segments
     * @param payload IDs
     * @returns types_Ack OK
     * @throws ApiError
     */
    public static segmentsDelete(
        payload: types_IDListRequest,
    ): CancelablePromise<types_Ack> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/segments:delete',
            body: payload,
            errors: {
                400: `Bad Request`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Simulate segment selection for context
     * @param payload Dry run
     * @returns types_SegmentDryRunResponse OK
     * @throws ApiError
     */
    public static segmentsDryRun(
        payload: types_SegmentDryRunRequest,
    ): CancelablePromise<types_SegmentDryRunResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/segments:dry-run',
            body: payload,
            errors: {
                400: `Bad Request`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Upsert a segment and its rules
     * @param payload Segment
     * @returns types_Ack OK
     * @throws ApiError
     */
    public static segmentsUpsert(
        payload: types_SegmentsUpsertRequest,
    ): CancelablePromise<types_Ack> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/v1/segments:upsert',
            body: payload,
            errors: {
                400: `Bad Request`,
                500: `Internal Server Error`,
            },
        });
    }
}
