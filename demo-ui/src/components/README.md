# Error Handling & UX Components

This module provides comprehensive error handling, user experience, and telemetry components for the demo UI application.

## Components

### AppErrorBoundary

Application-level error boundary with fallback UI and logging integration.

**Features:**
- Catches JavaScript errors anywhere in the component tree
- Displays user-friendly error screen with recovery options
- Logs errors with structured data for debugging
- Provides error reporting functionality
- Shows development error details in dev mode

**Usage:**
```tsx
import { AppErrorBoundary } from "./components/AppErrorBoundary";

function App() {
  return (
    <AppErrorBoundary>
      <YourApp />
    </AppErrorBoundary>
  );
}
```

### GuardedRoute

Route protection component for access control and authentication.

**Features:**
- Conditional rendering based on access permissions
- Custom fallback UI for denied access
- Automatic redirects for unauthorized users
- Logging of access attempts
- Ready for future authentication integration

**Usage:**
```tsx
import { GuardedRoute } from "./components/GuardedRoute";

function ProtectedPage() {
  return (
    <GuardedRoute
      canActivate={() => user.isAuthenticated}
      fallback={<AccessDeniedMessage />}
      onAccessDenied={() => redirectToLogin()}
    >
      <ProtectedContent />
    </GuardedRoute>
  );
}
```

### Toast Notifications

Non-blocking toast notification system for user feedback.

**Features:**
- Multiple toast types: success, error, warning, info
- Auto-dismiss with configurable duration
- Persistent toasts for important messages
- Action buttons for user interaction
- Automatic cleanup and size limits
- Smooth animations and transitions

**Usage:**
```tsx
import { useToast } from "../contexts/ToastContext";

function MyComponent() {
  const toast = useToast();

  const handleSuccess = () => {
    toast.showSuccess("Success!", "Operation completed");
  };

  const handleError = () => {
    toast.showError("Error!", "Something went wrong", {
      action: {
        label: "Retry",
        onClick: () => retryOperation(),
      },
    });
  };
}
```

## Contexts

### ToastContext

React context for managing toast notifications throughout the application.

**Features:**
- Centralized toast state management
- Automatic cleanup and size limits
- Persistent storage of toast preferences
- Integration with logging system

### Telemetry Context

Telemetry and analytics tracking system.

**Features:**
- Event tracking with structured data
- User identification and session tracking
- Page view tracking
- Feature flag integration
- No-op implementation by default
- Ready for analytics service integration

## Hooks

### useTelemetry

Hook for tracking events and analytics data.

**Features:**
- Track custom events with properties
- User identification and session management
- Page view tracking
- Feature flag integration
- Automatic event categorization

**Usage:**
```tsx
import { useTelemetry, TELEMETRY_EVENTS } from "../hooks/useTelemetry";

function MyComponent() {
  const telemetry = useTelemetry();

  const handleButtonClick = () => {
    telemetry.track(TELEMETRY_EVENTS.BUTTON_CLICK, {
      button: "submit",
      page: "checkout",
    });
  };

  const handleUserAction = () => {
    telemetry.identify("user_123", {
      name: "John Doe",
      plan: "premium",
    });
  };
}
```

### useToast

Convenience hook for toast notifications.

**Features:**
- Type-safe toast creation
- Automatic error handling
- Integration with telemetry
- Consistent UX patterns

## Integration

### HTTP Client Integration

Enhanced HTTP client with toast notifications for network failures.

**Features:**
- Automatic error toast display
- User-friendly error messages
- Retry functionality
- Error categorization
- Telemetry integration

**Usage:**
```tsx
import { useToastHttpInterceptors } from "../di/http/toast-integration";

function MyComponent() {
  const { errorInterceptor, successInterceptor } = useToastHttpInterceptors();
  
  // Add interceptors to HTTP client
  httpClient.addResponseInterceptor({
    onError: (error) => errorInterceptor.processError(error),
    onResponse: (response) => successInterceptor.processSuccess(response, request),
  });
}
```

## Error Handling Patterns

### 1. Error Boundaries

Use `AppErrorBoundary` to catch and handle JavaScript errors:

```tsx
<AppErrorBoundary
  onError={(error, errorInfo) => {
    // Custom error handling
    reportError(error, errorInfo);
  }}
>
  <App />
</AppErrorBoundary>
```

### 2. Route Protection

Use `GuardedRoute` for access control:

```tsx
<GuardedRoute
  canActivate={() => checkUserPermissions()}
  requireAuth={true}
  redirectTo="/login"
>
  <AdminPanel />
</GuardedRoute>
```

### 3. Toast Notifications

Use toasts for user feedback:

```tsx
const toast = useToast();

// Success feedback
toast.showSuccess("Data saved", "Your changes have been saved");

// Error feedback with retry
toast.showError("Save failed", "Unable to save your changes", {
  action: {
    label: "Retry",
    onClick: () => retrySave(),
  },
});
```

### 4. Telemetry Tracking

Track user interactions and system events:

```tsx
const telemetry = useTelemetry();

// Track user actions
telemetry.track("feature_used", {
  feature: "recommendations",
  algorithm: "collaborative_filtering",
});

// Track errors
telemetry.track(TELEMETRY_EVENTS.ERROR_OCCURRED, {
  error: error.message,
  component: "RecommendationEngine",
  severity: "high",
});
```

## Configuration

### Feature Flags

Control error handling and UX features with feature flags:

```tsx
const { isEnabled } = useFeatureFlags();

// Enable/disable analytics
const analyticsEnabled = isEnabled("analyticsEnabled");

// Enable/disable debug mode
const debugMode = isEnabled("debugMode");
```

### Environment Variables

Configure behavior with environment variables:

```bash
# Enable debug mode
NODE_ENV=development

# Enable analytics
VITE_ENABLE_ANALYTICS=true

# Configure toast duration
VITE_TOAST_DURATION=5000
```

## Best Practices

### 1. Error Handling

- Always wrap components in error boundaries
- Provide meaningful error messages to users
- Log errors with structured data
- Implement retry mechanisms for transient failures

### 2. User Experience

- Use toasts for non-blocking feedback
- Provide clear success and error messages
- Implement loading states for async operations
- Use route guards for access control

### 3. Telemetry

- Track important user interactions
- Monitor error rates and performance
- Use consistent event naming
- Respect user privacy preferences

### 4. Accessibility

- Ensure error messages are accessible
- Provide keyboard navigation for toasts
- Use proper ARIA labels and roles
- Test with screen readers

## Future Enhancements

- **Error Recovery**: Automatic retry mechanisms
- **User Preferences**: Customizable toast settings
- **Analytics Dashboard**: Real-time error monitoring
- **A/B Testing**: Feature flag experimentation
- **Performance Monitoring**: Automatic performance tracking
- **User Feedback**: In-app feedback collection
