package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/apperr"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/runtime"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/signals"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/catalog"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/logger"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/signalstore"
)

// ImportItemTags ingests catalog tags into the DB-backed signal store.
type ImportItemTags struct {
	rt               runtime.Runtime
	reader           catalog.Reader
	store            signalstore.Store
	maxItems         int
	defaultNamespace string
}

func NewImportItemTags(rt runtime.Runtime, reader catalog.Reader, store signalstore.Store, maxItems int) *ImportItemTags {
	return &ImportItemTags{rt: rt, reader: reader, store: store, maxItems: maxItems, defaultNamespace: "default"}
}

func (uc *ImportItemTags) Execute(ctx context.Context, tenant, namespace string) error {
	if uc == nil || uc.reader == nil || uc.store == nil {
		return fmt.Errorf("catalog reader and signal store are required")
	}
	if strings.TrimSpace(tenant) == "" {
		return fmt.Errorf("tenant is required")
	}
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		namespace = uc.defaultNamespace
	}

	start := uc.rt.Clock.NowUTC()
	uc.rt.Logger.Info(ctx, "catalog: start",
		logger.Field{Key: "tenant", Value: tenant},
		logger.Field{Key: "namespace", Value: namespace},
	)

	items, err := uc.reader.Read(ctx)
	if err != nil {
		return err
	}
	if uc.maxItems > 0 && len(items) > uc.maxItems {
		return apperr.New(
			apperr.KindLimitExceeded,
			fmt.Sprintf("catalog item limit exceeded: %d > %d", len(items), uc.maxItems),
			nil,
		)
	}

	filtered := make([]signals.ItemTag, 0, len(items))
	now := time.Now().UTC()
	for _, item := range items {
		if strings.TrimSpace(item.ItemID) == "" {
			continue
		}
		if strings.TrimSpace(item.Namespace) == "" {
			item.Namespace = namespace
		}
		if item.CreatedAt.IsZero() {
			item.CreatedAt = now
		}
		filtered = append(filtered, item)
	}
	if err := uc.store.UpsertItemTags(ctx, tenant, namespace, filtered); err != nil {
		return err
	}

	dur := uc.rt.Clock.NowUTC().Sub(start)
	uc.rt.Logger.Info(ctx, "catalog: done",
		logger.Field{Key: "items", Value: len(filtered)},
		logger.Field{Key: "duration_ms", Value: dur.Milliseconds()},
	)
	return nil
}
