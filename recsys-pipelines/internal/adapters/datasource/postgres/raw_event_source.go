package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/events"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/windows"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/datasource"
)

type RawEventSource struct {
	pool          *pgxpool.Pool
	tenantTable   string
	exposureTable string
}

var _ datasource.RawEventSource = (*RawEventSource)(nil)

type Config struct {
	DSN           string
	TenantTable   string
	ExposureTable string
}

func New(cfg Config) (*RawEventSource, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}
	pool, err := pgxpool.New(context.Background(), cfg.DSN)
	if err != nil {
		return nil, err
	}
	if cfg.TenantTable == "" {
		cfg.TenantTable = "tenants"
	}
	if cfg.ExposureTable == "" {
		cfg.ExposureTable = "exposure_events"
	}
	return &RawEventSource{
		pool:          pool,
		tenantTable:   cfg.TenantTable,
		exposureTable: cfg.ExposureTable,
	}, nil
}

func (s *RawEventSource) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

func (s *RawEventSource) ReadExposureEvents(
	ctx context.Context,
	tenant string,
	surface string,
	w windows.Window,
) (<-chan events.ExposureEvent, <-chan error) {
	out := make(chan events.ExposureEvent, 256)
	errs := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errs)

		if s == nil || s.pool == nil {
			errs <- fmt.Errorf("postgres source not configured")
			return
		}
		tenantID, err := s.resolveTenantID(ctx, tenant)
		if err != nil {
			errs <- err
			return
		}
		if tenantID == uuid.Nil {
			return
		}

		query := fmt.Sprintf(`
select occurred_at, request_id, surface, segment, session_id, response
  from %s
 where tenant_id = $1
   and surface = $2
   and occurred_at >= $3
   and occurred_at < $4
 order by occurred_at asc
`, s.exposureTable)

		rows, err := s.pool.Query(ctx, query, tenantID, surface, w.Start, w.End)
		if err != nil {
			errs <- err
			return
		}
		defer rows.Close()

		for rows.Next() {
			var occurredAt time.Time
			var requestID uuid.UUID
			var rowSurface, segment string
			var sessionID *string
			var responseRaw []byte
			if err := rows.Scan(&occurredAt, &requestID, &rowSurface, &segment, &sessionID, &responseRaw); err != nil {
				errs <- err
				return
			}
			if len(responseRaw) == 0 {
				continue
			}
			var payload struct {
				Items []struct {
					ItemID string `json:"item_id"`
					Rank   int    `json:"rank"`
				} `json:"items"`
			}
			if err := json.Unmarshal(responseRaw, &payload); err != nil {
				errs <- fmt.Errorf("decode exposure response: %w", err)
				return
			}
			for _, item := range payload.Items {
				ev := events.ExposureEvent{
					Version:   1,
					TS:        occurredAt.UTC(),
					Tenant:    tenant,
					Surface:   rowSurface,
					UserID:    "",
					SessionID: "",
					RequestID: requestID.String(),
					ItemID:    item.ItemID,
					Rank:      item.Rank,
				}
				if sessionID != nil {
					ev.SessionID = *sessionID
				}
				ev = ev.Normalized()
				if err := ev.Validate(); err != nil {
					errs <- fmt.Errorf("invalid exposure event: %w", err)
					return
				}
				out <- ev
			}
		}
		if err := rows.Err(); err != nil {
			errs <- err
			return
		}
	}()

	return out, errs
}

func (s *RawEventSource) resolveTenantID(ctx context.Context, tenant string) (uuid.UUID, error) {
	query := fmt.Sprintf(`select id from %s where external_id = $1 or id::text = $1`, s.tenantTable)
	var id uuid.UUID
	if err := s.pool.QueryRow(ctx, query, tenant).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, nil
		}
		return uuid.Nil, err
	}
	return id, nil
}
