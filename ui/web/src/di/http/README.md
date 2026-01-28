# HTTP Module

Advanced HTTP client with interceptors, retry logic, circuit breaker, and request tracing.

## Features

- **Request/Response Interceptors**: Pluggable middleware for authentication, logging, and error handling
- **Retry Logic**: Exponential backoff with jitter for resilient requests
- **Circuit Breaker**: Automatic failure detection and service protection
- **Request Tracing**: Automatic UUID generation for request correlation
- **Error Classification**: Unified ApiError type with retry and auth behavior classification
- **Cancellation Support**: AbortController integration for request cancellation

## Architecture

### Core Components

- **EnhancedHttpClient**: Main HTTP client with all advanced features
- **ApiError**: Unified error type with classification for retry/auth behaviors
- **CircuitBreaker**: Failure detection and service protection
- **RetryManager**: Exponential backoff with jitter
- **Interceptors**: Pluggable middleware for requests and responses

### Error Classification

The `ApiError` class provides automatic classification of errors:

```typescript
const error = ApiError.fromHttpError(originalError, status);
console.log(error.retryable);    // true for 5xx, network errors, timeouts
console.log(error.authError);    // true for 401/403
console.log(error.serverError);  // true for 5xx
console.log(error.clientError);  // true for 4xx
console.log(error.networkError); // true for network failures
console.log(error.timeout);      // true for timeouts
```

### Interceptors

#### Request Interceptors
- **AuthInterceptor**: Attach authentication headers (stub for future implementation)
- **RequestIdInterceptor**: Add unique request IDs for tracing

#### Response Interceptors
- **ErrorInterceptor**: Process and classify errors
- **ResponseLoggingInterceptor**: Log successful responses

### Circuit Breaker

The circuit breaker protects against cascading failures:

- **CLOSED**: Normal operation, requests pass through
- **OPEN**: Circuit is open, requests are blocked
- **HALF_OPEN**: Testing if service has recovered

Configuration:
```typescript
circuitBreaker: {
  enabled: true,
  failureThreshold: 5,    // Failures before opening
  timeout: 5000,          // Time to wait before half-open
  resetTimeout: 30000,    // Time before attempting reset
}
```

### Retry Logic

Exponential backoff with jitter for resilient requests:

```typescript
retry: {
  retries: 3,           // Maximum retry attempts
  retryDelay: 1000,     // Base delay in milliseconds
  retryBackoff: true,   // Enable exponential backoff
  jitter: true,         // Add random jitter to prevent thundering herd
}
```

## Usage

### Basic Usage

```typescript
import { getHttpClient } from '../di';

const httpClient = getHttpClient();

// Simple GET request
const response = await httpClient.get('/api/users');

// POST with data
const result = await httpClient.post('/api/users', { name: 'John' });

// Request with options
const data = await httpClient.get('/api/data', {
  headers: { 'Authorization': 'Bearer token' },
  timeout: 5000,
});
```

### Error Handling

```typescript
try {
  const response = await httpClient.get('/api/data');
} catch (error) {
  if (error instanceof ApiError) {
    if (error.authError) {
      // Handle authentication error
      redirectToLogin();
    } else if (error.retryable) {
      // Request can be retried
      console.log('Retrying request...');
    }
  }
}
```

### Custom Interceptors

```typescript
// Add custom request interceptor
httpClient.addRequestInterceptor({
  onRequest: (request) => {
    console.log('Making request:', request.method, request.url);
    return {
      ...request,
      headers: {
        ...request.headers,
        'X-Custom-Header': 'value',
      },
    };
  },
});

// Add custom response interceptor
httpClient.addResponseInterceptor({
  onResponse: (response) => {
    console.log('Response received:', response.status);
    return response;
  },
  onError: (error) => {
    console.error('Request failed:', error.message);
    return error;
  },
});
```

### Request Cancellation

```typescript
const token = httpClient.createCancellationToken();

// Cancel after 5 seconds
setTimeout(() => token.cancel(), 5000);

try {
  const response = await httpClient.get('/api/slow', {
    signal: token.cancelled ? undefined : new AbortController().signal,
  });
} catch (error) {
  if (error instanceof ApiError && error.timeout) {
    console.log('Request was cancelled');
  }
}
```

## Configuration

The HTTP client is configured through the DI container with sensible defaults:

```typescript
const httpClient = new EnhancedHttpClient({
  baseUrl: '/api',
  timeout: 30000,
  retries: 3,
  retryDelay: 1000,
  retryBackoff: true,
  jitter: true,
  circuitBreaker: {
    enabled: true,
    failureThreshold: 5,
    timeout: 5000,
    resetTimeout: 30000,
  },
});
```

## Testing

The HTTP client can be easily mocked for testing:

```typescript
import { setContainer } from '../di';

const mockHttpClient = {
  get: jest.fn(),
  post: jest.fn(),
  // ... other methods
};

setContainer({
  getHttpClient: () => mockHttpClient,
  // ... other services
});
```

## Future Enhancements

- **Authentication**: JWT token refresh, OAuth2 flows
- **Caching**: Request/response caching with TTL
- **Metrics**: Request timing, success rates, error rates
- **Rate Limiting**: Client-side rate limiting
- **Offline Support**: Request queuing for offline scenarios
