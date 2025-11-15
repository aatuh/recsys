/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { handlers_VersionResponse } from '../models/handlers_VersionResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class MetaService {
    /**
     * Describe the running build
     * @returns handlers_VersionResponse OK
     * @throws ApiError
     */
    public static getVersion(): CancelablePromise<handlers_VersionResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/version',
        });
    }
}
