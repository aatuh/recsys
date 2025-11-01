package store

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// DecisionTraceInsert represents a row to insert into rec_decisions.
type DecisionTraceInsert struct {
	DecisionID      uuid.UUID
	OrgID           uuid.UUID
	Timestamp       time.Time
	Namespace       string
	Surface         *string
	RequestID       *string
	UserHash        *string
	K               *int
	ConstraintsJSON []byte
	EffectiveConfig []byte
	BanditJSON      []byte
	CandidatesPre   []byte
	FinalItems      []byte
	MMRInfo         []byte
	Caps            []byte
	Extras          []byte
}

//go:embed queries/rec_decisions_insert.sql
var decisionInsertSQL string

//go:embed queries/rec_decisions_list.sql
var decisionListSQL string

//go:embed queries/rec_decisions_get.sql
var decisionGetSQL string

// InsertDecisionTraces batches inserts into rec_decisions.
func (s *Store) InsertDecisionTraces(ctx context.Context, rows []DecisionTraceInsert) error {
	if len(rows) == 0 {
		return nil
	}

	return s.withRetry(ctx, func(ctx context.Context) error {
		batch := &pgx.Batch{}
		for _, row := range rows {
			ts := row.Timestamp
			if ts.IsZero() {
				ts = time.Now().UTC()
			}
			var surface any
			if row.Surface != nil && *row.Surface != "" {
				surface = *row.Surface
			}
			var reqID any
			if row.RequestID != nil && *row.RequestID != "" {
				reqID = *row.RequestID
			}
			var userHash any
			if row.UserHash != nil && *row.UserHash != "" {
				userHash = *row.UserHash
			}
			var k any
			if row.K != nil {
				k = *row.K
			}
			batch.Queue(
				decisionInsertSQL,
				row.DecisionID,
				row.OrgID,
				ts,
				row.Namespace,
				surface,
				reqID,
				userHash,
				k,
				nullableJSON(row.ConstraintsJSON),
				row.EffectiveConfig,
				nullableJSON(row.BanditJSON),
				row.CandidatesPre,
				row.FinalItems,
				nullableJSON(row.MMRInfo),
				nullableJSON(row.Caps),
				nullableJSON(row.Extras),
			)
		}

		br := s.Pool.SendBatch(ctx, batch)
		defer br.Close()

		for range rows {
			if _, err := br.Exec(); err != nil {
				return err
			}
		}
		return nil
	})
}

func nullableJSON(data []byte) any {
	if len(data) == 0 {
		return nil
	}
	return data
}

type DecisionTraceFilter struct {
	From      *time.Time
	To        *time.Time
	UserHash  string
	RequestID string
	Limit     int
}

type DecisionTraceSummary struct {
	DecisionID     uuid.UUID
	OrgID          uuid.UUID
	Timestamp      time.Time
	Namespace      string
	Surface        *string
	RequestID      *string
	UserHash       *string
	K              *int
	FinalItemsJSON []byte
	ExtrasJSON     []byte
}

type DecisionTraceRecord struct {
	DecisionID          uuid.UUID
	OrgID               uuid.UUID
	Timestamp           time.Time
	Namespace           string
	Surface             *string
	RequestID           *string
	UserHash            *string
	K                   *int
	ConstraintsJSON     []byte
	EffectiveConfigJSON []byte
	BanditJSON          []byte
	CandidatesPreJSON   []byte
	FinalItemsJSON      []byte
	MMRInfoJSON         []byte
	CapsJSON            []byte
	ExtrasJSON          []byte
}

func (s *Store) ListDecisionTraces(
	ctx context.Context,
	orgID uuid.UUID,
	ns string,
	filter DecisionTraceFilter,
) ([]DecisionTraceSummary, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 50
	}

	var out []DecisionTraceSummary
	err := s.withRetry(ctx, func(ctx context.Context) error {
		rows, err := s.Pool.Query(
			ctx,
			decisionListSQL,
			orgID,
			ns,
			nullableTime(filter.From),
			nullableTime(filter.To),
			nullableString(filter.UserHash),
			nullableString(filter.RequestID),
			limit,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		list := make([]DecisionTraceSummary, 0, limit)
		for rows.Next() {
			var (
				surface    sql.NullString
				reqID      sql.NullString
				userHash   sql.NullString
				k          sql.NullInt32
				finalJSON  []byte
				extrasJSON []byte
				row        DecisionTraceSummary
			)
			if err := rows.Scan(
				&row.DecisionID,
				&row.OrgID,
				&row.Timestamp,
				&row.Namespace,
				&surface,
				&reqID,
				&userHash,
				&k,
				&finalJSON,
				&extrasJSON,
			); err != nil {
				return err
			}
			if surface.Valid {
				val := surface.String
				row.Surface = &val
			}
			if reqID.Valid {
				val := reqID.String
				row.RequestID = &val
			}
			if userHash.Valid {
				val := userHash.String
				row.UserHash = &val
			}
			if k.Valid {
				val := int(k.Int32)
				row.K = &val
			}
			row.FinalItemsJSON = append([]byte(nil), finalJSON...)
			row.ExtrasJSON = append([]byte(nil), extrasJSON...)
			list = append(list, row)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		out = list
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) GetDecisionTrace(
	ctx context.Context,
	orgID uuid.UUID,
	decisionID uuid.UUID,
) (*DecisionTraceRecord, error) {
	var (
		surface  sql.NullString
		reqID    sql.NullString
		userHash sql.NullString
		k        sql.NullInt32
		record   DecisionTraceRecord
	)
	err := s.withRetry(ctx, func(ctx context.Context) error {
		row := s.Pool.QueryRow(ctx, decisionGetSQL, orgID, decisionID)
		return row.Scan(
			&record.DecisionID,
			&record.OrgID,
			&record.Timestamp,
			&record.Namespace,
			&surface,
			&reqID,
			&userHash,
			&k,
			&record.ConstraintsJSON,
			&record.EffectiveConfigJSON,
			&record.BanditJSON,
			&record.CandidatesPreJSON,
			&record.FinalItemsJSON,
			&record.MMRInfoJSON,
			&record.CapsJSON,
			&record.ExtrasJSON,
		)
	})
	if err != nil {
		return nil, err
	}
	if surface.Valid {
		val := surface.String
		record.Surface = &val
	}
	if reqID.Valid {
		val := reqID.String
		record.RequestID = &val
	}
	if userHash.Valid {
		val := userHash.String
		record.UserHash = &val
	}
	if k.Valid {
		val := int(k.Int32)
		record.K = &val
	}
	return &record, nil
}

func nullableTime(t *time.Time) any {
	if t == nil || t.IsZero() {
		return nil
	}
	return *t
}

func nullableString(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
