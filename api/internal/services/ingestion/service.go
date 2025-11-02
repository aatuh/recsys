package ingestion

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"recsys/internal/store"
	specs "recsys/specs/types"
)

// ValidationError represents a client-side error during ingestion processing.
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

// Store defines the persistence contract required by the ingestion service.
type Store interface {
	UpsertItems(ctx context.Context, orgID uuid.UUID, namespace string, items []store.ItemUpsert) error
	UpsertUsers(ctx context.Context, orgID uuid.UUID, namespace string, users []store.UserUpsert) error
	InsertEvents(ctx context.Context, orgID uuid.UUID, namespace string, events []store.EventInsert) error
}

// Service handles ingestion domain mutations.
type Service struct {
	store         Store
	embeddingDims int
	now           func() time.Time
}

// Option configures the Service.
type Option func(*Service)

// WithNow overrides the wall clock used for default timestamps.
func WithNow(fn func() time.Time) Option {
	return func(s *Service) {
		if fn != nil {
			s.now = fn
		}
	}
}

// WithEmbeddingDims overrides the expected embedding dimensions.
func WithEmbeddingDims(dims int) Option {
	return func(s *Service) {
		if dims > 0 {
			s.embeddingDims = dims
		}
	}
}

// New constructs a Service backed by the provided store.
func New(st Store, opts ...Option) *Service {
	svc := &Service{
		store:         st,
		embeddingDims: store.EmbeddingDims,
		now:           time.Now,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// UpsertItems validates and persists item payloads.

func (s *Service) UpsertItems(ctx context.Context, orgID uuid.UUID, req specs.ItemsUpsertRequest) error {
	if len(req.Items) == 0 {
		return nil
	}
	batch := make([]store.ItemUpsert, 0, len(req.Items))
	for idx, it := range req.Items {
		var emb *[]float64
		if len(it.Embedding) > 0 {
			if len(it.Embedding) != s.embeddingDims {
				return ValidationError{
					Code:    "embedding_dim_mismatch",
					Message: fmt.Sprintf("embedding length must be %d", s.embeddingDims),
					Details: map[string]any{
						"index": idx,
						"got":   len(it.Embedding),
					},
				}
			}
			tmp := make([]float64, len(it.Embedding))
			copy(tmp, it.Embedding)
			emb = &tmp
		}
		batch = append(batch, store.ItemUpsert{
			ItemID:          it.ItemID,
			Available:       it.Available,
			Price:           it.Price,
			Tags:            it.Tags,
			Props:           it.Props,
			Embedding:       emb,
			Brand:           cloneStringPtr(it.Brand),
			Category:        cloneStringPtr(it.Category),
			CategoryPath:    cloneStringSlicePtr(it.CategoryPath),
			Description:     cloneStringPtr(it.Description),
			ImageURL:        cloneStringPtr(it.ImageURL),
			MetadataVersion: cloneStringPtr(it.MetadataVersion),
		})
	}
	if s.store == nil {
		return errors.New("ingestion service: store is nil")
	}
	return s.store.UpsertItems(ctx, orgID, req.Namespace, batch)
}

// UpsertUsers persists user metadata.
func (s *Service) UpsertUsers(ctx context.Context, orgID uuid.UUID, req specs.UsersUpsertRequest) error {
	if s.store == nil {
		return errors.New("ingestion service: store is nil")
	}
	if len(req.Users) == 0 {
		return nil
	}
	batch := make([]store.UserUpsert, 0, len(req.Users))
	for _, u := range req.Users {
		batch = append(batch, store.UserUpsert{UserID: u.UserID, Traits: u.Traits})
	}
	return s.store.UpsertUsers(ctx, orgID, req.Namespace, batch)
}

// InsertEvents validates and stores incoming events.
func (s *Service) InsertEvents(ctx context.Context, orgID uuid.UUID, req specs.EventsBatchRequest) error {
	if len(req.Events) == 0 {
		return nil
	}
	now := s.now
	if now == nil {
		now = time.Now
	}

	batch := make([]store.EventInsert, 0, len(req.Events))
	for idx, ev := range req.Events {
		ts := now().UTC()
		if ev.TS != "" {
			parsed, err := time.Parse(time.RFC3339, ev.TS)
			if err != nil {
				return ValidationError{
					Code:    "invalid_timestamp",
					Message: "ts must be RFC3339",
					Details: map[string]any{
						"index": idx,
						"value": ev.TS,
					},
				}
			}
			ts = parsed
		}
		batch = append(batch, store.EventInsert{
			UserID:        ev.UserID,
			ItemID:        ev.ItemID,
			Type:          ev.Type,
			Value:         ev.Value,
			TS:            ts,
			Meta:          ev.Meta,
			SourceEventID: ev.SourceEventID,
		})
	}
	if s.store == nil {
		return errors.New("ingestion service: store is nil")
	}
	return s.store.InsertEvents(ctx, orgID, req.Namespace, batch)
}

func cloneStringPtr(src *string) *string {
	if src == nil {
		return nil
	}
	val := *src
	return &val
}

func cloneStringSlicePtr(src *[]string) *[]string {
	if src == nil {
		return nil
	}
	cp := append([]string(nil), (*src)...)
	return &cp
}
