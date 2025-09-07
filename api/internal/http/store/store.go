package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ Pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{Pool: pool} }

type ItemUpsert struct {
	ItemID    string
	Available bool
	Price     *float64
	Tags      []string
	Props     any
}
type UserUpsert struct {
	UserID string
	Traits any
}
type EventInsert struct {
	UserID, ItemID string
	Type           int16
	Value          float64
	TS             time.Time
	Meta           any
}

func (s *Store) UpsertItems(ctx context.Context, orgID uuid.UUID, ns string, items []ItemUpsert) error {
	if len(items) == 0 {
		return nil
	}
	bat := &pgx.Batch{}
	for _, it := range items {
		bat.Queue(`
		INSERT INTO items (org_id, namespace, item_id, available, price, tags, props, created_at, updated_at)
		VALUES ($1,$2,$3,$4,
				$5,
				COALESCE($6, '{}'::text[]),
				COALESCE($7, '{}'::jsonb),
				now(), now())
		ON CONFLICT (org_id, namespace, item_id) DO UPDATE SET
		  available = EXCLUDED.available,
		  price     = COALESCE(EXCLUDED.price, items.price),
		  tags      = COALESCE(EXCLUDED.tags,  items.tags),
		  props     = COALESCE(EXCLUDED.props, items.props),
		  updated_at= now()
	  `, orgID, ns, it.ItemID, it.Available, it.Price, it.Tags, it.Props)

	}
	br := s.Pool.SendBatch(ctx, bat)
	defer br.Close()
	for range items {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) UpsertUsers(ctx context.Context, orgID uuid.UUID, ns string, users []UserUpsert) error {
	if len(users) == 0 {
		return nil
	}
	bat := &pgx.Batch{}
	for _, u := range users {
		bat.Queue(`
		INSERT INTO users (org_id, namespace, user_id, traits, created_at, updated_at)
		VALUES ($1,$2,$3, COALESCE($4, '{}'::jsonb), now(), now())
		ON CONFLICT (org_id, namespace, user_id) DO UPDATE SET
		  traits    = COALESCE(EXCLUDED.traits, users.traits),
		  updated_at= now()
	  `, orgID, ns, u.UserID, u.Traits)
	}
	br := s.Pool.SendBatch(ctx, bat)
	defer br.Close()
	for range users {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) InsertEvents(ctx context.Context, orgID uuid.UUID, ns string, evs []EventInsert) error {
	if len(evs) == 0 {
		return nil
	}
	bat := &pgx.Batch{}
	for _, e := range evs {
		bat.Queue(`
		INSERT INTO events (org_id, namespace, user_id, item_id, type, value, ts, meta)
		VALUES ($1,$2,$3,$4,$5,$6,$7, COALESCE($8, '{}'::jsonb))
	  `, orgID, ns, e.UserID, e.ItemID, e.Type, e.Value, e.TS, e.Meta)

	}
	br := s.Pool.SendBatch(ctx, bat)
	defer br.Close()
	for range evs {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

type ScoredItem struct {
	ItemID string
	Score  float64
}

// Time-decayed popularity with fixed type weights.
func (s *Store) PopularityTopK(ctx context.Context, orgID uuid.UUID, ns string, halfLifeDays float64, k int) ([]ScoredItem, error) {
	if k <= 0 {
		k = 20
	}
	rows, err := s.Pool.Query(ctx, `
WITH w(type, w) AS (VALUES (0,0.1::float8),(1,0.3),(2,0.7),(3,1.0),(4,0.2))
SELECT e.item_id,
       SUM( EXP( LN(0.5) * ( EXTRACT(EPOCH FROM (now() - e.ts)) / ($3 * 86400.0) ) ) * w.w ) AS score
FROM events e
JOIN w ON w.type = e.type
WHERE e.org_id = $1 AND e.namespace = $2
GROUP BY e.item_id
ORDER BY score DESC
LIMIT $4`, orgID, ns, halfLifeDays, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ScoredItem, 0, k)
	for rows.Next() {
		var it ScoredItem
		if err := rows.Scan(&it.ItemID, &it.Score); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// Naive co-visitation (same-user), last 30 days.
func (s *Store) CooccurrenceTopK(ctx context.Context, orgID uuid.UUID, ns, itemID string, k int) ([]ScoredItem, error) {
	if k <= 0 {
		k = 20
	}
	rows, err := s.Pool.Query(ctx, `
SELECT e2.item_id, COUNT(*)::float8 AS c
FROM events e1
JOIN events e2 ON e1.org_id=e2.org_id AND e1.namespace=e2.namespace AND e1.user_id=e2.user_id AND e2.item_id <> $3
WHERE e1.org_id=$1 AND e1.namespace=$2 AND e1.item_id=$3
  AND e1.ts > now() - interval '30 days'
  AND e2.ts > now() - interval '30 days'
GROUP BY e2.item_id
ORDER BY c DESC
LIMIT $4`, orgID, ns, itemID, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ScoredItem, 0, k)
	for rows.Next() {
		var it ScoredItem
		if err := rows.Scan(&it.ItemID, &it.Score); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}
