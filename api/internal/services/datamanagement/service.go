package datamanagement

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	specstypes "recsys/specs/types"
)

// ValidationError represents client-facing validation failures.
type ValidationError struct {
	Code    string
	Message string
	Details map[string]any
}

func (e ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

// Store defines the persistence contract required for data management.
type Store interface {
	ListUsers(ctx context.Context, orgID uuid.UUID, namespace string, limit, offset int, filters map[string]any) ([]map[string]any, int, error)
	ListItems(ctx context.Context, orgID uuid.UUID, namespace string, limit, offset int, filters map[string]any) ([]map[string]any, int, error)
	ListEvents(ctx context.Context, orgID uuid.UUID, namespace string, limit, offset int, filters map[string]any) ([]map[string]any, int, error)

	DeleteUsers(ctx context.Context, orgID uuid.UUID, namespace string, filters map[string]any) (int, error)
	DeleteItems(ctx context.Context, orgID uuid.UUID, namespace string, filters map[string]any) (int, error)
	DeleteEvents(ctx context.Context, orgID uuid.UUID, namespace string, filters map[string]any) (int, error)
}

// Service orchestrates list and delete operations for data management.
type Service struct {
	store    Store
	maxLimit int
}

// Option configures the Service.
type Option func(*Service)

// WithMaxLimit overrides the maximum page size enforced by the service.
func WithMaxLimit(limit int) Option {
	return func(s *Service) {
		if limit > 0 {
			s.maxLimit = limit
		}
	}
}

// New constructs a Service with the provided store.
func New(st Store, opts ...Option) *Service {
	svc := &Service{store: st, maxLimit: 1000}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// ListOptions provides shared pagination inputs.
type ListOptions struct {
	Namespace string
	Limit     int
	Offset    int
	Filters   map[string]any
}

func (s *Service) validateList(opts ListOptions) error {
	if opts.Namespace == "" {
		return ValidationError{Code: "missing_namespace", Message: "namespace parameter is required"}
	}
	if opts.Limit <= 0 {
		return ValidationError{Code: "invalid_limit", Message: "limit must be positive"}
	}
	if opts.Limit > s.maxLimit {
		return ValidationError{Code: "limit_too_large", Message: fmt.Sprintf("limit must be <= %d", s.maxLimit)}
	}
	if opts.Offset < 0 {
		return ValidationError{Code: "invalid_offset", Message: "offset must be >= 0"}
	}
	return nil
}

func buildListResponse(items []map[string]any, total, limit, offset int) specstypes.ListResponse {
	hasMore := offset+limit < total
	var nextOffset *int
	if hasMore {
		next := offset + limit
		nextOffset = &next
	}

	payload := make([]any, 0, len(items))
	for _, item := range items {
		payload = append(payload, item)
	}

	return specstypes.ListResponse{
		Items:      payload,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
		NextOffset: nextOffset,
	}
}

// ListUsers fetches paginated users with validation.
func (s *Service) ListUsers(ctx context.Context, orgID uuid.UUID, opts ListOptions) (specstypes.ListResponse, error) {
	if err := s.validateList(opts); err != nil {
		return specstypes.ListResponse{}, err
	}
	rows, total, err := s.store.ListUsers(ctx, orgID, opts.Namespace, opts.Limit, opts.Offset, opts.Filters)
	if err != nil {
		return specstypes.ListResponse{}, err
	}
	return buildListResponse(rows, total, opts.Limit, opts.Offset), nil
}

// ListItems fetches paginated items with validation.
func (s *Service) ListItems(ctx context.Context, orgID uuid.UUID, opts ListOptions) (specstypes.ListResponse, error) {
	if err := s.validateList(opts); err != nil {
		return specstypes.ListResponse{}, err
	}
	rows, total, err := s.store.ListItems(ctx, orgID, opts.Namespace, opts.Limit, opts.Offset, opts.Filters)
	if err != nil {
		return specstypes.ListResponse{}, err
	}
	return buildListResponse(rows, total, opts.Limit, opts.Offset), nil
}

// ListEvents fetches paginated events with validation.
func (s *Service) ListEvents(ctx context.Context, orgID uuid.UUID, opts ListOptions) (specstypes.ListResponse, error) {
	if err := s.validateList(opts); err != nil {
		return specstypes.ListResponse{}, err
	}
	rows, total, err := s.store.ListEvents(ctx, orgID, opts.Namespace, opts.Limit, opts.Offset, opts.Filters)
	if err != nil {
		return specstypes.ListResponse{}, err
	}
	return buildListResponse(rows, total, opts.Limit, opts.Offset), nil
}

// DeleteOptions captures namespace-filter inputs for delete operations.
type DeleteOptions struct {
	Namespace string
	Filters   map[string]any
}

func (s *Service) validateDelete(opts DeleteOptions) error {
	if opts.Namespace == "" {
		return ValidationError{Code: "missing_namespace", Message: "namespace parameter is required"}
	}
	return nil
}

// DeleteUsers removes users matching filters and returns the deleted count.
func (s *Service) DeleteUsers(ctx context.Context, orgID uuid.UUID, opts DeleteOptions) (int, error) {
	if err := s.validateDelete(opts); err != nil {
		return 0, err
	}
	return s.store.DeleteUsers(ctx, orgID, opts.Namespace, opts.Filters)
}

// DeleteItems removes items matching filters.
func (s *Service) DeleteItems(ctx context.Context, orgID uuid.UUID, opts DeleteOptions) (int, error) {
	if err := s.validateDelete(opts); err != nil {
		return 0, err
	}
	return s.store.DeleteItems(ctx, orgID, opts.Namespace, opts.Filters)
}

// DeleteEvents removes events matching filters.
func (s *Service) DeleteEvents(ctx context.Context, orgID uuid.UUID, opts DeleteOptions) (int, error) {
	if err := s.validateDelete(opts); err != nil {
		return 0, err
	}
	return s.store.DeleteEvents(ctx, orgID, opts.Namespace, opts.Filters)
}
