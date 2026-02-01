package algorithm

import (
	"context"
	"strings"

	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"
	"github.com/aatuh/recsys-suite/api/recsys-algo/rules"

	"github.com/google/uuid"
)

// DefaultAlgorithm wires the standard engine + similar-items logic.
type DefaultAlgorithm struct {
	id      string
	version string
	engine  *Engine
	similar *SimilarItemsEngine
}

// NewDefaultAlgorithm constructs the default algorithm implementation.
func NewDefaultAlgorithm(cfg Config, store recmodel.EngineStore, rulesManager *rules.Manager, opts ...EngineOption) *DefaultAlgorithm {
	if store == nil {
		store = noopEngineStore{}
	}
	engine := NewEngine(cfg, store, rulesManager, opts...)
	similar := NewSimilarItemsEngine(store, cfg.CoVisWindowDays)
	version := strings.TrimSpace(cfg.Version)
	if version == "" {
		version = "recsys-algo@local"
	}
	return &DefaultAlgorithm{
		id:      "default",
		version: version,
		engine:  engine,
		similar: similar,
	}
}

// ID returns the stable algorithm identifier.
func (a *DefaultAlgorithm) ID() string {
	if a == nil || strings.TrimSpace(a.id) == "" {
		return "default"
	}
	return a.id
}

// Version returns the algorithm build version label.
func (a *DefaultAlgorithm) Version() string {
	if a == nil || strings.TrimSpace(a.version) == "" {
		return "recsys-algo@local"
	}
	return a.version
}

// Recommend delegates to the core engine.
func (a *DefaultAlgorithm) Recommend(ctx context.Context, req Request) (*Response, *TraceData, error) {
	if a == nil || a.engine == nil {
		return nil, nil, nil
	}
	return a.engine.Recommend(ctx, req)
}

// Similar delegates to the similar-items engine.
func (a *DefaultAlgorithm) Similar(ctx context.Context, req SimilarItemsRequest) (*SimilarItemsResponse, error) {
	if a == nil || a.similar == nil {
		return nil, recmodel.ErrFeatureUnavailable
	}
	return a.similar.FindSimilar(ctx, req)
}

type noopEngineStore struct{}

func (noopEngineStore) PopularityTopK(
	context.Context,
	uuid.UUID,
	string,
	float64,
	int,
	*recmodel.PopConstraints,
) ([]recmodel.ScoredItem, error) {
	return nil, nil
}

func (noopEngineStore) ListItemsTags(
	context.Context,
	uuid.UUID,
	string,
	[]string,
) (map[string]recmodel.ItemTags, error) {
	return map[string]recmodel.ItemTags{}, nil
}

var _ Algorithm = (*DefaultAlgorithm)(nil)
var _ recmodel.EngineStore = (*noopEngineStore)(nil)
