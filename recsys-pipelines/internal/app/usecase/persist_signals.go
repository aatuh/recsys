package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/artifacts"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/signals"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/logger"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/signalstore"
)

// PersistSignals stores computed artifacts into DB signal tables.
type PersistSignals struct {
	rt    runtime.Runtime
	store signalstore.Store
}

func NewPersistSignals(rt runtime.Runtime, store signalstore.Store) *PersistSignals {
	return &PersistSignals{rt: rt, store: store}
}

func (uc *PersistSignals) Execute(ctx context.Context, popJSON, coocJSON []byte) error {
	if uc == nil || uc.store == nil {
		return nil
	}
	if len(popJSON) == 0 && len(coocJSON) == 0 {
		return nil
	}
	if len(popJSON) > 0 {
		if err := uc.persistPopularity(ctx, popJSON); err != nil {
			return err
		}
	}
	if len(coocJSON) > 0 {
		if err := uc.persistCooc(ctx, coocJSON); err != nil {
			return err
		}
	}
	return nil
}

func (uc *PersistSignals) persistPopularity(ctx context.Context, payload []byte) error {
	var pop artifacts.PopularityArtifactV1
	if err := json.Unmarshal(payload, &pop); err != nil {
		return fmt.Errorf("popularity json: %w", err)
	}
	day, err := parseWindowStart(pop.Window.Start)
	if err != nil {
		return err
	}
	items := make([]signals.PopularityItem, 0, len(pop.Items))
	for _, it := range pop.Items {
		if strings.TrimSpace(it.ItemID) == "" {
			continue
		}
		items = append(items, signals.PopularityItem{ItemID: it.ItemID, Score: float64(it.Count)})
	}
	uc.rt.Logger.Info(ctx, "signals: popularity",
		logger.Field{Key: "tenant", Value: pop.Tenant},
		logger.Field{Key: "surface", Value: pop.Surface},
		logger.Field{Key: "items", Value: len(items)},
	)
	return uc.store.UpsertPopularity(ctx, pop.Tenant, pop.Surface, day, items)
}

func (uc *PersistSignals) persistCooc(ctx context.Context, payload []byte) error {
	var cooc artifacts.CoocArtifactV1
	if err := json.Unmarshal(payload, &cooc); err != nil {
		return fmt.Errorf("cooc json: %w", err)
	}
	day, err := parseWindowStart(cooc.Window.Start)
	if err != nil {
		return err
	}
	items := make([]signals.CooccurrenceItem, 0, len(cooc.Neighbors))
	for _, row := range cooc.Neighbors {
		anchor := strings.TrimSpace(row.ItemID)
		if anchor == "" {
			continue
		}
		for _, neighbor := range row.Items {
			if strings.TrimSpace(neighbor.ItemID) == "" {
				continue
			}
			items = append(items, signals.CooccurrenceItem{
				ItemID:     anchor,
				NeighborID: neighbor.ItemID,
				Score:      float64(neighbor.Count),
			})
		}
	}
	uc.rt.Logger.Info(ctx, "signals: cooc",
		logger.Field{Key: "tenant", Value: cooc.Tenant},
		logger.Field{Key: "surface", Value: cooc.Surface},
		logger.Field{Key: "pairs", Value: len(items)},
	)
	return uc.store.UpsertCooccurrence(ctx, cooc.Tenant, cooc.Surface, day, items)
}

func parseWindowStart(start string) (time.Time, error) {
	if strings.TrimSpace(start) == "" {
		return time.Time{}, fmt.Errorf("window start missing")
	}
	parsed, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return time.Time{}, err
	}
	return parsed, nil
}
