# recsys-algo

Deterministic recommendation engine with explainable scoring, optional
personalization, and merchandising rules. It follows a ports-and-adapters
style: the `model` package defines store interfaces, and `algorithm` consumes
them to produce ranked outputs.

## What it does

- Blends popularity, co-visitation, and similarity signals.
- Applies personalization boosts from user tag profiles.
- Supports MMR-style diversification and brand/category caps.
- Enforces merchandising rules (pin/boost/block) with caching.
- Emits explain blocks and trace data for debugging and audits.

## What it does not do

- It does not store events, embeddings, or item data.
- It does not manage training pipelines for models or embeddings.
- It does not enforce business-specific availability beyond what your store returns.

## Quickstart

```bash
go get github.com/aatuh/recsys-suite/api/recsys-algo
```

### Minimal example (popularity-only)

```go
package main

import (
	"context"
	"fmt"

	"github.com/aatuh/recsys-suite/api/recsys-algo/algorithm"
	recmodel "github.com/aatuh/recsys-suite/api/recsys-algo/model"

	"github.com/google/uuid"
)

type popStore struct{}

func (popStore) PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string, halfLifeDays float64, k int, c *recmodel.PopConstraints) ([]recmodel.ScoredItem, error) {
	return []recmodel.ScoredItem{
		{ItemID: "a", Score: 10},
		{ItemID: "b", Score: 8},
		{ItemID: "c", Score: 6},
	}, nil
}

func (popStore) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]recmodel.ItemTags, error) {
	return map[string]recmodel.ItemTags{}, nil
}

func main() {
	engine := algorithm.NewEngine(algorithm.Config{BlendAlpha: 1}, popStore{}, nil)
	resp, _, err := engine.Recommend(context.Background(), algorithm.Request{
		OrgID:     uuid.New(),
		Namespace: "default",
		K:         3,
	})
	if err != nil {
		panic(err)
	}
	for _, item := range resp.Items {
		fmt.Printf("%s %.2f\n", item.ItemID, item.Score)
	}
}
```

See `examples/basic` for a runnable version.

### Full example (anchors + similarity + rules + MMR)

For a full pipeline with anchors, similarity signals, and merchandising rules,
see:

1. `examples/personalized` (user profile + personalization)
2. `examples/rules` (pin/boost/block rules)

## Signals, weights, and explain/trace

- `BlendAlpha` controls popularity weight.
- `BlendBeta` controls co-visitation weight.
- `BlendGamma` controls similarity weight (embedding/collab/content/session).
- `IncludeReasons` and `ExplainLevel` enrich the response; `TraceData` provides
  deeper diagnostics for audit logging.

## Implementing store ports

At minimum, implement:

- `model.PopularityStore` (popularity candidates)
- `model.TagStore` (item tags for filters and MMR/caps)

Optional capabilities enable more signals:

- `model.ProfileStore` (personalization)
- `model.CooccurrenceStore` and `model.HistoryStore` (co-visitation)
- `model.EmbeddingStore` (embedding similarity)
- `model.CollaborativeStore` (ALS/CF)
- `model.ContentStore` (tag overlap similarity)
- `model.SessionStore` (session sequences)
- `model.EventStore` (event-based exclusions)

If a capability is missing, the engine treats the signal as unavailable and
continues. You can also return `model.ErrFeatureUnavailable` from a method to
disable a signal at runtime.
