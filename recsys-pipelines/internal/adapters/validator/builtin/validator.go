package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/apperr"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/artifacts"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/lineage"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/datasource"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/validator"
)

type Options struct {
	// Canonical limits.
	MinEvents           int
	MaxEvents           int
	MaxDistinctItems    int
	MaxDistinctSessions int

	// Artifact limits.
	MaxPopularityItems int
	MaxCoocRows        int
	MaxCoocNeighbors   int
}

type Builtin struct {
	canonical datasource.CanonicalStore
	opts      Options
}

var _ validator.Validator = (*Builtin)(nil)

func New(canonical datasource.CanonicalStore, opts Options) *Builtin {
	return &Builtin{canonical: canonical, opts: opts}
}

func (v *Builtin) ValidateCanonical(ctx context.Context, tenant, surface string, w windows.Window) error {
	if err := w.Validate(); err != nil {
		return apperr.New(apperr.KindInvalidArgument, "invalid window", err)
	}
	evCh, errCh := v.canonical.ReadExposureEvents(ctx, tenant, surface, w)
	count := 0
	distItems := map[string]struct{}{}
	distSessions := map[string]struct{}{}
	for ev := range evCh {
		if err := ev.Validate(); err != nil {
			return apperr.New(apperr.KindValidationFailed, "invalid canonical event", err)
		}
		if !w.Contains(ev.TS) {
			return apperr.New(
				apperr.KindValidationFailed,
				"canonical event timestamp outside window",
				fmt.Errorf("ts=%s", ev.TS.UTC().Format(time.RFC3339)),
			)
		}
		count++
		if v.opts.MaxEvents > 0 && count > v.opts.MaxEvents {
			return apperr.New(
				apperr.KindValidationFailed,
				"canonical event count exceeds max",
				fmt.Errorf("events=%d max=%d", count, v.opts.MaxEvents),
			)
		}
		if v.opts.MaxDistinctItems > 0 {
			if _, ok := distItems[ev.ItemID]; !ok {
				distItems[ev.ItemID] = struct{}{}
				if len(distItems) > v.opts.MaxDistinctItems {
					return apperr.New(
						apperr.KindValidationFailed,
						"distinct item count exceeds max",
						fmt.Errorf("items=%d max=%d", len(distItems), v.opts.MaxDistinctItems),
					)
				}
			}
		}
		if v.opts.MaxDistinctSessions > 0 {
			if _, ok := distSessions[ev.SessionID]; !ok {
				distSessions[ev.SessionID] = struct{}{}
				if len(distSessions) > v.opts.MaxDistinctSessions {
					return apperr.New(
						apperr.KindValidationFailed,
						"distinct session count exceeds max",
						fmt.Errorf("sessions=%d max=%d", len(distSessions), v.opts.MaxDistinctSessions),
					)
				}
			}
		}
	}
	if err := <-errCh; err != nil {
		return apperr.New(apperr.KindDependency, "read canonical events", err)
	}
	if v.opts.MinEvents > 0 && count < v.opts.MinEvents {
		return apperr.New(
			apperr.KindValidationFailed,
			"canonical event count below min",
			fmt.Errorf("events=%d min=%d", count, v.opts.MinEvents),
		)
	}
	return nil
}

func (v *Builtin) ValidateArtifact(ctx context.Context, ref artifacts.Ref, artifactJSON []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := ref.Validate(); err != nil {
		return apperr.New(apperr.KindInvalidArgument, "invalid artifact ref", err)
	}
	switch ref.Key.Type {
	case artifacts.TypePopularity:
		return v.validatePopularity(ref, artifactJSON)
	case artifacts.TypeCooc:
		return v.validateCooc(ref, artifactJSON)
	case artifacts.TypeImplicit:
		return apperr.New(apperr.KindInvalidArgument, "implicit artifacts are not supported by pipelines yet", nil)
	default:
		return apperr.New(apperr.KindInvalidArgument, "unknown artifact type", fmt.Errorf("type=%s", ref.Key.Type))
	}
}

func (v *Builtin) validatePopularity(ref artifacts.Ref, b []byte) error {
	var a artifacts.PopularityArtifactV1
	if err := json.Unmarshal(b, &a); err != nil {
		return apperr.New(apperr.KindValidationFailed, "invalid popularity artifact json", err)
	}
	if a.V != 1 {
		return apperr.New(apperr.KindValidationFailed, "unsupported popularity artifact version", fmt.Errorf("v=%d", a.V))
	}
	if a.ArtifactType != string(artifacts.TypePopularity) {
		return apperr.New(apperr.KindValidationFailed, "popularity artifact_type mismatch", fmt.Errorf("type=%s", a.ArtifactType))
	}
	if a.Build.Version != ref.Version {
		return apperr.New(
			apperr.KindValidationFailed,
			"artifact version mismatch",
			fmt.Errorf("ref=%s payload=%s", ref.Version, a.Build.Version),
		)
	}
	if a.Tenant != ref.Key.Tenant || a.Surface != ref.Key.Surface || a.Segment != ref.Key.Segment {
		return apperr.New(
			apperr.KindValidationFailed,
			"artifact key mismatch",
			fmt.Errorf("ref=%s/%s/%s", ref.Key.Tenant, ref.Key.Surface, ref.Key.Segment),
		)
	}
	aw, err := parseWindow(a.Window.Start, a.Window.End)
	if err != nil {
		return err
	}
	if !windowsEqualUTC(aw, ref.Window) {
		return apperr.New(apperr.KindValidationFailed, "artifact window mismatch", nil)
	}
	if v.opts.MaxPopularityItems > 0 && len(a.Items) > v.opts.MaxPopularityItems {
		return apperr.New(apperr.KindValidationFailed, "popularity items exceeds max", fmt.Errorf("items=%d max=%d", len(a.Items), v.opts.MaxPopularityItems))
	}
	for _, it := range a.Items {
		if it.ItemID == "" {
			return apperr.New(apperr.KindValidationFailed, "popularity item_id is empty", nil)
		}
		if it.Count < 0 {
			return apperr.New(apperr.KindValidationFailed, "popularity item count negative", fmt.Errorf("item=%s", it.ItemID))
		}
	}
	ver, err := recomputePopularityVersion(a)
	if err != nil {
		return err
	}
	if ver != ref.Version {
		return apperr.New(
			apperr.KindValidationFailed,
			"popularity version hash mismatch",
			fmt.Errorf("computed=%s payload=%s", ver, ref.Version),
		)
	}
	return nil
}

