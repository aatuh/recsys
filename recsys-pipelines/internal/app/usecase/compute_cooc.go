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

type ComputeCooc struct {
	rt                 runtime.Runtime
	canonical          datasource.CanonicalStore
	maxNeighbors       int
	minSupport         int64
	maxItemsPerArt     int
	maxSessions        int
	maxItemsPerSession int
	maxDistinctItems   int
}

func NewComputeCooc(
	rt runtime.Runtime,
	canonical datasource.CanonicalStore,
	maxNeighbors int,
	minSupport int64,
	maxItemsPerArt int,
	maxSessions int,
	maxItemsPerSession int,
	maxDistinctItems int,
) *ComputeCooc {
	return &ComputeCooc{
		rt:                 rt,
		canonical:          canonical,
		maxNeighbors:       maxNeighbors,
		minSupport:         minSupport,
		maxItemsPerArt:     maxItemsPerArt,
		maxSessions:        maxSessions,
		maxItemsPerSession: maxItemsPerSession,
		maxDistinctItems:   maxDistinctItems,
	}
}

func (uc *ComputeCooc) Execute(
	ctx context.Context,
	tenant, surface, segment string,
	w windows.Window,
) (artifacts.Ref, []byte, error) {
	start := uc.rt.Clock.NowUTC()
	uc.rt.Logger.Info(ctx, "cooc: start",
		logger.Field{Key: "tenant", Value: tenant},
		logger.Field{Key: "surface", Value: surface},
	)

	evCh, errCh := uc.canonical.ReadExposureEvents(ctx, tenant, surface, w)

	sessions := map[string]map[string]struct{}{}
	distinct := map[string]struct{}{}
	for ev := range evCh {
		m := sessions[ev.SessionID]
		if m == nil {
			if uc.maxSessions > 0 && len(sessions) >= uc.maxSessions {
				return artifacts.Ref{}, nil, apperr.New(
					apperr.KindLimitExceeded,
					fmt.Sprintf("session limit exceeded: %d >= %d", len(sessions), uc.maxSessions),
					nil,
				)
			}
			m = map[string]struct{}{}
			sessions[ev.SessionID] = m
		}
		if _, ok := m[ev.ItemID]; !ok {
			if uc.maxItemsPerSession > 0 && len(m) >= uc.maxItemsPerSession {
				return artifacts.Ref{}, nil, apperr.New(
					apperr.KindLimitExceeded,
					fmt.Sprintf("items per session limit exceeded: %d >= %d", len(m), uc.maxItemsPerSession),
					nil,
				)
			}
			m[ev.ItemID] = struct{}{}
			if _, ok := distinct[ev.ItemID]; !ok {
				if uc.maxDistinctItems > 0 && len(distinct) >= uc.maxDistinctItems {
					return artifacts.Ref{}, nil, apperr.New(
						apperr.KindLimitExceeded,
						fmt.Sprintf("distinct item limit exceeded: %d >= %d", len(distinct), uc.maxDistinctItems),
						nil,
					)
				}
				distinct[ev.ItemID] = struct{}{}
			}
		}
	}
	if err := <-errCh; err != nil {
		return artifacts.Ref{}, nil, err
	}

	pairs := map[string]map[string]int64{}
	for _, itemSet := range sessions {
		items := make([]string, 0, len(itemSet))
		for id := range itemSet {
			items = append(items, id)
		}
		sort.Strings(items)

		for i := 0; i < len(items); i++ {
			for j := i + 1; j < len(items); j++ {
				a := items[i]
				b := items[j]
				if pairs[a] == nil {
					pairs[a] = map[string]int64{}
				}
				if pairs[b] == nil {
					pairs[b] = map[string]int64{}
				}
				pairs[a][b]++
				pairs[b][a]++
			}
		}
	}

	rows := make([]artifacts.CoocRow, 0, len(pairs))
	for itemID, neigh := range pairs {
		ns := make([]artifacts.CoocNeighbor, 0, len(neigh))
		for nid, c := range neigh {
			if c < uc.minSupport {
				continue
			}
			ns = append(ns, artifacts.CoocNeighbor{ItemID: nid, Count: c})
		}
		sort.Slice(ns, func(i, j int) bool {
			if ns[i].Count == ns[j].Count {
				return ns[i].ItemID < ns[j].ItemID
			}
			return ns[i].Count > ns[j].Count
		})
		if uc.maxNeighbors > 0 && len(ns) > uc.maxNeighbors {
			ns = ns[:uc.maxNeighbors]
		}
		rows = append(rows, artifacts.CoocRow{ItemID: itemID, Items: ns})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ItemID < rows[j].ItemID })
	if uc.maxItemsPerArt > 0 && len(rows) > uc.maxItemsPerArt {
		rows = rows[:uc.maxItemsPerArt]
	}

	noBuild := artifacts.NewCoocArtifact(tenant, surface, segment, w, rows, start, "", "")
	noBuild.Build = artifacts.BuildInfo{}
	noBuildBytes, err := json.Marshal(noBuild)
	if err != nil {
		return artifacts.Ref{}, nil, err
	}
	ver := lineage.SHA256Hex(noBuildBytes)

	payload := artifacts.NewCoocArtifact(tenant, surface, segment, w, rows, start, ver, ver)
	blob, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return artifacts.Ref{}, nil, err
	}

	ref := artifacts.Ref{
		Key:     artifacts.Key{Tenant: tenant, Surface: surface, Segment: segment, Type: artifacts.TypeCooc},
		Window:  w,
		Version: ver,
		BuiltAt: start,
	}

	dur := uc.rt.Clock.NowUTC().Sub(start)
	uc.rt.Logger.Info(ctx, "cooc: done",
		logger.Field{Key: "rows", Value: len(rows)},
		logger.Field{Key: "duration_ms", Value: dur.Milliseconds()},
	)
	return ref, blob, nil
}
