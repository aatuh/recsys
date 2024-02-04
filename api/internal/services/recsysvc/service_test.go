package recsysvc

import (
	"context"
	"testing"
)

type stubEngine struct {
	items []Item
}

func (s stubEngine) Recommend(ctx context.Context, req RecommendRequest) ([]Item, []Warning, error) {
	return append([]Item(nil), s.items...), nil, nil
}

func (s stubEngine) Similar(ctx context.Context, req SimilarRequest) ([]Item, []Warning, error) {
	return append([]Item(nil), s.items...), nil, nil
}

func (s stubEngine) Version() string {
	return "stub-engine"
}

func TestRecommendDeterministicOrdering(t *testing.T) {
	engine := stubEngine{items: []Item{
		{ItemID: "b", Score: 0.9},
		{ItemID: "a", Score: 0.9},
		{ItemID: "c", Score: 0.95},
	}}
	svc := New(engine)
	items, _, meta, err := svc.Recommend(context.Background(), RecommendRequest{Surface: "home", Segment: "default", K: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.AlgoVersion != "stub-engine" {
		t.Fatalf("expected algo version stub-engine, got %q", meta.AlgoVersion)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].ItemID != "c" || items[0].Rank != 1 {
		t.Fatalf("expected first item c rank 1, got %v rank %d", items[0].ItemID, items[0].Rank)
	}
	if items[1].ItemID != "a" || items[1].Rank != 2 {
		t.Fatalf("expected second item a rank 2, got %v rank %d", items[1].ItemID, items[1].Rank)
	}
	if items[2].ItemID != "b" || items[2].Rank != 3 {
		t.Fatalf("expected third item b rank 3, got %v rank %d", items[2].ItemID, items[2].Rank)
	}
}

func TestRecommendPinnedOrderingPreserved(t *testing.T) {
	engine := stubEngine{items: []Item{
		{ItemID: "pin_b", Score: 0.9, PinRank: 2},
		{ItemID: "top", Score: 0.99},
		{ItemID: "pin_a", Score: 0.1, PinRank: 1},
	}}
	svc := New(engine)
	items, _, _, err := svc.Recommend(context.Background(), RecommendRequest{Surface: "home", Segment: "default", K: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].ItemID != "pin_a" || items[0].Rank != 1 {
		t.Fatalf("expected first item pin_a rank 1, got %v rank %d", items[0].ItemID, items[0].Rank)
	}
	if items[1].ItemID != "pin_b" || items[1].Rank != 2 {
		t.Fatalf("expected second item pin_b rank 2, got %v rank %d", items[1].ItemID, items[1].Rank)
	}
	if items[2].ItemID != "top" || items[2].Rank != 3 {
		t.Fatalf("expected third item top rank 3, got %v rank %d", items[2].ItemID, items[2].Rank)
	}
}
