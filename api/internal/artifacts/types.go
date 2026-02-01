package artifacts

import (
	"errors"
	"fmt"
	"time"
)

const (
	TypePopularity = "popularity"
	TypeCooc       = "cooc"
	TypeImplicit   = "implicit"
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
	if m.UpdatedAt != "" {
		if _, ok := ParseTime(m.UpdatedAt); !ok {
			return errors.New("manifest updated_at must be RFC3339")
		}
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

func (a PopularityArtifactV1) Validate() error {
	if a.V != 1 {
		return fmt.Errorf("unsupported popularity artifact version: %d", a.V)
	}
	if a.ArtifactType != TypePopularity {
		return fmt.Errorf("popularity artifact_type mismatch: %s", a.ArtifactType)
	}
	if a.Tenant == "" || a.Surface == "" {
		return errors.New("popularity artifact tenant and surface required")
	}
	if _, ok := ParseTime(a.Window.Start); !ok {
		return errors.New("popularity artifact window.start must be RFC3339")
	}
	if _, ok := ParseTime(a.Window.End); !ok {
		return errors.New("popularity artifact window.end must be RFC3339")
	}
	return nil
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

func (a CoocArtifactV1) Validate() error {
	if a.V != 1 {
		return fmt.Errorf("unsupported cooc artifact version: %d", a.V)
	}
	if a.ArtifactType != TypeCooc {
		return fmt.Errorf("cooc artifact_type mismatch: %s", a.ArtifactType)
	}
	if a.Tenant == "" || a.Surface == "" {
		return errors.New("cooc artifact tenant and surface required")
	}
	if _, ok := ParseTime(a.Window.Start); !ok {
		return errors.New("cooc artifact window.start must be RFC3339")
	}
	if _, ok := ParseTime(a.Window.End); !ok {
		return errors.New("cooc artifact window.end must be RFC3339")
	}
	return nil
}

type ImplicitItem struct {
	ItemID string  `json:"item_id"`
	Score  float64 `json:"score"`
}

type ImplicitUser struct {
	UserID string         `json:"user_id"`
	Items  []ImplicitItem `json:"items"`
}

// ImplicitArtifactV1 stores user-item recommendation scores from implicit feedback.
type ImplicitArtifactV1 struct {
	V            int            `json:"v"`
	ArtifactType string         `json:"artifact_type"`
	Tenant       string         `json:"tenant"`
	Surface      string         `json:"surface"`
	Segment      string         `json:"segment,omitempty"`
	Window       windowJSON     `json:"window"`
	Users        []ImplicitUser `json:"users"`
	Build        BuildInfo      `json:"build"`
}

func (a ImplicitArtifactV1) Validate() error {
	if a.V != 1 {
		return fmt.Errorf("unsupported implicit artifact version: %d", a.V)
	}
	if a.ArtifactType != TypeImplicit {
		return fmt.Errorf("implicit artifact_type mismatch: %s", a.ArtifactType)
	}
	if a.Tenant == "" || a.Surface == "" {
		return errors.New("implicit artifact tenant and surface required")
	}
	if _, ok := ParseTime(a.Window.Start); !ok {
		return errors.New("implicit artifact window.start must be RFC3339")
	}
	if _, ok := ParseTime(a.Window.End); !ok {
		return errors.New("implicit artifact window.end must be RFC3339")
	}
	return nil
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
