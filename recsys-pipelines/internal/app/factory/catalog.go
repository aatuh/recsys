package factory

import (
	"fmt"
	"path/filepath"
	"strings"

	catalogcsv "github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/catalog/csv"
	catalogjsonl "github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/catalog/jsonl"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/catalog"
)

func BuildCatalogReader(cfg config.EnvConfig) (catalog.Reader, error) {
	path := strings.TrimSpace(cfg.Catalog.Path)
	if path == "" {
		return nil, fmt.Errorf("catalog.path is required")
	}
	format := strings.ToLower(strings.TrimSpace(cfg.Catalog.Format))
	if format == "" {
		switch strings.ToLower(filepath.Ext(path)) {
		case ".csv":
			format = "csv"
		case ".jsonl":
			format = "jsonl"
		default:
			return nil, fmt.Errorf("catalog.format is required when the file extension is not csv or jsonl")
		}
	}
	switch format {
	case "csv":
		return catalogcsv.New(path), nil
	case "jsonl":
		return catalogjsonl.New(path), nil
	default:
		return nil, fmt.Errorf("unsupported catalog format: %s", cfg.Catalog.Format)
	}
}
