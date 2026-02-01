package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/apperr"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/artifacts"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/lineage"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/datasource"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/logger"
)

type ComputePopularity struct {
	rt        runtime.Runtime
	canonical datasource.CanonicalStore
	topN      int
	maxItems  int
}

func NewComputePopularity(
	rt runtime.Runtime,
	canonical datasource.CanonicalStore,
	topN int,
	maxItems int,
) *ComputePopularity {
	return &ComputePopularity{rt: rt, canonical: canonical, topN: topN, maxItems: maxItems}
}

func (uc *ComputePopularity) Execute(
	ctx context.Context,
	tenant, surface, segment string,
	w windows.Window,
) (artifacts.Ref, []byte, error) {
	start := uc.rt.Clock.NowUTC()
	uc.rt.Logger.Info(ctx, "popularity: start",
		logger.Field{Key: "tenant", Value: tenant},
		logger.Field{Key: "surface", Value: surface},
	)

	evCh, errCh := uc.canonical.ReadExposureEvents(ctx, tenant, surface, w)

	counts := map[string]int64{}
	for ev := range evCh {
		if _, ok := counts[ev.ItemID]; !ok {
			if uc.maxItems > 0 && len(counts) >= uc.maxItems {
				return artifacts.Ref{}, nil, apperr.New(
					apperr.KindLimitExceeded,
					fmt.Sprintf("distinct item limit exceeded: %d >= %d", len(counts), uc.maxItems),
					nil,
				)
			}
		}
		counts[ev.ItemID]++
	}
	if err := <-errCh; err != nil {
		return artifacts.Ref{}, nil, err
	}

	items := make([]artifacts.PopularityItem, 0, len(counts))
	for id, c := range counts {
		items = append(items, artifacts.PopularityItem{ItemID: id, Count: c})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].ItemID < items[j].ItemID
		}
		return items[i].Count > items[j].Count
	})
	limit := uc.topN
	if limit <= 0 {
		limit = 1000
	}
	if len(items) > limit {
		items = items[:limit]
	}

	// Stable version hash from payload without build info.
	noBuild := artifacts.NewPopularityArtifact(tenant, surface, segment, w, items, start, "", "")
	noBuild.Build = artifacts.BuildInfo{}
	noBuildBytes, err := json.Marshal(noBuild)
	if err != nil {
		return artifacts.Ref{}, nil, err
	}
	ver := lineage.SHA256Hex(noBuildBytes)

	payload := artifacts.NewPopularityArtifact(tenant, surface, segment, w, items, start, ver, ver)
	blob, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return artifacts.Ref{}, nil, err
	}

	ref := artifacts.Ref{
		Key:     artifacts.Key{Tenant: tenant, Surface: surface, Segment: segment, Type: artifacts.TypePopularity},
		Window:  w,
		Version: ver,
		BuiltAt: start,
	}

	dur := uc.rt.Clock.NowUTC().Sub(start)
	uc.rt.Logger.Info(ctx, "popularity: done",
		logger.Field{Key: "items", Value: len(items)},
		logger.Field{Key: "duration_ms", Value: dur.Milliseconds()},
	)
	return ref, blob, nil
}
