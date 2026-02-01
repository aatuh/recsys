package workflow

import (
	"fmt"
	"strings"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/datasource/duckdb"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/datasource/jsonl"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/datasource/postgres"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/app/usecase"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/datasource"
)

// BuildExposureReader instantiates an exposure reader from config.
func BuildExposureReader(cfg usecase.SourceConfig) (datasource.ExposureReader, error) {
	switch strings.ToLower(cfg.Type) {
	case "jsonl":
		return jsonl.NewExposureReader(cfg.Path), nil
	case "postgres":
		return postgres.NewExposureReader(cfg.DSN, cfg.Query), nil
	case "duckdb":
		if !duckdb.Supported() {
			return nil, duckdb.ErrUnsupported
		}
		return duckdb.NewExposureReader(cfg.DSN, cfg.Query), nil
	default:
		return nil, fmt.Errorf("unsupported exposure source type: %s", cfg.Type)
	}
}

// BuildOutcomeReader instantiates an outcome reader from config.
func BuildOutcomeReader(cfg usecase.SourceConfig) (datasource.OutcomeReader, error) {
	switch strings.ToLower(cfg.Type) {
	case "jsonl":
		return jsonl.NewOutcomeReader(cfg.Path), nil
	case "postgres":
		return postgres.NewOutcomeReader(cfg.DSN, cfg.Query), nil
	case "duckdb":
		if !duckdb.Supported() {
			return nil, duckdb.ErrUnsupported
		}
		return duckdb.NewOutcomeReader(cfg.DSN, cfg.Query), nil
	default:
		return nil, fmt.Errorf("unsupported outcome source type: %s", cfg.Type)
	}
}

// BuildAssignmentReader instantiates an assignment reader from config.
func BuildAssignmentReader(cfg usecase.SourceConfig) (datasource.AssignmentReader, error) {
	switch strings.ToLower(cfg.Type) {
	case "jsonl":
		return jsonl.NewAssignmentReader(cfg.Path), nil
	case "postgres":
		return postgres.NewAssignmentReader(cfg.DSN, cfg.Query), nil
	case "duckdb":
		if !duckdb.Supported() {
			return nil, duckdb.ErrUnsupported
		}
		return duckdb.NewAssignmentReader(cfg.DSN, cfg.Query), nil
	default:
		return nil, fmt.Errorf("unsupported assignment source type: %s", cfg.Type)
	}
}

// BuildRankListReader instantiates a rank list reader from config.
func BuildRankListReader(cfg usecase.SourceConfig) (datasource.RankListReader, error) {
	switch strings.ToLower(cfg.Type) {
	case "jsonl":
		return jsonl.NewRankListReader(cfg.Path), nil
	case "postgres":
		return postgres.NewRankListReader(cfg.DSN, cfg.Query), nil
	case "duckdb":
		if !duckdb.Supported() {
			return nil, duckdb.ErrUnsupported
		}
		return duckdb.NewRankListReader(cfg.DSN, cfg.Query), nil
	default:
		return nil, fmt.Errorf("unsupported rank list source type: %s", cfg.Type)
	}
}