func (v *Builtin) validateCooc(ref artifacts.Ref, b []byte) error {
	var a artifacts.CoocArtifactV1
	if err := json.Unmarshal(b, &a); err != nil {
		return apperr.New(apperr.KindValidationFailed, "invalid cooc artifact json", err)
	}
	if a.V != 1 {
		return apperr.New(apperr.KindValidationFailed, "unsupported cooc artifact version", fmt.Errorf("v=%d", a.V))
	}
	if a.ArtifactType != string(artifacts.TypeCooc) {
		return apperr.New(apperr.KindValidationFailed, "cooc artifact_type mismatch", fmt.Errorf("type=%s", a.ArtifactType))
	}
	if a.Build.Version != ref.Version {
		return apperr.New(apperr.KindValidationFailed, "artifact version mismatch", fmt.Errorf("ref=%s payload=%s", ref.Version, a.Build.Version))
	}
	if a.Tenant != ref.Key.Tenant || a.Surface != ref.Key.Surface || a.Segment != ref.Key.Segment {
		return apperr.New(apperr.KindValidationFailed, "artifact key mismatch", nil)
	}
	aw, err := parseWindow(a.Window.Start, a.Window.End)
	if err != nil {
		return err
	}
	if !windowsEqualUTC(aw, ref.Window) {
		return apperr.New(apperr.KindValidationFailed, "artifact window mismatch", nil)
	}
	if v.opts.MaxCoocRows > 0 && len(a.Neighbors) > v.opts.MaxCoocRows {
		return apperr.New(apperr.KindValidationFailed, "cooc rows exceeds max", fmt.Errorf("rows=%d max=%d", len(a.Neighbors), v.opts.MaxCoocRows))
	}
	for _, r := range a.Neighbors {
		if r.ItemID == "" {
			return apperr.New(apperr.KindValidationFailed, "cooc row item_id is empty", nil)
		}
		if v.opts.MaxCoocNeighbors > 0 && len(r.Items) > v.opts.MaxCoocNeighbors {
			return apperr.New(apperr.KindValidationFailed, "cooc neighbors exceeds max", fmt.Errorf("item=%s neighbors=%d max=%d", r.ItemID, len(r.Items), v.opts.MaxCoocNeighbors))
		}
		for _, n := range r.Items {
			if n.ItemID == "" {
				return apperr.New(apperr.KindValidationFailed, "cooc neighbor item_id is empty", fmt.Errorf("item=%s", r.ItemID))
			}
			if n.Count < 0 {
				return apperr.New(apperr.KindValidationFailed, "cooc neighbor count negative", fmt.Errorf("item=%s neighbor=%s", r.ItemID, n.ItemID))
			}
		}
	}
	ver, err := recomputeCoocVersion(a)
	if err != nil {
		return err
	}
	if ver != ref.Version {
		return apperr.New(apperr.KindValidationFailed, "cooc version hash mismatch", fmt.Errorf("computed=%s payload=%s", ver, ref.Version))
	}
	return nil
}
func recomputePopularityVersion(a artifacts.PopularityArtifactV1) (string, error) {
	a.Build = artifacts.BuildInfo{}
	b, err := json.Marshal(a)
	if err != nil {
		return "", apperr.New(apperr.KindValidationFailed, "marshal popularity for version", err)
	}
	return lineage.SHA256Hex(b), nil
}
func recomputeCoocVersion(a artifacts.CoocArtifactV1) (string, error) {
	a.Build = artifacts.BuildInfo{}
	b, err := json.Marshal(a)
	if err != nil {
		return "", apperr.New(apperr.KindValidationFailed, "marshal cooc for version", err)
	}
	return lineage.SHA256Hex(b), nil
}

func parseWindow(startStr, endStr string) (windows.Window, error) {
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return windows.Window{}, apperr.New(
			apperr.KindValidationFailed,
			"invalid window.start",
			err,
		)
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		return windows.Window{}, apperr.New(
			apperr.KindValidationFailed,
			"invalid window.end",
			err,
		)
	}
	win := windows.Window{Start: start.UTC(), End: end.UTC()}
	if err := win.Validate(); err != nil {
		return windows.Window{}, apperr.New(
			apperr.KindValidationFailed,
			"invalid window",
			err,
		)
	}
	return win, nil
}

func windowsEqualUTC(a, b windows.Window) bool {
	return a.Start.UTC().Equal(b.Start.UTC()) && a.End.UTC().Equal(b.End.UTC())
}
