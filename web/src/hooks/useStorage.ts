/**
 * React hook for using the enhanced storage abstraction.
 * Provides a clean interface for persisting UI state with TTL and metadata support.
 */

import { useState, useEffect, useCallback } from "react";
import { getStorage } from "../di";
import type { Storage } from "../di/storage/types";

export interface UseStorageOptions {
  /** Storage key for persistence */
  key: string;
  /** Default value if nothing is stored */
  defaultValue?: any;
  /** Time to live in milliseconds (optional) */
  ttl?: number;
  /** Whether to parse JSON values */
  parseJson?: boolean;
}

export interface UseStorageReturn<T> {
  /** Current stored value */
  value: T;
  /** Set a new value */
  setValue: (value: T) => void;
  /** Remove the stored value */
  removeValue: () => void;
  /** Check if a value exists */
  hasValue: boolean;
  /** Storage instance for advanced operations */
  storage: Storage;
}

/**
 * Hook for using enhanced storage with automatic JSON serialization/deserialization.
 * Supports TTL, metadata, and multiple storage backends.
 */
export function useStorage<T>(options: UseStorageOptions): UseStorageReturn<T> {
  const { key, defaultValue, ttl, parseJson = true } = options;
  const storage = getStorage();

  // Initialize state with stored value or default
  const [value, setValueState] = useState<T>(() => {
    try {
      const stored = storage.getItem(key);
      if (stored === null) {
        return defaultValue;
      }

      if (parseJson) {
        return JSON.parse(stored);
      }

      return stored as T;
    } catch (error) {
      console.warn(`Failed to parse stored value for key "${key}":`, error);
      return defaultValue;
    }
  });

  // Check if value exists in storage
  const [hasValue, setHasValue] = useState(() => storage.hasItem(key));

  // Update storage when value changes
  const setValue = useCallback(
    (newValue: T) => {
      try {
        const serialized = parseJson
          ? JSON.stringify(newValue)
          : String(newValue);
        storage.setItem(key, serialized, ttl);
        setValueState(newValue);
        setHasValue(true);
      } catch (error) {
        console.warn(`Failed to store value for key "${key}":`, error);
      }
    },
    [key, ttl, parseJson, storage]
  );

  // Remove value from storage
  const removeValue = useCallback(() => {
    storage.removeItem(key);
    setValueState(defaultValue);
    setHasValue(false);
  }, [key, defaultValue, storage]);

  // Listen for storage changes (useful for cross-tab synchronization)
  useEffect(() => {
    const handleStorageChange = (e: any) => {
      const storageEvent = e as any;
      if (
        storageEvent.key === key &&
        storageEvent.storageArea === localStorage
      ) {
        try {
          const newValue = storageEvent.newValue;
          if (newValue === null) {
            setValueState(defaultValue);
            setHasValue(false);
          } else {
            const parsed = parseJson ? JSON.parse(newValue) : newValue;
            setValueState(parsed);
            setHasValue(true);
          }
        } catch (error) {
          console.warn(
            `Failed to parse storage change for key "${key}":`,
            error
          );
        }
      }
    };

    window.addEventListener("storage", handleStorageChange);
    return () => window.removeEventListener("storage", handleStorageChange);
  }, [key, defaultValue, parseJson]);

  return {
    value,
    setValue,
    removeValue,
    hasValue,
    storage,
  };
}

/**
 * Hook for simple string storage without JSON parsing.
 */
export function useStringStorage(
  key: string,
  defaultValue: string = ""
): UseStorageReturn<string> {
  return useStorage({
    key,
    defaultValue,
    parseJson: false,
  });
}

/**
 * Hook for JSON storage with TTL support.
 */
export function useJsonStorage<T>(
  key: string,
  defaultValue: T,
  ttl?: number
): UseStorageReturn<T> {
  return useStorage({
    key,
    defaultValue,
    ttl,
    parseJson: true,
  });
}
