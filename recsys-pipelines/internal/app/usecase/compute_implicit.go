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

type ComputeImplicit struct {
	rt              runtime.Runtime
	canonical       datasource.CanonicalStore
	topN            int
	maxUsers        int
	maxItemsPerUser int
}

func NewComputeImplicit(
	rt runtime.Runtime,
	canonical datasource.CanonicalStore,
	topN int,
	maxUsers int,
	maxItemsPerUser int,
) *ComputeImplicit {
	return &ComputeImplicit{
		rt:              rt,
		canonical:       canonical,
		topN:            topN,
		maxUsers:        maxUsers,
		maxItemsPerUser: maxItemsPerUser,
	}
}

func (uc *ComputeImplicit) Execute(
	ctx context.Context,
	tenant, surface, segment string,
	w windows.Window,
) (artifacts.Ref, []byte, error) {
	start := uc.rt.Clock.NowUTC()
	uc.rt.Logger.Info(ctx, "implicit: start",
		logger.Field{Key: "tenant", Value: tenant},
		logger.Field{Key: "surface", Value: surface},
	)

	evCh, errCh := uc.canonical.ReadExposureEvents(ctx, tenant, surface, w)
	users := map[string]map[string]int64{}
	for ev := range evCh {
		userID := strings.TrimSpace(ev.UserID)
		if userID == "" {
			continue
		}
		items, ok := users[userID]
		if !ok {
			if uc.maxUsers > 0 && len(users) >= uc.maxUsers {
				return artifacts.Ref{}, nil, apperr.New(
					apperr.KindLimitExceeded,
					fmt.Sprintf("distinct user limit exceeded: %d >= %d", len(users), uc.maxUsers),
					nil,
				)
			}
			items = map[string]int64{}
			users[userID] = items
		}
		if uc.maxItemsPerUser > 0 {
			if _, ok := items[ev.ItemID]; !ok && len(items) >= uc.maxItemsPerUser {
				return artifacts.Ref{}, nil, apperr.New(
					apperr.KindLimitExceeded,
					fmt.Sprintf("items per user limit exceeded: %d >= %d", len(items), uc.maxItemsPerUser),
					nil,
				)
			}
		}
		items[ev.ItemID]++
	}
	if err := <-errCh; err != nil {
		return artifacts.Ref{}, nil, err
	}

	userRows := make([]artifacts.ImplicitUser, 0, len(users))
	limit := uc.topN
	if limit <= 0 {
		limit = 200
	}
	for userID, items := range users {
		list := make([]artifacts.ImplicitItem, 0, len(items))
		for id, c := range items {
			if id == "" {
				continue
			}
			list = append(list, artifacts.ImplicitItem{ItemID: id, Score: float64(c)})
		}
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].Score == list[j].Score {
				return list[i].ItemID < list[j].ItemID
			}
			return list[i].Score > list[j].Score
		})
		if len(list) > limit {
			list = list[:limit]
		}
		userRows = append(userRows, artifacts.ImplicitUser{UserID: userID, Items: list})
	}
	sort.SliceStable(userRows, func(i, j int) bool { return userRows[i].UserID < userRows[j].UserID })

	noBuild := artifacts.NewImplicitArtifact(tenant, surface, segment, w, userRows, start, "", "")
	noBuild.Build = artifacts.BuildInfo{}
	noBuildBytes, err := json.Marshal(noBuild)
	if err != nil {
		return artifacts.Ref{}, nil, err
	}
	ver := lineage.SHA256Hex(noBuildBytes)

	payload := artifacts.NewImplicitArtifact(tenant, surface, segment, w, userRows, start, ver, ver)
	blob, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return artifacts.Ref{}, nil, err
	}

	ref := artifacts.Ref{
		Key:     artifacts.Key{Tenant: tenant, Surface: surface, Segment: segment, Type: artifacts.TypeImplicit},
		Window:  w,
		Version: ver,
		BuiltAt: start,
	}

	dur := uc.rt.Clock.NowUTC().Sub(start)
	uc.rt.Logger.Info(ctx, "implicit: done",
		logger.Field{Key: "users", Value: len(userRows)},
		logger.Field{Key: "duration_ms", Value: dur.Milliseconds()},
	)
	return ref, blob, nil
}
