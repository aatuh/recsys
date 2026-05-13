package artifacts

import (
	"fmt"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/pathsafe"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
)

type Type string

const (
	TypePopularity Type = "popularity"
	TypeCooc       Type = "cooc"
	TypeImplicit   Type = "implicit"
	TypeContentSim Type = "content_sim"
	TypeSessionSeq Type = "session_seq"
)

type Key struct {
	Tenant  string
	Surface string
	Segment string
	Type    Type
}

func (k Key) Validate() error {
	if _, err := pathsafe.Segment("tenant", k.Tenant); err != nil {
		return err
	}
	if _, err := pathsafe.Segment("surface", k.Surface); err != nil {
		return err
	}
	if _, err := pathsafe.Segment("segment", k.Segment); err != nil {
		return err
	}
	if k.Type == "" {
		return fmt.Errorf("type must be set")
	}
	if _, err := pathsafe.Segment("artifact type", string(k.Type)); err != nil {
		return err
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
	if _, err := pathsafe.Segment("version", r.Version); err != nil {
		return err
	}
	return nil
}
