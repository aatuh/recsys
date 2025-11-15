package ingestion

import (
	"context"
	"testing"

	"recsys/internal/store"
	spectypes "recsys/specs/types"

	"github.com/google/uuid"
)

type stubStore struct {
	lastItems []store.ItemUpsert
}

func (s *stubStore) UpsertItems(ctx context.Context, orgID uuid.UUID, namespace string, items []store.ItemUpsert) error {
	s.lastItems = append([]store.ItemUpsert(nil), items...)
	return nil
}

func (s *stubStore) UpsertUsers(ctx context.Context, orgID uuid.UUID, namespace string, users []store.UserUpsert) error {
	return nil
}

func (s *stubStore) InsertEvents(ctx context.Context, orgID uuid.UUID, namespace string, events []store.EventInsert) error {
	return nil
}

func TestUpsertItems_GeneratesFallbackEmbedding(t *testing.T) {
	stub := &stubStore{}
	svc := New(stub)

	err := svc.UpsertItems(context.Background(), uuid.New(), spectypes.ItemsUpsertRequest{
		Namespace: "default",
		Items: []spectypes.Item{
			{
				ItemID:      "sku-1",
				Available:   true,
				Brand:       ptr("Voltify"),
				Category:    ptr("Electronics"),
				Description: ptr("Smart home hub"),
			},
		},
	})
	if err != nil {
		t.Fatalf("UpsertItems returned error: %v", err)
	}

	if len(stub.lastItems) != 1 {
		t.Fatalf("expected 1 item captured, got %d", len(stub.lastItems))
	}
	emb := stub.lastItems[0].Embedding
	if emb == nil {
		t.Fatalf("expected fallback embedding to be generated")
	}
	if len(*emb) != store.EmbeddingDims {
		t.Fatalf("embedding length mismatch: got %d want %d", len(*emb), store.EmbeddingDims)
	}
}

func TestUpsertItems_NoEmbeddingWhenInsufficientMetadata(t *testing.T) {
	stub := &stubStore{}
	svc := New(stub)

	err := svc.UpsertItems(context.Background(), uuid.New(), spectypes.ItemsUpsertRequest{
		Namespace: "default",
		Items: []spectypes.Item{
			{ItemID: "sku-plain", Available: true},
		},
	})
	if err != nil {
		t.Fatalf("UpsertItems returned error: %v", err)
	}

	if len(stub.lastItems) != 1 {
		t.Fatalf("expected 1 item captured, got %d", len(stub.lastItems))
	}
	if stub.lastItems[0].Embedding != nil {
		t.Fatalf("did not expect embedding when metadata is empty")
	}
}

func ptr[T any](v T) *T { return &v }
