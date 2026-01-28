// Package internal contains private application code following clean architecture.
//
// This package implements the core business logic with clear separation of concerns:
// - adapters: External interface implementations (HTTP, database, etc.)
// - config: Configuration management
// - domain: Core business logic and models
// - ports: Interface definitions for dependency inversion
// - service: Application service orchestration
// - store: Data access layer
// - types: Type definitions
//
// The structure supports both simple and complex APIs with proper
// configuration management, database migrations, and operational tooling.
package internal
