import { OpenAPI } from "../core/OpenAPI";
import { request as __request } from "../core/request";
import type { CancelablePromise } from "../core/CancelablePromise";
import type {
  RulePayload,
  RuleResponse,
  RulesListResponse,
  RuleDryRunRequest,
  RuleDryRunResponse,
} from "../models/specs_types_RulePayload";

export class RuleService {
  /**
   * Create a new rule
   */
  public static postV1AdminRules(
    payload: RulePayload
  ): CancelablePromise<RuleResponse> {
    return __request(OpenAPI, {
      method: "POST",
      url: "/v1/admin/rules",
      body: payload,
    });
  }

  /**
   * Update an existing rule
   */
  public static putV1AdminRulesRuleId(
    ruleId: string,
    payload: RulePayload
  ): CancelablePromise<RuleResponse> {
    return __request(OpenAPI, {
      method: "PUT",
      url: `/v1/admin/rules/${ruleId}`,
      body: payload,
    });
  }

  /**
   * List rules with optional filters
   */
  public static getV1AdminRules(params?: {
    namespace?: string;
    surface?: string;
    segment_id?: string;
    enabled?: boolean;
    active_now?: boolean;
    action?: string;
    target_type?: string;
  }): CancelablePromise<RulesListResponse> {
    const searchParams = new URLSearchParams();
    if (params?.namespace) searchParams.set("namespace", params.namespace);
    if (params?.surface) searchParams.set("surface", params.surface);
    if (params?.segment_id) searchParams.set("segment_id", params.segment_id);
    if (params?.enabled !== undefined)
      searchParams.set("enabled", params.enabled.toString());
    if (params?.active_now !== undefined)
      searchParams.set("active_now", params.active_now.toString());
    if (params?.action) searchParams.set("action", params.action);
    if (params?.target_type)
      searchParams.set("target_type", params.target_type);

    const queryString = searchParams.toString();
    const url = queryString
      ? `/v1/admin/rules?${queryString}`
      : "/v1/admin/rules";

    return __request(OpenAPI, {
      method: "GET",
      url,
    });
  }

  /**
   * Dry run rule evaluation
   */
  public static postV1AdminRulesDryRun(
    payload: RuleDryRunRequest
  ): CancelablePromise<RuleDryRunResponse> {
    return __request(OpenAPI, {
      method: "POST",
      url: "/v1/admin/rules/dry-run",
      body: payload,
    });
  }
}
