package artifacts

import (
	"fmt"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

type Type string

const (
	TypePopularity Type = "popularity"
	TypeCooc       Type = "cooc"
	TypeImplicit   Type = "implicit"
)

type Key struct {
	Tenant  string
	Surface string
	Segment string
	Type    Type
}

func (k Key) Validate() error {
	if strings.TrimSpace(k.Tenant) == "" {
		return fmt.Errorf("tenant must be set")
	}
	if strings.TrimSpace(k.Surface) == "" {
		return fmt.Errorf("surface must be set")
	}
	if k.Type == "" {
		return fmt.Errorf("type must be set")
	}
	return nil
}

type Ref struct {
	Key     Key
	Window  windows.Window
	Version string
	URI     string
	BuiltAt time.Time
}

func (r Ref) Validate() error {
	if err := r.Key.Validate(); err != nil {
		return err
	}
	if err := r.Window.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(r.Version) == "" {
		return fmt.Errorf("version must be set")
	}
	return nil
}
