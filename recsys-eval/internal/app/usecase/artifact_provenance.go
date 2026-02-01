package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/objectstore/file"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/adapters/objectstore/s3"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/domain/report"
	"github.com/aatuh/recsys-suite/recsys-eval/internal/ports/objectstore"
)

type manifestV1 struct {
	V         int               `json:"v"`
	Tenant    string            `json:"tenant"`
	Surface   string            `json:"surface"`
	Current   map[string]string `json:"current"`
	UpdatedAt string            `json:"updated_at"`
}

type artifactMeta struct {
	ArtifactType string `json:"artifact_type"`
	Build        struct {
		Version    string `json:"version"`
		SourceHash string `json:"source_hash"`
		BuiltAt    string `json:"built_at"`
	} `json:"build"`
}

// ErrArtifactProvenanceDisabled indicates that no manifest URI was provided.
var ErrArtifactProvenanceDisabled = errors.New("artifact provenance disabled")

// ResolveArtifactProvenance loads a manifest and resolves artifact metadata.
func ResolveArtifactProvenance(ctx context.Context, cfg ArtifactConfig) (*report.ArtifactProvenance, error) {
	if strings.TrimSpace(cfg.ManifestURI) == "" {
		return nil, ErrArtifactProvenanceDisabled
	}
	store, err := buildObjectStore(cfg.ObjectStore)
	if err != nil {
		return nil, err
	}
	manifestBytes, err := fetchURI(ctx, cfg.ManifestURI, store)
	if err != nil {
		return nil, err
	}
	var manifest manifestV1
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("manifest parse: %w", err)
	}

	prov := &report.ArtifactProvenance{
		ManifestURI: cfg.ManifestURI,
		Tenant:      manifest.Tenant,
		Surface:     manifest.Surface,
		UpdatedAt:   manifest.UpdatedAt,
	}

	for typ, uri := range manifest.Current {
		if strings.TrimSpace(uri) == "" {
			continue
		}
		info := report.ArtifactRef{Type: typ, URI: uri}
		blob, err := fetchURI(ctx, uri, store)
		if err != nil {
			prov.Warnings = append(prov.Warnings, fmt.Sprintf("artifact fetch failed: %s (%v)", typ, err))
			prov.Artifacts = append(prov.Artifacts, info)
			continue
		}
		h := sha256.Sum256(blob)
		info.Checksum = hex.EncodeToString(h[:])
		var meta artifactMeta
		if err := json.Unmarshal(blob, &meta); err != nil {
			prov.Warnings = append(prov.Warnings, fmt.Sprintf("artifact parse failed: %s (%v)", typ, err))
			prov.Artifacts = append(prov.Artifacts, info)
			continue
		}
		info.Version = meta.Build.Version
		info.SourceHash = meta.Build.SourceHash
		info.BuiltAt = meta.Build.BuiltAt
		prov.Artifacts = append(prov.Artifacts, info)
	}

	return prov, nil
}

func buildObjectStore(cfg ObjectStoreConfig) (objectstore.Store, error) {
	storeType := strings.ToLower(strings.TrimSpace(cfg.Type))
	switch storeType {
	case "", "file", "fs":
		return file.New(), nil
	case "s3", "minio":
		return s3.New(s3.Config{
			Endpoint:  cfg.S3.Endpoint,
			Bucket:    cfg.S3.Bucket,
			AccessKey: cfg.S3.AccessKey,
			SecretKey: cfg.S3.SecretKey,
			Region:    cfg.S3.Region,
			UseSSL:    cfg.S3.UseSSL,
		})
	default:
		return nil, fmt.Errorf("unsupported object store type: %s", cfg.Type)
	}
}

func fetchURI(ctx context.Context, uri string, store objectstore.Store) ([]byte, error) {
	if strings.TrimSpace(uri) == "" {
		return nil, fmt.Errorf("empty uri")
	}
	if strings.HasPrefix(uri, "file://") || strings.HasPrefix(uri, "/") {
		return file.New().Get(ctx, uri)
	}
	if strings.HasPrefix(uri, "s3://") {
		if store == nil {
			return nil, fmt.Errorf("s3 store not configured")
		}
		return store.Get(ctx, uri)
	}
	if store != nil {
		return store.Get(ctx, uri)
	}
	return file.New().Get(ctx, uri)
}
