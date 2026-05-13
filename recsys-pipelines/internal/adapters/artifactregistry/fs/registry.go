package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/fsutil"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/artifacts"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/domain/pathsafe"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/artifactregistry"
)

type FSRegistry struct {
	baseDir string
	mu      sync.Mutex
}

var _ artifactregistry.Registry = (*FSRegistry)(nil)

func New(baseDir string) *FSRegistry {
	return &FSRegistry{baseDir: baseDir}
}

func (r *FSRegistry) Record(ctx context.Context, ref artifacts.Ref) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ref.Validate(); err != nil {
		return err
	}
	if ref.URI == "" {
		return fmt.Errorf("ref uri must be set")
	}
	tenant, _ := pathsafe.Segment("tenant", ref.Key.Tenant)
	surface, _ := pathsafe.Segment("surface", ref.Key.Surface)
	artifactType, _ := pathsafe.Segment("artifact type", string(ref.Key.Type))
	version, _ := pathsafe.Segment("version", ref.Version)

	path := filepath.Join(r.baseDir, "records", tenant, surface, artifactType, version+".json")
	if _, err := os.Stat(path); err == nil {
		// Records are append-only and version-addressed. Re-publishing the same
		// version should not mutate the original record.
		return nil
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(ref, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(path, b, 0o644)
}

func (r *FSRegistry) LoadManifest(ctx context.Context, tenant, surface string) (artifacts.ManifestV1, bool, error) {
	select {
	case <-ctx.Done():
		return artifacts.ManifestV1{}, false, ctx.Err()
	default:
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	path, err := r.manifestPath(tenant, surface)
	if err != nil {
		return artifacts.ManifestV1{}, false, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return artifacts.ManifestV1{}, false, nil
		}
		return artifacts.ManifestV1{}, false, err
	}
	var m artifacts.ManifestV1
	if err := json.Unmarshal(b, &m); err != nil {
		return artifacts.ManifestV1{}, false, err
	}
	return m, true, nil
}

func (r *FSRegistry) SwapManifest(ctx context.Context, tenant, surface string, next artifacts.ManifestV1) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	path, err := r.manifestPath(tenant, surface)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(path, b, 0o644)
}

func (r *FSRegistry) manifestPath(tenant, surface string) (string, error) {
	tenant, err := pathsafe.Segment("tenant", tenant)
	if err != nil {
		return "", err
	}
	surface, err = pathsafe.Segment("surface", surface)
	if err != nil {
		return "", err
	}
	return filepath.Join(r.baseDir, "current", tenant, surface, "manifest.json"), nil
}
