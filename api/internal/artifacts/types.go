package artifacts

import (
	"errors"
	"fmt"
	"time"
)

const (
	TypePopularity = "popularity"
	TypeCooc       = "cooc"
)

type ManifestV1 struct {
	V         int               `json:"v"`
	Tenant    string            `json:"tenant"`
	Surface   string            `json:"surface"`
	Current   map[string]string `json:"current"`
	UpdatedAt string            `json:"updated_at"`
}

func (m ManifestV1) Validate() error {
	if m.V != 1 {
		return fmt.Errorf("unsupported manifest version: %d", m.V)
	}
	if m.Tenant == "" || m.Surface == "" {
		return errors.New("manifest tenant and surface required")
	}
	return nil
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

type windowJSON struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

func ParseTime(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
