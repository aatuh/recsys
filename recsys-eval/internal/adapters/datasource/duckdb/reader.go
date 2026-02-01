//go:build duckdb

package duckdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
)

// Supported reports whether this build includes DuckDB support.
func Supported() bool { return true }

// ExposureReader reads exposures from DuckDB rows containing JSON.
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

// OutcomeReader reads outcomes from DuckDB rows containing JSON.
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

// AssignmentReader reads assignments from DuckDB rows containing JSON.
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

// RankListReader reads rank lists from DuckDB rows containing JSON.
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
	if dsn == "" {
		dsn = ":memory:"
	}
	db, err := sql.Open("duckdb", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []T
	idx := 0
	for rows.Next() {
		idx++
		var raw any
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("scan error at row %d: %w", idx, err)
		}
		var data []byte
		switch v := raw.(type) {
		case []byte:
			data = v
		case string:
			data = []byte(v)
		default:
			return nil, fmt.Errorf("unexpected duckdb row type %T at row %d", raw, idx)
		}
		var item T
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, fmt.Errorf("json parse error at row %d: %w", idx, err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
