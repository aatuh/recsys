package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/apperr"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/artifacts"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/lineage"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/datasource"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/logger"
)

type ComputeSessionSeq struct {
	rt              runtime.Runtime
	canonical       datasource.CanonicalStore
	topN            int
	maxUsers        int
	maxItemsPerUser int
}

func NewComputeSessionSeq(
	rt runtime.Runtime,
	canonical datasource.CanonicalStore,
	topN int,
	maxUsers int,
	maxItemsPerUser int,
) *ComputeSessionSeq {
	return &ComputeSessionSeq{
		rt:              rt,
		canonical:       canonical,
		topN:            topN,
		maxUsers:        maxUsers,
		maxItemsPerUser: maxItemsPerUser,
	}
}

type userEvent struct {
	ts   int64
	item string
}

func (uc *ComputeSessionSeq) Execute(
	ctx context.Context,
	tenant, surface, segment string,
	w windows.Window,
) (artifacts.Ref, []byte, error) {
	start := uc.rt.Clock.NowUTC()
	uc.rt.Logger.Info(ctx, "session_seq: start",
		logger.Field{Key: "tenant", Value: tenant},
		logger.Field{Key: "surface", Value: surface},
	)

	evCh, errCh := uc.canonical.ReadExposureEvents(ctx, tenant, surface, w)
	events := map[string][]userEvent{}
	for ev := range evCh {
		userID := strings.TrimSpace(ev.UserID)
		if userID == "" {
			continue
		}
		if _, ok := events[userID]; !ok {
			if uc.maxUsers > 0 && len(events) >= uc.maxUsers {
				return artifacts.Ref{}, nil, apperr.New(
					apperr.KindLimitExceeded,
					fmt.Sprintf("distinct user limit exceeded: %d >= %d", len(events), uc.maxUsers),
					nil,
				)
			}
		}
		events[userID] = append(events[userID], userEvent{
			ts:   ev.TS.UTC().UnixNano(),
			item: strings.TrimSpace(ev.ItemID),
		})
	}
	if err := <-errCh; err != nil {
		return artifacts.Ref{}, nil, err
	}

	users := make([]artifacts.SessionSeqUser, 0, len(events))
	limit := uc.topN
	if limit <= 0 {
		limit = 200
	}
	for userID, list := range events {
		sort.SliceStable(list, func(i, j int) bool { return list[i].ts < list[j].ts })
		counts := map[string]int64{}
		for i := 0; i+1 < len(list); i++ {
			next := list[i+1].item
			if next == "" {
				continue
			}
			if uc.maxItemsPerUser > 0 {
				if _, ok := counts[next]; !ok && len(counts) >= uc.maxItemsPerUser {
					return artifacts.Ref{}, nil, apperr.New(
						apperr.KindLimitExceeded,
						fmt.Sprintf("items per user limit exceeded: %d >= %d", len(counts), uc.maxItemsPerUser),
						nil,
					)
				}
			}
			counts[next]++
		}
		items := make([]artifacts.SessionSeqItem, 0, len(counts))
		for id, c := range counts {
			items = append(items, artifacts.SessionSeqItem{ItemID: id, Score: float64(c)})
		}
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].Score == items[j].Score {
				return items[i].ItemID < items[j].ItemID
			}
			return items[i].Score > items[j].Score
		})
		if len(items) > limit {
			items = items[:limit]
		}
		users = append(users, artifacts.SessionSeqUser{UserID: userID, Items: items})
	}
	sort.SliceStable(users, func(i, j int) bool { return users[i].UserID < users[j].UserID })

	noBuild := artifacts.NewSessionSeqArtifact(tenant, surface, segment, w, users, start, "", "")
	noBuild.Build = artifacts.BuildInfo{}
	noBuildBytes, err := json.Marshal(noBuild)
	if err != nil {
		return artifacts.Ref{}, nil, err
	}
	ver := lineage.SHA256Hex(noBuildBytes)

	payload := artifacts.NewSessionSeqArtifact(tenant, surface, segment, w, users, start, ver, ver)
	blob, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return artifacts.Ref{}, nil, err
	}

	ref := artifacts.Ref{
		Key:     artifacts.Key{Tenant: tenant, Surface: surface, Segment: segment, Type: artifacts.TypeSessionSeq},
		Window:  w,
		Version: ver,
		BuiltAt: start,
	}

	dur := uc.rt.Clock.NowUTC().Sub(start)
	uc.rt.Logger.Info(ctx, "session_seq: done",
		logger.Field{Key: "users", Value: len(users)},
		logger.Field{Key: "duration_ms", Value: dur.Milliseconds()},
	)
	return ref, blob, nil
}
