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
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/catalog"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/logger"
)

type ComputeContentSim struct {
	rt       runtime.Runtime
	reader   catalog.Reader
	maxItems int
}

func NewComputeContentSim(rt runtime.Runtime, reader catalog.Reader, maxItems int) *ComputeContentSim {
	return &ComputeContentSim{rt: rt, reader: reader, maxItems: maxItems}
}

func (uc *ComputeContentSim) Execute(
	ctx context.Context,
	tenant, surface, segment string,
	w windows.Window,
) (artifacts.Ref, []byte, error) {
	if uc == nil || uc.reader == nil {
		return artifacts.Ref{}, nil, fmt.Errorf("catalog reader is required")
	}
	start := uc.rt.Clock.NowUTC()
	uc.rt.Logger.Info(ctx, "content_sim: start",
		logger.Field{Key: "tenant", Value: tenant},
		logger.Field{Key: "surface", Value: surface},
	)

	items, err := uc.reader.Read(ctx)
	if err != nil {
		return artifacts.Ref{}, nil, err
	}
	if uc.maxItems > 0 && len(items) > uc.maxItems {
		return artifacts.Ref{}, nil, apperr.New(
			apperr.KindLimitExceeded,
			fmt.Sprintf("catalog item limit exceeded: %d > %d", len(items), uc.maxItems),
			nil,
		)
	}

	out := make([]artifacts.ContentItem, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.ItemID) == "" {
			continue
		}
		if item.Namespace != "" && strings.TrimSpace(item.Namespace) != surface {
			continue
		}
		tags := normalizeTags(item.Tags)
		if len(tags) == 0 {
			continue
		}
		out = append(out, artifacts.ContentItem{ItemID: strings.TrimSpace(item.ItemID), Tags: tags})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ItemID < out[j].ItemID })

	noBuild := artifacts.NewContentArtifact(tenant, surface, segment, w, out, start, "", "")
	noBuild.Build = artifacts.BuildInfo{}
	noBuildBytes, err := json.Marshal(noBuild)
	if err != nil {
		return artifacts.Ref{}, nil, err
	}
	ver := lineage.SHA256Hex(noBuildBytes)

	payload := artifacts.NewContentArtifact(tenant, surface, segment, w, out, start, ver, ver)
	blob, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return artifacts.Ref{}, nil, err
	}

	ref := artifacts.Ref{
		Key:     artifacts.Key{Tenant: tenant, Surface: surface, Segment: segment, Type: artifacts.TypeContentSim},
		Window:  w,
		Version: ver,
		BuiltAt: start,
	}

	dur := uc.rt.Clock.NowUTC().Sub(start)
	uc.rt.Logger.Info(ctx, "content_sim: done",
		logger.Field{Key: "items", Value: len(out)},
		logger.Field{Key: "duration_ms", Value: dur.Milliseconds()},
	)
	return ref, blob, nil
}

func normalizeTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		t := strings.ToLower(strings.TrimSpace(tag))
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}
