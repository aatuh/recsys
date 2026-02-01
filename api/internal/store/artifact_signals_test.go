package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aatuh/api-toolkit/authorization"
	"github.com/aatuh/recsys-suite/api/internal/artifacts"
	"github.com/aatuh/recsys-suite/api/internal/objectstore"
	"github.com/google/uuid"
)

func TestArtifactContentSimilarity(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	manifestTpl := filepath.Join(tmp, "manifest-{tenant}-{surface}.json")

	contentPath := filepath.Join(tmp, "content.json")
	content := struct {
		V            int    `json:"v"`
		ArtifactType string `json:"artifact_type"`
		Tenant       string `json:"tenant"`
		Surface      string `json:"surface"`
		Window       struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"window"`
		Items []struct {
			ItemID string   `json:"item_id"`
			Tags   []string `json:"tags"`
		} `json:"items"`
		Build artifacts.BuildInfo `json:"build"`
	}{
		V:            1,
		ArtifactType: artifacts.TypeContentSim,
		Tenant:       "demo",
		Surface:      "home",
		Items: []struct {
			ItemID string   `json:"item_id"`
			Tags   []string `json:"tags"`
		}{
			{ItemID: "a", Tags: []string{"sports", "news"}},
			{ItemID: "b", Tags: []string{"news"}},
			{ItemID: "c", Tags: []string{"music"}},
		},
		Build: artifacts.BuildInfo{BuiltAt: time.Now().UTC().Format(time.RFC3339), Version: "v1"},
	}
	content.Window.Start = time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	content.Window.End = time.Now().UTC().Format(time.RFC3339)
	writeJSON(t, contentPath, content)

	manifest := artifacts.ManifestV1{
		V:         1,
		Tenant:    "demo",
		Surface:   "home",
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Current: map[string]string{
			artifacts.TypeContentSim: contentPath,
		},
	}
	writeJSON(t, filepath.Join(tmp, "manifest-demo-home.json"), manifest)

	loader := artifacts.NewLoader(objectstore.NewFSReader(0), artifacts.LoaderConfig{
		ManifestTemplate: manifestTpl,
	})
	store := NewArtifactAlgoStore(loader, nil)

	ctx := authorization.WithScope(context.Background(), authorization.Scope{TenantID: "demo"})
	out, err := store.ContentSimilarityTopK(ctx, uuid.Nil, "home", []string{"news"}, 10, nil)
	if err != nil {
		t.Fatalf("ContentSimilarityTopK error: %v", err)
	}
	if len(out) != 2 || out[0].ItemID != "a" || out[1].ItemID != "b" {
		t.Fatalf("unexpected content result: %+v", out)
	}
}

func TestArtifactSessionSequence(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	manifestTpl := filepath.Join(tmp, "manifest-{tenant}-{surface}.json")

	sessionPath := filepath.Join(tmp, "session.json")
	session := struct {
		V            int    `json:"v"`
		ArtifactType string `json:"artifact_type"`
		Tenant       string `json:"tenant"`
		Surface      string `json:"surface"`
		Window       struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"window"`
		Users []struct {
			UserID string `json:"user_id"`
			Items  []struct {
				ItemID string  `json:"item_id"`
				Score  float64 `json:"score"`
			} `json:"items"`
		} `json:"users"`
		Build artifacts.BuildInfo `json:"build"`
	}{
		V:            1,
		ArtifactType: artifacts.TypeSessionSeq,
		Tenant:       "demo",
		Surface:      "home",
		Users: []struct {
			UserID string `json:"user_id"`
			Items  []struct {
				ItemID string  `json:"item_id"`
				Score  float64 `json:"score"`
			} `json:"items"`
		}{
			{UserID: "u1", Items: []struct {
				ItemID string  `json:"item_id"`
				Score  float64 `json:"score"`
			}{{ItemID: "x", Score: 2}, {ItemID: "y", Score: 1}}},
		},
		Build: artifacts.BuildInfo{BuiltAt: time.Now().UTC().Format(time.RFC3339), Version: "v1"},
	}
	session.Window.Start = time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	session.Window.End = time.Now().UTC().Format(time.RFC3339)
	writeJSON(t, sessionPath, session)

	manifest := artifacts.ManifestV1{
		V:         1,
		Tenant:    "demo",
		Surface:   "home",
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Current: map[string]string{
			artifacts.TypeSessionSeq: sessionPath,
		},
	}
	writeJSON(t, filepath.Join(tmp, "manifest-demo-home.json"), manifest)

	loader := artifacts.NewLoader(objectstore.NewFSReader(0), artifacts.LoaderConfig{
		ManifestTemplate: manifestTpl,
	})
	store := NewArtifactAlgoStore(loader, nil)

	ctx := authorization.WithScope(context.Background(), authorization.Scope{TenantID: "demo"})
	out, err := store.SessionSequenceTopK(ctx, uuid.Nil, "home", "u1", 5, 30, nil, 10)
	if err != nil {
		t.Fatalf("SessionSequenceTopK error: %v", err)
	}
	if len(out) != 2 || out[0].ItemID != "x" || out[1].ItemID != "y" {
		t.Fatalf("unexpected session result: %+v", out)
	}
}

func writeJSON(t *testing.T, path string, payload any) {
	t.Helper()
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write json: %v", err)
	}
}
