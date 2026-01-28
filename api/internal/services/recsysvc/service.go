package recsysvc

import (
	"context"
	"sort"
)

// Engine defines the algorithm interface for recommendations.
type Engine interface {
	Recommend(ctx context.Context, req RecommendRequest) ([]Item, []Warning, error)
	Similar(ctx context.Context, req SimilarRequest) ([]Item, []Warning, error)
}

// Service orchestrates recommendation requests.
type Service struct {
	engine Engine
}

// New constructs a new Service.
func New(engine Engine) *Service {
	return &Service{engine: engine}
}

// Recommend returns ranked recommendations.
func (s *Service) Recommend(ctx context.Context, req RecommendRequest) ([]Item, []Warning, error) {
	if s == nil || s.engine == nil {
		return nil, nil, nil
	}
	items, warnings, err := s.engine.Recommend(ctx, req)
	if err != nil {
		return nil, warnings, err
	}
	applyDeterministicOrdering(items)
	return items, warnings, nil
}

// Similar returns similar items for a given item.
func (s *Service) Similar(ctx context.Context, req SimilarRequest) ([]Item, []Warning, error) {
	if s == nil || s.engine == nil {
		return nil, nil, nil
	}
	items, warnings, err := s.engine.Similar(ctx, req)
	if err != nil {
		return nil, warnings, err
	}
	applyDeterministicOrdering(items)
	return items, warnings, nil
}

// NewNoopEngine returns an engine that always returns empty results.
func NewNoopEngine() Engine {
	return noopEngine{}
}

type noopEngine struct{}

func (noopEngine) Recommend(ctx context.Context, req RecommendRequest) ([]Item, []Warning, error) {
	return nil, nil, nil
}

func (noopEngine) Similar(ctx context.Context, req SimilarRequest) ([]Item, []Warning, error) {
	return nil, nil, nil
}

func applyDeterministicOrdering(items []Item) {
	if len(items) == 0 {
		return
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			return items[i].ItemID < items[j].ItemID
		}
		return items[i].Score > items[j].Score
	})
	for i := range items {
		items[i].Rank = i + 1
	}
}
