package signals

import "time"

// ItemTag describes tag metadata to store in the signal tables.
type ItemTag struct {
	ItemID    string
	Tags      []string
	Price     *float64
	CreatedAt time.Time
	Namespace string
}

// PopularityItem captures a daily popularity score for an item.
type PopularityItem struct {
	ItemID string
	Score  float64
}

// CooccurrenceItem captures a daily co-visitation score for an anchor->neighbor pair.
type CooccurrenceItem struct {
	ItemID     string
	NeighborID string
	Score      float64
}
