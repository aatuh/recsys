package types

import "time"

// UserRecord represents a stored user and their traits.
type UserRecord struct {
	UserID    string
	Traits    map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}
