package adminsvc

import (
	"encoding/json"
	"net"
	"time"

	"github.com/google/uuid"
)

// TenantConfig represents the stored config document for a tenant.
type TenantConfig struct {
	TenantID string
	Version  string
	Raw      json.RawMessage
}

// TenantRules represents the stored rules document for a tenant.
type TenantRules struct {
	TenantID string
	Version  string
	Raw      json.RawMessage
}

// CacheInvalidateRequest describes cache invalidation inputs.
type CacheInvalidateRequest struct {
	Targets []string
	Surface string
}

// CacheInvalidateResult captures the invalidation outcome.
type CacheInvalidateResult struct {
	TenantID    string
	Targets     []string
	Surface     string
	Status      string
	Invalidated map[string]int
}

// Actor identifies the admin actor performing a change.
type Actor struct {
	ID   string
	Type string
}

// RequestMeta captures request context for audit/invalidation logging.
type RequestMeta struct {
	RequestID string
	IP        net.IP
	UserAgent string
}

// CacheInvalidationEvent persists a cache invalidation request.
type CacheInvalidationEvent struct {
	TenantID    uuid.UUID
	RequestID   *uuid.UUID
	ActorID     string
	Targets     []string
	Surface     string
	Status      string
	ErrorDetail string
}

// AuditQuery filters audit log listings.
type AuditQuery struct {
	Limit    int
	Before   time.Time
	BeforeID int64
}

// AuditEntry describes a stored audit event.
type AuditEntry struct {
	ID         int64
	OccurredAt time.Time
	TenantID   string
	ActorSub   string
	ActorType  string
	Action     string
	EntityType string
	EntityID   string
	RequestID  string
	IP         net.IP
	UserAgent  string
	Before     json.RawMessage
	After      json.RawMessage
	Extra      json.RawMessage
}

// AuditLog is a paginated audit log response.
type AuditLog struct {
	TenantID     string
	Entries      []AuditEntry
	NextBefore   time.Time
	NextBeforeID int64
}

// AuditEvent captures a change to be recorded.
type AuditEvent struct {
	TenantID   uuid.UUID
	Actor      Actor
	Meta       RequestMeta
	Action     string
	EntityType string
	EntityID   string
	Before     json.RawMessage
	After      json.RawMessage
	Extra      json.RawMessage
}
