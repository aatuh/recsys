package store

import (
	"context"
	"fmt"
	"recsys/internal/types"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Standard embedding dimension. Keep in sync with migration and handler.
const EmbeddingDims = 384

// Options configures Store behaviour.
type Options struct {
	QueryTimeout        time.Duration
	RetryAttempts       int
	RetryInitialBackoff time.Duration
	RetryMaxBackoff     time.Duration
}

// Store wraps the pgx pool with retry and timeout controls.
type Store struct {
	Pool *pgxpool.Pool
	opts Options
}

// New constructs a Store with default options.
func New(pool *pgxpool.Pool) *Store {
	return NewWithOptions(pool, Options{})
}

// NewWithOptions constructs a Store with the provided options merged onto defaults.
func NewWithOptions(pool *pgxpool.Pool, opts Options) *Store {
	defaults := Options{
		QueryTimeout:        5 * time.Second,
		RetryAttempts:       3,
		RetryInitialBackoff: 50 * time.Millisecond,
		RetryMaxBackoff:     500 * time.Millisecond,
	}

	if opts.QueryTimeout > 0 {
		defaults.QueryTimeout = opts.QueryTimeout
	}
	if opts.RetryAttempts > 0 {
		defaults.RetryAttempts = opts.RetryAttempts
	}
	if opts.RetryInitialBackoff > 0 {
		defaults.RetryInitialBackoff = opts.RetryInitialBackoff
	}
	if opts.RetryMaxBackoff > 0 {
		defaults.RetryMaxBackoff = opts.RetryMaxBackoff
	}
	if defaults.RetryAttempts < 1 {
		defaults.RetryAttempts = 1
	}
	if defaults.RetryMaxBackoff < defaults.RetryInitialBackoff {
		defaults.RetryMaxBackoff = defaults.RetryInitialBackoff
	}

	return &Store{
		Pool: pool,
		opts: defaults,
	}
}

func (s *Store) withRetry(ctx context.Context, fn func(context.Context) error) error {
	attempts := s.opts.RetryAttempts
	if attempts < 1 {
		attempts = 1
	}

	initialDelay := s.opts.RetryInitialBackoff
	if initialDelay < 0 {
		initialDelay = 0
	}
	maxDelay := s.opts.RetryMaxBackoff
	if maxDelay < initialDelay {
		maxDelay = initialDelay
	}

	var err error
	delay := initialDelay
	for attempt := 0; attempt < attempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		runCtx := ctx
		var cancel context.CancelFunc
		if s.opts.QueryTimeout > 0 {
			runCtx, cancel = context.WithTimeout(ctx, s.opts.QueryTimeout)
		}

		err = fn(runCtx)

		if cancel != nil {
			cancel()
		}

		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if attempt == attempts-1 || !pgconn.SafeToRetry(err) {
			return err
		}

		sleep := delay
		if maxDelay > 0 && sleep > maxDelay {
			sleep = maxDelay
		}
		if sleep > 0 {
			timer := time.NewTimer(sleep)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			}
		}

		if delay == 0 {
			delay = initialDelay
		} else {
			delay *= 2
			if maxDelay > 0 && delay > maxDelay {
				delay = maxDelay
			}
		}
	}
	return err
}

