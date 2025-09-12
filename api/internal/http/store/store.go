package store

import (
	"context"
	"fmt"
	"recsys/internal/types"
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

// Time-decayed popularity with fixed type weights.
func (s *Store) PopularityTopK(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	halfLifeDays float64,
	k int,
	c *types.PopConstraints,
) ([]types.ScoredItem, error) {
	if k <= 0 {
		k = 20
	}
	rows, err := s.Pool.Query(ctx, popularitySQL,
		orgID, ns, halfLifeDays, k,
		// $5..$10:
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
		func() any {
			return time.Now()
		}(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]types.ScoredItem, 0, k)
	for rows.Next() {
		var it types.ScoredItem
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
) ([]types.ScoredItem, error) {
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
	out := make([]types.ScoredItem, 0, k)
	for rows.Next() {
		var it types.ScoredItem
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

// ListItemsMeta returns tags for the given item IDs.
func (s *Store) ListItemsMeta(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	itemIDs []string,
) (map[string]types.ItemMeta, error) {
	if len(itemIDs) == 0 {
		return map[string]types.ItemMeta{}, nil
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

	out := make(map[string]types.ItemMeta, len(itemIDs))
	for rows.Next() {
		var id string
		var tags []string
		if err := rows.Scan(&id, &tags); err != nil {
			return nil, err
		}
		out[id] = types.ItemMeta{ItemID: id, Tags: tags}
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

// List and Delete methods

// ListUsers returns a paginated list of users with optional filtering
func (s *Store) ListUsers(ctx context.Context, orgID uuid.UUID, ns string, limit, offset int, filters map[string]interface{}) ([]map[string]interface{}, int, error) {
	// Build WHERE clause
	whereClause := "WHERE org_id = $1 AND namespace = $2"
	args := []interface{}{orgID, ns}
	argIndex := 3

	// Add filters
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		whereClause += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if createdAfter, ok := filters["created_after"].(string); ok && createdAfter != "" {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, createdAfter)
		argIndex++
	}

	if createdBefore, ok := filters["created_before"].(string); ok && createdBefore != "" {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, createdBefore)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var total int
	err := s.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT user_id, traits, created_at, updated_at
		FROM users %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var userID string
		var traits []byte
		var createdAt, updatedAt time.Time

		err := rows.Scan(&userID, &traits, &createdAt, &updatedAt)
		if err != nil {
			return nil, 0, err
		}

		user := map[string]interface{}{
			"user_id":    userID,
			"traits":     string(traits),
			"created_at": createdAt.Format(time.RFC3339),
			"updated_at": updatedAt.Format(time.RFC3339),
		}
		users = append(users, user)
	}

	return users, total, rows.Err()
}

// ListItems returns a paginated list of items with optional filtering
func (s *Store) ListItems(ctx context.Context, orgID uuid.UUID, ns string, limit, offset int, filters map[string]interface{}) ([]map[string]interface{}, int, error) {
	// Build WHERE clause
	whereClause := "WHERE org_id = $1 AND namespace = $2"
	args := []interface{}{orgID, ns}
	argIndex := 3

	// Add filters
	if itemID, ok := filters["item_id"].(string); ok && itemID != "" {
		whereClause += fmt.Sprintf(" AND item_id = $%d", argIndex)
		args = append(args, itemID)
		argIndex++
	}

	if createdAfter, ok := filters["created_after"].(string); ok && createdAfter != "" {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, createdAfter)
		argIndex++
	}

	if createdBefore, ok := filters["created_before"].(string); ok && createdBefore != "" {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, createdBefore)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM items %s", whereClause)
	var total int
	err := s.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT item_id, available, price, tags, props, created_at, updated_at
		FROM items %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var itemID string
		var available bool
		var price *float64
		var tags []string
		var props []byte
		var createdAt, updatedAt time.Time

		err := rows.Scan(&itemID, &available, &price, &tags, &props, &createdAt, &updatedAt)
		if err != nil {
			return nil, 0, err
		}

		item := map[string]interface{}{
			"item_id":    itemID,
			"available":  available,
			"price":      price,
			"tags":       tags,
			"props":      string(props),
			"created_at": createdAt.Format(time.RFC3339),
			"updated_at": updatedAt.Format(time.RFC3339),
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

// ListEvents returns a paginated list of events with optional filtering
func (s *Store) ListEvents(ctx context.Context, orgID uuid.UUID, ns string, limit, offset int, filters map[string]interface{}) ([]map[string]interface{}, int, error) {
	// Build WHERE clause
	whereClause := "WHERE org_id = $1 AND namespace = $2"
	args := []interface{}{orgID, ns}
	argIndex := 3

	// Add filters
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		whereClause += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if itemID, ok := filters["item_id"].(string); ok && itemID != "" {
		whereClause += fmt.Sprintf(" AND item_id = $%d", argIndex)
		args = append(args, itemID)
		argIndex++
	}

	if eventType, ok := filters["event_type"].(int16); ok {
		whereClause += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, eventType)
		argIndex++
	}

	if createdAfter, ok := filters["created_after"].(string); ok && createdAfter != "" {
		whereClause += fmt.Sprintf(" AND ts >= $%d", argIndex)
		args = append(args, createdAfter)
		argIndex++
	}

	if createdBefore, ok := filters["created_before"].(string); ok && createdBefore != "" {
		whereClause += fmt.Sprintf(" AND ts <= $%d", argIndex)
		args = append(args, createdBefore)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM events %s", whereClause)
	var total int
	err := s.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT user_id, item_id, type, value, ts, meta, source_event_id
		FROM events %s
		ORDER BY ts DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var userID, itemID string
		var eventType int16
		var value float64
		var ts time.Time
		var meta []byte
		var sourceEventID *string

		err := rows.Scan(&userID, &itemID, &eventType, &value, &ts, &meta, &sourceEventID)
		if err != nil {
			return nil, 0, err
		}

		event := map[string]interface{}{
			"user_id":         userID,
			"item_id":         itemID,
			"type":            eventType,
			"value":           value,
			"ts":              ts.Format(time.RFC3339),
			"meta":            string(meta),
			"source_event_id": sourceEventID,
		}
		events = append(events, event)
	}

	return events, total, rows.Err()
}

// DeleteUsers deletes users based on filters
func (s *Store) DeleteUsers(ctx context.Context, orgID uuid.UUID, ns string, filters map[string]interface{}) (int, error) {
	// Build WHERE clause
	whereClause := "WHERE org_id = $1 AND namespace = $2"
	args := []interface{}{orgID, ns}
	argIndex := 3

	// Add filters
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		whereClause += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if createdAfter, ok := filters["created_after"].(string); ok && createdAfter != "" {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, createdAfter)
		argIndex++
	}

	if createdBefore, ok := filters["created_before"].(string); ok && createdBefore != "" {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, createdBefore)
		argIndex++
	}

	query := fmt.Sprintf("DELETE FROM users %s", whereClause)
	result, err := s.Pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return int(result.RowsAffected()), nil
}

// DeleteItems deletes items based on filters
func (s *Store) DeleteItems(ctx context.Context, orgID uuid.UUID, ns string, filters map[string]interface{}) (int, error) {
	// Build WHERE clause
	whereClause := "WHERE org_id = $1 AND namespace = $2"
	args := []interface{}{orgID, ns}
	argIndex := 3

	// Add filters
	if itemID, ok := filters["item_id"].(string); ok && itemID != "" {
		whereClause += fmt.Sprintf(" AND item_id = $%d", argIndex)
		args = append(args, itemID)
		argIndex++
	}

	if createdAfter, ok := filters["created_after"].(string); ok && createdAfter != "" {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, createdAfter)
		argIndex++
	}

	if createdBefore, ok := filters["created_before"].(string); ok && createdBefore != "" {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, createdBefore)
		argIndex++
	}

	query := fmt.Sprintf("DELETE FROM items %s", whereClause)
	result, err := s.Pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return int(result.RowsAffected()), nil
}

// DeleteEvents deletes events based on filters
func (s *Store) DeleteEvents(ctx context.Context, orgID uuid.UUID, ns string, filters map[string]interface{}) (int, error) {
	// Build WHERE clause
	whereClause := "WHERE org_id = $1 AND namespace = $2"
	args := []interface{}{orgID, ns}
	argIndex := 3

	// Add filters
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		whereClause += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if itemID, ok := filters["item_id"].(string); ok && itemID != "" {
		whereClause += fmt.Sprintf(" AND item_id = $%d", argIndex)
		args = append(args, itemID)
		argIndex++
	}

	if eventType, ok := filters["event_type"].(int16); ok {
		whereClause += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, eventType)
		argIndex++
	}

	if createdAfter, ok := filters["created_after"].(string); ok && createdAfter != "" {
		whereClause += fmt.Sprintf(" AND ts >= $%d", argIndex)
		args = append(args, createdAfter)
		argIndex++
	}

	if createdBefore, ok := filters["created_before"].(string); ok && createdBefore != "" {
		whereClause += fmt.Sprintf(" AND ts <= $%d", argIndex)
		args = append(args, createdBefore)
		argIndex++
	}

	query := fmt.Sprintf("DELETE FROM events %s", whereClause)
	result, err := s.Pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return int(result.RowsAffected()), nil
}
