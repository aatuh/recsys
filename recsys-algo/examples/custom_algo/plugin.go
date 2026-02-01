package main

import (
	"context"

	"github.com/aatuh/recsys-suite/api/recsys-algo/algorithm"
	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"
	"github.com/aatuh/recsys-suite/api/recsys-algo/rules"
)

// RecsysAlgorithmPlugin exposes the plugin entrypoint for the service.
var RecsysAlgorithmPlugin = algorithm.Plugin{
	ContractVersion: algorithm.ContractVersion,
	New: func(store recmodel.EngineStore, _ *rules.Manager, cfg algorithm.Config) (algorithm.Algorithm, error) {
		return &HelloAlgorithm{store: store, version: cfg.Version}, nil
	},
}

// HelloAlgorithm is a minimal popularity-only algorithm for demo purposes.
type HelloAlgorithm struct {
	store   recmodel.EngineStore
	version string
}

func (a *HelloAlgorithm) ID() string {
	return "hello"
}

func (a *HelloAlgorithm) Version() string {
	if a.version == "" {
		return "hello@dev"
	}
	return a.version
}

func (a *HelloAlgorithm) Recommend(
	ctx context.Context,
	req algorithm.Request,
) (*algorithm.Response, *algorithm.TraceData, error) {
	if a.store == nil {
		return &algorithm.Response{Items: nil}, nil, nil
	}
	items, err := a.store.PopularityTopK(ctx, req.OrgID, req.Namespace, 30, req.K, req.Constraints)
	if err != nil {
		return nil, nil, err
	}
	out := make([]algorithm.ScoredItem, 0, len(items))
	for _, item := range items {
		out = append(out, algorithm.ScoredItem{ItemID: item.ItemID, Score: item.Score})
	}
	return &algorithm.Response{
		ModelVersion: "hello_v1",
		Items:        out,
		SegmentID:    req.SegmentID,
	}, nil, nil
}

func (a *HelloAlgorithm) Similar(
	ctx context.Context,
	req algorithm.SimilarItemsRequest,
) (*algorithm.SimilarItemsResponse, error) {
	_ = ctx
	_ = req
	return &algorithm.SimilarItemsResponse{}, recmodel.ErrFeatureUnavailable
}

var _ algorithm.Algorithm = (*HelloAlgorithm)(nil)
