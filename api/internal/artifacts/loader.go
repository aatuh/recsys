package artifacts

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aatuh/recsys-suite/api/internal/cache"
	"github.com/aatuh/recsys-suite/api/internal/objectstore"
)

type LoaderConfig struct {
	ManifestTemplate string
	ManifestTTL      time.Duration
	ArtifactTTL      time.Duration
	MaxBytes         int
}

type Loader struct {
	reader           objectstore.Reader
	manifestTemplate string
	manifestCache    *cache.TTLCache[manifestKey, ManifestV1]
	popCache         *cache.TTLCache[string, PopularityArtifactV1]
	coocCache        *cache.TTLCache[string, CoocArtifactV1]
	manifestTTL      time.Duration
	artifactTTL      time.Duration
	maxBytes         int
}

type manifestKey struct {
	Tenant  string
	Surface string
}

func NewLoader(reader objectstore.Reader, cfg LoaderConfig) *Loader {
	return &Loader{
		reader:           reader,
		manifestTemplate: strings.TrimSpace(cfg.ManifestTemplate),
		manifestCache:    cache.NewTTL[manifestKey, ManifestV1](nil),
		popCache:         cache.NewTTL[string, PopularityArtifactV1](nil),
		coocCache:        cache.NewTTL[string, CoocArtifactV1](nil),
		manifestTTL:      cfg.ManifestTTL,
		artifactTTL:      cfg.ArtifactTTL,
		maxBytes:         cfg.MaxBytes,
	}
}

func (l *Loader) Invalidate(tenant, surface string) int {
	if l == nil {
		return 0
	}
	n := 0
	if l.manifestCache != nil {
		n += l.manifestCache.Invalidate(func(k manifestKey, _ ManifestV1) bool {
			if tenant != "" && k.Tenant != tenant {
				return false
			}
			if surface != "" && k.Surface != surface {
				return false
			}
			return true
		})
	}
	if l.popCache != nil {
		n += l.popCache.Invalidate(func(key string, _ PopularityArtifactV1) bool { return true })
	}
	if l.coocCache != nil {
		n += l.coocCache.Invalidate(func(key string, _ CoocArtifactV1) bool { return true })
	}
	return n
}

func (l *Loader) ManifestURI(tenant, surface string) (string, error) {
	if l == nil {
		return "", fmt.Errorf("loader is nil")
	}
	if l.manifestTemplate == "" {
		return "", fmt.Errorf("manifest template is required")
	}
	out := l.manifestTemplate
	out = strings.ReplaceAll(out, "{tenant}", tenant)
	out = strings.ReplaceAll(out, "{surface}", surface)
	return out, nil
}

func (l *Loader) LoadManifest(ctx context.Context, tenant, surface string) (ManifestV1, bool, error) {
	if l == nil {
		return ManifestV1{}, false, fmt.Errorf("loader is nil")
	}
	key := manifestKey{Tenant: tenant, Surface: surface}
	if val, ok := l.manifestCache.Get(key); ok {
		return val, true, nil
	}
	uri, err := l.ManifestURI(tenant, surface)
	if err != nil {
		return ManifestV1{}, false, err
	}
	data, err := l.reader.Get(ctx, uri)
	if err != nil {
		if _, ok := err.(objectstore.ErrNotFound); ok {
			return ManifestV1{}, false, nil
		}
		return ManifestV1{}, false, err
	}
	var manifest ManifestV1
	if err := json.Unmarshal(data, &manifest); err != nil {
		return ManifestV1{}, false, err
	}
	if err := manifest.Validate(); err != nil {
		return ManifestV1{}, false, err
	}
	if l.manifestTTL > 0 {
		l.manifestCache.Set(key, manifest, l.manifestTTL)
	}
	return manifest, true, nil
}

func (l *Loader) LoadPopularity(ctx context.Context, uri string) (PopularityArtifactV1, bool, error) {
	if l == nil {
		return PopularityArtifactV1{}, false, fmt.Errorf("loader is nil")
	}
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return PopularityArtifactV1{}, false, nil
	}
	if val, ok := l.popCache.Get(uri); ok {
		return val, true, nil
	}
	data, err := l.reader.Get(ctx, uri)
	if err != nil {
		if _, ok := err.(objectstore.ErrNotFound); ok {
			return PopularityArtifactV1{}, false, nil
		}
		return PopularityArtifactV1{}, false, err
	}
	var art PopularityArtifactV1
	if err := json.Unmarshal(data, &art); err != nil {
		return PopularityArtifactV1{}, false, err
	}
	if art.V != 1 || art.ArtifactType != TypePopularity {
		return PopularityArtifactV1{}, false, fmt.Errorf("invalid popularity artifact")
	}
	if l.artifactTTL > 0 {
		l.popCache.Set(uri, art, l.artifactTTL)
	}
	return art, true, nil
}

func (l *Loader) LoadCooc(ctx context.Context, uri string) (CoocArtifactV1, bool, error) {
	if l == nil {
		return CoocArtifactV1{}, false, fmt.Errorf("loader is nil")
	}
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return CoocArtifactV1{}, false, nil
	}
	if val, ok := l.coocCache.Get(uri); ok {
		return val, true, nil
	}
	data, err := l.reader.Get(ctx, uri)
	if err != nil {
		if _, ok := err.(objectstore.ErrNotFound); ok {
			return CoocArtifactV1{}, false, nil
		}
		return CoocArtifactV1{}, false, err
	}
	var art CoocArtifactV1
	if err := json.Unmarshal(data, &art); err != nil {
		return CoocArtifactV1{}, false, err
	}
	if art.V != 1 || art.ArtifactType != TypeCooc {
		return CoocArtifactV1{}, false, fmt.Errorf("invalid cooc artifact")
	}
	if l.artifactTTL > 0 {
		l.coocCache.Set(uri, art, l.artifactTTL)
	}
	return art, true, nil
}
