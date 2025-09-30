/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type types_DeleteRequest = {
    /**
     * Date range filters
     */
    created_after?: string;
    created_before?: string;
    event_type?: number;
    item_id?: string;
    namespace?: string;
    /**
     * Optional filters - if not provided, deletes all data in namespace
     */
    user_id?: string;
};

