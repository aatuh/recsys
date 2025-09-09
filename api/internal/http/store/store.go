package store

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Standard embedding dimension. Keep in sync with migration and handler.
const EmbeddingDims = 384

type Store struct{ Pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{Pool: pool} }

type ItemUpsert struct {
	ItemID    string
	Available bool
	Price     *float64
	Tags      []string
	Props     any
	// Optional embedding. If provided, stored in items.embedding (pgvector).
	Embedding *[]float64
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
	SourceEventID  *string
}

type EventTypeConfig struct {
	Type         int16
	Name         *string
	Weight       float64
	HalfLifeDays *float64
	IsActive     *bool
}

func (s *Store) UpsertItems(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	items []ItemUpsert,
) error {
	if len(items) == 0 {
		return nil
	}
	bat := &pgx.Batch{}
	for _, it := range items {
		// Prepare optional vector literal like: "[0.1,0.2,...]"
		var embText *string
		if it.Embedding != nil && len(*it.Embedding) > 0 {
			t := vectorLiteral(*it.Embedding)
			embText = &t
		}

		bat.Queue(`
		INSERT INTO items (
		  org_id, namespace, item_id, available, price, tags, props, embedding,
		  created_at, updated_at
		)
		VALUES (
		  $1,$2,$3,$4,
		  $5,
		  COALESCE($6, '{}'::text[]),
		  COALESCE($7, '{}'::jsonb),
		  CASE WHEN $8::text IS NULL
		       THEN NULL
		       ELSE CAST($8 AS vector(`+strconv.Itoa(EmbeddingDims)+`))
		  END,
		  now(), now()
		)
		ON CONFLICT (org_id, namespace, item_id) DO UPDATE SET
		  available = EXCLUDED.available,
		  price     = COALESCE(EXCLUDED.price, items.price),
		  tags      = COALESCE(EXCLUDED.tags,  items.tags),
		  props     = COALESCE(EXCLUDED.props, items.props),
		  embedding = CASE WHEN EXCLUDED.embedding IS NULL
		                  THEN items.embedding
		                  ELSE EXCLUDED.embedding
		             END,
		  updated_at= now()
	  `, orgID, ns, it.ItemID, it.Available, it.Price, it.Tags, it.Props, embText)
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
		INSERT INTO events (org_id, namespace, user_id, item_id, type, value, ts, meta, source_event_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7, COALESCE($8, '{}'::jsonb), $9)
		ON CONFLICT (org_id, namespace, source_event_id) DO NOTHING
	  `, orgID, ns, e.UserID, e.ItemID, e.Type, e.Value, e.TS, e.Meta, e.SourceEventID)

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

type PopConstraints struct {
	IncludeTagsAny     []string
	MinPrice, MaxPrice *float64
	CreatedAfter       *time.Time
	ExcludeItemIDs     []string
}

type ScoredItem struct {
	ItemID string
	Score  float64
}

// Time-decayed popularity with fixed type weights.
func (s *Store) PopularityTopK(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	halfLifeDays float64,
	k int,
	c *PopConstraints,
) ([]ScoredItem, error) {
	if k <= 0 {
		k = 20
	}
	rows, err := s.Pool.Query(ctx, popularitySQL,
		orgID, ns, halfLifeDays, k,
		// $5..$9:
		func() any {
			if c != nil && c.CreatedAfter != nil {
				return *c.CreatedAfter
			}
			return nil
		}(),
		func() any {
			if c != nil && c.MinPrice != nil {
				return *c.MinPrice
			}
			return nil
		}(),
		func() any {
			if c != nil && c.MaxPrice != nil {
				return *c.MaxPrice
			}
			return nil
		}(),
		func() any {
			if c != nil && len(c.IncludeTagsAny) > 0 {
				return c.IncludeTagsAny
			}
			return []string{}
		}(),
		func() any {
			if c != nil && len(c.ExcludeItemIDs) > 0 {
				return c.ExcludeItemIDs
			}
			return []string{}
		}(),
	)
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

// CooccurrenceTopKWithin returns co-vis neighbors since a cutoff timestamp.
func (s *Store) CooccurrenceTopKWithin(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	itemID string,
	k int,
	since time.Time,
) ([]ScoredItem, error) {
	if k <= 0 {
		k = 20
	}
	rows, err := s.Pool.Query(ctx, `
SELECT e2.item_id, COUNT(*)::float8 AS c
FROM events e1
JOIN events e2
  ON e1.org_id = e2.org_id
 AND e1.namespace = e2.namespace
 AND e1.user_id = e2.user_id
 AND e2.item_id <> $3
WHERE e1.org_id = $1
  AND e1.namespace = $2
  AND e1.item_id = $3
  AND e1.ts > $5
  AND e2.ts > $5
GROUP BY e2.item_id
ORDER BY c DESC
LIMIT $4
`, orgID, ns, itemID, k, since)
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

// Upsert tenant overrides (batch).
func (s *Store) UpsertEventTypeConfig(ctx context.Context, orgID uuid.UUID, ns string, rows []EventTypeConfig) error {
	if len(rows) == 0 {
		return nil
	}
	bat := &pgx.Batch{}
	for _, r := range rows {
		bat.Queue(`
		INSERT INTO event_type_config (org_id, namespace, type, name, weight, half_life_days, is_active, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,COALESCE($7,true), now())
		ON CONFLICT (org_id, namespace, type) DO UPDATE SET
		  name          = COALESCE(EXCLUDED.name, event_type_config.name),
		  weight        = EXCLUDED.weight,
		  half_life_days= EXCLUDED.half_life_days,
		  is_active     = COALESCE(EXCLUDED.is_active, event_type_config.is_active),
		  updated_at    = now()
	  `, orgID, ns, r.Type, r.Name, r.Weight, r.HalfLifeDays, r.IsActive)
	}
	br := s.Pool.SendBatch(ctx, bat)
	defer br.Close()
	for range rows {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

type EventTypeConfigRow struct {
	Type         int16
	Name         *string
	Weight       float64
	HalfLifeDays *float64
	IsActive     bool
	Source       string // "tenant" or "default"
}

// Effective view (tenant override if exists; else default).
func (s *Store) ListEventTypeConfigEffective(ctx context.Context, orgID uuid.UUID, ns string) ([]EventTypeConfigRow, error) {
	rows, err := s.Pool.Query(ctx, `
	  SELECT COALESCE(tc.type, d.type) AS type,
			 COALESCE(tc.name, d.name) AS name,
			 COALESCE(tc.weight, d.weight) AS weight,
			 COALESCE(tc.half_life_days, d.half_life_days) AS half_life_days,
			 COALESCE(tc.is_active, true) AS is_active,
			 CASE WHEN tc.type IS NULL THEN 'default' ELSE 'tenant' END AS source
	  FROM event_type_defaults d
	  FULL OUTER JOIN event_type_config tc
		ON tc.org_id=$1 AND tc.namespace=$2 AND tc.type=d.type
	  WHERE COALESCE(tc.is_active, true)=true
	  ORDER BY type
	`, orgID, ns)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EventTypeConfigRow{}
	for rows.Next() {
		var r EventTypeConfigRow
		if err := rows.Scan(&r.Type, &r.Name, &r.Weight, &r.HalfLifeDays, &r.IsActive, &r.Source); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) EnsureEventTypeDefaults(ctx context.Context) error {
	_, err := s.Pool.Exec(ctx, `
    INSERT INTO event_type_defaults(type, name, weight, half_life_days) VALUES
      (0,'view',0.1,NULL),(1,'click',0.3,NULL),(2,'add',0.7,NULL),(3,'purchase',1.0,NULL),(4,'custom',0.2,NULL)
    ON CONFLICT (type) DO UPDATE
      SET name=EXCLUDED.name,
          weight=EXCLUDED.weight,
          half_life_days=EXCLUDED.half_life_days;
  `)
	return err
}

// ItemMeta holds lightweight metadata required for diversity and caps.
type ItemMeta struct {
	ItemID string
	Tags   []string
}

// ListItemsMeta returns tags for the given item IDs.
func (s *Store) ListItemsMeta(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	itemIDs []string,
) (map[string]ItemMeta, error) {
	if len(itemIDs) == 0 {
		return map[string]ItemMeta{}, nil
	}
	rows, err := s.Pool.Query(ctx, `
SELECT item_id, tags
FROM items
WHERE org_id = $1
  AND namespace = $2
  AND item_id = ANY($3::text[])
`, orgID, ns, itemIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]ItemMeta, len(itemIDs))
	for rows.Next() {
		var id string
		var tags []string
		if err := rows.Scan(&id, &tags); err != nil {
			return nil, err
		}
		out[id] = ItemMeta{ItemID: id, Tags: tags}
	}
	return out, rows.Err()
}

// ListUserPurchasedSince returns distinct item IDs the user purchased on/after
// the given timestamp. We consider type=3 as "purchase".
func (s *Store) ListUserPurchasedSince(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	userID string,
	since time.Time,
) ([]string, error) {
	rows, err := s.Pool.Query(ctx, `
SELECT DISTINCT item_id
FROM events
WHERE org_id = $1
  AND namespace = $2
  AND user_id = $3
  AND type = 3
  AND ts >= $4
`, orgID, ns, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// vectorLiteral formats floats as a pgvector textual literal: "[x,y,...]".
// Uses %g to keep things compact while remaining precise enough for ranking.
func vectorLiteral(v []float64) string {
	sb := strings.Builder{}
	sb.Grow(2 + len(v)*10)
	sb.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
	}
	sb.WriteByte(']')
	return sb.String()
}
