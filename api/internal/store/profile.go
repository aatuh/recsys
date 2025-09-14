package store

import (
	"context"
	"time"

	_ "embed"

	"github.com/google/uuid"
)

//go:embed queries/user_tag_profile.sql
var userTagProfileSQL string

// BuildUserTagProfile builds a lightweight user profile as decayed, weighted
// tag preferences from the user's own events. It returns a normalized map
// "tag -> weight" whose values sum to 1 (unless empty).
//
// Notes:
// - Uses tenant effective event-type config (weight, half-life) like popularity.
// - Optional "since" via windowDays; pass <=0 to not limit by time window.
// - Limits to topN strongest tags server-side for performance.
//
// This is intentionally simple, no training or user factors: just decayed
// counts on tags of items the user interacted with.
func (s *Store) BuildUserTagProfile(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	userID string,
	windowDays float64,
	topN int,
) (map[string]float64, error) {
	if userID == "" || topN <= 0 {
		return map[string]float64{}, nil
	}

	// Compute optional since timestamp if windowDays > 0.
	var since *time.Time
	if windowDays > 0 {
		d := time.Duration(windowDays*24.0) * time.Hour
		t := time.Now().UTC().Add(-d)
		since = &t
	}

	rows, err := s.Pool.Query(ctx, userTagProfileSQL, orgID, ns, userID, since, topN)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	raw := map[string]float64{}
	total := 0.0
	for rows.Next() {
		var tag string
		var score float64
		if err := rows.Scan(&tag, &score); err != nil {
			return nil, err
		}
		// Guard against negatives or NaN (shouldn't happen).
		if score > 0 {
			raw[tag] = score
			total += score
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Normalize to sum to 1 for a stable dot-product scale.
	if total <= 0 {
		return map[string]float64{}, nil
	}
	out := make(map[string]float64, len(raw))
	for k, v := range raw {
		out[k] = v / total
	}
	return out, nil
}
