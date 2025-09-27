/**
 * Global type declarations for browser APIs.
 */

declare global {
  interface Window {
    crypto: any;
  }

  const crypto: any;
  const AbortController: any;
  const AbortSignal: any;

  // Node.js globals for compatibility
  const process: {
    env: Record<string, string>;
    browser: boolean;
  };

  const require: (id: string) => any;

  // ESLint globals
  const queryClient: any;
}

export {};
