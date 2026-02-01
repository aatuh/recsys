package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/signals"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/signalstore"
)

// Store implements DB-backed signal persistence for the service.
type Store struct {
	pool          *pgxpool.Pool
	createTenant  bool
	statementTTL  time.Duration
	maxBatchItems int
}

// Option configures the Store.
type Option func(*Store)

// WithCreateTenant enables auto-creation of tenant rows.
func WithCreateTenant(enabled bool) Option {
	return func(s *Store) {
		s.createTenant = enabled
	}
}

// WithStatementTimeout sets a per-operation timeout.
func WithStatementTimeout(d time.Duration) Option {
	return func(s *Store) {
		s.statementTTL = d
	}
}

// WithMaxBatchItems caps the batch size for inserts.
func WithMaxBatchItems(n int) Option {
	return func(s *Store) {
		if n > 0 {
			s.maxBatchItems = n
		}
	}
}

// New constructs a Store from an existing pgx pool.
func New(pool *pgxpool.Pool, opts ...Option) *Store {
	st := &Store{pool: pool, maxBatchItems: 500}
	for _, opt := range opts {
		if opt != nil {
			opt(st)
		}
	}
	return st
}

// NewFromDSN constructs a Store by creating its own pgx pool.
func NewFromDSN(ctx context.Context, dsn string, opts ...Option) (*Store, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("postgres dsn is required")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return New(pool, opts...), nil
}

// Close releases the underlying pool.
func (s *Store) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) UpsertItemTags(ctx context.Context, tenant, namespace string, items []signals.ItemTag) error {
	if s == nil || s.pool == nil {
		return errors.New("signal store not configured")
	}
	if len(items) == 0 {
		return nil
	}
	if namespace == "" {
		namespace = "default"
	}
	tenantID, err := s.resolveTenant(ctx, tenant)
	if err != nil {
		return err
	}
	query := `
insert into item_tags (tenant_id, namespace, item_id, tags, price, created_at)
values ($1, $2, $3, $4, $5, $6)
on conflict (tenant_id, namespace, item_id)
do update set tags = excluded.tags,
              price = excluded.price,
              created_at = excluded.created_at,
              updated_at = now();
`
	return s.execBatch(ctx, len(items), func(batch *pgx.Batch, offset, limit int) error {
		for i := 0; i < limit; i++ {
			it := items[offset+i]
			createdAt := it.CreatedAt
			if createdAt.IsZero() {
				createdAt = time.Now().UTC()
			}
			itemNS := namespace
			if strings.TrimSpace(it.Namespace) != "" {
				itemNS = strings.TrimSpace(it.Namespace)
			}
			batch.Queue(query, tenantID, itemNS, it.ItemID, it.Tags, it.Price, createdAt)
		}
		return nil
	})
}

func (s *Store) UpsertPopularity(ctx context.Context, tenant, namespace string, day time.Time, items []signals.PopularityItem) error {
	if s == nil || s.pool == nil {
		return errors.New("signal store not configured")
	}
	if len(items) == 0 {
		return nil
	}
	if namespace == "" {
		namespace = "default"
	}
	tenantID, err := s.resolveTenant(ctx, tenant)
	if err != nil {
		return err
	}
	day = truncateDay(day)
	query := `
insert into item_popularity_daily (tenant_id, namespace, item_id, day, score)
values ($1, $2, $3, $4, $5)
on conflict (tenant_id, namespace, item_id, day)
do update set score = excluded.score,
              updated_at = now();
`
	return s.execBatch(ctx, len(items), func(batch *pgx.Batch, offset, limit int) error {
		for i := 0; i < limit; i++ {
			it := items[offset+i]
			batch.Queue(query, tenantID, namespace, it.ItemID, day, it.Score)
		}
		return nil
	})
}

func (s *Store) UpsertCooccurrence(ctx context.Context, tenant, namespace string, day time.Time, items []signals.CooccurrenceItem) error {
	if s == nil || s.pool == nil {
		return errors.New("signal store not configured")
	}
	if len(items) == 0 {
		return nil
	}
	if namespace == "" {
		namespace = "default"
	}
	tenantID, err := s.resolveTenant(ctx, tenant)
	if err != nil {
		return err
	}
	day = truncateDay(day)
	query := `
insert into item_covisit_daily (tenant_id, namespace, item_id, neighbor_id, day, score)
values ($1, $2, $3, $4, $5, $6)
on conflict (tenant_id, namespace, item_id, neighbor_id, day)
do update set score = excluded.score,
              updated_at = now();
`
	return s.execBatch(ctx, len(items), func(batch *pgx.Batch, offset, limit int) error {
		for i := 0; i < limit; i++ {
			it := items[offset+i]
			batch.Queue(query, tenantID, namespace, it.ItemID, it.NeighborID, day, it.Score)
		}
		return nil
	})
}

func (s *Store) execBatch(ctx context.Context, total int, queue func(batch *pgx.Batch, offset, limit int) error) error {
	if total == 0 {
		return nil
	}
	maxBatch := s.maxBatchItems
	if maxBatch <= 0 {
		maxBatch = total
	}
	for offset := 0; offset < total; offset += maxBatch {
		limit := total - offset
		if limit > maxBatch {
			limit = maxBatch
		}
		batch := &pgx.Batch{}
		if err := queue(batch, offset, limit); err != nil {
			return err
		}
		ctxTimeout, cancel := applyTimeout(ctx, s.statementTTL)
		res := s.pool.SendBatch(ctxTimeout, batch)
		for i := 0; i < limit; i++ {
			if _, err := res.Exec(); err != nil {
				res.Close()
				if cancel != nil {
					cancel()
				}
				return err
			}
		}
		if err := res.Close(); err != nil {
			if cancel != nil {
				cancel()
			}
			return err
		}
		if cancel != nil {
			cancel()
		}
	}
	return nil
}

func (s *Store) resolveTenant(ctx context.Context, tenant string) (uuid.UUID, error) {
	if strings.TrimSpace(tenant) == "" {
		return uuid.Nil, errors.New("tenant is required")
	}
	const lookup = `select id from tenants where external_id = $1 or id::text = $1`
	var id uuid.UUID
	ctxTimeout, cancel := applyTimeout(ctx, s.statementTTL)
	if cancel != nil {
		defer cancel()
	}
	err := s.pool.QueryRow(ctxTimeout, lookup, tenant).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, err
	}
	if !s.createTenant {
		return uuid.Nil, fmt.Errorf("tenant not found: %s", tenant)
	}
	const insertQ = `insert into tenants (external_id, name, status) values ($1, $2, 'active') returning id`
	ctxTimeout, cancel = applyTimeout(ctx, s.statementTTL)
	if cancel != nil {
		defer cancel()
	}
	if err := s.pool.QueryRow(ctxTimeout, insertQ, tenant, tenant).Scan(&id); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func applyTimeout(ctx context.Context, ttl time.Duration) (context.Context, context.CancelFunc) {
	if ttl <= 0 {
		return ctx, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, ttl)
	return ctx, cancel
}

func truncateDay(day time.Time) time.Time {
	if day.IsZero() {
		return time.Now().UTC().Truncate(24 * time.Hour)
	}
	day = day.UTC()
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
}

var _ signalstore.Store = (*Store)(nil)
