package algorithm

import "context"

// ContractVersion identifies the plugin contract version for custom algorithms.
const ContractVersion = "v1"

// Algorithm defines the interface for recommendation algorithms.
type Algorithm interface {
	ID() string
	Version() string
	Recommend(ctx context.Context, req Request) (*Response, *TraceData, error)
	Similar(ctx context.Context, req SimilarItemsRequest) (*SimilarItemsResponse, error)
}
