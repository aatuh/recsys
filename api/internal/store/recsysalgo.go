package store

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/aatuh/api-toolkit-contrib/adapters/txpostgres"
	"github.com/aatuh/api-toolkit/ports"
	recmodel "github.com/aatuh/recsys-algo/model"
	"github.com/google/uuid"
)

// AlgoStore is a Postgres-backed adapter for recsys-algo data access.
type AlgoStore struct {
	Pool ports.DatabasePool
}

func NewAlgoStore(pool ports.DatabasePool) *AlgoStore {
	return &AlgoStore{Pool: pool}
}

// PopularityTopK returns candidates ordered by decayed popularity.
func (s *AlgoStore) PopularityTopK(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	halfLifeDays float64,
	k int,
	c *recmodel.PopConstraints,
) ([]recmodel.ScoredItem, error) {
	if k <= 0 {
		return nil, nil
	}
	tenantID, err := resolveTenantID(ctx, s.Pool, orgID)
	if err != nil || tenantID == uuid.Nil {
		return nil, err
	}
	ns = strings.TrimSpace(ns)
	if ns == "" {
		ns = "default"
	}
	if halfLifeDays <= 0 {
		halfLifeDays = 30
	}

	var excludeIDs []string
	var includeTags []string
	var minPrice *float64
	var maxPrice *float64
	var createdAfter *time.Time
	if c != nil {
		if len(c.ExcludeItemIDs) > 0 {
			excludeIDs = c.ExcludeItemIDs
		}
		if len(c.IncludeTagsAny) > 0 {
			includeTags = c.IncludeTagsAny
		}
		minPrice = c.MinPrice
		maxPrice = c.MaxPrice
		createdAfter = c.CreatedAfter
	}

	db := txpostgres.FromCtx(ctx, s.Pool)
	const q = `
with scored as (
  select p.item_id,
         sum(p.score * exp(-ln(2) * greatest((current_date - p.day), 0)::numeric / $4)) as score
    from item_popularity_daily p
    left join item_tags t
      on t.tenant_id = p.tenant_id
     and t.namespace = p.namespace
     and t.item_id = p.item_id
   where p.tenant_id = $1
     and p.namespace = $2
     and ($5::text[] is null or not (p.item_id = any($5)))
     and ($6::text[] is null or t.tags && $6)
     and ($7::numeric is null or t.price >= $7)
     and ($8::numeric is null or t.price <= $8)
     and ($9::timestamptz is null or t.created_at >= $9)
   group by p.item_id
)
select item_id, score
  from scored
 order by score desc
 limit $3;
`
	query := func(namespace string) ([]recmodel.ScoredItem, error) {
		rows, err := db.Query(ctx, q, tenantID, namespace, k, halfLifeDays, excludeIDs, includeTags, minPrice, maxPrice, createdAfter)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		out := make([]recmodel.ScoredItem, 0, k)
		for rows.Next() {
			var itemID string
			var score float64
			if err := rows.Scan(&itemID, &score); err != nil {
				return nil, err
			}
			out = append(out, recmodel.ScoredItem{ItemID: itemID, Score: score})
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return out, nil
	}

	out, err := query(ns)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 && ns != "default" {
		return query("default")
	}
	return out, nil
}

// ListItemsTags returns tag metadata for requested items.
func (s *AlgoStore) ListItemsTags(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]recmodel.ItemTags, error) {
	if len(itemIDs) == 0 {
		return map[string]recmodel.ItemTags{}, nil
	}
	tenantID, err := resolveTenantID(ctx, s.Pool, orgID)
	if err != nil || tenantID == uuid.Nil {
		return nil, err
	}
	ns = strings.TrimSpace(ns)
	if ns == "" {
		ns = "default"
	}
	db := txpostgres.FromCtx(ctx, s.Pool)
	const q = `
select item_id, tags, price, created_at
  from item_tags
 where tenant_id = $1
   and namespace = $2
   and item_id = any($3);
`
	query := func(namespace string) (map[string]recmodel.ItemTags, error) {
		rows, err := db.Query(ctx, q, tenantID, namespace, itemIDs)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		out := make(map[string]recmodel.ItemTags, len(itemIDs))
		for rows.Next() {
			var itemID string
			var tags []string
			var price sql.NullFloat64
			var createdAt time.Time
			if err := rows.Scan(&itemID, &tags, &price, &createdAt); err != nil {
				return nil, err
			}
			tagInfo := recmodel.ItemTags{
				ItemID:    itemID,
				Tags:      tags,
				CreatedAt: createdAt,
			}
			if price.Valid {
				tagInfo.Price = &price.Float64
			}
			out[itemID] = tagInfo
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return out, nil
	}

	out, err := query(ns)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 && ns != "default" {
		return query("default")
	}
	return out, nil
}

// CooccurrenceTopKWithin returns top co-visitation neighbors for an anchor item.
func (s *AlgoStore) CooccurrenceTopKWithin(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	anchor string,
	k int,
	since time.Time,
) ([]recmodel.ScoredItem, error) {
	if k <= 0 || strings.TrimSpace(anchor) == "" {
		return nil, nil
	}
	tenantID, err := resolveTenantID(ctx, s.Pool, orgID)
	if err != nil || tenantID == uuid.Nil {
		return nil, err
	}
	ns = strings.TrimSpace(ns)
	if ns == "" {
		ns = "default"
	}
	if since.IsZero() {
		since = time.Now().UTC().Add(-30 * 24 * time.Hour)
	}
	day := since.UTC()

	db := txpostgres.FromCtx(ctx, s.Pool)
	const q = `
select neighbor_id, sum(score)::float8 as score
  from item_covisit_daily
 where tenant_id = $1
   and namespace = $2
   and item_id = $3
   and day >= $4::date
 group by neighbor_id
 order by score desc
 limit $5;
`
	query := func(namespace string) ([]recmodel.ScoredItem, error) {
		rows, err := db.Query(ctx, q, tenantID, namespace, anchor, day, k)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		out := make([]recmodel.ScoredItem, 0, k)
		for rows.Next() {
			var itemID string
			var score float64
			if err := rows.Scan(&itemID, &score); err != nil {
				return nil, err
			}
			out = append(out, recmodel.ScoredItem{ItemID: itemID, Score: score})
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return out, nil
	}

	out, err := query(ns)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 && ns != "default" {
		return query("default")
	}
	return out, nil
}

// ListItemsAvailability marks all requested items as available.
func (s *AlgoStore) ListItemsAvailability(ctx context.Context, orgID uuid.UUID, ns string, itemIDs []string) (map[string]bool, error) {
	out := make(map[string]bool, len(itemIDs))
	for _, id := range itemIDs {
		if id != "" {
			out[id] = true
		}
	}
	return out, nil
}

var _ recmodel.EngineStore = (*AlgoStore)(nil)
