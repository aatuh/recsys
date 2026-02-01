package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
)

// ExposureReader reads exposures from Postgres rows containing JSON.
type ExposureReader struct {
	dsn   string
	query string
}

func NewExposureReader(dsn, query string) ExposureReader {
	return ExposureReader{dsn: dsn, query: query}
}

func (r ExposureReader) Read(ctx context.Context) ([]dataset.Exposure, error) {
	return readJSONRows[dataset.Exposure](ctx, r.dsn, r.query)
}

// OutcomeReader reads outcomes from Postgres rows containing JSON.
type OutcomeReader struct {
	dsn   string
	query string
}

func NewOutcomeReader(dsn, query string) OutcomeReader {
	return OutcomeReader{dsn: dsn, query: query}
}

func (r OutcomeReader) Read(ctx context.Context) ([]dataset.Outcome, error) {
	return readJSONRows[dataset.Outcome](ctx, r.dsn, r.query)
}

// AssignmentReader reads assignments from Postgres rows containing JSON.
type AssignmentReader struct {
	dsn   string
	query string
}

func NewAssignmentReader(dsn, query string) AssignmentReader {
	return AssignmentReader{dsn: dsn, query: query}
}

func (r AssignmentReader) Read(ctx context.Context) ([]dataset.Assignment, error) {
	return readJSONRows[dataset.Assignment](ctx, r.dsn, r.query)
}

// RankListReader reads rank lists from Postgres rows containing JSON.
type RankListReader struct {
	dsn   string
	query string
}

func NewRankListReader(dsn, query string) RankListReader {
	return RankListReader{dsn: dsn, query: query}
}

func (r RankListReader) Read(ctx context.Context) ([]dataset.RankList, error) {
	return readJSONRows[dataset.RankList](ctx, r.dsn, r.query)
}

func readJSONRows[T any](ctx context.Context, dsn, query string) ([]T, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []T
	idx := 0
	for rows.Next() {
		idx++
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("scan error at row %d: %w", idx, err)
		}
		var item T
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, fmt.Errorf("json parse error at row %d: %w", idx, err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
