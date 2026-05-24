package factory

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
)

func TestBuildCatalogReaderDetectsCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalog.csv")
	if err := os.WriteFile(path, []byte("item_id,tags\nsku-1,brand:a\n"), 0o600); err != nil {
		t.Fatalf("write catalog: %v", err)
	}

	reader, err := BuildCatalogReader(config.EnvConfig{Catalog: config.CatalogConfig{Path: path}})
	if err != nil {
		t.Fatalf("BuildCatalogReader() error = %v", err)
	}
	items, err := reader.Read(context.Background())
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(items) != 1 || items[0].ItemID != "sku-1" {
		t.Fatalf("items = %+v, want sku-1", items)
	}
}

func TestBuildCatalogReaderRejectsUnknownFormat(t *testing.T) {
	_, err := BuildCatalogReader(config.EnvConfig{
		Catalog: config.CatalogConfig{Path: "catalog.data", Format: "xml"},
	})
	if err == nil {
		t.Fatalf("BuildCatalogReader() error = nil")
	}
	if !strings.Contains(err.Error(), "unsupported catalog format") {
		t.Fatalf("BuildCatalogReader() error = %q, want unsupported catalog format", err)
	}
}
