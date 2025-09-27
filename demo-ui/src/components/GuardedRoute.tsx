/**
 * GuardedRoute component for protecting routes with access control.
 * Currently defaults to allowing all access, ready for future auth integration.
 */

import React, { ReactNode } from "react";
import { getLogger } from "../di";

export interface GuardedRouteProps {
  children: ReactNode;
  canActivate?: () => boolean;
  fallback?: ReactNode;
  onAccessDenied?: () => void;
  requireAuth?: boolean;
  redirectTo?: string;
}

export function GuardedRoute({
  children,
  canActivate = () => true, // Default to allowing access
  fallback,
  onAccessDenied,
  requireAuth = false,
  redirectTo,
}: GuardedRouteProps) {
  const logger = getLogger().child({ component: "GuardedRoute" });

  // Check if access is allowed
  const hasAccess = canActivate();

  // Log access attempts
  React.useEffect(() => {
    if (hasAccess) {
      logger.debug("Route access granted");
    } else {
      logger.warn("Route access denied", {
        requireAuth,
        redirectTo,
      });

      // Call access denied handler
      if (onAccessDenied) {
        onAccessDenied();
      }
    }
  }, [hasAccess, requireAuth, redirectTo, onAccessDenied, logger]);

  // Handle redirect
  React.useEffect(() => {
    if (!hasAccess && redirectTo) {
      logger.info("Redirecting to", { redirectTo });
      window.location.href = redirectTo;
    }
  }, [hasAccess, redirectTo, logger]);

  // If access is denied, show fallback or nothing
  if (!hasAccess) {
    if (fallback) {
      return <>{fallback}</>;
    }

    // Default fallback for access denied
    return (
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          minHeight: "400px",
          padding: "20px",
          textAlign: "center",
        }}
      >
        <div
          style={{
            fontSize: "48px",
            marginBottom: "20px",
            color: "#6c757d",
          }}
        >
          🔒
        </div>

        <h2
          style={{
            fontSize: "20px",
            fontWeight: "bold",
            marginBottom: "12px",
            color: "#212529",
          }}
        >
          Access Denied
        </h2>

        <p
          style={{
            fontSize: "14px",
            color: "#6c757d",
            marginBottom: "20px",
            maxWidth: "400px",
          }}
        >
          You don't have permission to access this page.
        </p>

        <button
          onClick={() => window.history.back()}
          style={{
            padding: "10px 20px",
            backgroundColor: "#007bff",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: "pointer",
            fontSize: "14px",
          }}
        >
          Go Back
        </button>
      </div>
    );
  }

  // Access granted, render children
  return <>{children}</>;
}

/**
 * Higher-order component for creating guarded routes.
 * Useful for wrapping components with route protection.
 */
export function withGuardedRoute<P extends object>(
  Component: React.ComponentType<P>,
  guardOptions: Omit<GuardedRouteProps, "children">
) {
  return function GuardedComponent(props: P) {
    return (
      <GuardedRoute {...guardOptions}>
        <Component {...props} />
      </GuardedRoute>
    );
  };
}

/**
 * Hook for checking route access permissions.
 * Useful for conditional rendering based on access rights.
 */
export function useRouteAccess(canActivate: () => boolean) {
  const [hasAccess, setHasAccess] = React.useState(false);
  const logger = getLogger().child({ hook: "useRouteAccess" });

  React.useEffect(() => {
    try {
      const access = canActivate();
      setHasAccess(access);

      if (access) {
        logger.debug("Route access check passed");
      } else {
        logger.warn("Route access check failed");
      }
    } catch (error) {
      logger.error("Route access check failed with error", { error });
      setHasAccess(false);
    }
  }, [canActivate, logger]);

  return hasAccess;
}
