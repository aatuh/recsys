# Storage Module

Enhanced storage abstraction with multiple backends, TTL support, and metadata tracking.

## Features

- **Multiple Backends**: In-memory, localStorage, sessionStorage, and hybrid storage
- **TTL Support**: Time-to-live for automatic expiration of stored items
- **Metadata Tracking**: Store and retrieve metadata about stored items
- **Size Management**: Automatic cleanup and size limits
- **Fallback Support**: Graceful degradation when storage is unavailable

## Architecture

### Storage Backends

- **InMemoryStorageBackend**: Map-based storage, lost on page refresh
- **LocalStorageBackend**: Persistent storage across browser sessions
- **SessionStorageBackend**: Session-only storage, lost when tab closes
- **HybridStorageBackend**: Tries multiple backends in order with fallback

### Enhanced Storage

The `EnhancedStorage` class provides additional features on top of basic storage:

- **TTL Support**: Items can have time-to-live for automatic expiration
- **Metadata**: Store creation time, TTL, and other metadata
- **Size Limits**: Automatic cleanup when storage exceeds limits
- **Key Prefixing**: Namespace storage keys to avoid conflicts

## Usage

### Basic Usage

```typescript
import { createStorage } from '../di/storage';

// Create storage with hybrid backend (recommended)
const storage = createStorage("hybrid", {
  prefix: "myapp_",
  defaultTTL: 24 * 60 * 60 * 1000, // 24 hours
  maxSize: 1000,
});

// Basic operations
storage.setItem("key", "value");
const value = storage.getItem("key");
storage.removeItem("key");
storage.clear();
```

### TTL Support

```typescript
// Set item with 5 second TTL
storage.setItem("temp", "data", 5000);

// Item will be automatically removed after 5 seconds
setTimeout(() => {
  const value = storage.getItem("temp"); // null
}, 6000);
```

### Metadata Operations

```typescript
// Set item with metadata
storage.setItemWithMetadata({
  key: "user",
  value: "john",
  timestamp: Date.now(),
  ttl: 3600000, // 1 hour
});

// Get item with metadata
const item = storage.getItemWithMetadata("user");
console.log(item.timestamp); // Creation time
console.log(item.ttl); // Time to live
```

### Storage Management

```typescript
// Check if item exists
const exists = storage.hasItem("key");

// Get all keys
const keys = storage.keys();

// Get storage size
const size = storage.size();

// List all items with metadata
keys.forEach(key => {
  const item = storage.getItemWithMetadata(key);
  console.log(`${key}: ${item.value} (created: ${new Date(item.timestamp)})`);
});
```

## Configuration

### Storage Types

```typescript
// In-memory only (lost on refresh)
const memoryStorage = createStorage("memory");

// LocalStorage only (persistent)
const localStorage = createStorage("local");

// SessionStorage only (session-only)
const sessionStorage = createStorage("session");

// Hybrid (recommended - tries local, then session, then memory)
const hybridStorage = createStorage("hybrid");
```

### Configuration Options

```typescript
const storage = createStorage("hybrid", {
  prefix: "myapp_",           // Key prefix to avoid conflicts
  defaultTTL: 86400000,       // Default TTL in milliseconds
  maxSize: 1000,              // Maximum number of items
});
```

## Backend Selection

The hybrid backend tries backends in this order:

1. **LocalStorage**: Persistent across sessions
2. **SessionStorage**: Session-only storage
3. **InMemoryStorage**: Fallback when others fail

This provides the best user experience with graceful degradation.

## Error Handling

All storage operations are designed to fail gracefully:

```typescript
// These operations will never throw
storage.setItem("key", "value"); // Silently fails if storage unavailable
const value = storage.getItem("key"); // Returns null if unavailable
storage.removeItem("key"); // Silently fails if storage unavailable
```

## Performance Considerations

- **Automatic Cleanup**: Expired items are removed during operations
- **Size Limits**: Oldest items are removed when limits are exceeded
- **Lazy Loading**: Metadata is only loaded when needed
- **Efficient Backends**: Uses native browser APIs when available

## Testing

Storage can be easily mocked for testing:

```typescript
import { setContainer } from '../di';

const mockStorage = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
  keys: jest.fn(),
  size: jest.fn(),
  hasItem: jest.fn(),
  getItemWithMetadata: jest.fn(),
  setItemWithMetadata: jest.fn(),
};

setContainer({
  getStorage: () => mockStorage,
  // ... other services
});
```

## Future Enhancements

- **Encryption**: Encrypt sensitive data before storage
- **Compression**: Compress large values to save space
- **Sync**: Synchronize storage across browser tabs
- **Analytics**: Track storage usage and performance
- **Quotas**: Respect browser storage quotas
