package systemclock

import "time"

type SystemClock struct{}

func (SystemClock) NowUTC() time.Time {
	return time.Now().UTC()
}
