/**
 * URL and query parameter validation and sanitization utilities.
 * Provides a single parseSearchParams() utility with comprehensive validation.
 */

import { z } from "zod";

/**
 * Common query parameter schemas for validation.
 */
export const QueryParamSchemas = {
  // String parameters
  string: z.string().min(1).max(1000),
  optionalString: z.string().max(1000).optional(),

  // Numeric parameters
  positiveInt: z.coerce.number().int().positive(),
  nonNegativeInt: z.coerce.number().int().min(0),
  optionalPositiveInt: z.coerce.number().int().positive().optional(),
  optionalNonNegativeInt: z.coerce.number().int().min(0).optional(),

  // Boolean parameters
  boolean: z.coerce.boolean(),
  optionalBoolean: z.coerce.boolean().optional(),

  // Enum parameters
  enum: <T extends readonly [string, ...string[]]>(values: T) => z.enum(values),
  optionalEnum: <T extends readonly [string, ...string[]]>(values: T) =>
    z.enum(values).optional(),

  // Date parameters
  isoDate: z.string().datetime(),
  optionalIsoDate: z.string().datetime().optional(),

  // Array parameters
  stringArray: z.string().transform((val) => val.split(",").filter(Boolean)),
  optionalStringArray: z
    .string()
    .transform((val) => val.split(",").filter(Boolean))
    .optional(),

  // UUID parameters
  uuid: z.string().uuid(),
  optionalUuid: z.string().uuid().optional(),

  // URL parameters
  url: z.string().url(),
  optionalUrl: z.string().url().optional(),
} as const;

/**
 * Configuration for query parameter parsing.
 */
export interface ParseSearchParamsConfig {
  /** Whether to strip unknown parameters */
  stripUnknown?: boolean;
  /** Whether to validate all parameters strictly */
  strict?: boolean;
  /** Custom error messages */
  errorMessages?: Record<string, string>;
  /** Maximum number of parameters allowed */
  maxParams?: number;
  /** Maximum length of parameter values */
  maxValueLength?: number;
}

/**
 * Result of parsing search parameters.
 */
export interface ParseSearchParamsResult<T> {
  /** Parsed and validated parameters */
  data: T;
  /** Any validation errors */
  errors: Record<string, string[]>;
  /** Whether parsing was successful */
  success: boolean;
  /** Raw search parameters before parsing */
  raw: Record<string, string>;
  /** Sanitized search parameters */
  sanitized: Record<string, string>;
}

/**
 * Parse and validate search parameters from a URL or URLSearchParams.
 *
 * @param source - URL string, URLSearchParams, or object to parse
 * @param schema - Zod schema for validation
 * @param config - Optional configuration
 * @returns Parsed and validated parameters with error information
 */
export function parseSearchParams<T>(
  source: string | URLSearchParams | Record<string, string>,
  schema: z.ZodSchema<T>,
  config: ParseSearchParamsConfig = {}
): ParseSearchParamsResult<T> {
  const {
    stripUnknown: _stripUnknown = true,
    strict: _strict = true,
    errorMessages = {},
    maxParams = 50,
    maxValueLength = 1000,
  } = config;

  // Convert source to URLSearchParams
  let searchParams: URLSearchParams;
  if (typeof source === "string") {
    try {
      const url = new URL(source);
      searchParams = url.searchParams;
    } catch {
      // If not a full URL, treat as query string
      searchParams = new URLSearchParams(source);
    }
  } else if (source instanceof URLSearchParams) {
    searchParams = source;
  } else {
    searchParams = new URLSearchParams(source);
  }

  // Convert to object and sanitize
  const raw: Record<string, string> = {};
  const sanitized: Record<string, string> = {};

  for (const [key, value] of searchParams.entries()) {
    // Basic sanitization
    const sanitizedKey = key.trim().slice(0, 100); // Limit key length
    const sanitizedValue = value.trim().slice(0, maxValueLength); // Limit value length

    // Skip empty values
    if (!sanitizedValue) continue;

    // Check parameter count limit
    if (Object.keys(raw).length >= maxParams) {
      break;
    }

    raw[sanitizedKey] = value;
    sanitized[sanitizedKey] = sanitizedValue;
  }

  // Validate with Zod schema
  const result = schema.safeParse(sanitized);

  if (result.success) {
    return {
      data: result.data,
      errors: {},
      success: true,
      raw,
      sanitized,
    };
  }

  // Handle validation errors
  const errors: Record<string, string[]> = {};
  if (result.error) {
    for (const issue of result.error.issues) {
      const path = issue.path.join(".");
      if (!errors[path]) {
        errors[path] = [];
      }

      const message = errorMessages[path] || issue.message;
      errors[path].push(message);
    }
  }

  return {
    data: {} as T, // Empty object on failure
    errors,
    success: false,
    raw,
    sanitized,
  };
}

/**
 * Create a query parameter parser with predefined schema.
 */
export function createQueryParser<T>(schema: z.ZodSchema<T>) {
  return (
    source: string | URLSearchParams | Record<string, string>,
    config?: ParseSearchParamsConfig
  ) => parseSearchParams(source, schema, config);
}

