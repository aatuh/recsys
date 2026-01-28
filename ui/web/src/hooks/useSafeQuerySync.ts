/**
 * Enhanced useQuerySync hook with URL and query parameter validation.
 * Replaces the basic useQuerySync with comprehensive validation and sanitization.
 */

import { useEffect, useState, useCallback } from "react";
import {
  parseSearchParams,
  type ParseSearchParamsResult,
} from "../utils/urlValidation";
import { z } from "zod";

export interface UseSafeQuerySyncConfig<T> {
  /** Zod schema for validation */
  schema: z.ZodSchema<T>;
  /** Initial value if validation fails */
  initial: T;
  /** Whether to validate on every change */
  validateOnChange?: boolean;
  /** Whether to strip invalid parameters */
  stripInvalid?: boolean;
  /** Custom error handler */
  onError?: (errors: Record<string, string[]>) => void;
  /** Storage key for persistence */
  storageKey?: string;
  /** Whether to persist to localStorage */
  persist?: boolean;
}

/**
 * Enhanced useQuerySync hook with comprehensive validation.
 *
 * Features:
 * - Zod schema validation
 * - Parameter sanitization
 * - Error handling
 * - Persistence support
 * - Type safety
 */
export function useSafeQuerySync<T>(
  key: string,
  config: UseSafeQuerySyncConfig<T>
): [T, (value: T) => void, ParseSearchParamsResult<T>] {
  const {
    schema,
    initial,
    validateOnChange = true,
    stripInvalid = true,
    onError,
    storageKey,
    persist = false,
  } = config;

  // Parse current URL parameters
  const parseCurrentParams = useCallback((): ParseSearchParamsResult<T> => {
    const url = new URL(window.location.href);
    const searchParams = url.searchParams;

    // Extract the specific parameter we're interested in
    const paramValue = searchParams.get(key);
    if (!paramValue) {
      return {
        data: initial,
        errors: {},
        success: true,
        raw: {},
        sanitized: {},
      };
    }

    // Create a schema for just this parameter
    const paramSchema = z.object({ [key]: schema });
    const result = parseSearchParams({ [key]: paramValue }, paramSchema, {
      stripUnknown: stripInvalid,
      strict: validateOnChange,
    });

    return result as ParseSearchParamsResult<T>;
  }, [key, schema, initial, stripInvalid, validateOnChange]);

  // Initialize state
  const [value, setValue] = useState<T>(() => {
    // Try to get from URL first
    const urlResult = parseCurrentParams();
    if (urlResult.success) {
      return (urlResult.data as any)[key] || initial;
    }

    // Fall back to localStorage if available
    if (persist && storageKey) {
      try {
        const saved = localStorage.getItem(storageKey);
        if (saved) {
          const parsed = JSON.parse(saved);
          const validation = schema.safeParse(parsed);
          if (validation.success) {
            return validation.data;
          }
        }
      } catch {
        // Ignore localStorage errors
      }
    }

    return initial;
  });

  // Parse result state
  const [parseResult, setParseResult] = useState<ParseSearchParamsResult<T>>(
    () => parseCurrentParams()
  );

  // Update URL and localStorage when value changes
  const updateValue = useCallback(
    (newValue: T) => {
      setValue(newValue);

      // Update URL
      const url = new URL(window.location.href);
      url.searchParams.set(key, String(newValue));
      window.history.replaceState({}, "", url.toString());

      // Update localStorage if enabled
      if (persist && storageKey) {
        try {
          localStorage.setItem(storageKey, JSON.stringify(newValue));
        } catch {
          // Ignore localStorage errors
        }
      }
    },
    [key, storageKey, persist]
  );

  // Listen for URL changes
  useEffect(() => {
    const handleUrlChange = () => {
      const result = parseCurrentParams();
      setParseResult(result);

      if (result.success) {
        const newValue = (result.data as any)[key] || initial;
        if (newValue !== value) {
          setValue(newValue);
        }
      } else if (onError) {
        onError(result.errors);
      }
    };

    // Listen for popstate events (back/forward navigation)
    window.addEventListener("popstate", handleUrlChange);

    // Also check on mount
    handleUrlChange();

    return () => {
      window.removeEventListener("popstate", handleUrlChange);
    };
  }, [parseCurrentParams, onError, key, initial, value]);

  return [value, updateValue, parseResult];
}

/**
 * Simplified hook for single parameter validation.
 */
export function useSafeQueryParam<T>(
  key: string,
  schema: z.ZodSchema<T>,
  initial: T,
  options: {
    onError?: (error: string) => void;
    storageKey?: string;
    persist?: boolean;
  } = {}
): [T, (value: T) => void, { success: boolean; error?: string }] {
  const [value, setValue, parseResult] = useSafeQuerySync(key, {
    schema,
    initial,
    onError: options.onError
      ? (errors) => {
          const errorMessages = Object.values(errors).flat();
          options.onError?.(errorMessages[0] || "Validation failed");
        }
      : undefined,
    storageKey: options.storageKey,
    persist: options.persist,
  });

  return [
    value,
    setValue,
    {
      success: parseResult.success,
      error: parseResult.success
        ? undefined
        : Object.values(parseResult.errors).flat()[0],
    },
  ];
}

/**
 * Hook for validating multiple query parameters at once.
 */
export function useSafeQueryParams<T>(
  schema: z.ZodSchema<T>,
  initial: T,
  options: {
    onError?: (errors: Record<string, string[]>) => void;
    storageKey?: string;
    persist?: boolean;
  } = {}
): [T, (value: T) => void, ParseSearchParamsResult<T>] {
  const [value, setValue, parseResult] = useSafeQuerySync("", {
    schema,
    initial,
    onError: options.onError,
    storageKey: options.storageKey,
    persist: options.persist,
  });

  const updateValue = useCallback(
    (newValue: T) => {
      setValue(newValue);

      // Update URL with all parameters
      const url = new URL(window.location.href);

      // Clear existing parameters
      url.search = "";

      // Add new parameters
      Object.entries(newValue as Record<string, any>).forEach(([key, val]) => {
        if (val !== undefined && val !== null) {
          url.searchParams.set(key, String(val));
        }
      });

      window.history.replaceState({}, "", url.toString());

      // Update localStorage if enabled
      if (options.persist && options.storageKey) {
        try {
          localStorage.setItem(options.storageKey, JSON.stringify(newValue));
        } catch {
          // Ignore localStorage errors
        }
      }
    },
    [options.persist, options.storageKey]
  );

  return [value, updateValue, parseResult];
}

export default {
  useSafeQuerySync,
  useSafeQueryParam,
  useSafeQueryParams,
};
