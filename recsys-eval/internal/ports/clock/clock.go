package clock

import "time"

// Clock provides time for deterministic runs.
type Clock interface {
	Now() time.Time
}