/**
 * Common query parameter schemas for the application.
 */
export const AppQuerySchemas = {
  // View parameter (for navigation)
  view: QueryParamSchemas.enum([
    "namespace-seed",
    "recommendations-playground",
    "bandit-playground",
    "user-session",
    "data-management",
    "rules",
    "documentation",
    "explain-llm",
    "privacy-policy",
  ] as const),

  // Namespace parameter
  namespace: QueryParamSchemas.string,

  // Pagination parameters
  page: QueryParamSchemas.optionalPositiveInt,
  limit: QueryParamSchemas.optionalPositiveInt,
  offset: QueryParamSchemas.optionalNonNegativeInt,

  // Search parameters
  search: QueryParamSchemas.optionalString,
  query: QueryParamSchemas.optionalString,

  // Filter parameters
  userId: QueryParamSchemas.optionalUuid,
  itemId: QueryParamSchemas.optionalUuid,
  eventType: QueryParamSchemas.optionalNonNegativeInt,

  // Date filters
  createdAfter: QueryParamSchemas.optionalIsoDate,
  createdBefore: QueryParamSchemas.optionalIsoDate,

  // Boolean flags
  enabled: QueryParamSchemas.optionalBoolean,
  active: QueryParamSchemas.optionalBoolean,

  // Array parameters
  tags: QueryParamSchemas.optionalString,
  categories: QueryParamSchemas.optionalString,

  // Sort parameters
  sortBy: QueryParamSchemas.optionalString,
  sortOrder: QueryParamSchemas.optionalEnum(["asc", "desc"] as const),
} as const;

/**
 * Application-specific query parameter parser.
 */
export const parseAppSearchParams = createQueryParser(
  z.object({
    view: AppQuerySchemas.view.optional(),
    namespace: AppQuerySchemas.namespace.optional(),
    page: AppQuerySchemas.page,
    limit: AppQuerySchemas.limit,
    offset: AppQuerySchemas.offset,
    search: AppQuerySchemas.search,
    query: AppQuerySchemas.query,
    userId: AppQuerySchemas.userId,
    itemId: AppQuerySchemas.itemId,
    eventType: AppQuerySchemas.eventType,
    createdAfter: AppQuerySchemas.createdAfter,
    createdBefore: AppQuerySchemas.createdBefore,
    enabled: AppQuerySchemas.enabled,
    active: AppQuerySchemas.active,
    tags: AppQuerySchemas.tags,
    categories: AppQuerySchemas.categories,
    sortBy: AppQuerySchemas.sortBy,
    sortOrder: AppQuerySchemas.sortOrder,
  })
);

/**
 * Validate and sanitize a single query parameter.
 */
export function validateQueryParam<T>(
  value: string,
  schema: z.ZodSchema<T>,
  paramName: string
): { success: boolean; data?: T; error?: string } {
  try {
    const result = schema.parse(value);
    return { success: true, data: result };
  } catch (error) {
    if (error instanceof z.ZodError) {
      return {
        success: false,
        error: `${paramName}: ${error.errors[0]?.message || "Invalid value"}`,
      };
    }
    return {
      success: false,
      error: `${paramName}: ${
        error instanceof Error ? error.message : "Unknown error"
      }`,
    };
  }
}

/**
 * Sanitize a query parameter value.
 */
export function sanitizeQueryParam(
  value: string,
  maxLength: number = 1000
): string {
  return value
    .trim()
    .slice(0, maxLength)
    .replace(/[<>]/g, "") // Remove potential HTML tags
    .replace(/javascript:/gi, "") // Remove javascript: protocol
    .replace(/data:/gi, "") // Remove data: protocol
    .replace(/vbscript:/gi, ""); // Remove vbscript: protocol
}

/**
 * Create a safe URL with validated query parameters.
 */
export function createSafeUrl(
  baseUrl: string,
  params: Record<string, string | number | boolean | undefined>,
  schema?: z.ZodSchema<Record<string, any>>
): string {
  const url = new URL(baseUrl);

  // Filter out undefined values and convert to strings
  const cleanParams = Object.entries(params)
    .filter(([, value]) => value !== undefined)
    .reduce((acc, [key, value]) => {
      acc[key] = String(value);
      return acc;
    }, {} as Record<string, string>);

  // Validate if schema provided
  if (schema) {
    const result = schema.safeParse(cleanParams);
    if (!result.success) {
      throw new Error(
        `Invalid query parameters: ${result.error.errors
          .map((e) => e.message)
          .join(", ")}`
      );
    }
  }

  // Add parameters to URL
  Object.entries(cleanParams).forEach(([key, value]) => {
    url.searchParams.set(key, value);
  });

  return url.toString();
}

export default {
  parseSearchParams,
  createQueryParser,
  parseAppSearchParams,
  validateQueryParam,
  sanitizeQueryParam,
  createSafeUrl,
  QueryParamSchemas,
  AppQuerySchemas,
};
