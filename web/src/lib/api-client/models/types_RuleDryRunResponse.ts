/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { types_RuleDryRunPinnedItem } from './types_RuleDryRunPinnedItem';
import type { types_RuleItemEffectResponse } from './types_RuleItemEffectResponse';
import type { types_RuleMatchResponse } from './types_RuleMatchResponse';
export type types_RuleDryRunResponse = {
    pinned_items?: Array<types_RuleDryRunPinnedItem>;
    reason_tags?: Record<string, Array<string>>;
    rule_effects_per_item?: Record<string, types_RuleItemEffectResponse>;
    rules_evaluated?: Array<string>;
    rules_matched?: Array<types_RuleMatchResponse>;
};

