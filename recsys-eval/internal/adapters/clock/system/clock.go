package system

import "time"

// Clock uses the system time.
type Clock struct{}

func (Clock) Now() time.Time {
	return time.Now().UTC()
}
