package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

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

	rows, err := s.Pool.Query(ctx, `
WITH w AS (
  SELECT
    COALESCE(tc.type, d.type) AS type,
    COALESCE(tc.weight, d.weight) AS w,
    COALESCE(
      NULLIF(tc.half_life_days, 0),
      d.half_life_days,
      14
    ) AS hl
  FROM event_type_defaults d
  FULL OUTER JOIN event_type_config tc
    ON tc.org_id = $1 AND tc.namespace = $2 AND tc.type = d.type
  WHERE COALESCE(tc.is_active, true) = true
)
SELECT t.tag,
       SUM(
         POWER(0.5, EXTRACT(EPOCH FROM (now() - e.ts))
               / (NULLIF(w.hl,0) * 86400.0))
         * w.w * COALESCE(e.value, 1)
       ) AS s
FROM events e
JOIN w ON w.type = e.type
JOIN items i
  ON i.org_id = e.org_id
 AND i.namespace = e.namespace
 AND i.item_id = e.item_id
CROSS JOIN LATERAL UNNEST(i.tags) AS t(tag)
WHERE e.org_id = $1
  AND e.namespace = $2
  AND e.user_id = $3
  AND ($4::timestamptz IS NULL OR e.ts >= $4)
GROUP BY t.tag
ORDER BY s DESC
LIMIT $5
`, orgID, ns, userID, since, topN)
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