type ItemUpsert struct {
	ItemID    string
	Available bool
	Price     *float64
	Tags      []string
	Props     any
	// Optional embedding. If provided, stored in items.embedding (pgvector).
	Embedding       *[]float64
	Brand           *string
	Category        *string
	CategoryPath    *[]string
	Description     *string
	ImageURL        *string
	MetadataVersion *string
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

//go:embed queries/popularity.sql
var popularitySQL string

//go:embed queries/items_upsert.sql
var itemsUpsertSQL string

//go:embed queries/users_upsert.sql
var usersUpsertSQL string

//go:embed queries/events_insert.sql
var eventsInsertSQL string

//go:embed queries/cooccurrence_top_k.sql
var cooccurrenceTopKSQL string

//go:embed queries/event_type_config_upsert.sql
var eventTypeConfigUpsertSQL string

//go:embed queries/event_type_config_effective.sql
var eventTypeConfigEffectiveSQL string

//go:embed queries/items_tags.sql
var itemsTagsSQL string

//go:embed queries/user_events_since.sql
var userEventsSinceSQL string

func (s *Store) UpsertItems(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	items []ItemUpsert,
) error {
	if len(items) == 0 {
		return nil
	}
	return s.withRetry(ctx, func(ctx context.Context) error {
		bat := &pgx.Batch{}
		for _, it := range items {
			// Prepare optional vector literal like: "[0.1,0.2,...]"
			var embText *string
			if it.Embedding != nil && len(*it.Embedding) > 0 {
				t := vectorLiteral(*it.Embedding)
				embText = &t
			}

			var catPath interface{}
			if it.CategoryPath != nil {
				catPath = *it.CategoryPath
			}

			bat.Queue(
				itemsUpsertSQL,
				orgID,
				ns,
				it.ItemID,
				it.Available,
				it.Price,
				it.Tags,
				it.Props,
				embText,
				it.Brand,
				it.Category,
				catPath,
				it.Description,
				it.ImageURL,
				it.MetadataVersion,
			)
		}
		br := s.Pool.SendBatch(ctx, bat)
		defer br.Close()
		for range items {
			if _, err := br.Exec(); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) UpsertUsers(ctx context.Context, orgID uuid.UUID, ns string, users []UserUpsert) error {
	if len(users) == 0 {
		return nil
	}
	return s.withRetry(ctx, func(ctx context.Context) error {
		bat := &pgx.Batch{}
		for _, u := range users {
			bat.Queue(usersUpsertSQL, orgID, ns, u.UserID, u.Traits)
		}
		br := s.Pool.SendBatch(ctx, bat)
		defer br.Close()
		for range users {
			if _, err := br.Exec(); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) InsertEvents(ctx context.Context, orgID uuid.UUID, ns string, evs []EventInsert) error {
	if len(evs) == 0 {
		return nil
	}
	return s.withRetry(ctx, func(ctx context.Context) error {
		bat := &pgx.Batch{}
		for _, e := range evs {
			bat.Queue(eventsInsertSQL, orgID, ns, e.UserID, e.ItemID, e.Type, e.Value, e.TS, e.Meta, e.SourceEventID)
		}
		br := s.Pool.SendBatch(ctx, bat)
		defer br.Close()
		for range evs {
			if _, err := br.Exec(); err != nil {
				return err
			}
		}
		return nil
	})
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
	var out []types.ScoredItem
	err := s.withRetry(ctx, func(ctx context.Context) error {
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
			return err
		}
		defer rows.Close()

		items := make([]types.ScoredItem, 0, k)
		for rows.Next() {
			var it types.ScoredItem
			if err := rows.Scan(&it.ItemID, &it.Score); err != nil {
				return err
			}
			items = append(items, it)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		out = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
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
	var out []types.ScoredItem
	err := s.withRetry(ctx, func(ctx context.Context) error {
		rows, err := s.Pool.Query(ctx, cooccurrenceTopKSQL, orgID, ns, itemID, k, since)
		if err != nil {
			return err
		}
		defer rows.Close()

		items := make([]types.ScoredItem, 0, k)
		for rows.Next() {
			var it types.ScoredItem
			if err := rows.Scan(&it.ItemID, &it.Score); err != nil {
				return err
			}
			items = append(items, it)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		out = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Upsert tenant overrides (batch).
func (s *Store) UpsertEventTypeConfig(ctx context.Context, orgID uuid.UUID, ns string, rows []EventTypeConfig) error {
	if len(rows) == 0 {
		return nil
	}
	return s.withRetry(ctx, func(ctx context.Context) error {
		bat := &pgx.Batch{}
		for _, r := range rows {
			bat.Queue(eventTypeConfigUpsertSQL, orgID, ns, r.Type, r.Name, r.Weight, r.HalfLifeDays, r.IsActive)
		}
		br := s.Pool.SendBatch(ctx, bat)
		defer br.Close()
		for range rows {
			if _, err := br.Exec(); err != nil {
				return err
			}
		}
		return nil
	})
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
	var out []EventTypeConfigRow
	err := s.withRetry(ctx, func(ctx context.Context) error {
		rows, err := s.Pool.Query(ctx, eventTypeConfigEffectiveSQL, orgID, ns)
		if err != nil {
			return err
		}
		defer rows.Close()

		items := make([]EventTypeConfigRow, 0)
		for rows.Next() {
			var r EventTypeConfigRow
			if err := rows.Scan(&r.Type, &r.Name, &r.Weight, &r.HalfLifeDays, &r.IsActive, &r.Source); err != nil {
				return err
			}
			items = append(items, r)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		out = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListItemsTags returns tags for the given item IDs.
func (s *Store) ListItemsTags(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	itemIDs []string,
) (map[string]types.ItemTags, error) {
	if len(itemIDs) == 0 {
		return map[string]types.ItemTags{}, nil
	}
	var out map[string]types.ItemTags
	err := s.withRetry(ctx, func(ctx context.Context) error {
		rows, err := s.Pool.Query(ctx, itemsTagsSQL, orgID, ns, itemIDs)
		if err != nil {
			return err
		}
		defer rows.Close()

		res := make(map[string]types.ItemTags, len(itemIDs))
		for rows.Next() {
			var id string
			var tags []string
			if err := rows.Scan(&id, &tags); err != nil {
				return err
			}
			res[id] = types.ItemTags{ItemID: id, Tags: tags}
		}
		if err := rows.Err(); err != nil {
			return err
		}
		out = res
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListUserEventsSince returns distinct item IDs for the user's events on/after
// the given timestamp, filtered by the provided event types
// (nil or empty means any).
func (s *Store) ListUserEventsSince(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	userID string,
	since time.Time,
	eventTypes []int16,
) ([]string, error) {
	var evParam any
	if len(eventTypes) > 0 {
		evParam = eventTypes
	}
	var ids []string
	err := s.withRetry(ctx, func(ctx context.Context) error {
		rows, err := s.Pool.Query(
			ctx, userEventsSinceSQL, orgID, ns, userID, since, evParam,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		list := make([]string, 0, 64)
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			list = append(list, id)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		ids = list
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
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
	baseArgs := []interface{}{orgID, ns}
	argIndex := 3

	// Add filters
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		whereClause += fmt.Sprintf(" AND user_id = $%d", argIndex)
		baseArgs = append(baseArgs, userID)
		argIndex++
	}

	if createdAfter, ok := filters["created_after"].(string); ok && createdAfter != "" {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		baseArgs = append(baseArgs, createdAfter)
		argIndex++
	}

	if createdBefore, ok := filters["created_before"].(string); ok && createdBefore != "" {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		baseArgs = append(baseArgs, createdBefore)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	query := fmt.Sprintf(`
		SELECT user_id, traits, created_at, updated_at
		FROM users %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	var (
		total int
		users []map[string]interface{}
	)
	err := s.withRetry(ctx, func(ctx context.Context) error {
		countArgs := append([]interface{}(nil), baseArgs...)
		if err := s.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
			return err
		}

		queryArgs := append([]interface{}(nil), baseArgs...)
		queryArgs = append(queryArgs, limit, offset)

		rows, err := s.Pool.Query(ctx, query, queryArgs...)
		if err != nil {
			return err
		}
		defer rows.Close()

		result := make([]map[string]interface{}, 0)
		for rows.Next() {
			var userID string
			var traits []byte
			var createdAt, updatedAt time.Time

			if err := rows.Scan(&userID, &traits, &createdAt, &updatedAt); err != nil {
				return err
			}

			result = append(result, map[string]interface{}{
				"user_id":    userID,
				"traits":     string(traits),
				"created_at": createdAt.Format(time.RFC3339),
				"updated_at": updatedAt.Format(time.RFC3339),
			})
		}
		if err := rows.Err(); err != nil {
			return err
		}
		users = result
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// ListItems returns a paginated list of items with optional filtering
func (s *Store) ListItems(ctx context.Context, orgID uuid.UUID, ns string, limit, offset int, filters map[string]interface{}) ([]map[string]interface{}, int, error) {
	// Build WHERE clause
	whereClause := "WHERE org_id = $1 AND namespace = $2"
	baseArgs := []interface{}{orgID, ns}
	argIndex := 3

	// Add filters
	if itemID, ok := filters["item_id"].(string); ok && itemID != "" {
		whereClause += fmt.Sprintf(" AND item_id = $%d", argIndex)
		baseArgs = append(baseArgs, itemID)
		argIndex++
	}

	if createdAfter, ok := filters["created_after"].(string); ok && createdAfter != "" {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		baseArgs = append(baseArgs, createdAfter)
		argIndex++
	}

	if createdBefore, ok := filters["created_before"].(string); ok && createdBefore != "" {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		baseArgs = append(baseArgs, createdBefore)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM items %s", whereClause)
	query := fmt.Sprintf(`
		SELECT item_id,
		       available,
		       price,
		       tags,
		       props,
		       brand,
		       category,
		       category_path,
		       description,
		       image_url,
		       metadata_version,
		       created_at,
		       updated_at
		FROM items %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	var (
		total int
		items []map[string]interface{}
	)
	err := s.withRetry(ctx, func(ctx context.Context) error {
		countArgs := append([]interface{}(nil), baseArgs...)
		if err := s.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
			return err
		}

		queryArgs := append([]interface{}(nil), baseArgs...)
		queryArgs = append(queryArgs, limit, offset)

		rows, err := s.Pool.Query(ctx, query, queryArgs...)
		if err != nil {
			return err
		}
		defer rows.Close()

		result := make([]map[string]interface{}, 0)
		for rows.Next() {
			var (
				itemID      string
				available   bool
				price       *float64
				tags        []string
				props       []byte
				brand       *string
				category    *string
				catPath     []string
				description *string
				imageURL    *string
				metadataVer *string
				createdAt   time.Time
				updatedAt   time.Time
			)

			if err := rows.Scan(&itemID, &available, &price, &tags, &props, &brand, &category, &catPath, &description, &imageURL, &metadataVer, &createdAt, &updatedAt); err != nil {
				return err
			}

			entry := map[string]interface{}{
				"item_id":       itemID,
				"available":     available,
				"price":         price,
				"tags":          tags,
				"props":         string(props),
				"category_path": catPath,
				"created_at":    createdAt.Format(time.RFC3339),
				"updated_at":    updatedAt.Format(time.RFC3339),
			}
			if brand != nil {
				entry["brand"] = *brand
			}
			if category != nil {
				entry["category"] = *category
			}
			if description != nil {
				entry["description"] = *description
			}
			if imageURL != nil {
				entry["image_url"] = *imageURL
			}
			if metadataVer != nil {
				entry["metadata_version"] = *metadataVer
			}

			result = append(result, entry)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		items = result
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ListEvents returns a paginated list of events with optional filtering
func (s *Store) ListEvents(ctx context.Context, orgID uuid.UUID, ns string, limit, offset int, filters map[string]interface{}) ([]map[string]interface{}, int, error) {
	// Build WHERE clause
	whereClause := "WHERE org_id = $1 AND namespace = $2"
	baseArgs := []interface{}{orgID, ns}
	argIndex := 3

	// Add filters
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		whereClause += fmt.Sprintf(" AND user_id = $%d", argIndex)
		baseArgs = append(baseArgs, userID)
		argIndex++
	}

	if itemID, ok := filters["item_id"].(string); ok && itemID != "" {
		whereClause += fmt.Sprintf(" AND item_id = $%d", argIndex)
		baseArgs = append(baseArgs, itemID)
		argIndex++
	}

	if eventType, ok := filters["event_type"].(int16); ok {
		whereClause += fmt.Sprintf(" AND type = $%d", argIndex)
		baseArgs = append(baseArgs, eventType)
		argIndex++
	}

	if createdAfter, ok := filters["created_after"].(string); ok && createdAfter != "" {
		whereClause += fmt.Sprintf(" AND ts >= $%d", argIndex)
		baseArgs = append(baseArgs, createdAfter)
		argIndex++
	}

	if createdBefore, ok := filters["created_before"].(string); ok && createdBefore != "" {
		whereClause += fmt.Sprintf(" AND ts <= $%d", argIndex)
		baseArgs = append(baseArgs, createdBefore)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM events %s", whereClause)
	query := fmt.Sprintf(`
		SELECT user_id, item_id, type, value, ts, meta, source_event_id
		FROM events %s
		ORDER BY ts DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	var (
		total  int
		events []map[string]interface{}
	)
	err := s.withRetry(ctx, func(ctx context.Context) error {
		countArgs := append([]interface{}(nil), baseArgs...)
		if err := s.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
			return err
		}

		queryArgs := append([]interface{}(nil), baseArgs...)
		queryArgs = append(queryArgs, limit, offset)

		rows, err := s.Pool.Query(ctx, query, queryArgs...)
		if err != nil {
			return err
		}
		defer rows.Close()

		result := make([]map[string]interface{}, 0)
		for rows.Next() {
			var userID, itemID string
			var eventType int16
			var value float64
			var ts time.Time
			var meta []byte
			var sourceEventID *string

			if err := rows.Scan(&userID, &itemID, &eventType, &value, &ts, &meta, &sourceEventID); err != nil {
				return err
			}

			result = append(result, map[string]interface{}{
				"user_id":         userID,
				"item_id":         itemID,
				"type":            eventType,
				"value":           value,
				"ts":              ts.Format(time.RFC3339),
				"meta":            string(meta),
				"source_event_id": sourceEventID,
			})
		}
		if err := rows.Err(); err != nil {
			return err
		}
		events = result
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

// CountEventsByName returns the number of events with a given logical name
// (e.g. 'click') within a time window for a namespace. If itemID is non-empty,
// the count is restricted to that item.
func (s *Store) CountEventsByName(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	from, to time.Time,
	itemID string,
	name string,
) (int, error) {
	args := []any{orgID, ns, from, to, name}
	where := "WHERE e.org_id = $1 AND e.namespace = $2 AND e.ts >= $3 AND e.ts <= $4 AND etc.name = $5"
	if itemID != "" {
		where += " AND e.item_id = $6"
		args = append(args, itemID)
	}
	query := `
        SELECT COUNT(*)
        FROM events e
        JOIN event_type_config etc
          ON etc.org_id = e.org_id AND etc.namespace = e.namespace AND etc.type = e.type
        ` + where
	var total int
	err := s.withRetry(ctx, func(ctx context.Context) error {
		return s.Pool.QueryRow(ctx, query, args...).Scan(&total)
	})
	if err != nil {
		return 0, err
	}
	return total, nil
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
	}

	query := fmt.Sprintf("DELETE FROM users %s", whereClause)
	var affected int
	err := s.withRetry(ctx, func(ctx context.Context) error {
		result, err := s.Pool.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
		affected = int(result.RowsAffected())
		return nil
	})
	if err != nil {
		return 0, err
	}
	return affected, nil
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
	}

	query := fmt.Sprintf("DELETE FROM items %s", whereClause)
	var affected int
	err := s.withRetry(ctx, func(ctx context.Context) error {
		result, err := s.Pool.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
		affected = int(result.RowsAffected())
		return nil
	})
	if err != nil {
		return 0, err
	}
	return affected, nil
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
	}

	query := fmt.Sprintf("DELETE FROM events %s", whereClause)
	var affected int
	err := s.withRetry(ctx, func(ctx context.Context) error {
		result, err := s.Pool.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
		affected = int(result.RowsAffected())
		return nil
	})
	if err != nil {
		return 0, err
	}
	return affected, nil
}
