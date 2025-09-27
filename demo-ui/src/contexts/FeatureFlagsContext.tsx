/**
 * Feature flags context for conditional UI rendering and behavior.
 */

import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react";
import { getStorage } from "../di";

export interface FeatureFlags {
  authEnabled: boolean;
  useRemoteEmbeddings: boolean;
  circuitBreakerEnabled: boolean;
  retryEnabled: boolean;
  analyticsEnabled: boolean;
  debugMode: boolean;
}

export interface FeatureFlagsContextType {
  flags: FeatureFlags;
  updateFlag: <K extends keyof FeatureFlags>(
    key: K,
    value: FeatureFlags[K]
  ) => void;
  updateFlags: (flags: Partial<FeatureFlags>) => void;
  resetFlags: () => void;
  isEnabled: (flag: keyof FeatureFlags) => boolean;
}

const defaultFlags: FeatureFlags = {
  authEnabled: false,
  useRemoteEmbeddings: false,
  circuitBreakerEnabled: true,
  retryEnabled: true,
  analyticsEnabled: false,
  debugMode: false,
};

const FeatureFlagsContext = createContext<FeatureFlagsContextType | undefined>(
  undefined
);

export interface FeatureFlagsProviderProps {
  children: ReactNode;
  initialFlags?: Partial<FeatureFlags>;
  persistFlags?: boolean;
}

export function FeatureFlagsProvider({
  children,
  initialFlags = {},
  persistFlags = true,
}: FeatureFlagsProviderProps) {
  const [flags, setFlags] = useState<FeatureFlags>({
    ...defaultFlags,
    ...initialFlags,
  });

  const storage = getStorage();
  const storageKey = "feature_flags";

  // Load flags from storage on mount
  useEffect(() => {
    if (persistFlags) {
      try {
        const storedFlags = storage.getItem(storageKey);
        if (storedFlags) {
          const parsedFlags = JSON.parse(storedFlags);
          setFlags((prev) => ({ ...prev, ...parsedFlags }));
        }
      } catch (error) {
        console.warn("Failed to load feature flags from storage:", error);
      }
    }
  }, [persistFlags, storage]);

  // Save flags to storage when they change
  useEffect(() => {
    if (persistFlags) {
      try {
        storage.setItem(storageKey, JSON.stringify(flags));
      } catch (error) {
        console.warn("Failed to save feature flags to storage:", error);
      }
    }
  }, [flags, persistFlags, storage]);

  const updateFlag = <K extends keyof FeatureFlags>(
    key: K,
    value: FeatureFlags[K]
  ) => {
    setFlags((prev) => ({ ...prev, [key]: value }));
  };

  const updateFlags = (newFlags: Partial<FeatureFlags>) => {
    setFlags((prev) => ({ ...prev, ...newFlags }));
  };

  const resetFlags = () => {
    setFlags(defaultFlags);
  };

  const isEnabled = (flag: keyof FeatureFlags): boolean => {
    return flags[flag];
  };

  const contextValue: FeatureFlagsContextType = {
    flags,
    updateFlag,
    updateFlags,
    resetFlags,
    isEnabled,
  };

  return (
    <FeatureFlagsContext.Provider value={contextValue}>
      {children}
    </FeatureFlagsContext.Provider>
  );
}

export function useFeatureFlags(): FeatureFlagsContextType {
  const context = useContext(FeatureFlagsContext);
  if (context === undefined) {
    throw new Error(
      "useFeatureFlags must be used within a FeatureFlagsProvider"
    );
  }
  return context;
}

// Convenience hooks for specific flags
export function useAuthEnabled(): boolean {
  const { isEnabled } = useFeatureFlags();
  return isEnabled("authEnabled");
}

export function useRemoteEmbeddings(): boolean {
  const { isEnabled } = useFeatureFlags();
  return isEnabled("useRemoteEmbeddings");
}

export function useCircuitBreakerEnabled(): boolean {
  const { isEnabled } = useFeatureFlags();
  return isEnabled("circuitBreakerEnabled");
}

export function useRetryEnabled(): boolean {
  const { isEnabled } = useFeatureFlags();
  return isEnabled("retryEnabled");
}

export function useAnalyticsEnabled(): boolean {
  const { isEnabled } = useFeatureFlags();
  return isEnabled("analyticsEnabled");
}

export function useDebugMode(): boolean {
  const { isEnabled } = useFeatureFlags();
  return isEnabled("debugMode");
}
