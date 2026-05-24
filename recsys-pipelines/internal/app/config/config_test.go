package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadEnvConfigDefaultsArtifactKinds(t *testing.T) {
	path := writeConfig(t, `{}`)

	cfg, err := LoadEnvConfig(path)
	if err != nil {
		t.Fatalf("LoadEnvConfig() error = %v", err)
	}
	got := strings.Join(cfg.ArtifactKinds, ",")
	if got != "popularity,cooc" {
		t.Fatalf("ArtifactKinds = %q, want popularity,cooc", got)
	}
	selection, err := ParseArtifactSelection(cfg.ArtifactKinds)
	if err != nil {
		t.Fatalf("ParseArtifactSelection() error = %v", err)
	}
	if !selection.Popularity || !selection.Cooc {
		t.Fatalf("default selection = %+v, want popularity and cooc", selection)
	}
	if selection.Implicit || selection.ContentSim || selection.SessionSeq {
		t.Fatalf("default selection enabled rich artifacts: %+v", selection)
	}
}

func TestLoadEnvConfigExplicitArtifactKindsAndCatalog(t *testing.T) {
	path := writeConfig(t, `{
		"artifact_kinds": ["popularity", "cooc", "implicit", "content_sim", "session_seq"],
		"catalog": {"path": "catalog.csv", "format": "csv"}
	}`)

	cfg, err := LoadEnvConfig(path)
	if err != nil {
		t.Fatalf("LoadEnvConfig() error = %v", err)
	}
	selection, err := ParseArtifactSelection(cfg.ArtifactKinds)
	if err != nil {
		t.Fatalf("ParseArtifactSelection() error = %v", err)
	}
	if !selection.Popularity || !selection.Cooc || !selection.Implicit || !selection.ContentSim || !selection.SessionSeq {
		t.Fatalf("selection = %+v, want all artifacts enabled", selection)
	}
	if cfg.Catalog.Path != "catalog.csv" || cfg.Catalog.Format != "csv" {
		t.Fatalf("Catalog = %+v, want configured catalog", cfg.Catalog)
	}
}

func TestLoadEnvConfigRejectsInvalidArtifactKind(t *testing.T) {
	path := writeConfig(t, `{"artifact_kinds": ["popularity", "unknown"]}`)

	_, err := LoadEnvConfig(path)
	if err == nil {
		t.Fatalf("LoadEnvConfig() error = nil")
	}
	if !strings.Contains(err.Error(), "unsupported artifact kind") {
		t.Fatalf("LoadEnvConfig() error = %q, want unsupported artifact kind", err)
	}
}

func TestLoadEnvConfigRequiresCatalogForContentArtifact(t *testing.T) {
	path := writeConfig(t, `{"artifact_kinds": ["content_sim"]}`)

	_, err := LoadEnvConfig(path)
	if err == nil {
		t.Fatalf("LoadEnvConfig() error = nil")
	}
	if !strings.Contains(err.Error(), "catalog.path") {
		t.Fatalf("LoadEnvConfig() error = %q, want catalog.path", err)
	}
}

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
