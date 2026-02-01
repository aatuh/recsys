package events

import "fmt"

func (e ExposureEvent) Validate() error {
	if e.Version != 1 {
		return fmt.Errorf("unsupported version: %d", e.Version)
	}
	if e.TS.IsZero() {
		return fmt.Errorf("ts must be set")
	}
	if e.Tenant == "" {
		return fmt.Errorf("tenant must be set")
	}
	if e.Surface == "" {
		return fmt.Errorf("surface must be set")
	}
	if e.SessionID == "" {
		return fmt.Errorf("session_id must be set")
	}
	if e.ItemID == "" {
		return fmt.Errorf("item_id must be set")
	}
	if e.Rank < 0 {
		return fmt.Errorf("rank must be >= 0")
	}
	return nil
}
