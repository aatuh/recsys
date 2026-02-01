package clock

import "time"

type Clock interface {
	NowUTC() time.Time
}
