// Package services provides composable business services.
//
// This package contains reusable business services that can be composed
// across different API services. Services are designed to be:
// - Database-agnostic (use interfaces for data access)
// - Stateless and thread-safe
// - Testable with mock dependencies
// - Composable with other services
//
// Services should not directly depend on database implementations
// but instead use interfaces defined in the ports package.
//
// Example services:
// - UserService: User management operations
// - ItemService: Item catalog operations
// - RecommendationService: Recommendation algorithms
// - AuditService: Audit logging operations
package services
