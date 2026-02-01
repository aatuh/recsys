//go:build !duckdb

package duckdb

import (
	"context"
	"errors"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/dataset"
)

// ErrUnsupported indicates DuckDB support is not enabled in this build.
var ErrUnsupported = errors.New("duckdb support is not enabled (build with -tags duckdb)")

// Supported reports whether this build includes DuckDB support.
func Supported() bool { return false }

// ExposureReader reads exposures from DuckDB.
type ExposureReader struct{}

func NewExposureReader(_, _ string) ExposureReader { return ExposureReader{} }

func (ExposureReader) Read(context.Context) ([]dataset.Exposure, error) { return nil, ErrUnsupported }

// OutcomeReader reads outcomes from DuckDB.
type OutcomeReader struct{}

func NewOutcomeReader(_, _ string) OutcomeReader { return OutcomeReader{} }

func (OutcomeReader) Read(context.Context) ([]dataset.Outcome, error) { return nil, ErrUnsupported }

// AssignmentReader reads assignments from DuckDB.
type AssignmentReader struct{}

func NewAssignmentReader(_, _ string) AssignmentReader { return AssignmentReader{} }

func (AssignmentReader) Read(context.Context) ([]dataset.Assignment, error) {
	return nil, ErrUnsupported
}

// RankListReader reads rank lists from DuckDB.
type RankListReader struct{}

func NewRankListReader(_, _ string) RankListReader { return RankListReader{} }

func (RankListReader) Read(context.Context) ([]dataset.RankList, error) { return nil, ErrUnsupported }
