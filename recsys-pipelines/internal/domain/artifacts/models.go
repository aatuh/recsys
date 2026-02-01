package artifacts

import (
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

type windowJSON struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type BuildInfo struct {
	BuiltAt    string `json:"built_at"`
	Version    string `json:"version"`
	SourceHash string `json:"source_hash,omitempty"`
}

type PopularityItem struct {
	ItemID string `json:"item_id"`
	Count  int64  `json:"count"`
}

type PopularityArtifactV1 struct {
	V            int              `json:"v"`
	ArtifactType string           `json:"artifact_type"`
	Tenant       string           `json:"tenant"`
	Surface      string           `json:"surface"`
	Segment      string           `json:"segment,omitempty"`
	Window       windowJSON       `json:"window"`
	Items        []PopularityItem `json:"items"`
	Build        BuildInfo        `json:"build"`
}

func NewPopularityArtifact(
	tenant, surface, segment string,
	w windows.Window,
	items []PopularityItem,
	builtAt time.Time,
	version string,
	sourceHash string,
) PopularityArtifactV1 {
	return PopularityArtifactV1{
		V:            1,
		ArtifactType: string(TypePopularity),
		Tenant:       tenant,
		Surface:      surface,
		Segment:      segment,
		Window: windowJSON{
			Start: w.Start.UTC().Format(time.RFC3339),
			End:   w.End.UTC().Format(time.RFC3339),
		},
		Items: items,
		Build: BuildInfo{
			BuiltAt:    builtAt.UTC().Format(time.RFC3339),
			Version:    version,
			SourceHash: sourceHash,
		},
	}
}

type CoocNeighbor struct {
	ItemID string `json:"item_id"`
	Count  int64  `json:"count"`
}

type CoocRow struct {
	ItemID string         `json:"item_id"`
	Items  []CoocNeighbor `json:"items"`
}

type CoocArtifactV1 struct {
	V            int        `json:"v"`
	ArtifactType string     `json:"artifact_type"`
	Tenant       string     `json:"tenant"`
	Surface      string     `json:"surface"`
	Segment      string     `json:"segment,omitempty"`
	Window       windowJSON `json:"window"`
	Neighbors    []CoocRow  `json:"neighbors"`
	Build        BuildInfo  `json:"build"`
}

func NewCoocArtifact(
	tenant, surface, segment string,
	w windows.Window,
	rows []CoocRow,
	builtAt time.Time,
	version string,
	sourceHash string,
) CoocArtifactV1 {
	return CoocArtifactV1{
		V:            1,
		ArtifactType: string(TypeCooc),
		Tenant:       tenant,
		Surface:      surface,
		Segment:      segment,
		Window: windowJSON{
			Start: w.Start.UTC().Format(time.RFC3339),
			End:   w.End.UTC().Format(time.RFC3339),
		},
		Neighbors: rows,
		Build: BuildInfo{
			BuiltAt:    builtAt.UTC().Format(time.RFC3339),
			Version:    version,
			SourceHash: sourceHash,
		},
	}
}

type ManifestV1 struct {
	V         int               `json:"v"`
	Tenant    string            `json:"tenant"`
	Surface   string            `json:"surface"`
	Current   map[string]string `json:"current"`
	UpdatedAt string            `json:"updated_at"`
}

func NewManifest(tenant, surface string, current map[string]string, now time.Time) ManifestV1 {
	cpy := make(map[string]string, len(current))
	for k, v := range current {
		cpy[k] = v
	}
	return ManifestV1{
		V:         1,
		Tenant:    tenant,
		Surface:   surface,
		Current:   cpy,
		UpdatedAt: now.UTC().Format(time.RFC3339),
	}
}
