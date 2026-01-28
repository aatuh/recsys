/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type types_RecommendConstraints = {
    /**
     * Optional ISO8601 timestamp; only consider events on/after this instant.
     */
    created_after?: string;
    /**
     * Exclude these item IDs from results.
     */
    exclude_item_ids?: Array<string>;
    /**
     * Match if item.tags overlaps these (any). Empty/omitted = no tag filter.
     */
    include_tags_any?: Array<string>;
    /**
     * Optional price bounds: [min, max]. Either end may be omitted.
     */
    price_between?: Array<number>;
};

